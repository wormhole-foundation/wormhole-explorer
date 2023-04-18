# syntax=docker.io/docker/dockerfile:1.3@sha256:42399d4635eddd7a9b8a24be879d2f9a930d0ed040a61324cfdf59ef1357b3b2
FROM --platform=linux/amd64 docker.io/golang:1.19.8@sha256:a507d049fcda8f10e5ea6e84784905defbf66b48af434d85ec443fa6aee0bfc9 AS build

WORKDIR /app

COPY parser parser
COPY common common

# Build the Go app
RUN cd parser && CGO_ENABLED=0 GOOS=linux go build -o "./parser-pipeline" cmd/main.go

############################
# STEP 2 build a small image
############################
FROM alpine
#Copy certificates
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy our static executable.
COPY --from=build "/app/parser/parser-pipeline" "/parser-pipeline"
# Run the binary.
ENTRYPOINT ["/parser-pipeline"]
