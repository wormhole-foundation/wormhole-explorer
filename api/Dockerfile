FROM --platform=linux/amd64 docker.io/golang:1.20.7-bullseye@sha256:74b09b3b6fa5aa542df8ef974cb745eb477be72f6fcf821517fb410aff532b00 AS build
WORKDIR /app

COPY api api
COPY common common

# Build the Go app
RUN cd api && make build

############################
# STEP 2 build a small image
############################
FROM alpine
#Copy certificates
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy our static executable.
COPY --from=build "/app/api/api" "/api"
# Run the binary.
ENTRYPOINT ["/api"]
