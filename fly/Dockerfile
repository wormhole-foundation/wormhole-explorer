FROM --platform=linux/amd64 docker.io/golang:1.20.7-bullseye@sha256:74b09b3b6fa5aa542df8ef974cb745eb477be72f6fcf821517fb410aff532b00 AS build

WORKDIR /app

COPY fly fly
COPY common common

# Build the Go app
RUN cd fly && CGO_ENABLED=1 GOOS=linux go build -o "./fly" main.go && \
  go get github.com/CosmWasm/wasmvm@v1.0.0 && \
  cp /go/pkg/mod/github.com/!cosm!wasm/wasmvm@v1.0.0/api/libwasmvm.x86_64.so /usr/lib/


############################
# STEP 2 build a small image
############################
FROM alpine
#Copy certificates
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
#Copy libraries
COPY --from=build /lib/* /lib/
COPY --from=build /lib64/* /lib64/
COPY --from=build /usr/lib/libwasmvm.x86_64.so /usr/lib/
#Copy our static executable.
COPY --from=build "/app/fly/fly" "/fly"

# Run the binary.
ENTRYPOINT ["/fly"]
