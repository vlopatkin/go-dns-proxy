package main

import (
	"fmt"
	"github.com/gobwas/glob"
	"strings"
)

type HostMap map[string]string

type CompiledHost struct {
	Key   glob.Glob
	Value string
}

type CompiledHostMap []CompiledHost

func (m HostMap) Compile() (c CompiledHostMap, err error) {
	for k, v := range m {
		if !strings.HasSuffix(k, ".") {
			k = k + "."
		}
		g, err := glob.Compile(k, '.')
		if err != nil {
			return c, err
		}
		c = append(c, CompiledHost{g, v})
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

func (m CompiledHostMap) Find(name string) string {
	for _, v := range m {
		if v.Key.Match(name) {
			return v.Value
		}
	}
	return ""
}
