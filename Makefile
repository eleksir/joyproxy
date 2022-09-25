#!/usr/bin/env gmake -f

GOOPTS=CGO_ENABLED=0
BUILDOPTS=-ldflags="-s -w" -a -gcflags=all=-l

all: clean build

build:
	${GOOPTS} go build ${BUILDOPTS}

clean:
	go clean

wipe:
	go clean
	rm -rf go.{mod,sum}
	rm -rf vendor

prep:
	go mod init joyproxy
	go mod tidy
	go mod vendor

# vim: set ft=make noet ai ts=4 sw=4 sts=4:
