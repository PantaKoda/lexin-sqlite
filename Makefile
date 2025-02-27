# Variables
APP_NAME := lexin-sqlite
VERSION := 0.1.0
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

.PHONY: all build clean run test

all: build

# Build the application
build:
	mkdir -p bin
	go build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/lexin

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f lexin.db

# Run the application
run: build
	./bin/$(APP_NAME) $(ARGS)

# Run tests
test:
	go test -v ./...

# Example usage
example: build
	@echo "Example usage:"
	@echo "./bin/$(APP_NAME) -file swedish-english.xml -target english"
