# DNS Proxy
A simple DNS proxy written in go based on [github.com/miekg/dns](https://github.com/miekg/dns)

## How to use it


## Docker

```shell
$ docker run -p 53:53/udp vlopatkin/go-dns-proxy:latest -use-outbound -json-config='{
    "default_dns": "8.8.8.8:53",
    "servers": {
        "google.com" : "8.8.8.8:53"
    },
    "domains": {
        "**.test.com" : "127.0.0.1"
    }
}'
```

## Download executables

[Download](https://github.com/vlopatkin/go-dns-proxy/releases)

## Go get

```shell
$ go get github.com/vlopatkin/go-dns-proxy
$ go-dns-proxy -use-outbound -json-config='{
    "default_dns": "8.8.8.8:53",
    "servers": {
        "google.com" : "8.8.8.8:53"
    },
    "domains": {
        "**.test.com" : "127.0.0.1"
    }
}'
```

## Arguments

```
  -expiration int
    	cache expiration time in seconds (default -1)
  -file string
    	config filename (default "config.json")
  -json-config string
    	config in json format
  -log-level string
    	log level, accepts err, info, none (default "info")
  -port int
    	UDP port, use with use-outbound flag (default 53)
  -use-outbound
    	use outbound address of the host for incoming connections
```

## Config file format

```json
{
    "host": "localhost:35353",
    "default_dns": "8.8.8.8:53",
    "servers": {
        "google.com" : "8.8.8.8:53"
    },
    "domains": {
        "**.test.com" : "127.0.0.1"
    }
}
```
