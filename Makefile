.PHONY: web build test dev e2e release

RELEASE_DIR := dist/release
RELEASE_MAX_BYTES := 31457280

web:
	npm --prefix web run build

build: web
	go build -ldflags "-s -w" -o dist/befrest ./cmd/befrest

test: web
	go test ./...

dev: web
	go run ./cmd/befrest

e2e: build
	npm --prefix e2e ci
	npm --prefix e2e test

release: web
	@set -e; \
	os=$$(go env GOOS); \
	arch=$$(go env GOARCH); \
	extension=; \
	if [ "$$os" = windows ]; then extension=.exe; fi; \
	mkdir -p $(RELEASE_DIR); \
	output=$(RELEASE_DIR)/befrest-$$os-$$arch$$extension; \
	echo "building $$output"; \
	go build -ldflags "-s -w" -o $$output ./cmd/befrest; \
	size=$$(wc -c < $$output); \
	if [ $$size -ge $(RELEASE_MAX_BYTES) ]; then \
		echo "$$output is $$size bytes; release binaries must be smaller than $(RELEASE_MAX_BYTES) bytes" >&2; \
		exit 1; \
	fi; \
	echo "released $$output ($$size bytes)"
