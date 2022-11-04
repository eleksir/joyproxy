#!/usr/bin/env gmake -f

GOOPTS=CGO_ENABLED=0
BUILDOPTS=-ldflags="-s -w" -a -gcflags=all=-l -trimpath

all: clean build

build:
	${GOOPTS} go build ${BUILDOPTS}

clean:
	go clean

upgrade:
	rm -rf vendor
	go get -d -u -t ./...
	go mod vendor

# vim: set ft=make noet ai ts=4 sw=4 sts=4:
