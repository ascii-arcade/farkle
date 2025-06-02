.PHONY: build run

GIT_TAG=$(shell git rev-parse --short HEAD)

build:
	@echo "Building Docker image with GIT_TAG=$(GIT_TAG)"
	@docker compose build --build-arg GIT_TAG=$(GIT_TAG)
run: build
	@docker compose up -d
