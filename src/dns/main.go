package main

import (
	"fmt"
	"github.com/miekg/dns"
	"net"
)

type DNSS struct {
}

func main() {
	fmt.Println("DNS SERVER")
	D := &DNSS{}

	ds := dns.ListenAndServe("0.0.0.0:53", "udp", D)
	fmt.Println(ds)
}

func (ds *DNSS) GetA(fqdn string) *dns.A {
	return &dns.A{
		Hdr: dns.RR_Header{
			Name:   fqdn,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			// 0 TTL results in UB for DNS resolvers and generally causes problems.
			Ttl: 1,
		},
		A: net.ParseIP("192.168.66.66"),
	}
}

func (ds *DNSS) GetNS(fqdn string) []*dns.NS {
	records := []*dns.NS{}
	r := &dns.NS{
		Hdr: dns.RR_Header{
			Name:   fqdn,
			Rrtype: dns.TypeNS,
			Class:  dns.ClassINET,
			Ttl:    1,
		},
		Ns: "192.168.5.1.",
	}
	records = append(records, r)
	return records
}

func (ds *DNSS) GetSRV(fqdn string) []*dns.SRV {
	return []*dns.SRV{}
}

func (ds *DNSS) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	m := &dns.Msg{}
	m.SetReply(r)

	answers := []dns.RR{}

	fmt.Println(ds, w, r)
	for _, question := range r.Question {
		fmt.Println(question)
		switch question.Qtype {
		case dns.TypeA:
			a := ds.GetA(question.Name)
			if a != nil {
				answers = append(answers, a)
			}

		case dns.TypeNS:
			ns := ds.GetNS(question.Name)
			if ns != nil {
				for _, record := range ns {
					answers = append(answers, record)
				}
			}
		case dns.TypeSRV:
			srv := ds.GetSRV(question.Name)

			if srv != nil {
				for _, record := range srv {
					answers = append(answers, record)
				}
			}

		}
	}
	m.Authoritative = true
	m.RecursionAvailable = true
	m.Answer = answers

	w.WriteMsg(m)
}
