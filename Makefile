.PHONY: all build build-frontend build-backend test test-go test-frontend lint dev docker-build clean

BINARY := bin/clabnoc
DOCKER_IMAGE := ghcr.io/tjst-t/clabnoc
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

all: build

## Build

build: build-frontend build-backend

build-frontend:
	cd frontend && npm ci && npm run build
	cp -r frontend/dist/. internal/frontend/dist/

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

## Dev

dev:
	@echo "Starting development mode..."
	@echo "Frontend: cd frontend && npm run dev"
	@echo "Backend:  go run ./cmd/clabnoc --dev"

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
	find internal/frontend/dist/ -not -name '.gitkeep' -delete 2>/dev/null || true
