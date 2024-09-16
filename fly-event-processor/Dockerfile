ARG BUILDPLATFORM="linux/amd64"
FROM --platform=${BUILDPLATFORM} docker.io/golang:1.21.9-bullseye@sha256:311468bffa9fa4747a334b94e6ce3681b564126d653675a6adc46698b2b88d35 AS build

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
