GOCMD         := go
GOBUILD       := $(GOCMD) build
GOCLEAN       := $(GOCMD) clean
GOGET         := $(GOCMD) get
INSTALL       := install
UPX           := upx
OUTDIR        := out
INSTALL_PATH  := /usr/bin/gr

BUILDDATE     := $(shell date --rfc-3339=seconds)
VERSION_PROD  := $(shell git describe --exact-match --abbrev=0 2>/dev/null)
VERSION_DEV   := dev-$(shell git rev-parse --short HEAD)
VERSION       := $(or $(VERSION_PROD),$(VERSION_DEV))

BINARY_386    := $(OUTDIR)/gr-$(VERSION)-386
BINARY_AMD64  := $(OUTDIR)/gr-$(VERSION)-amd64
BINARY_NATIVE := $(OUTDIR)/gr-$(VERSION)-$(shell go env GOARCH)
BINARIES      := $(sort $(BINARY_386) $(BINARY_AMD64) $(BINARY_NATIVE))
LDFLAGS       := -s -w -X 'main.Version=$(VERSION)' -X 'main.BuildDate=$(BUILDDATE)'

.PHONY: all
all: $(BINARIES)

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARIES)
	rmdir $(OUTDIR)

.PHONY: deps
deps:
	$(GOGET) -d -t -v ./...

$(OUTDIR):
	mkdir -p $@

$(BINARY_386)   : export GOARCH = 386
$(BINARY_AMD64) : export GOARCH = amd64
$(BINARIES) : | deps $(OUTDIR)
	$(GOBUILD) -ldflags="$(LDFLAGS)" -o $@
	$(UPX) -9 $@
	sha1sum $@ | awk '{print $$1}' > $@.sha1sum

.PHONY: test
test:
	echo "Not implemented"

.PHONY: install
install: $(BINARY_NATIVE)
	$(INSTALL) -o root -g root -m 0755 $(BINARY_NATIVE) $(INSTALL_PATH)

.PHONY: uninstall
uninstall:
	rm -f $(INSTALL_PATH)
