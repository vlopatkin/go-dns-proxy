package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"time"
)

type DNSConfig struct {
	Host       string  `json:"host"`
	DefaultDns string  `json:"default_dns"`
	Servers    HostMap `json:"servers"`
	Domains    HostMap `json:"domains"`
}

type Config struct {
	DNSConfigs      DNSConfig
	CacheExpiration time.Duration
	UseOutbound     bool
	Port            int
	LogLevel        string
}

func InitConfig() (Config, error) {
	fileName := flag.String("file", "config.json", "config filename")
	logLevel := flag.String("log-level", "info", "log level, accepts err, info, none")
	expiration := flag.Int64("expiration", -1, "cache expiration time in seconds")
	useOutbound := flag.Bool("use-outbound", false, "use outbound address of the host for incoming connections")
	port := flag.Int("port", 53, "UDP port, use with use-outbound flag")
	cliConfigs := flag.String("json-config", "", "config in json format")
	flag.Parse()

	var dnsConfigs DNSConfig
	if *cliConfigs == "" {
		var err error
		err = parseFile(*fileName, &dnsConfigs)
		if err != nil {
			return Config{}, err
		}
	} else {
		if err := json.Unmarshal([]byte(*cliConfigs), &dnsConfigs); err != nil {
			return Config{}, err
		}
	}

	return Config{
		DNSConfigs:      dnsConfigs,
		CacheExpiration: time.Duration(*expiration) * time.Second,
		UseOutbound:     *useOutbound,
		Port:            *port,
		LogLevel:        *logLevel,
	}, nil
}

func parseFile(filePath string, v interface{}) error {
	body, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, v); err != nil {
		return err
	}

	return nil
}
