.PHONY: build build-ee run run-ee test test-ee test-integration lint vet fmt clean \
       docker-up docker-down docker-build docker-reset \
       migrate-up migrate-down migrate-create \
       sqlc-generate generate-api web-generate-types \
       web-dev web-build \
       test-pgtap test-pgtap-setup \
       test-pgtap-ee test-pgtap-ee-setup test-pgtap-all

-include .env
export

# ─── Variables ───────────────────────────────────────────────
APP_NAME    := crm-api
BIN_DIR     := bin
GO_FILES    := $(shell find . -name '*.go' -not -path './vendor/*' -not -path './web/*')
MIGRATE_DIR := migrations
DB_DSN      ?= postgres://crm:crm_secret@localhost:5433/crm?sslmode=disable
DB_TEST_DSN ?= postgres://crm:crm_secret@localhost:5433/crm_test?sslmode=disable

# ─── Go ──────────────────────────────────────────────────────
build:
	go build -ldflags="-s -w" -o $(BIN_DIR)/$(APP_NAME) ./cmd/api

build-ee:
	go build -tags enterprise -ldflags="-s -w" -o $(BIN_DIR)/$(APP_NAME) ./cmd/api

run:
	-pkill -f 'go-build.*/api$$' 2>/dev/null || true
	go run ./cmd/api

run-ee:
	-pkill -f 'go-build.*/api$$' 2>/dev/null || true
	go run -tags enterprise ./cmd/api

test:
	go test ./... -race -cover -coverprofile=coverage.out
	go tool cover -func=coverage.out

test-ee:
	go test -tags enterprise ./... -race -cover -coverprofile=coverage.out
	go tool cover -func=coverage.out

test-integration:
	DB_TEST_DSN="$(DB_TEST_DSN)" go test -tags integration ./... -race -v -count=1

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
	psql "$(DB_TEST_DSN)" -c "DROP SCHEMA IF EXISTS ee CASCADE; DROP SCHEMA IF EXISTS security CASCADE; DROP SCHEMA IF EXISTS iam CASCADE; DROP SCHEMA IF EXISTS metadata CASCADE;"
	migrate -path $(MIGRATE_DIR) -database "$(DB_TEST_DSN)" drop -f
	migrate -path $(MIGRATE_DIR) -database "$(DB_TEST_DSN)" up

test-pgtap: test-pgtap-setup
	docker compose exec postgres pg_prove -U crm -d crm_test --ext .sql --recurse /tests/pgtap/

# ─── pgTAP (Enterprise) ─────────────────────────────────────
EE_MIGRATE_DSN = $(DB_TEST_DSN)&x-migrations-table=ee_schema_migrations

test-pgtap-ee-setup: test-pgtap-setup
	migrate -path ee/migrations -database "$(EE_MIGRATE_DSN)" up

test-pgtap-ee: test-pgtap-ee-setup
	docker compose exec postgres pg_prove -U crm -d crm_test --ext .sql --recurse /ee/tests/pgtap/

test-pgtap-all: test-pgtap-ee-setup
	docker compose exec postgres pg_prove -U crm -d crm_test --ext .sql --recurse /tests/pgtap/ /ee/tests/pgtap/

# ─── Migrations ──────────────────────────────────────────────
migrate-up:
	migrate -path $(MIGRATE_DIR) -database "$(DB_DSN)" up

migrate-up-ee: migrate-up
	migrate -path ee/migrations -database "$(DB_DSN)&x-migrations-table=ee_schema_migrations" up

migrate-down:
	migrate -path $(MIGRATE_DIR) -database "$(DB_DSN)" down 1

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir $(MIGRATE_DIR) -seq $$name

# ─── sqlc ────────────────────────────────────────────────────
sqlc-generate:
	sqlc generate -f sqlc/sqlc.yaml

# ─── OpenAPI code generation ─────────────────────────────────
generate-api:
	oapi-codegen -generate gin,types,spec -package api -o internal/api/openapi_gen.go api/openapi.yaml
	cd web && npx openapi-typescript ../api/openapi.yaml -o src/types/openapi.d.ts

web-generate-types:
	cd web && npx openapi-typescript ../api/openapi.yaml -o src/types/openapi.d.ts

# ─── Frontend ────────────────────────────────────────────────
web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

web-lint:
	cd web && npm run lint

web-test:
	cd web && npm run test
