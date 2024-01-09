SHELL := /bin/bash


build:
	CGO_ENABLED=0 GOOS=linux go build -o notional cmd/main.go
	
doc:
	swag init -pd


test:
	go test -v -cover ./...


.PHONY: build doc test