ARG BUILDPLATFORM="linux/amd64"
FROM --platform=${BUILDPLATFORM} docker.io/golang:1.21.9-bullseye@sha256:311468bffa9fa4747a334b94e6ce3681b564126d653675a6adc46698b2b88d35 AS build

WORKDIR /app

COPY pipeline pipeline
COPY common common

# Build the Go app
RUN cd pipeline && make build

############################
# STEP 2 build a small image
############################
FROM alpine
#Copy certificates
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy our static executable.
COPY --from=build "/app/pipeline/pipeline-app" "/pipeline"
# Run the binary.
ENTRYPOINT ["/pipeline"]
