FROM golang:1.13.8 AS builder
WORKDIR /go/src/github.com/gabemontero/obu
COPY . .
RUN go build ./cmd/obu/

FROM registry.access.redhat.com/ubi8/ubi:8.1-397
COPY --from=builder /go/src/github.com/gabemontero/obu/obu /usr/bin/
# tekton current seems to mess with home/passwd stuff, so the new user is commented out
#RUN useradd obu-user
#USER obu-user
ENTRYPOINT []
CMD ["/usr/bin/obu"]