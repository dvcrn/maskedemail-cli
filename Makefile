
GO_SRC := $(wildcard ./*.go ./pkg/*.go)
INSTALL_DIR := ${HOME}/.local/bin
TARGET_BIN := maskedemail-cli

all: build

install: bin/${TARGET_BIN}
	cp -f $< ${INSTALL_DIR}/${TARGET_BIN}

build: bin/${TARGET_BIN}

bin/%: ${GO_SRC}
	mkdir bin/ &> /dev/null || true
	go build -o $@

.PHONY: clean
clean:
	rm -rf bin/
