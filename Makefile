.PHONY: run dev test build

run:
	go run .

dev:
	@which air > /dev/null 2>&1 || (echo "Installing air for live reload..." && go install github.com/air-verse/air@latest)
	air

test:
	go test -v ./...

build:
	go build -o bin/server .
