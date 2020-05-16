.PHONY: all test build-binary install clean

GO := CGO_ENABLED=0 go
VERSION := $(shell git describe --tags HEAD)
GIT_COMMIT := $(shell git rev-parse HEAD)

all: test build-binary

test:
	$(GO) test ./...

build-binary:
	$(GO) build -tags netgo -ldflags "-w -X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT)" -o nextcloud-exporter .

install:
	install nextcloud-exporter /usr/local/bin/
	install contrib/nextcloud-exporter.service /etc/systemd/system/

clean:
	rm -f nextcloud-exporter
