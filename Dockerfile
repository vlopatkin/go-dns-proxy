FROM golang:1.12-alpine as pacpac
RUN apk add --update --no-cache git
WORKDIR /src/go-dns-proxy
ADD * ./
RUN go get -v -t -d ./...
RUN go build -o go-dns-proxy

FROM alpine:latest
RUN apk --no-cache add ca-certificates
EXPOSE 53/udp
WORKDIR /root/
COPY --from=pacpac /src/go-dns-proxy/go-dns-proxy .
COPY --from=pacpac /src/go-dns-proxy/config.json .
ENTRYPOINT ["./go-dns-proxy"]