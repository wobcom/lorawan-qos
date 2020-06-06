.PHONY: build serve
VERSION := $(shell git describe --always |sed -e "s/^v//")

build: internal/migrations
	mkdir -p build
	go generate internal/migrations/migrations.go
	go build $(GO_EXTRA_BUILD_ARGS) -ldflags "-s -w -X main.version=$(VERSION)" -o build/network-qos-service cmd/main.go

# shortcuts for development

start:
	@echo "Starting network-qos-service"
	 ./build/network-qos-service


internal/migrations:
	@echo "Generating migrations files"
	go generate internal/migrations/migrations.go