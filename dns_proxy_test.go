package main

import (
	"strings"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

type DnsCalls struct {
	calls []struct {
		m      *dns.Msg
		server string
	}
}

func (c *DnsCalls) Exchange(m *dns.Msg, server string) (r *dns.Msg, err error) {
	c.calls = append(c.calls, struct {
		m      *dns.Msg
		server string
	}{m: m, server: server})
	return &dns.Msg{Answer: []dns.RR{&dns.A{}}}, nil
}

func query(host string) *dns.Msg {
	if !strings.HasSuffix(host, ".") {
		host += "."
	}
	return &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Opcode:            dns.OpcodeQuery,
			Rcode:             dns.RcodeSuccess,
			RecursionDesired:  true,
			AuthenticatedData: true,
		},
		Question: []dns.Question{
			dns.Question{
				Name:   host,
				Qtype:  dns.TypeA,
				Qclass: dns.ClassINET,
			},
		},
	}
}

func TestGetResponse_DNSResolve(t *testing.T) {
	//arrange
	proxy := DnsProxy{defaultServer: "8.8.8.8:53"}
	msg := query("google.com")

	//act
	r, err := proxy.getResponse(msg)

	//assert
	assert.Nil(t, err)
	if assert.Len(t, r.Answer, 1) {
		assert.IsType(t, &dns.A{}, r.Answer[0])
	}
}

func TestGetResponse_LocalResolve(t *testing.T) {
	//arrange
	mock := &DnsCalls{}
	host := "test.com"
	resolution := "1.2.3.4"
	proxy := DnsProxy{DnsApi: mock, domains: HostMap{host: resolution}.ShouldCompile()}
	msg := query(host)

	//act
	r, err := proxy.getResponse(msg)

	//assert
	assert.Nil(t, err)
	assert.Empty(t, mock.calls, "calls: %+v", mock.calls)
	if assert.Len(t, r.Answer, 1) {
		assert.IsType(t, &dns.A{}, r.Answer[0])
		assert.Equal(t, resolution, r.Answer[0].(*dns.A).A.String(), "answers: %+v", r.Answer)
	}
}

func TestGetResponse_RerouteRequest(t *testing.T) {
	//arrange
	mock := &DnsCalls{}
	host := "test.com"
	server := "5.6.7.8"
	proxy := DnsProxy{
		DnsApi:  mock,
		servers: HostMap{host: server}.ShouldCompile(),
	}
	msg := query(host)

	//act
	proxy.getResponse(msg)

	//assert
	if assert.Len(t, mock.calls, 1) {
		assert.Equal(t, server, mock.calls[0].server)
	}
}

func TestGetResponse_GlobCheck(t *testing.T) {
	//arrange
	mock := &DnsCalls{}
	resolution := "1.2.3.4"
	proxy := DnsProxy{
		DnsApi: mock,
		domains: HostMap{
			"**.com": resolution,
			"**.net": "5.6.7.8",
		}.ShouldCompile(),
	}
	msg := query("yo.test.com")

	//act
	r, _ := proxy.getResponse(msg)

	//assert
	if assert.Len(t, r.Answer, 1, "answers: %+v", r.Answer) {
		assert.Equal(t, resolution, r.Answer[0].(*dns.A).A.String())
	}
}

func TestGetResponse_Redirect(t *testing.T) {
	//arrange
	mock := &DnsCalls{}
	resolution := "1.2.3.4"
	proxy := DnsProxy{
		DnsApi: mock,
		domains: HostMap{
			"a.com": "b.net",
			"b.net": resolution,
		}.ShouldCompile(),
	}
	msg := query("a.com")

	//act
	r, err := proxy.getResponse(msg)

	//assert
	assert.Nil(t, err)
	if assert.Len(t, r.Answer, 1, "answers: %+v", r.Answer) {
		assert.Equal(t, resolution, r.Answer[0].(*dns.A).A.String())
	}
}

func TestGetResponse_RedirectStopRecursion(t *testing.T) {
	//arrange
	mock := &DnsCalls{}
	proxy := DnsProxy{
		DnsApi: mock,
		domains: HostMap{
			"a.com": "b.net",
			"b.net": "a.com",
		}.ShouldCompile(),
	}
	msg := query("a.com")

	//act
	_, err := proxy.getResponse(msg)

	//assert
	assert.EqualError(t, err, ErrNotFound.Error())
}

func TestGetResponse_Vars(t *testing.T) {
	//arrange
	resolution := "1.2.3.4"
	mock := &DnsCalls{}
	proxy := DnsProxy{
		DnsApi: mock,
		domains: HostMap{
			"a.com": "var1",
		}.ShouldCompile(),
		vars: HostMap{
			"var1": resolution,
		}.ShouldCompile(),
	}
	msg := query("a.com")

	//act
	r, err := proxy.getResponse(msg)

	//assert
	assert.Nil(t, err)
	if assert.Len(t, r.Answer, 1, "answers: %+v", r.Answer) {
		assert.Equal(t, resolution, r.Answer[0].(*dns.A).A.String())
	}
}
