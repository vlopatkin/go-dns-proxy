package main

import (
	"fmt"
	"net"

	"github.com/miekg/dns"
)

type DnsApi interface {
	Exchange(m *dns.Msg, server string) (r *dns.Msg, err error)
}

type DnsProxy struct {
	DnsApi
	Cache         *Cache
	domains       CompiledHostMap
	servers       CompiledHostMap
	vars          CompiledHostMap
	defaultServer string
}

var ErrNotFound = fmt.Errorf("not found")

func (proxy *DnsProxy) getResponse(requestMsg *dns.Msg) (*dns.Msg, error) {
	responseMsg := new(dns.Msg)
	if len(requestMsg.Question) > 0 {
		question := requestMsg.Question[0]

		dnsServer := proxy.dnsServer(question.Name)

		switch question.Qtype {
		case dns.TypeA:
			answer, err := proxy.processTypeA(dnsServer, &question, requestMsg)
			if err != nil {
				return responseMsg, err
			}
			responseMsg.Answer = append(responseMsg.Answer, *answer)

		default:
			answer, err := proxy.processOtherTypes(dnsServer, &question, requestMsg)
			if err != nil {
				return responseMsg, err
			}
			responseMsg.Answer = append(responseMsg.Answer, *answer)
		}
	}

	return responseMsg, nil
}

func (proxy *DnsProxy) dnsServer(name string) string {
	dnsServer := proxy.servers.Find(name)
	if dnsServer.Alias != "" {
		return dnsServer.Alias
	}

	return proxy.defaultServer
}

func (proxy *DnsProxy) processOtherTypes(dnsServer string, q *dns.Question, requestMsg *dns.Msg) (*dns.RR, error) {
	queryMsg := new(dns.Msg)
	requestMsg.CopyTo(queryMsg)
	queryMsg.Question = []dns.Question{*q}

	msg, err := proxy.lookup(dnsServer, queryMsg)
	if err != nil {
		return nil, err
	}

	if len(msg.Answer) > 0 {
		return &msg.Answer[0], nil
	}
	return nil, ErrNotFound
}

func (proxy *DnsProxy) processTypeA(dnsServer string, q *dns.Question, requestMsg *dns.Msg) (*dns.RR, error) {
	localResolve := CompiledHost{Alias: q.Name}
	cacheMsg, found := proxy.Cache.Get(q.Name)
	if !found {
		for c := 0; localResolve.Aliased(); c++ {
			if c >= 50 {
				return nil, ErrNotFound
			}
			domainResolve := proxy.domains.Find(localResolve.Alias)
			if domainResolve.Empty() {
				localResolve = proxy.vars.Find(localResolve.Alias)
				break
			}
			localResolve = domainResolve
		}
	}

	if !localResolve.Resolved() && !found {
		queryMsg := new(dns.Msg)
		requestMsg.CopyTo(queryMsg)
		queryMsg.Question = []dns.Question{*q}

		msg, err := proxy.lookup(dnsServer, queryMsg)
		if err != nil {
			return nil, err
		}

		if len(msg.Answer) > 0 {
			proxy.Cache.Set(q.Name, &msg.Answer[len(msg.Answer)-1])
			return &msg.Answer[len(msg.Answer)-1], nil
		}

	} else if found {
		return cacheMsg.(*dns.RR), nil
	} else if localResolve.Resolved() {
		answer, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, localResolve.IP))
		if err != nil {
			return nil, err
		}
		return &answer, nil

	}
	return nil, ErrNotFound
}

func (proxy *DnsProxy) lookup(server string, requestMsg *dns.Msg) (r *dns.Msg, err error) {
	api := proxy.DnsApi
	if api == nil {
		api = defaultApi
	}
	return api.Exchange(requestMsg, server)
}

func GetOutboundIP() (net.IP, error) {

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}

type DefaultApi struct {
}

func (*DefaultApi) Exchange(m *dns.Msg, a string) (r *dns.Msg, err error) {
	return dns.Exchange(m, a)
}

var defaultApi = &DefaultApi{}
