# Log Detect - Makefile for Database Migrations

.PHONY: help migrate-up migrate-down migrate-version migrate-goto migrate-force migrate-create

help: ## 顯示幫助訊息
	@echo "Log Detect - Database Migration Commands"
	@echo "========================================"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

migrate-up: ## 執行所有 pending migrations
	@echo "🚀 Running migrations..."
	go run cmd/migrate/main.go -action=up

migrate-down: ## 回滾最後一個 migration
	@echo "⏪ Rolling back migration..."
	go run cmd/migrate/main.go -action=down

migrate-version: ## 顯示當前 migration 版本
	@echo "📊 Checking migration version..."
	go run cmd/migrate/main.go -action=version

migrate-goto: ## 遷移到指定版本 (使用: make migrate-goto VERSION=3)
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ Please specify VERSION. Example: make migrate-goto VERSION=3"; \
		exit 1; \
	fi
	@echo "🎯 Migrating to version $(VERSION)..."
	go run cmd/migrate/main.go -action=goto -version=$(VERSION)

migrate-force: ## 強制設定版本 (使用: make migrate-force VERSION=3)
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ Please specify VERSION. Example: make migrate-force VERSION=3"; \
		exit 1; \
	fi
	@echo "⚠️  WARNING: Forcing to version $(VERSION)..."
	go run cmd/migrate/main.go -action=force -version=$(VERSION)

migrate-create: ## 建立新的 migration 檔案 (使用: make migrate-create NAME=add_users_table DB=mysql)
	@if [ -z "$(NAME)" ]; then \
		echo "❌ Please specify NAME. Example: make migrate-create NAME=add_users_table DB=mysql"; \
		exit 1; \
	fi
	@if [ -z "$(DB)" ]; then \
		echo "❌ Please specify DB (mysql or timescaledb). Example: make migrate-create NAME=add_users_table DB=mysql"; \
		exit 1; \
	fi
	@NEXT_VERSION=$$(ls migrations/$(DB)/*.up.sql 2>/dev/null | wc -l); \
	NEXT_VERSION=$$(printf "%06d" $$((NEXT_VERSION + 1))); \
	UP_FILE="migrations/$(DB)/$${NEXT_VERSION}_$(NAME).up.sql"; \
	DOWN_FILE="migrations/$(DB)/$${NEXT_VERSION}_$(NAME).down.sql"; \
	echo "-- Migration: $(NAME)" > $$UP_FILE; \
	echo "-- Created: $$(date)" >> $$UP_FILE; \
	echo "" >> $$UP_FILE; \
	echo "-- Add your UP migration SQL here" >> $$UP_FILE; \
	echo "" >> $$UP_FILE; \
	echo "-- Migration: $(NAME)" > $$DOWN_FILE; \
	echo "-- Created: $$(date)" >> $$DOWN_FILE; \
	echo "" >> $$DOWN_FILE; \
	echo "-- Add your DOWN migration SQL here" >> $$DOWN_FILE; \
	echo ""; \
	echo "✅ Created migration files:"; \
	echo "   📄 $$UP_FILE"; \
	echo "   📄 $$DOWN_FILE"

# Development commands
run: ## 啟動應用程式
	go run main.go

build: ## 編譯應用程式
	go build -o bin/log-detect main.go

test: ## 執行測試
	go test ./...

clean: ## 清理編譯產物
	rm -rf bin/

.DEFAULT_GOAL := help
