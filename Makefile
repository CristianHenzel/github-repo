GOCMD         := go
GOBUILD       := $(GOCMD) build
GOCLEAN       := $(GOCMD) clean
GOGET         := $(GOCMD) get
INSTALL       := install
UPX           := upx
OUTDIR        := out
INSTALL_PATH  := /usr/bin/gr

BUILDDATE     := $(shell date -u "+%Y-%m-%d %H:%M:%S UTC")
VERSION_PROD  := $(shell git describe --all --exact-match --abbrev=0 2>/dev/null | sed "s/tags\/v//")
VERSION_DEV   := $(shell git describe --all | sed "s/-/+/" | sed "s/-/./")
VERSION       := $(or $(VERSION_PROD),$(VERSION_DEV))

BINARY_386    := $(OUTDIR)/gr_linux_386
BINARY_AMD64  := $(OUTDIR)/gr_linux_amd64
BINARY_NATIVE := $(OUTDIR)/gr_$(shell go env GOOS)_$(shell go env GOARCH)
BINARIES      := $(sort $(BINARY_386) $(BINARY_AMD64) $(BINARY_NATIVE))
BUILD_TAGS    := osusergo netgo static_build
LDFLAGS       := -s -w -X 'main.Version=$(VERSION)' -X 'main.BuildDate=$(BUILDDATE)' -extldflags=-static

.DEFAULT_GOAL := $(BINARY_NATIVE)

.PHONY: all
all: $(BINARIES)

.PHONY: check
check:
	docker run --rm -v $(shell pwd):/app -w /app golang gofmt -d ./
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint golangci-lint run

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARIES)
	rm -rf $(OUTDIR)/*.sha256
	rmdir $(OUTDIR)

.PHONY: deps
deps:
	$(GOGET) -d -t -v ./...

$(OUTDIR):
	mkdir -p $@

$(BINARY_386)   : export GOARCH = 386
	export GOOS = linux
$(BINARY_AMD64) : export GOARCH = amd64
	export GOOS = linux
$(BINARIES) : | deps $(OUTDIR)
	$(GOBUILD) -ldflags="$(LDFLAGS)" -tags="$(BUILD_TAGS)" -o "$@.uncompressed"
	sha256sum "$@.uncompressed" | awk '{print $$1}' > "$@.uncompressed.sha256"
	$(UPX) --best "$@.uncompressed" -o"$@"
	sha256sum "$@" | awk '{print $$1}' > "$@.sha256"

.PHONY: test
test:
	echo "Not implemented"

.PHONY: install
install: $(BINARY_NATIVE)
	$(INSTALL) -o root -g root -m 0755 $(BINARY_NATIVE) $(INSTALL_PATH)

.PHONY: uninstall
uninstall:
	rm -f $(INSTALL_PATH)
