.PHONY: build run test lint vet fmt clean \
       docker-up docker-down docker-build docker-reset \
       migrate-up migrate-down migrate-create \
       sqlc-generate web-dev web-build \
       test-pgtap test-pgtap-setup

# ─── Variables ───────────────────────────────────────────────
APP_NAME    := crm-api
BIN_DIR     := bin
GO_FILES    := $(shell find . -name '*.go' -not -path './vendor/*' -not -path './web/*')
MIGRATE_DIR := migrations
DB_DSN      ?= postgres://crm:crm_secret@localhost:5432/crm?sslmode=disable
DB_TEST_DSN ?= postgres://crm:crm_secret@localhost:5432/crm_test?sslmode=disable

# ─── Go ──────────────────────────────────────────────────────
build:
	go build -ldflags="-s -w" -o $(BIN_DIR)/$(APP_NAME) ./cmd/api

run:
	go run ./cmd/api

test:
	go test ./... -race -cover -coverprofile=coverage.out
	go tool cover -func=coverage.out

lint:
	golangci-lint run ./...

vet:
	go vet ./...

fmt:
	gofmt -s -w $(GO_FILES)

clean:
	rm -rf $(BIN_DIR) coverage.out coverage.html

# ─── Docker ──────────────────────────────────────────────────
docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-build:
	docker compose build

docker-reset:
	docker compose down -v && docker compose up -d

# ─── pgTAP ──────────────────────────────────────────────────
test-pgtap-setup:
	migrate -path $(MIGRATE_DIR) -database "$(DB_TEST_DSN)" drop -f
	migrate -path $(MIGRATE_DIR) -database "$(DB_TEST_DSN)" up

test-pgtap: test-pgtap-setup
	docker compose exec postgres pg_prove -U crm -d crm_test --recurse /tests/pgtap/

# ─── Migrations ──────────────────────────────────────────────
migrate-up:
	migrate -path $(MIGRATE_DIR) -database "$(DB_DSN)" up

migrate-down:
	migrate -path $(MIGRATE_DIR) -database "$(DB_DSN)" down 1

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir $(MIGRATE_DIR) -seq $$name

# ─── sqlc ────────────────────────────────────────────────────
sqlc-generate:
	sqlc generate -f sqlc/sqlc.yaml

# ─── Frontend ────────────────────────────────────────────────
web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

web-lint:
	cd web && npm run lint

web-test:
	cd web && npm run test
