FROM golang:1.16.5-alpine3.12 as builder
WORKDIR /go/src/github.com/mendersoftware/iot-manager
RUN apk add --no-cache \
    xz-dev \
    musl-dev \
    gcc \
    ca-certificates
COPY ./ .
RUN CGO_ENABLED=0 go build

FROM scratch
WORKDIR /etc/iot-manager
EXPOSE 8080
COPY ./config.yaml .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/mendersoftware/iot-manager/iot-manager /usr/bin/

ENTRYPOINT ["/usr/bin/iot-manager"]
