SHELL := /bin/bash

# These will be provided to the target
GIT := github.com/wormhole-foundation/wormhole-explorer/api/internal/
VERSION := 1.0.0
BUILD := `git rev-parse --short HEAD`
AUTHOR := `whoami`
BUILD_DATE := `date +%Y%m%d%H%M%S`
BRANCH := `git branch --show-current`
MACHINE := `uname -n`


# Use linker flags to provide version/build settings to the target
LDFLAGS=-ldflags "-X=$(GIT)build.Version=$(VERSION)\
				  -X=$(GIT)build.Time=$(BUILD_DATE)\
				  -X=$(GIT)build.Build=$(BUILD)\
 				  -X=$(GIT)build.Branch=$(BRANCH)\
				  -X=$(GIT)build.Machine=$(MACHINE)\
				  -X=$(GIT)build.User=$(AUTHOR)"


## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

build:
	CGO_ENABLED=0 GOOS=linux go build -v $(LDFLAGS) -o api main.go
	
doc:
	swag init -pd


test:
	go test -v -cover ./...


.PHONY: build doc test