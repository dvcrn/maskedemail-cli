
GO_SRC := $(wildcard ./*.go)
INSTALL_DIR := ${HOME}/.local/bin


all: build

build: bin/maskedemail-cli

install: build
	mv $< ${INSTALL_DIR}/.

bin/maskedemail-cli: ${GO_SRC}
	mkdir bin/ || true
	go build -o $@

.PHONY: clean
clean:
	rm -rf bin/
