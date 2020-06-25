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
VERSION_DEV   := $(shell git describe | sed "s/-/+/" | sed "s/-/./")
VERSION       := $(or $(VERSION_PROD),$(VERSION_DEV))

BINARY_386    := $(OUTDIR)/gr_linux_386
BINARY_AMD64  := $(OUTDIR)/gr_linux_amd64
BINARY_NATIVE := $(OUTDIR)/gr_$(shell go env GOOS)_$(shell go env GOARCH)
BINARIES      := $(sort $(BINARY_386) $(BINARY_AMD64) $(BINARY_NATIVE))
LDFLAGS       := -s -w -X 'main.Version=$(VERSION)' -X 'main.BuildDate=$(BUILDDATE)'

COVERAGE_OUTPUT := coverage.out
COVERAGE_HTML   := coverage.html

.DEFAULT_GOAL := $(BINARY_NATIVE)

.PHONY: all
all: $(BINARIES)

.PHONY: check
check:
	#docker run --rm -v $(shell pwd):/app:ro -w /app golang gofmt -d ./
	docker run --rm -v $(shell pwd):/app:ro -w /app golangci/golangci-lint golangci-lint run
	#docker run --rm -v $(shell pwd):/app:ro -w /app georgeok/goreportcard-cli /bin/sh -c "go get -d ./... && go get github.com/georgeok/goreportcard/cmd/goreportcard-cli && goreportcard-cli -v"

lint:
	$(GOGET) -v github.com/alecthomas/gometalinter
	$(GOGET) -v github.com/gordonklaus/ineffassign
	$(GOGET) -v github.com/client9/misspell/cmd/misspell
	$(GOGET) -v golang.org/x/lint/golint
	$(GOGET) -v github.com/fzipp/gocyclo
	$(GOGET) -v github.com/gojp/goreportcard/cmd/goreportcard-cli
	$(GOGET) -v github.com/golangci/golangci-lint/cmd/golangci-lint
	goreportcard-cli -v
	golangci-lint run

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
	$(GOBUILD) -ldflags="$(LDFLAGS)" -o $@
	$(UPX) -9 $@
	sha256sum $@ | awk '{print $$1}' > $@.sha256

.PHONY: test
test: deps
	go test -v -cover -covermode=atomic -coverprofile=$(COVERAGE_OUTPUT) ./cmd
	go tool cover -html=$(COVERAGE_OUTPUT) -o=$(COVERAGE_HTML)
	cp $(COVERAGE_HTML) /data/html/index.html

.PHONY: install
install: $(BINARY_NATIVE)
	$(INSTALL) -o root -g root -m 0755 $(BINARY_NATIVE) $(INSTALL_PATH)

.PHONY: uninstall
uninstall:
	rm -f $(INSTALL_PATH)
