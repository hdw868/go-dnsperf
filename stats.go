package main

import (
	"fmt"
	"math"
	"time"

	"github.com/miekg/dns"
)

type Stats struct {
	RunTime            time.Duration
	TotalLatency       time.Duration
	TotalLatencySquare float64
	TotalRequestSize   int64
	TotalResponseSize  int64
	MinResponseTime    time.Duration
	MaxResponseTime    time.Duration
	NumSend            int64
	NumCompleted       int64
	NumTimeout         int64
	NumInterrupted     int64
	ResponseCodes      [16]int64
	QPS                float64
	startTime          time.Time
	endTime            time.Time
}

func (s *Stats) printSummary() {
	s.RunTime = s.endTime.Sub(s.startTime)
	fmt.Printf("Statistics:\n\n")

	// print basic info
	fmt.Printf(" Queries sent:         %d\n", s.NumSend)
	fmt.Printf(" Queries completed:    %d (%.2f%%)\n", s.NumCompleted,
		SafeDiv(100*s.NumCompleted, s.NumSend))
	fmt.Printf(" Queries lost:         %d (%.2f%%)\n", s.NumTimeout,
		SafeDiv(100*s.NumTimeout, s.NumSend))
	if s.NumInterrupted > 0 {
		fmt.Printf(" Queries interrupted:  %d (%.2f%%)\n", s.NumTimeout,
			SafeDiv(100*s.NumInterrupted, s.NumSend))
	}
	fmt.Printf("\n")

	// print rcodes info
	s.printRcodes()

	// print QPS and misc
	fmt.Printf(" Average packet size: request %.f, response %.f\n",
		SafeDiv(s.TotalRequestSize, s.NumSend),
		SafeDiv(s.TotalResponseSize, s.NumCompleted))
	fmt.Printf(" Run time (s):         %.6f\n", s.RunTime.Seconds())
	fmt.Printf(" Query per second:     %.6f\n",
		SafeDiv(s.NumCompleted, int64(s.RunTime.Seconds())))
	fmt.Printf(" Average Latency (s):  %.6f (min %.6f, max %.6f)\n",
		SafeDiv(int64(s.TotalLatency.Seconds()), s.NumCompleted),
		s.MinResponseTime.Seconds(),
		s.MaxResponseTime.Seconds())

	if s.NumCompleted > 1 {
		latencyStdDev := StdDev(s.TotalLatencySquare, s.TotalLatency.Seconds(), float64(s.NumCompleted))
		fmt.Printf(" Latency StdDev (s):   %.6f\n", latencyStdDev)
	}
	fmt.Printf("\n")
}

func (s *Stats) printRcodes() {
	fmt.Printf(" Response codes:    ")
	firstRcode := true
	for i := 0; i < 16; i++ {
		if s.ResponseCodes[i] == 0 {
			continue
		}
		if firstRcode {
			firstRcode = false
		} else {
			fmt.Print(", ")
		}
		fmt.Printf("%s %d (%.2f%%)",
			dns.RcodeToString[i],
			s.ResponseCodes[i],
			SafeDiv(s.ResponseCodes[i]*100, s.NumCompleted))
	}
}

func SafeDiv(m, n int64) float64 {
	if n == 0 {
		return 0
	}
	return float64(m) / float64(n)
}

func StdDev(sumOfSquares, sum, total float64) float64 {
	var (
		squared  float64
		variance float64
	)
	squared = sum * sum
	variance = (sumOfSquares - (squared / float64(total))) / (total - 1)
	return math.Sqrt(variance)
}

func MaxDuration(d1, d2 time.Duration) time.Duration {
	if d1 > d2 {
		return d1
	} else {
		return d2
	}
}

func MinDuration(d1, d2 time.Duration) time.Duration {
	if d1 < d2 {
		return d1
	} else {
		return d2
	}
}

func MaxTime(t1, t2 time.Time) time.Time {
	if t1.UnixNano() > t2.UnixNano() {
		return t1
	} else {
		return t2
	}
}

func MinTime(t1, t2 time.Time) time.Time {
	if t1.UnixNano() < t2.UnixNano() {
		return t1
	} else {
		return t2
	}
}
