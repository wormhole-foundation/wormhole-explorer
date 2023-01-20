SHELL := /bin/bash


## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

build:
	make -C api/ build
	make -C fly/ build
	make -C spy/ build
	
doc:
	swag init -pd


test:
	cd api && go test -v -cover ./...
	cd fly && go test -v -cover ./...
	cd spy && go test -v -cover ./...


.PHONY: build doc test
