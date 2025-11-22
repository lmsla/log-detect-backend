# Issue #002: 資料庫 Migration 機制

**狀態**: ✅ 已完成
**建立日期**: 2025-11-18
**完成日期**: 2025-11-22

---

## 目標

程式啟動時自動執行資料庫 migration，確保 schema 一致。

---

## 方案

### 運作流程

```
程式啟動
    ↓
連接 MySQL / TimescaleDB
    ↓
services.RunMigrations()
├─ 建立 schema_migrations 表（如不存在）
├─ 讀取已執行的版本
├─ 掃描 migrations/ 目錄
└─ 執行尚未跑過的 .up.sql
    ↓
初始化其他服務（ES、Auth 等）
    ↓
啟動 HTTP Server
```

### 部署方式

```bash
./log-detect   # 一個指令，migration 自動完成
```

---

## 目錄結構

```
migrations/
├── mysql/
│   ├── 001_initial_schema.up.sql      # 建立所有 MySQL 表
│   └── 001_initial_schema.down.sql    # 回滾用
└── timescaledb/
    ├── 001_initial_schema.up.sql      # 建立 TimescaleDB 表
    └── 001_initial_schema.down.sql    # 回滾用
```

### Migration 檔案命名規則

- 格式：`{版本號}_{描述}.{up|down}.sql`
- 版本號：三位數字，遞增（001, 002, 003...）
- up.sql：執行變更
- down.sql：回滾變更

---

## 實作內容

### 1. Migration 執行器

**位置**: `services/migration.go`

主要函數：
- `RunMigrations()` - 程式啟動時呼叫
- `runMySQLMigrations()` - 執行 MySQL migrations
- `runTimescaleDBMigrations()` - 執行 TimescaleDB migrations
- `executeMigrations()` - 核心執行邏輯

### 2. 版本追蹤表

```sql
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 3. main.go 整合

```go
// 執行資料庫 migrations
if err := services.RunMigrations(); err != nil {
    log.Fatalf("Database migration failed: %v", err)
}
```

---

## 已完成事項

- [x] 建立 `services/migration.go` - migration 執行邏輯
- [x] 建立 `migrations/mysql/001_initial_schema.up.sql` - 完整建表 SQL
- [x] 建立 `migrations/mysql/001_initial_schema.down.sql` - 回滾 SQL
- [x] 建立 `migrations/timescaledb/001_initial_schema.up.sql`
- [x] 建立 `migrations/timescaledb/001_initial_schema.down.sql`
- [x] 修改 `main.go` - 啟動時呼叫 RunMigrations()
- [x] 移除 `services/sqltable.go`（GORM AutoMigrate）
- [x] 移除 `cmd/migrate/` 目錄（不需要獨立 CLI）
- [x] 移除 `utils/migration_manager.go`（過度設計）
- [x] 更新 `migrations/README.md`
- [x] 簡化 `Makefile`

---

## 已移除的檔案

| 檔案/目錄 | 原因 |
|-----------|------|
| `cmd/migrate/main.go` | 不需要獨立 CLI |
| `utils/migration_manager.go` | 過度複雜 |
| `services/sqltable.go` | 改用 SQL migration |

---

## 優點

1. **一致性**：部署只要一個指令
2. **可追蹤**：每次 schema 變更都有記錄
3. **可回滾**：保留 down.sql 以備不時之需
4. **簡單**：沒有額外的工具或指令

---

## Commits

- `43f2bc3` - refactor: 簡化 Migration 機制 - 啟動時自動執行

---

## 相關

- Issue #001: ES 連線管理架構（依賴 es_connections 表）
