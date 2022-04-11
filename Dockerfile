FROM golang:1.17.6-alpine3.15 as builder
RUN apk add --no-cache \
    xz-dev \
    musl-dev \
    gcc
RUN mkdir -p /go/src/github.com/mendersoftware/reporting
COPY . /go/src/github.com/mendersoftware/reporting
RUN cd /go/src/github.com/mendersoftware/reporting && env CGO_ENABLED=1 go build

FROM alpine:3.15.4
RUN apk add --no-cache ca-certificates xz
RUN mkdir -p /etc/reporting
COPY ./config.yaml /etc/reporting
COPY --from=builder /go/src/github.com/mendersoftware/reporting/reporting /usr/bin
ENTRYPOINT ["/usr/bin/reporting", "--config", "/etc/reporting/config.yaml"]

EXPOSE 8080
