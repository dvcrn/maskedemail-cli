
GO_SRC := $(wildcard ./*.go ./pkg/*.go)
INSTALL_DIR := ${HOME}/.local/bin
TARGET_BIN := bin/maskedemail-cli

all: build

install: build
	mv ${TARGET_BIN} ${INSTALL_DIR}/.

build: ${TARGET_BIN}

bin/%: ${GO_SRC}
	mkdir bin/ &> /dev/null || true
	go build -o $@

.PHONY: clean
clean:
	rm -rf bin/
