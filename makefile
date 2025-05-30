test:
	@go test ./... -v

docker-build: test
	@docker build -t farkle:latest .

docker-run: docker-build
	@docker run --rm -d farkle:latest