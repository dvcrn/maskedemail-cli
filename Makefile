
GO_SRC := $(wildcard ./*.go)
INSTALL_DIR := ${HOME}/.local/bin


all: install

install: bin/maskedemail-cli
	mv $< ${INSTALL_DIR}/.

bin/maskedemail-cli: ${GO_SRC}
	mkdir bin/ || true
	go build -o $@