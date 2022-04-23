package main

import (
	"flag"
	"fmt"

	"github.com/miekg/dns"
)

var (
	helpFlag   bool
	goroutines int
	round      int
	timeouts   int
	timelimit  int
	dfile      string
	server     string
	port       int
	mode       string
	qps        int
)

func init() {
	flag.BoolVar(&helpFlag, "h", false, "print help")
	flag.IntVar(&goroutines, "c", 1, "the number of clients to act as (default: 1)")
	flag.IntVar(&round, "n", 1, "run through input at most N times")
	flag.IntVar(&timeouts, "t", 5, "the timeout for query completion in seconds (default: 5)")
	flag.IntVar(&timelimit, "l", 0, "run for at most this many seconds")
	flag.StringVar(&dfile, "d", "", "the input data file")
	flag.StringVar(&server, "s", "127.0.0.1", "the input data file (default: 127.0.0.1)")
	flag.IntVar(&port, "p", 53, "the port on which to query the server (default: 53)")
	flag.StringVar(&mode, "m", "udp", "set transport mode: udp/tcp (default: udp)")
	flag.IntVar(&qps, "Q", 0, "limit the number of queries per second")
}

func printDefaults() {
	fmt.Println(
		"Usage: go-dnsperf [-m mode] [-s server_addr] [-p port] [-d datafile]\n" +
			"                  [-c clients] [-n maxruns] [-l timelimit] [-t timeout]\n" +
			"                  [-Q max_qps] [-h]")
	flag.VisitAll(func(flag *flag.Flag) {
		fmt.Println(" -"+flag.Name, " ", flag.Usage)
	})
}

func main() {
	flag.Parse()

	if helpFlag || len(dfile) == 0 {
		printDefaults()
		return
	}

	// init loader
	loadCfg := LoadCfg{
		Server:     server,
		Port:       port,
		DataFile:   dfile,
		Mode:       mode,
		MaxQPS:     qps,
		Timeouts:   timeouts,
		Duration:   timelimit,
		Round:      int64(round),
		Goroutines: goroutines,
	}
	loader := NewLoadSession(loadCfg)

	// producer
	bufferSize := 10000
	requests := make(chan *dns.Msg, bufferSize)
	go func() {
		LoadRequests(loader.DataFile, int(loader.Round), requests)
	}()

	// consumer
	loader.RunLoadSessions(requests)
}
