.PHONY: build run

GIT_TAG=$(shell git rev-parse --short HEAD)

build:
	@echo "Building Docker image with GIT_TAG=$(GIT_TAG)"
	@GIT_TAG=$(GIT_TAG) docker compose build
run: build
	@GIT_TAG=$(GIT_TAG) docker compose up -d
