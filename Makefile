SHELL := /bin/bash
GO ?= go
GO_CMD := CGO_ENABLED=0 $(GO)
GIT_VERSION := $(shell git describe --tags --dirty)
VERSION := $(GIT_VERSION:v%=%)
GIT_COMMIT := $(shell git rev-parse HEAD)
DOCKER_REPO ?= xperimental/nextcloud-exporter
DOCKER_TAG ?= dev

include .bingo/Variables.mk

.PHONY: all
all: test build-binary

.PHONY: test
test:
	$(GO_CMD) test -cover ./...

.PHONY: lint
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run --fix

.PHONY: build-binary
build-binary:
	$(GO_CMD) build -tags netgo -ldflags "-w -X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT)" -o nextcloud-exporter .

.PHONY: deb
deb: build-binary
	mkdir -p dist/deb/DEBIAN dist/deb/usr/bin
	sed 's/%VERSION%/$(VERSION)/' contrib/debian/control > dist/deb/DEBIAN/control
	cp nextcloud-exporter dist/deb/usr/bin/
	fakeroot dpkg-deb --build dist/deb dist

.PHONY: install
install:
	install -D -t $(DESTDIR)/usr/bin/ nextcloud-exporter
	install -D -m 0644 -t $(DESTDIR)/lib/systemd/system/ contrib/nextcloud-exporter.service

.PHONY: image
image:
	docker buildx build -t "ghcr.io/$(DOCKER_REPO):$(DOCKER_TAG)" --load .

.PHONY: all-images
all-images:
	docker buildx build -t "ghcr.io/$(DOCKER_REPO):$(DOCKER_TAG)" -t "docker.io/$(DOCKER_REPO):$(DOCKER_TAG)" --platform linux/amd64,linux/arm64 --push .

.PHONY: tools
tools: $(BINGO) $(GOLANGCI_LINT)
	@echo Tools built.

.PHONY: clean
clean:
	rm -f nextcloud-exporter
	rm -r dist
