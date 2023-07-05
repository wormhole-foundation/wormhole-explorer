SHELL := /bin/bash


build:
	go build -o bin/service cmd/main.go
	
test:
	go test -v -cover ./...


.PHONY: build doc test
