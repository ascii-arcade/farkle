test:
	@go test ./... -v

build: test
	@go build -o bin/client ./cmd/client

run: build
	@./bin/client
