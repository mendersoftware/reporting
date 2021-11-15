FROM golang:1.16.5-alpine3.12 as builder
RUN apk add --no-cache \
    xz-dev \
    musl-dev \
    gcc
RUN mkdir -p /go/src/github.com/mendersoftware/reporting
COPY . /go/src/github.com/mendersoftware/reporting
RUN cd /go/src/github.com/mendersoftware/reporting && env CGO_ENABLED=1 go build

FROM alpine:3.14.3
RUN apk add --no-cache ca-certificates xz
RUN mkdir -p /etc/reporting
COPY ./config.yaml /etc/reporting
COPY --from=builder /go/src/github.com/mendersoftware/reporting/reporting /usr/bin
ENTRYPOINT ["/usr/bin/reporting", "--config", "/etc/reporting/config.yaml"]

EXPOSE 8080
