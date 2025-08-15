# Makefile for the ILC project

# Variables
GO := go

# Targets
.PHONY: all build test clean

all: build test

build:
	@echo "Building the application..."
	$(GO) build -o ilc

test:
	@echo "Running tests..."
	$(GO) test ./...

clean:
	@echo "Cleaning up..."
	rm -f ilc
