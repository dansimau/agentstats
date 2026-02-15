.PHONY: build build-all install test clean

# Detect current platform for local builds.
OS   := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
PLATFORM := $(OS)-$(ARCH)
BINARY := bin/agentstats-$(PLATFORM)

build:
	@mkdir -p bin
	go build -o $(BINARY) ./cmd/agentstats
	@chmod +x $(BINARY) bin/agentstats
	@echo "Built $(BINARY)"

install: build
	@go install ./cmd/agentstats
	@echo "Installed agentstats to $$(go env GOPATH)/bin/agentstats"

build-all:
	@bash scripts/build.sh

test:
	go test -race ./...

clean:
	rm -f bin/agentstats-darwin-* bin/agentstats-linux-*
