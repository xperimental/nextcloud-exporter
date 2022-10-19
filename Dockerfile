FROM golang:1.19.2 AS builder

WORKDIR /build

COPY go.mod go.sum /build/
RUN go mod download
RUN go mod verify

COPY . /build/
RUN make

FROM busybox
LABEL maintainer="Robert Jacob <xperimental@solidproject.de>"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /build/nextcloud-exporter /bin/nextcloud-exporter

USER nobody
EXPOSE 9205

ENTRYPOINT ["/bin/nextcloud-exporter"]
