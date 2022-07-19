# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

export GO111MODULE=on

all: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
        go build -a -ldflags '-extldflags "-static"' \
        -o ./monitor main.go

build: clean test cover
deploy: build

.PHONY: test
test: clean
	@go test -v ./...

format:
	 @go fmt ./...

clean:
	@rm -f ./monitor

.PHONY: cover
cover:
	@go test -p 2 ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html

