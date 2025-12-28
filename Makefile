.PHONY: test lint fmt build clean install-hooks

GO ?= go
GOLANGCI_LINT ?= golangci-lint

test:
	$(GO) test -v -race -coverprofile=coverage.txt ./...

lint:
	$(GOLANGCI_LINT) run

fmt:
	$(GO) fmt ./...
	@command -v goimports >/dev/null 2>&1 && goimports -w . || true

build:
	$(GO) build ./...

clean:
	rm -f coverage.txt

install-hooks:
	./scripts/install-hooks.sh

.DEFAULT_GOAL := test
