GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOGET = $(GOCMD) get
INSTALL = install
UPX = upx

BINARY_NAME = gr
INSTALL_PATH = /usr/bin/gr

all: $(BINARY_NAME)

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

.PHONY: deps
deps:
	$(GOGET) -d -v ./...

$(BINARY_NAME) :
	$(GOGET) -d -v ./...
	$(GOBUILD) -ldflags="-s -w" -o $(BINARY_NAME)
	$(UPX) -9 $(BINARY_NAME)

.PHONY: install
install: $(BINARY_NAME)
	$(INSTALL) -o root -g root -m 0755 $(BINARY_NAME) $(INSTALL_PATH)

.PHONY: uninstall
uninstall:
	rm -f $(INSTALL_PATH)
