#!/bin/sh
GOOS="linux" GOARCH="amd64" go build -o tcp-proxy cmd/tcp-proxy/main.go