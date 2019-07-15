package main

import (
	"fmt"
	"log"
	"net"

	"github.com/miekg/dns"
)

func main() {
	appConfigs, err := InitConfig()
	if err != nil {
		log.Fatalf("Failed to load configs: %s", err)
	}

	dnsCache := NewCache(appConfigs.CacheExpiration)
	domains, err := appConfigs.DNSConfigs.Domains.Compile()
	if err != nil {
		log.Fatalf("Failed to parse domains: %s", err)
	}
	servers, err := appConfigs.DNSConfigs.Servers.Compile()
	if err != nil {
		log.Fatalf("Failed to parse servers: %s", err)
	}

	dnsProxy := DnsProxy{
		Cache:         &dnsCache,
		domains:       domains,
		servers:       servers,
		defaultServer: appConfigs.DNSConfigs.DefaultDns,
	}

	logger := NewLogger(appConfigs.LogLevel)
	host := appConfigs.DNSConfigs.Host
	if host == "" || appConfigs.UseOutbound {
		ip, err := GetOutboundIP()
		if err != nil {
			logger.Fatalf("Failed to get outbound address: %s\n ", err.Error())
		}
		host = fmt.Sprintf("%s:%d", ip, appConfigs.Port)
	}

	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		switch r.Opcode {
		case dns.OpcodeQuery:
			m, err := dnsProxy.getResponse(r)
			if err != nil {
				logger.Errorf("Failed lookup for %s with error: %s\n", r, err.Error())
				m.SetReply(r)
				w.WriteMsg(m)
				return
			}
			if len(m.Answer) > 0 && logger.level >= infoLevel {
				var ip net.IP = nil
				switch a := m.Answer[0].(type) {
				case *dns.A:
					ip = a.A
				case *dns.AAAA:
					ip = a.AAAA
				case *dns.L32:
					ip = a.Locator32
				}

				if len(ip) > 0 {
					logger.Infof("Lookup for %s with ip %s\n", m.Answer[0].Header().Name, ip)
				} else {
					logger.Infof("Lookup for %s with response %s\n", m.Answer[0].Header().Name, m.Answer[0])
				}
			}
			m.SetReply(r)
			w.WriteMsg(m)
		}
	})

	server := &dns.Server{Addr: host, Net: "udp"}
	logger.Infof("Starting at %s\n", host)
	err = server.ListenAndServe()
	if err != nil {
		logger.Errorf("Failed to start server: %s\n ", err.Error())
	}
}
