FROM golang:1.24.0-alpine AS builder

RUN apk update && \
    apk add --no-cache bash

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./tun-sniffer ./cmd

FROM alpine:3.19

RUN apk update && \
    apk add --no-cache iproute2 iptables bash

WORKDIR /app

COPY --from=builder /app/tun-sniffer /app/tun-sniffer

RUN printf '#!/bin/sh\n\
    set -e\n\
    mkdir -p /dev/net\n\
    [ -c /dev/net/tun ] || mknod /dev/net/tun c 10 200\n\
    chmod 600 /dev/net/tun\n\
    exec ./tun-sniffer \\\n\
    -tunIP="$TUN_IP" \\\n\
    -tunRoute="$TUN_ROUTE"\n' \
    > /app/entrypoint.sh && chmod +x /app/entrypoint.sh

ARG TUN_IP=10.0.0.1/24
ARG TUN_ROUTE=10.0.0.0/24

ENV TUN_IP=${TUN_IP}
ENV TUN_ROUTE=${TUN_ROUTE}

ENTRYPOINT [ "/app/entrypoint.sh" ]
