VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo unknown)

.PHONY: build debug debug-run test lint format clean bootstrap docs docs-check

build:
	CGO_ENABLED=0 go build -ldflags "-X main.version=$(VERSION)" -o build/qcloud ./cmd/qcloud

debug:
	go build -gcflags="all=-N -l" -ldflags "-X main.version=$(VERSION)" -o build/qcloud-debug ./cmd/qcloud

debug-run: debug
	dlv exec ./build/qcloud-debug --headless --listen=:2345 --api-version=2 -- $(ARGS)

test:
	go test ./...

coverage:
	go test -coverpkg=./internal/... -coverprofile=build/coverage.txt -v -race ./...
	go tool cover -html=build/coverage.txt

lint:
	golangci-lint run

format:
	golangci-lint run --fix

clean:
	rm -rf build/

bootstrap:
	@command -v mise > /dev/null 2>&1 || \
		{ echo "mise is not installed. Install it from https://mise.jdx.dev/installing-mise.html"; exit 1; }
	mise install

docs:
	go run ./cmd/docgen ./docs/reference

docs-check: docs
	@git diff --exit-code docs/reference/ || \
		(echo "docs/reference is out of date — run 'make docs' and commit the result" && exit 1)

