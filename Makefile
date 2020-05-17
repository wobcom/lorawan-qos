.PHONY: build serve
VERSION := $(shell git describe --always |sed -e "s/^v//")

build:
	mkdir -p build
	go build $(GO_EXTRA_BUILD_ARGS) -ldflags "-s -w -X main.version=$(VERSION)" -o build/network-qos-service cmd/main.go

# shortcuts for development

start:
	@echo "Starting network-qos-service"
	./build/network-qos-service


