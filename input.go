package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/miekg/dns"
)

var ErrInvalidFormat = errors.New("invalid format")

func inputTransformer(input string) (*dns.Msg, error) {
	var domain, qType, subnet string

	n, err := fmt.Sscanf(input, "%s %s %s", &domain, &qType, &subnet)
	if err != nil && n < 2 {
		return nil, fmt.Errorf("%w, want:\"Domain Qtype [Subnet]\", got: %q", ErrInvalidFormat, input)
	}

	msgType, ok := dns.StringToType[strings.ToUpper(qType)]
	if !ok {
		return nil, fmt.Errorf("%w, invalid Qtype: %q", ErrInvalidFormat, qType)
	}

	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), msgType)

	// add subnet option
	if len(subnet) > 0 {
		o := new(dns.OPT)
		o.Hdr.Name = "."
		o.Hdr.Rrtype = dns.TypeOPT
		setSubnet(o, subnet)
		msg.Extra = append(msg.Extra, o)
	}

	return msg, nil
}

func parseDatafile(dfile string) []*dns.Msg {
	input, err := os.Open(dfile)
	if err != nil {
		fmt.Println(err)
	}
	defer input.Close()

	requests := make([]*dns.Msg, 0)
	scanner := bufio.NewScanner(input)

	for scanner.Scan() {
		line := scanner.Text()
		request, err := inputTransformer(line)

		if err != nil {
			fmt.Printf("Warning: Error parsing input line %q: %v\n", line, err)
		} else {
			requests = append(requests, request)
		}
	}
	return requests
}

func RequestGenerator(filename string, i int) *dns.Msg {
	data := parseDatafile(filename)
	if len(data) < 1 {
		fmt.Printf("%v: at least one valid input line is required", ErrInvalidFormat)
	}

	request := data[i%len(data)]
	return request
}

func LoadRequests(filename string, round int, out chan *dns.Msg) error {
	data := parseDatafile(filename)
	if len(data) < 1 {
		return fmt.Errorf("%v: at least one valid input line is required", ErrInvalidFormat)
	}
	for i := 0; i < round*len(data); i++ {
		out <- data[i%len(data)]
	}
	close(out)
	return nil
}
