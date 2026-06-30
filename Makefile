BINARY  := valence
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

.PHONY: build test clean install release $(PLATFORMS)

## build: compile for the current platform (version injected from git)
build:
	go build $(LDFLAGS) -o $(BINARY) .
	@echo "Built $(BINARY) $(VERSION)"

## test: run the full test suite
test:
	go test ./... -v

## install: build and copy to /usr/local/bin
install: build
	install -m 0755 $(BINARY) /usr/local/bin/$(BINARY)
	@echo "Installed to /usr/local/bin/$(BINARY)"

## release: cross-compile for all platforms and create .tar.gz archives in dist/
release: test
	@mkdir -p dist
	@$(foreach platform,$(PLATFORMS), \
		$(MAKE) _build_platform GOOS=$(word 1,$(subst /, ,$(platform))) GOARCH=$(word 2,$(subst /, ,$(platform)));)
	@echo "Release archives in dist/ for $(VERSION)"

_build_platform:
	@echo "  Building $(GOOS)/$(GOARCH)…"
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o dist/$(BINARY)_$(GOOS)_$(GOARCH) .
	tar -czf dist/$(BINARY)_$(GOOS)_$(GOARCH).tar.gz -C dist $(BINARY)_$(GOOS)_$(GOARCH)
	@rm dist/$(BINARY)_$(GOOS)_$(GOARCH)

## clean: remove built binaries and dist/ archives
clean:
	@rm -f $(BINARY)
	@rm -rf dist/
	@echo "Cleaned"
