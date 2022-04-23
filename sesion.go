package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"github.com/miekg/dns"
)

type LoadCfg struct {
	Server     string
	Port       int
	DataFile   string
	Mode       string
	MaxQPS     int
	Timeouts   int
	Duration   int
	Round      int64
	Goroutines int
}

type LoadSession struct {
	LoadCfg
	TotalStats      *Stats
	statsAggregator chan *Stats
	interrupted     int32
}

func NewLoadSession(cfg LoadCfg) (s *LoadSession) {
	ch := make(chan *Stats, cfg.Goroutines)
	s = &LoadSession{
		LoadCfg: cfg,
		TotalStats: &Stats{
			MinResponseTime: time.Minute,
		},
		statsAggregator: ch,
		interrupted:     0,
	}
	return
}

func DoSend(r *DNSRunner, req *dns.Msg) (respCode, respSize int, respDuration time.Duration, reqSize int) {
	respSize = -1
	respCode = -1
	resp, respDuration, err := r.client.Exchange(req, r.NameServer)
	if err != nil {
		fmt.Printf("Warning: Error occured doing request: %v\n", err)
		return
	}
	reqSize = req.Len()
	respSize = resp.Len()
	respCode = resp.Rcode
	return
}

func (s *LoadSession) RunSingleLoadSession(inputChan chan *dns.Msg) {
	stats := &Stats{MinResponseTime: time.Minute}
	timeouts := time.Duration(s.Timeouts) * time.Second
	r := NewDNSRunner(s.Server, s.Port, timeouts, s.Mode)

	if s.Duration == 0 {
		s.Duration = math.MaxInt64
	}
	maxQPS := SafeDiv(int64(s.MaxQPS), int64(s.Goroutines))
	stats.startTime = time.Now()

	// do query
	for time.Since(stats.startTime).Seconds() <= float64(s.Duration) && atomic.LoadInt32(&s.interrupted) == 0 {
		m, ok := <-inputChan
		if !ok {
			break
		}

		// rate limit
		if maxQPS > 0 {
			runDuration := time.Since(stats.startTime)
			expectSecond := float64(stats.NumSend) / maxQPS
			expectDuration := time.Duration(expectSecond * float64(time.Second))
			if expectDuration > runDuration {
				waitDuration := expectDuration - runDuration
				time.Sleep(time.Duration(waitDuration))
				continue
			}
		}

		// do send
		respCode, respSize, respDuration, reqSize := DoSend(r, m)

		// do statistic
		stats.NumSend++
		if respCode >= 0 {
			stats.TotalRequestSize += int64(reqSize)
			stats.TotalResponseSize += int64(respSize)
			stats.TotalLatency += respDuration
			stats.TotalLatencySquare += respDuration.Seconds() * respDuration.Seconds()
			stats.MaxResponseTime = MaxDuration(respDuration, stats.MaxResponseTime)
			stats.MinResponseTime = MinDuration(respDuration, stats.MinResponseTime)
			stats.ResponseCodes[respCode]++
			stats.NumCompleted++
		} else {
			stats.NumTimeout++
		}
	}
	// add result to channel
	stats.endTime = time.Now()
	s.statsAggregator <- stats
}

func (s *LoadSession) RunLoadSessions(inputs chan *dns.Msg) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	s.TotalStats.startTime = time.Now()

	// consumer
	for i := 0; i < goroutines; i++ {
		go s.RunSingleLoadSession(inputs)
	}

	responser := 0
	for responser < s.LoadCfg.Goroutines {
		select {
		case <-sigChan:
			s.Stop()
			fmt.Println("Stopping...")
		case stats := <-s.statsAggregator:
			s.AggregateStats(stats)
			responser++
		}
	}
	if s.TotalStats.NumSend == 0 {
		fmt.Println("Error: No statistics collected / no requests found")
		return
	}
	s.TotalStats.printSummary()

}

func (s *LoadSession) AggregateStats(stats *Stats) {
	aggStats := s.TotalStats
	aggStats.MaxResponseTime = MaxDuration(aggStats.MaxResponseTime, stats.MaxResponseTime)
	aggStats.MinResponseTime = MinDuration(aggStats.MinResponseTime, stats.MinResponseTime)
	aggStats.endTime = MaxTime(aggStats.endTime, stats.endTime)
	aggStats.NumCompleted += stats.NumCompleted
	aggStats.NumInterrupted += stats.NumInterrupted
	aggStats.NumSend += stats.NumSend
	aggStats.NumTimeout += stats.NumTimeout
	for i := 0; i < 16; i++ {
		aggStats.ResponseCodes[i] += stats.ResponseCodes[i]
	}
	aggStats.TotalLatency += stats.TotalLatency
	aggStats.TotalLatencySquare += stats.TotalLatencySquare
	aggStats.TotalRequestSize += stats.TotalRequestSize
	aggStats.TotalResponseSize += stats.TotalResponseSize
}

func (s *LoadSession) Stop() {
	atomic.StoreInt32(&s.interrupted, 1)
}
