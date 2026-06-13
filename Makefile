# Purpose: Build and test tasks for sin-websearch.
# Docs: Makefile.doc.md

BINARY := sin-websearch

.PHONY: build test clean install lint

build:
	go build -o $(BINARY) ./cmd/sin-websearch

test:
	go test ./...

lint:
	gofmt -w .

clean:
	rm -f $(BINARY)

install:
	go install ./cmd/sin-websearch
