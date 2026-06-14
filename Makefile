# Purpose: Common development tasks for sin-websearch.
# Docs: Makefile.doc.md

.PHONY: build test cover vet lint sec audit clean

build:
	go build ./cmd/sin-websearch

test:
	go test ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

vet:
	go vet ./...

lint:
	golangci-lint run ./... --timeout=5m

sec:
	gosec ./...
	govulncheck ./...

audit:
	bash ~/.config/opencode/skills/ceo-audit/scripts/audit.sh .

sbom:
	syft scan dir:. -o spdx-json=sbom.spdx.json -o cyclonedx-json=bom.json

clean:
	rm -f coverage.out sin-websearch
