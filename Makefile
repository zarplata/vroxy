.PHONY: all clean-all build clean-deps deps version

DATE := $(shell git log -1 --format="%cd" --date=short | sed s/-//g)
COUNT := $(shell git rev-list --count HEAD)
COMMIT := $(shell git rev-parse --short HEAD)

BINARY_NAME := vroxy
VERSION := "${DATE}.${COUNT}_${COMMIT}"
LDFLAGS := "-X main.version=${VERSION}"

default: all

all: clean-all deps build

version:
	@echo ${VERSION}

clean-all: clean-deps
	@echo Cleanup build artifacts
	rm -rf .out/
	@echo Done

build:
	@echo Build
	ln -s ${PWD}/vendor/ ${PWD}/vendor/src
	GOPATH="${PWD}/vendor" go build -v -o .out/${BINARY_NAME} -ldflags ${LDFLAGS} *.go
	@echo Done

clean-deps:
	@echo Clean dependencies
	rm -rf vendor/*

deps:
	@echo Fetch dependencies
	git submodule update --init
