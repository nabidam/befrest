.PHONY: web build test dev e2e

web:
	npm --prefix web run build

build: web
	go build -ldflags "-s -w" -o dist/befrest ./cmd/befrest

test: web
	go test ./...

dev: web
	go run ./cmd/befrest

e2e: build
	npm --prefix e2e install --no-package-lock --no-audit --no-fund
	npm --prefix e2e test
