# go-dnsperf
A tool similar to dnsperf but written in golang, with extra support of subnet option.

## Build
```
env GOOS=linux GOARCH=amd64 go build
```

## Usage
The usage and output are intended to keep consistent with dnsperf:
```
Usage: gatling [-m mode] [-s server_addr] [-p port] [-d datafile]
               [-c clients] [-n maxruns] [-l timelimit] [-t timeout]
               [-Q max_qps] [-h]
    -Q limit the number of queries per second
    -c the number of clients to act as
    -d the input data file
    -h print help
    -l run for at most this many seconds
    -m set transport mode: udp, tcp (default: udp)
    -n run through the input at most N times
    -p the port on which to query the server (default: udp/tcp 53)
    -s the server to query (default: 127.0.0.1)
    -t the timeout for query completion in seconds (default: 5)
```
and the output could be like:
```
Statistics:

 Queries sent:         30
 Queries completed:    30 (100.00%)
 Queries lost:         0 (0.00%)

 Response codes:    NOERROR 30 (100.00%)
 Average packet size: request 54, response 149
 Run time (s):         0.578596
 Query per second:     0.000000
 Average Latency (s):  0.000000 (min 0.001072, max 0.023456)
 Latency StdDev (s):   0.007992
```
## dataFile
The data file should be in format of `Domain QueryType [subnet]`, for example:
```
example.com AAAA
www.baidu.com A 1.1.1.1
```
## Benchmark
The QPS of single thread(-c 1) could reach 10K; And the MaxQPS could reach 100K with 12 concurrent threads, but it could differ with different CPU resource.
