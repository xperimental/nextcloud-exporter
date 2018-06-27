#!/usr/bin/env bash

set -e -u -o pipefail

VERSION=$(git rev-parse --short HEAD)
readonly VERSION

docker build --pull -t "xperimental/nextcloud-exporter:${VERSION}" .
