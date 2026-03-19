VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || "undefined")

.PHONY: build debug debug-run test lint format clean bootstrap generate generate-verify

build:
	CGO_ENABLED=0 go build -ldflags "-X main.version=$(VERSION)-dev" -o build/qcloud ./cmd/qcloud

debug:
	go build -gcflags="all=-N -l" -ldflags "-X main.version=$(VERSION)-dev" -o build/qcloud-debug ./cmd/qcloud

debug-run: debug
	dlv exec ./build/qcloud-debug --headless --listen=:2345 --api-version=2 -- $(ARGS)

test:
	go test ./...

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

generate:
	mise exec -- mockery

generate-verify:
	@git diff --exit-code internal/testutil/mocks/ || \
		{ echo "Generated mocks are out of date. Run 'make generate'."; exit 1; }

