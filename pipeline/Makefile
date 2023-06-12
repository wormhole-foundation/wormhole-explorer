SHELL := /bin/bash

build:
	CGO_ENABLED=0 GOOS=linux go build -o "./bin/pipeline" cmd/main.go

test:
	go test -v -cover ./...

.PHONY: build test
