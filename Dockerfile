FROM golang:1 AS builder

RUN apt-get update && apt-get install -y upx

ENV PACKAGE=github.com/xperimental/nextcloud-exporter

RUN mkdir -p /go/src/${PACKAGE}
WORKDIR /go/src/${PACKAGE}

ENV LD_FLAGS="-w"
ENV CGO_ENABLED=0

COPY . /go/src/${PACKAGE}
RUN echo "-- TEST" \
 && go test ./... \
 && echo "-- BUILD" \
 && go install -tags netgo -ldflags "${LD_FLAGS}" . \
 && echo "-- PACK" \
 && upx -9 /go/bin/nextcloud-exporter

FROM busybox
LABEL maintainer="Robert Jacob <xperimental@solidproject.de>"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/bin/nextcloud-exporter /bin/nextcloud-exporter

USER nobody
EXPOSE 9205

ENTRYPOINT ["/bin/nextcloud-exporter"]
