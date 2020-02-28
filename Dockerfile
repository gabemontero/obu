FROM golang:1.13.8 AS builder
WORKDIR /go/src/github.com/gabemontero/obu
COPY . .
RUN go build ./cmd/obu/
RUN pwd
RUN ls -la

FROM registry.access.redhat.com/ubi8/ubi:8.1-397
COPY --from=builder /go/src/github.com/gabemontero/obu/obu /usr/bin/
RUN useradd obu-user
USER obu-user
ENTRYPOINT []
CMD ["/usr/bin/obu"]