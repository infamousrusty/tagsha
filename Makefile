.PHONY: all build test lint clean docker-build docker-up docker-down frontend-build help

SHELL := /bin/bash
VERSION ?= dev

all: lint test

build:
	cd backend && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o tagsha-api ./cmd/server

test-backend:
	cd backend && go test -race -coverprofile=coverage.out ./...

test-frontend:
	cd frontend && npm ci && npm run test

test: test-backend test-frontend

lint-backend:
	cd backend && go vet ./...

lint-frontend:
	cd frontend && npm run lint

lint: lint-backend lint-frontend

frontend-build:
	cd frontend && npm ci && npm run build

frontend-dev:
	cd frontend && npm run dev

docker-build:
	docker build --build-arg VERSION=$(VERSION) -t tagsha-api:$(VERSION) backend/

docker-up: frontend-build
	docker compose -f infrastructure/docker-compose.yml up -d

docker-down:
	docker compose -f infrastructure/docker-compose.yml down

docker-dev:
	docker compose -f infrastructure/docker-compose.yml -f infrastructure/docker-compose.dev.yml up --build

clean:
	rm -f backend/tagsha-api backend/healthcheck backend/coverage.out
	rm -rf frontend/dist frontend/node_modules

secrets-init:
	@mkdir -p secrets
	@test -f secrets/github_token || echo 'REPLACE_WITH_GITHUB_TOKEN' > secrets/github_token
	@test -f secrets/grafana_admin_password || echo 'changeme123' > secrets/grafana_admin_password
	@echo 'Initialised secrets/. Update secrets/github_token before running.'

help:
	@grep -E '^[a-z_-]+:' $(MAKEFILE_LIST) | sed 's/:/\t/' | column -t
