
build:
	CGO_ENABLED=0 GOOS=linux go build -o "./bin/service" cmd/service/main.go

test:
	go test -v -cover ./...


.PHONY: build test