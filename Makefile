.PHONY: all test build-binary install clean

GO ?= go
GO_CMD := CGO_ENABLED=0 $(GO)
VERSION := $(shell git describe --tags --broken)
GIT_COMMIT := $(shell git rev-parse HEAD)

all: test build-binary

test:
	$(GO_CMD) test ./...

build-binary:
	$(GO_CMD) build -tags netgo -ldflags "-w -X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT)" -o nextcloud-exporter .

install:
	install -D -t $(DESTDIR)/usr/bin/ nextcloud-exporter
	install -D -m 0644 -t $(DESTDIR)/lib/systemd/system/ contrib/nextcloud-exporter.service

clean:
	rm -f nextcloud-exporter
