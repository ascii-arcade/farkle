.PHONY: build run
build:
	@echo "Building Docker image with GIT_TAG=$(shell git rev-parse --short HEAD)"
	@docker compose build --build-arg GIT_TAG=$(shell git rev-parse --short HEAD)
run: build
	@docker compose up -d
