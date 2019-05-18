GOCMD   := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOGET   := $(GOCMD) get
INSTALL := install
UPX     := upx

VERSION := $(shell git describe --exact-match --abbrev=0 2>/dev/null)
ifeq ($(VERSION),)
	VERSION := dev-$(shell git rev-parse --short HEAD)
endif

OUTDIR        := out
ARCH_NATIVE   := $(shell go env GOARCH)
BINARY_NATIVE := $(OUTDIR)/gr-$(ARCH_NATIVE)-$(VERSION)
ifeq ($(ARCH_NATIVE),amd64)
	ARCH_FOREIGN := 386
else
	ARCH_FOREIGN := amd64
endif
BINARY_FOREIGN := $(OUTDIR)/gr-$(ARCH_FOREIGN)-$(VERSION)
INSTALL_PATH   := /usr/bin/gr

.PHONY: all
all: $(BINARY_NATIVE) $(BINARY_FOREIGN)

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NATIVE) $(BINARY_FOREIGN)
	rmdir $(OUTDIR)

.PHONY: deps
deps:
	$(GOGET) -d -v ./...

out:
	mkdir -p $@

$(BINARY_FOREIGN) : export GOARCH = $(ARCH_FOREIGN)
$(BINARY_NATIVE) $(BINARY_FOREIGN) : | deps $(OUTDIR)
	$(GOBUILD) -ldflags="-s -w" -o $@
	$(UPX) -9 $@

.PHONY: test
test:
	echo "Not implemented"

.PHONY: install
install: $(BINARY_NATIVE)
	$(INSTALL) -o root -g root -m 0755 $(BINARY_NATIVE) $(INSTALL_PATH)

.PHONY: uninstall
uninstall:
	rm -f $(INSTALL_PATH)
