.PHONY: build install clean

BINARY_NAME=doman
INSTALL_PATH=/usr/local/bin

build:
	go build -o bin/$(BINARY_NAME) .

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
