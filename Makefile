.PHONY: all test build-binary install clean

GO := CGO_ENABLED=0 go

all: test build-binary

test:
	$(GO) test ./...

build-binary:
	$(GO) build -tags netgo -ldflags "-w" -o nextcloud-exporter .

install:
	install nextcloud-exporter /usr/local/bin/
	install contrib/nextcloud-exporter.service /etc/systemd/system/

clean:
	rm -f nextcloud-exporter
