.PHONY: build install clean

BINARY_NAME=doman
INSTALL_PATH=/usr/local/bin

COMMIT_HASH := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

build:
	go build -o bin/$(BINARY_NAME) . -ldflags "-X 'github.com/connordoman/doman/version.CommitHash=$(COMMIT_HASH)' -X 'github.com/connordoman/doman/version.BuildTime=$(BUILD_TIME)'"

install: build
	sudo cp bin/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)

install-user: build
	mkdir -p $(HOME)/.local/bin
	cp bin/$(BINARY_NAME) $(HOME)/.local/bin/$(BINARY_NAME)

clean:
	rm -rf bin/

uninstall:
	sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)

uninstall-user:
	rm -f $(HOME)/.local/bin/$(BINARY_NAME)
