package main

import (
	"fmt"
	"github.com/gobwas/glob"
	"net"
	"strings"
)

type HostMap map[string]string

type CompiledHost struct {
	IP    net.IP
	Alias string
}

type CompiledHostEntry struct {
	Key glob.Glob
	CompiledHost
}

type CompiledHostMap []CompiledHostEntry

func fixQName(name string) string {
	if !strings.HasSuffix(name, ".") {
		name += "."
	}
	return name
}

func (m HostMap) Compile() (c CompiledHostMap, err error) {
	for k, v := range m {
		k = fixQName(k)
		g, err := glob.Compile(k, '.')
		if err != nil {
			return c, err
		}

		ip := net.ParseIP(v)
		alias := v
		if ip == nil {
			alias = fixQName(v)
		}

		c = append(c, CompiledHostEntry{g, CompiledHost{ip, alias}})
	}

	return c, nil
}

func (m HostMap) ShouldCompile() CompiledHostMap {
	c, err := m.Compile()
	if err != nil {
		panic(fmt.Errorf("host map compilation failed: %s", err))
	}
	return c
}

func (m CompiledHostMap) Find(name string) CompiledHost {
	for _, v := range m {
		if v.Key.Match(name) {
			return v.CompiledHost
		}
	}
	return CompiledHost{}
}

func (m CompiledHost) Empty() bool {
	return m.IP == nil && m.Alias == ""
}

func (m CompiledHost) Resolved() bool {
	return m.IP != nil
}

func (m CompiledHost) Aliased() bool {
	return m.IP == nil && m.Alias != ""
}

func (m CompiledHost) String() string {
	return fmt.Sprintf("ip: %s, alias: %s", m.IP, m.Alias)
}
