SHELL := /bin/bash


## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

build:
	go build -o spy cmd/main.go
	
doc:
	swag init -pd


test:
	go test -v -cover ./...


.PHONY: build doc test
