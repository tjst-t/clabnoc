.PHONY: all build build-frontend embed-frontend build-backend test test-go test-frontend lint dev dev-stop docker-build docker-run clean

BINARY := bin/clabnoc
DOCKER_IMAGE := ghcr.io/tjst-t/clabnoc
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
DEV_ADDR := :8888
DEV_PID_FILE := /tmp/clabnoc-dev.pid

all: build

## Build

build: build-frontend embed-frontend build-backend

build-frontend:
	cd frontend && npm ci && npm run build

embed-frontend: build-frontend
	rm -rf internal/frontend/dist/assets
	cp -r frontend/dist/* internal/frontend/dist/

build-backend:
	CGO_ENABLED=0 go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY) ./cmd/clabnoc

## Test

test: test-go test-frontend

test-go:
	go test ./... -v -count=1 -race

test-frontend:
	cd frontend && npm run test

## Lint

lint:
	go vet ./...
	cd frontend && npm run lint

## Dev server

dev: build
	@if [ -f $(DEV_PID_FILE) ] && kill -0 $$(cat $(DEV_PID_FILE)) 2>/dev/null; then \
		echo "Stopping old dev server (PID $$(cat $(DEV_PID_FILE)))..."; \
		kill $$(cat $(DEV_PID_FILE)) 2>/dev/null; \
		sleep 1; \
	fi
	@nohup ./$(BINARY) -dev -addr $(DEV_ADDR) > /tmp/clabnoc.log 2>&1 & echo $$! > $(DEV_PID_FILE)
	@sleep 1
	@echo "clabnoc dev server started (PID $$(cat $(DEV_PID_FILE)), addr $(DEV_ADDR))"
	@echo "Log: tail -f /tmp/clabnoc.log"

dev-stop:
	@if [ -f $(DEV_PID_FILE) ] && kill -0 $$(cat $(DEV_PID_FILE)) 2>/dev/null; then \
		echo "Stopping dev server (PID $$(cat $(DEV_PID_FILE)))..."; \
		kill $$(cat $(DEV_PID_FILE)); \
		rm -f $(DEV_PID_FILE); \
		echo "Stopped."; \
	else \
		echo "No dev server running."; \
		rm -f $(DEV_PID_FILE); \
	fi

## Docker

docker-build:
	docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .

docker-run:
	docker run --rm --network host \
		-v /var/run/docker.sock:/var/run/docker.sock:ro \
		-v /tmp/containerlab:/tmp/containerlab:ro \
		$(DOCKER_IMAGE):latest

## Clean

clean:
	rm -rf bin/ frontend/dist/
	rm -f $(DEV_PID_FILE)
