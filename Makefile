SHELL := /bin/bash
GO ?= go
GO_CMD := CGO_ENABLED=0 $(GO)
GIT_VERSION := $(shell git describe --tags --dirty)
VERSION := $(GIT_VERSION:v%=%)
GIT_COMMIT := $(shell git rev-parse HEAD)
GITHUB_REF ?= refs/heads/master
DOCKER_TAG != if [[ "$(GITHUB_REF)" == "refs/heads/master" ]]; then \
		echo "latest"; \
	else \
		echo "$(VERSION)"; \
	fi

.PHONY: all
all: test build-binary

.PHONY: test
test:
	$(GO_CMD) test -cover ./...

.PHONY: lint
lint:
	golangci-lint run --fix

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
	docker build -t "xperimental/nextcloud-exporter:$(DOCKER_TAG)" .

.PHONY: all-images
all-images:
	docker buildx build -t "ghcr.io/xperimental/nextcloud-exporter:$(DOCKER_TAG)" -t "xperimental/nextcloud-exporter:$(DOCKER_TAG)" --platform linux/amd64,linux/arm64 --push .

.PHONY: clean
clean:
	rm -f nextcloud-exporter
	rm -r dist
