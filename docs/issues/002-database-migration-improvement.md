# Issue #002: 資料庫 Migration 機制

**狀態**: 🚧 進行中
**建立日期**: 2025-11-18
**最後更新**: 2025-11-22

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
自動執行 migrations
├─ 讀取 migrations/ 目錄的 SQL 檔案
├─ 檢查 schema_migrations 表（記錄已執行版本）
└─ 執行尚未跑過的 migration
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
│   ├── 001_initial_schema.up.sql      # 建立所有表
│   ├── 001_initial_schema.down.sql    # 回滾用
│   ├── 002_xxx.up.sql                 # 未來新增的變更
│   └── 002_xxx.down.sql
└── timescaledb/
    ├── 001_create_es_metrics.up.sql
    └── 001_create_es_metrics.down.sql
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

```go
// RunMigrations 在程式啟動時自動執行
func RunMigrations() error {
    // 1. 確保 schema_migrations 表存在
    // 2. 讀取已執行的版本
    // 3. 掃描 migrations/ 目錄
    // 4. 依序執行未跑過的 .up.sql
    // 5. 記錄已執行的版本
}
```

### 2. 版本追蹤表

```sql
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 3. main.go 整合

```go
func main() {
    utils.LoadEnvironment()
    clients.LoadDatabase()
    clients.LoadTimescaleDB()

    // 自動執行 migrations
    if err := services.RunMigrations(); err != nil {
        log.Fatalf("Migration failed: %v", err)
    }

    // 繼續初始化其他服務...
}
```

---

## 待辦事項

- [ ] 建立 `services/migration.go` - migration 執行邏輯
- [ ] 建立 `migrations/mysql/001_initial_schema.up.sql` - 完整建表 SQL
- [ ] 建立 `migrations/mysql/001_initial_schema.down.sql` - 回滾 SQL
- [ ] 建立 `migrations/timescaledb/001_create_es_metrics.up.sql`
- [ ] 修改 `main.go` - 啟動時呼叫 RunMigrations()
- [ ] 移除 `services/sqltable.go` 中的 GORM AutoMigrate
- [ ] 移除 `cmd/migrate/` 目錄（不需要獨立 CLI）
- [ ] 移除 `utils/migration_manager.go`（過度設計）
- [ ] 測試：空資料庫啟動
- [ ] 測試：已有資料庫啟動（應跳過已執行的 migration）

---

## 移除的東西

以下是之前過度設計的部分，應移除：

| 檔案/目錄 | 原因 |
|-----------|------|
| `cmd/migrate/main.go` | 不需要獨立 CLI |
| `utils/migration_manager.go` | 過度複雜 |
| `Makefile` 中的 migrate 指令 | 不需要 |
| GORM AutoMigrate | 改用 SQL migration |

---

## 優點

1. **一致性**：部署只要一個指令
2. **可追蹤**：每次 schema 變更都有記錄
3. **可回滾**：保留 down.sql 以備不時之需
4. **簡單**：沒有額外的工具或指令

---

## 相關

- Issue #001: ES 連線管理架構（依賴 es_connections 表）
