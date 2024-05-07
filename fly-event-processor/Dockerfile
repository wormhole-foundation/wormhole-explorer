# syntax=docker.io/docker/dockerfile:1.3@sha256:42399d4635eddd7a9b8a24be879d2f9a930d0ed040a61324cfdf59ef1357b3b2
FROM --platform=linux/amd64 docker.io/golang:1.20.7-bullseye@sha256:74b09b3b6fa5aa542df8ef974cb745eb477be72f6fcf821517fb410aff532b00 AS build
WORKDIR /app

COPY fly-event-processor fly-event-processor
COPY common common

# Build the Go app
RUN cd fly-event-processor && CGO_ENABLED=0 GOOS=linux go build -o "./fly-event-processor" cmd/main.go

############################
# STEP 2 build a small image
############################
FROM alpine
#Copy certificates
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy our static executable.
COPY --from=build "/app/fly-event-processor/fly-event-processor" "/fly-event-processor"
# Run the binary.
ENTRYPOINT ["/fly-event-processor"]
