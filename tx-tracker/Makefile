SHELL := /bin/bash


## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

build:
	CGO_ENABLED=0 GOOS=linux go build -o "./tx-tracker" cmd/service/main.go

test:
	go test -v -cover ./...


.PHONY: build test
