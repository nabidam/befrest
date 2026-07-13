.PHONY: web build test dev

web:
	npm --prefix web run build

build: web
	go build -ldflags "-s -w" -o dist/befrest ./cmd/befrest

test: web
	go test ./...

dev: web
	go run ./cmd/befrest
