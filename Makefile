.PHONY: all test build-binary install clean

GO ?= go
GO_CMD := CGO_ENABLED=0 $(GO)
GIT_VERSION := $(shell git describe --tags --dirty)
VERSION := $(GIT_VERSION:v%=%)
GIT_COMMIT := $(shell git rev-parse HEAD)

all: test build-binary

test:
	$(GO_CMD) test -cover ./...

build-binary:
	$(GO_CMD) build -tags netgo -ldflags "-w -X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT)" -o nextcloud-exporter .

deb: build-binary
	mkdir -p dist/deb/DEBIAN dist/deb/usr/bin
	sed 's/%VERSION%/$(VERSION)/' contrib/debian/control > dist/deb/DEBIAN/control
	cp nextcloud-exporter dist/deb/usr/bin/
	fakeroot dpkg-deb --build dist/deb dist

install:
	install -D -t $(DESTDIR)/usr/bin/ nextcloud-exporter
	install -D -m 0644 -t $(DESTDIR)/lib/systemd/system/ contrib/nextcloud-exporter.service

clean:
	rm -f nextcloud-exporter
	rm -r dist
