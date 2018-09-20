VERSION ?= $(shell git describe --tags --dirty --always)

BUILD_DIR ?= ./bin

LDFLAGS = -ldflags "-w -X github.com/glynternet/mon/cmd/moncli/cmd.version=$(VERSION)"
GOBUILD_FLAGS ?= -installsuffix cgo -a $(LDFLAGS)
GOBUILD_ENVVARS ?= CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH)
GOBUILD_CMD ?= $(GOBUILD_ENVVARS) go build $(GOBUILD_FLAGS)

SERVE_NAME = monserve
CLI_NAME = moncli

OS ?= linux
ARCH ?= amd64

all: build install clean

build: monserve moncli

install:
	cp -v $(BUILD_DIR)/* $(GOPATH)/bin/

clean:
	rm $(BUILD_DIR)/*

monserve: monserve-binary monserve-image

monserve-binary:
	$(MAKE) binary APP_NAME=monserve

monserve-image:
	docker build --tag $(SERVE_NAME):$(VERSION) .

moncli: moncli-binary

moncli-binary:
	$(MAKE) binary APP_NAME=moncli

binary:
	$(GOBUILD_CMD) -o $(BUILD_DIR)/$(APP_NAME) ./cmd/$(APP_NAME)