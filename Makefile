.PHONY: build install clean

BINARY_NAME=doman
INSTALL_PATH=/usr/local/bin

MODULE := github.com/connordoman/doman
CONFIG_PATH := internal/config

DEFAULT_VERSION := 0.0.0
DEFAULT_BUILD := dev

VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo v$(DEFAULT_VERSION))
COMMIT_HASH := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
BUILD ?= $(DEFAULT_BUILD)

BRANCH_NAME := $(shell git branch --show-current)

LD_FLAGS := -X '$(MODULE)/$(CONFIG_PATH).CommitHash=$(COMMIT_HASH)' \
			-X '$(MODULE)/$(CONFIG_PATH).BuildDate=$(BUILD_TIME)' \
			-X '$(MODULE)/$(CONFIG_PATH).Version=$(VERSION)' \
			-X '$(MODULE)/$(CONFIG_PATH).Build=$(BUILD)'

ldflags:
	@echo $(LD_FLAGS)

build:
	go build -ldflags "$(LD_FLAGS)" .

install:
	go install -ldflags "$(LD_FLAGS)" .

install-user:
	go install -ldflags "$(LD_FLAGS)" .

clean:
	go clean -cache

uninstall:
	sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)

uninstall-user:
	rm -f $(HOME)/.local/bin/$(BINARY_NAME)
