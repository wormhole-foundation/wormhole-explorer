FROM --platform=linux/amd64 docker.io/golang:1.21.9-bullseye@sha256:311468bffa9fa4747a334b94e6ce3681b564126d653675a6adc46698b2b88d35 AS build

WORKDIR /app

COPY fly fly
COPY common common

RUN --mount=type=cache,target=/root/.go --mount=type=cache,target=/go \
  cd fly && CGO_ENABLED=1 GOOS=linux go build -ldflags=-extldflags=-Wl,--allow-multiple-definition -o "./fly" main.go && \
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
