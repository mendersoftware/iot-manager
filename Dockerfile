FROM golang:1.16.5-alpine3.12 as builder
RUN apk add --no-cache \
    xz-dev \
    musl-dev \
    gcc
RUN mkdir -p /go/src/github.com/mendersoftware/iot-manager
COPY . /go/src/github.com/mendersoftware/iot-manager
RUN cd /go/src/github.com/mendersoftware/iot-manager && env CGO_ENABLED=1 go build

FROM alpine:3.15.0
RUN apk add --no-cache ca-certificates xz
RUN mkdir -p /etc/iot-manager
COPY ./config.yaml /etc/iot-manager
COPY --from=builder /go/src/github.com/mendersoftware/iot-manager/iot-manager /usr/bin
ENTRYPOINT ["/usr/bin/iot-manager"]

EXPOSE 8080
