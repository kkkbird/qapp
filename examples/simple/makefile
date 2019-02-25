APP_NAME = .
APP_VERSION = 0.0.1

LDFLAGS += -X 'github.com/kkkbird/bshark.Version=$(APP_VERSION)'
LDFLAGS += -X 'github.com/kkkbird/bshark.BuildTime=$(shell date -u '+%Y-%m-%d %I:%M:%S %Z')'
LDFLAGS += -X 'github.com/kkkbird/bshark.GitHash=$(shell git rev-parse HEAD)'
LDFLAGS += -X 'github.com/kkkbird/bshark.GoVersion=$(shell go version)'

OS := $(shell uname)

BUILD_FLAGS = -v

NOW = $(shell date -u '+%Y%m%d%I%M%S')

.PHONY: all build build-race test clean

all: test build-release

build:
	go build $(BUILD_FLAGS) $(APP_NAME)

build-race: $(GENERATED) govet
	go build $(BUILD_FLAGS) -race $(APP_NAME)

build-release:
	go build $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" $(APP_NAME)

clean:
	go clean -i ./...

test:
	go test -cover -race ./...