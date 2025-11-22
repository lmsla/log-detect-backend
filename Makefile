# Log Detect - Makefile

.PHONY: help run build test clean migrate-create

help: ## 顯示幫助訊息
	@echo "Log Detect - Commands"
	@echo "====================="
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

run: ## 啟動應用程式（自動執行 migrations）
	go run main.go

build: ## 編譯應用程式
	go build -o bin/log-detect main.go

test: ## 執行測試
	go test ./...

clean: ## 清理編譯產物
	rm -rf bin/

migrate-create: ## 建立新的 migration 檔案 (使用: make migrate-create NAME=add_xxx DB=mysql)
	@if [ -z "$(NAME)" ]; then \
		echo "❌ 請指定 NAME。範例: make migrate-create NAME=add_users_table DB=mysql"; \
		exit 1; \
	fi
	@if [ -z "$(DB)" ]; then \
		echo "❌ 請指定 DB (mysql 或 timescaledb)。範例: make migrate-create NAME=add_users_table DB=mysql"; \
		exit 1; \
	fi
	@NEXT_VERSION=$$(ls migrations/$(DB)/*.up.sql 2>/dev/null | wc -l); \
	NEXT_VERSION=$$(printf "%03d" $$((NEXT_VERSION + 1))); \
	UP_FILE="migrations/$(DB)/$${NEXT_VERSION}_$(NAME).up.sql"; \
	DOWN_FILE="migrations/$(DB)/$${NEXT_VERSION}_$(NAME).down.sql"; \
	echo "-- Migration: $(NAME)" > $$UP_FILE; \
	echo "-- Version: $${NEXT_VERSION}" >> $$UP_FILE; \
	echo "-- Created: $$(date '+%Y-%m-%d')" >> $$UP_FILE; \
	echo "" >> $$UP_FILE; \
	echo "-- Add your SQL here" >> $$UP_FILE; \
	echo "" >> $$UP_FILE; \
	echo "-- Rollback: $(NAME)" > $$DOWN_FILE; \
	echo "-- Version: $${NEXT_VERSION}" >> $$DOWN_FILE; \
	echo "" >> $$DOWN_FILE; \
	echo "-- Add rollback SQL here" >> $$DOWN_FILE; \
	echo ""; \
	echo "✅ 已建立:"; \
	echo "   📄 $$UP_FILE"; \
	echo "   📄 $$DOWN_FILE"

.DEFAULT_GOAL := help
