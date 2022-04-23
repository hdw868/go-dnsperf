package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type DNSRunner struct {
	NameServer string
	client     *dns.Client
}

func NewDNSRunner(server string, port int, timeout time.Duration, protocol string) (r *DNSRunner) {
	r = &DNSRunner{
		NameServer: server + ":" + strconv.Itoa(port),
		client: &dns.Client{
			Timeout: timeout,
			Net:     protocol,
		}}
	return
}

func (r *DNSRunner) Query(domain, queryType, subnet string) (resp *dns.Msg, rtt time.Duration, reqSize int, err error) {
	// set option
	o := new(dns.OPT)
	o.Hdr.Name = "."
	o.Hdr.Rrtype = dns.TypeOPT

	// set client subnet
	if len(subnet) != 0 {
		err := setSubnet(o, subnet)
		if err != nil {
			return nil, 0, 0, err
		}
	}

	// set message content
	m := new(dns.Msg)
	canonicalDomain := dns.Fqdn(domain)
	m.SetQuestion(canonicalDomain, dns.StringToType[strings.ToUpper(queryType)])
	m.Extra = append(m.Extra, o)
	reqSize = m.Len()
	resp, rtt, err = r.client.Exchange(m, r.NameServer)
	return
}

func setSubnet(o *dns.OPT, subnet string) error {
	ext := &dns.EDNS0_SUBNET{
		Code:          dns.EDNS0SUBNET,
		Address:       net.ParseIP(subnet),
		Family:        1, // IPv4
		SourceNetmask: net.IPv4len * 8,
	}

	if ext.Address == nil {
		fmt.Fprintf(os.Stderr, "fail to parse IP address: %s\n", subnet)
		return errors.New("invalid subnet address")
	}

	if ext.Address.To4() == nil {
		ext.Family = 2 // IPv6
		ext.SourceNetmask = net.IPv6len * 8
	}
	o.Option = append(o.Option, ext)
	return nil
}
