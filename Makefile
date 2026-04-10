.PHONY: build test clean

VERSION ?= dev
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/dbx-agent ./cmd/dbx-agent

test:
	go test ./... -v -count=1

clean:
	rm -rf bin/
