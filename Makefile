VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

.PHONY: build test lint format clean bootstrap

build:
	CGO_ENABLED=0 go build -ldflags "-X main.version=$(VERSION)" -o build/qcloud ./cmd/qcloud

test:
	go test ./...

lint:
	golangci-lint run

format:
	$(GOLANGCI_LINT) run --fix

clean:
	rm -rf build/

bootstrap:
	@command -v mise > /dev/null 2>&1 || \
		{ echo "mise is not installed. Install it from https://mise.jdx.dev/installing-mise.html"; exit 1; }
	mise install

