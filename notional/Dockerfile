# syntax=docker.io/docker/dockerfile:1.3@sha256:42399d4635eddd7a9b8a24be879d2f9a930d0ed040a61324cfdf59ef1357b3b2
FROM --platform=linux/amd64 docker.io/golang:1.19.2@sha256:0467d7d12d170ed8d998a2dae4a09aa13d0aa56e6d23c4ec2b1e4faacf86a813 AS build

WORKDIR /app

COPY notional notional
COPY common common

# Build the Go app
RUN cd notional && make build

############################
# STEP 2 build a small image
############################
FROM alpine
#Copy certificates
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy our static executable.
COPY --from=build "/app/notional/notional" "/notional"
# Run the binary.
ENTRYPOINT ["/notional"]
