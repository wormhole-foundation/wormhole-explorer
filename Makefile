SHELL := /bin/bash


## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

build:
	make -C analytics/ build
	make -C api/ build
	make -C fly/ build
	make -C fly-event-processor/ build
	make -C jobs/ build
	make -C parser/ build
	make -C pipeline/ build
	make -C spy/ build
	make -C tx-tracker/ build
	
doc:
	swag init -pd

test:
	cd analytics && go test -v -cover ./...
	cd api && go test -v -cover ./...
	cd fly && go test -v -cover ./...
	cd spy && go test -v -cover ./...
	cd parser && go test -v -cover ./...
	cd tx-tracker && go test -v -cover ./...

.PHONY: build doc test
