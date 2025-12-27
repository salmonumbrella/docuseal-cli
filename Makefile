.PHONY: build run test fmt lint clean setup

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE)

setup:
	@command -v lefthook >/dev/null || (echo "Install lefthook: brew install lefthook" && exit 1)
	lefthook install

build:
	go build -ldflags "$(LDFLAGS)" -o bin/docuseal ./cmd/docuseal

run: build
	./bin/docuseal

test:
	go test ./...

fmt:
	go fmt ./...

lint:
	go vet ./...

clean:
	rm -rf bin/
