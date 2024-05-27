SHELL := /bin/bash


## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

build:
	go build -ldflags=-extldflags=-Wl,--allow-multiple-definition -o fly main.go
	
doc:
	swag init -pd


test:
	go test -ldflags=-extldflags=-Wl,--allow-multiple-definition -v -cover ./...


.PHONY: build doc test
