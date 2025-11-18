# Issue #002: 資料庫 Migration 機制完善

**狀態**: 🚧 進行中 (Phase 1 完成)
**優先級**: 🟡 中高
**建立日期**: 2025-11-18
**負責人**: 待指派
**預計時程**: 3-4 天
**目前進度**: Phase 1/5 完成 (20%)

---

## 📌 議題描述

目前專案使用 **GORM AutoMigrate** 管理資料庫 schema，但存在以下問題：

1. ❌ **缺少版本控制**：無法追蹤哪些 migration 已執行
2. ❌ **無法回滾**：出錯時無法回退到先前版本
3. ❌ **遺漏表結構**：部分 entity 未加入 AutoMigrate
4. ❌ **外鍵約束遺失**：GORM AutoMigrate 不會自動建立所有外鍵
5. ❌ **部署困難**：生產環境部署時無法確保 schema 一致性
6. ❌ **數據遷移不支援**：無法處理需要數據轉換的 schema 變更

---

## 🔍 現狀分析

### 當前實作

**位置**: `services/sqltable.go:9-32`

```go
func CreateTable() {
    global.Mysql.Exec("USE logdetect")
    err := global.Mysql.AutoMigrate(
        &entities.User{},
        &entities.Role{},
        &entities.Permission{},
        &entities.Device{},
        &entities.Receiver{},
        &entities.Index{},
        &entities.Target{},
        &entities.History{},
        &entities.HistoryArchive{},
        &entities.HistoryDailyStats{},
        &entities.MailHistory{},
        &entities.AlertHistory{},
        &entities.CronList{},
        &entities.ElasticsearchMonitor{},
    )
}
```

### 缺失的 Entity

經過完整檢查，以下 entity 已定義但**未加入 AutoMigrate**：

#### 1. ESConnection（關鍵缺失！）
- **定義位置**: `entities/es_connection.go:9-66`
- **影響**: Issue #001 的核心功能無法正常運作
- **表名**: `es_connections`
- **用途**: ES 連線配置管理

#### 2. IndicesTargets
- **定義位置**: `entities/targets.go:23-27`
- **表名**: `indices_targets`
- **用途**: Index 與 Target 的 many-to-many 關聯表
- **狀態**: GORM 可能自動處理，需驗證

#### 3. Module
- **定義位置**: `entities/menu.go:21-28`
- **表名**: `modules`
- **用途**: 系統模組管理

### 外鍵約束問題

**已在 Phase 1 定義但未執行的 SQL**：

```sql
-- migrations/002_alter_indices_add_es_connection.up.sql
ALTER TABLE `indices`
ADD CONSTRAINT `fk_indices_es_connection`
  FOREIGN KEY (`es_connection_id`)
  REFERENCES `es_connections`(`id`)
  ON DELETE RESTRICT;

-- migrations/003_alter_elasticsearch_monitors_add_es_connection.up.sql
ALTER TABLE `elasticsearch_monitors`
ADD CONSTRAINT `fk_es_monitors_connection`
  FOREIGN KEY (`es_connection_id`)
  REFERENCES `es_connections`(`id`)
  ON DELETE SET NULL;
```

**問題**：
- GORM AutoMigrate 會添加欄位，但**不保證**建立外鍵約束
- `migrations/` 目錄中的 SQL 檔案**沒有執行機制**

### TimescaleDB 表

以下表存儲在 **TimescaleDB**（非 MySQL），需要獨立的 migration：

1. **es_metrics** - ES 監控指標時序數據
2. **es_alerts** - ES 告警歷史記錄

**狀態**: 目前由 `clients/timescaledb.go` 手動建立，缺乏版本控制

---

## ✅ 解決方案

### 方案選擇：golang-migrate/migrate

採用業界標準的 **golang-migrate/migrate** 工具，理由：

1. ✅ 支援版本控制（up/down migration）
2. ✅ 支援多種資料庫（MySQL, PostgreSQL, TimescaleDB）
3. ✅ 可整合到 CI/CD pipeline
4. ✅ 豐富的社群支援
5. ✅ 支援純 SQL 與 Go 程式碼兩種方式

**專案**: https://github.com/golang-migrate/migrate

### 架構設計

```
log-detect-backend/
├── migrations/
│   ├── mysql/                          # MySQL migrations
│   │   ├── 000001_initial_schema.up.sql
│   │   ├── 000001_initial_schema.down.sql
│   │   ├── 000002_add_es_connections.up.sql
│   │   ├── 000002_add_es_connections.down.sql
│   │   ├── 000003_add_foreign_keys.up.sql
│   │   ├── 000003_add_foreign_keys.down.sql
│   │   └── ...
│   └── timescaledb/                    # TimescaleDB migrations
│       ├── 000001_create_es_metrics.up.sql
│       ├── 000001_create_es_metrics.down.sql
│       └── ...
├── utils/
│   └── migrate.go                      # Migration 執行工具
└── cmd/
    └── migrate/
        └── main.go                     # CLI migration tool
```

### 核心功能

#### 1. Migration 執行器

```go
// utils/migrate.go

package utils

import (
    "database/sql"
    "fmt"
    "log"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/mysql"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

type MigrationManager struct {
    mysqlMigrate      *migrate.Migrate
    timescaleDBMigrate *migrate.Migrate
}

func NewMigrationManager(mysqlDB *sql.DB, timescaleDB *sql.DB) (*MigrationManager, error) {
    // MySQL migrations
    mysqlDriver, err := mysql.WithInstance(mysqlDB, &mysql.Config{})
    if err != nil {
        return nil, err
    }

    mysqlMigrate, err := migrate.NewWithDatabaseInstance(
        "file://migrations/mysql",
        "mysql",
        mysqlDriver,
    )
    if err != nil {
        return nil, err
    }

    // TimescaleDB migrations (使用 postgres driver)
    // ... similar setup

    return &MigrationManager{
        mysqlMigrate:      mysqlMigrate,
        timescaleDBMigrate: timescaleDBMigrate,
    }, nil
}

func (m *MigrationManager) MigrateUp() error {
    // Migrate MySQL
    if err := m.mysqlMigrate.Up(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("MySQL migration failed: %w", err)
    }

    // Migrate TimescaleDB
    if err := m.timescaleDBMigrate.Up(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("TimescaleDB migration failed: %w", err)
    }

    return nil
}

func (m *MigrationManager) MigrateDown() error { /* ... */ }
func (m *MigrationManager) MigrateTo(version uint) error { /* ... */ }
func (m *MigrationManager) GetVersion() (uint, bool, error) { /* ... */ }
```

#### 2. CLI 工具

```go
// cmd/migrate/main.go

package main

import (
    "flag"
    "log"
    "log-detect/clients"
    "log-detect/utils"
)

func main() {
    action := flag.String("action", "up", "Migration action: up, down, version")
    steps := flag.Int("steps", 0, "Number of migrations to apply (0 = all)")
    flag.Parse()

    // Initialize databases
    clients.LoadDatabase()
    clients.LoadTimescaleDB()

    // Create migration manager
    manager, err := utils.NewMigrationManager(global.Mysql.DB(), global.TimescaleDB)
    if err != nil {
        log.Fatal(err)
    }

    // Execute migration
    switch *action {
    case "up":
        err = manager.MigrateUp()
    case "down":
        err = manager.MigrateDown()
    case "version":
        version, dirty, err := manager.GetVersion()
        log.Printf("Version: %d, Dirty: %v", version, dirty)
        return
    }

    if err != nil {
        log.Fatal(err)
    }
}
```

---

## 📋 實作計畫

### Phase 1：基礎建設（1天）✅ 已完成
**完成日期**: 2025-11-18
**Commit**: 待提交

#### 1.1 安裝 Migration 工具
- [x] 添加 `github.com/golang-migrate/migrate/v4` 到 go.mod
- [x] 添加 MySQL 和 PostgreSQL driver
- [x] 建立 `migrations/mysql/` 和 `migrations/timescaledb/` 目錄
- [x] 建立 Migration 執行器 `utils/migration_manager.go`
- [x] 建立 CLI 工具 `cmd/migrate/main.go`
- [x] 建立 Makefile 簡化操作

#### 1.2 移動現有 Migration 檔案
- [x] 移動 Phase 1 建立的 SQL 檔案到 `migrations/mysql/`
  - `001_create_es_connections.up/down.sql`
  - `002_alter_indices_add_es_connection.up/down.sql`
  - `003_alter_elasticsearch_monitors_add_es_connection.up/down.sql`

#### 1.3 建立工具與文件
- [x] Makefile 指令
  - `make migrate-up` - 執行 migrations
  - `make migrate-down` - 回滾 migration
  - `make migrate-version` - 查看版本
  - `make migrate-create` - 建立新 migration
  - `make migrate-goto` - 遷移到指定版本
  - `make migrate-force` - 強制設定版本

#### 成果
- 新增 3 個檔案（utils/migration_manager.go, cmd/migrate/main.go, Makefile）
- 移動 6 個 migration 檔案到正確位置
- 完整的 CLI 工具支援 up/down/version/goto/force 操作
- Makefile 簡化日常操作

### Phase 2：補齊缺失內容（0.5天）✅ 已完成
**完成日期**: 2025-11-18
**Commit**: 待提交

#### 2.1 修復 AutoMigrate 遺漏
- [x] 修改 `services/sqltable.go`
  - [x] 添加 `&entities.ESConnection{}`
  - [x] 添加 `&entities.Module{}`
  - [x] 重新組織程式碼，添加分類註解
- [x] 驗證編譯無錯誤

#### 成果
- 修復了 Issue #001 無法運作的關鍵問題
- ESConnection 表現在會自動建立
- Module 表也加入 AutoMigrate

### Phase 3：整合與測試（待進行）

#### 3.1 整合到啟動流程
- [ ] 修改 `main.go` 添加自動 migration 選項
- [ ] 添加環境變數配置 `database.auto_migrate`

#### 2.2 建立外鍵約束 Migration
- [ ] **000002_add_foreign_keys.up.sql**
  ```sql
  -- indices -> es_connections
  ALTER TABLE `indices`
  ADD CONSTRAINT `fk_indices_es_connection`
    FOREIGN KEY (`es_connection_id`)
    REFERENCES `es_connections`(`id`)
    ON DELETE RESTRICT;

  -- elasticsearch_monitors -> es_connections
  ALTER TABLE `elasticsearch_monitors`
  ADD CONSTRAINT `fk_es_monitors_connection`
    FOREIGN KEY (`es_connection_id`)
    REFERENCES `es_connections`(`id`)
    ON DELETE SET NULL;

  -- 其他外鍵...
  ```
- [ ] **000002_add_foreign_keys.down.sql**

### Phase 3：整合到專案啟動流程（0.5天）

#### 3.1 修改 main.go
- [ ] 在 `clients.LoadDatabase()` 之後執行 migration
  ```go
  // Auto-run migrations on startup
  if global.EnvConfig.Database.AutoMigrate {
      migrationManager, err := utils.NewMigrationManager(...)
      if err := migrationManager.MigrateUp(); err != nil {
          log.Printf("WARNING: Migration failed: %v", err)
      }
  }
  ```

#### 3.2 環境變數配置
- [ ] 添加 `database.auto_migrate` 到 setting.yml
- [ ] 添加 `database.migration_path` 配置

### Phase 4：文件與工具（1天）

#### 4.1 文件更新
- [ ] 建立 `docs/guides/database-migration.md`
  - Migration 使用指南
  - 如何建立新的 migration
  - 回滾步驟
- [ ] 更新 `docs/specs_cn/04-資料庫設計.md`
  - 更新完整的表結構
  - 添加 ER 圖（包含所有外鍵）

#### 4.2 開發工具
- [ ] 建立 Makefile 指令
  ```makefile
  migrate-up:
      go run cmd/migrate/main.go -action=up

  migrate-down:
      go run cmd/migrate/main.go -action=down

  migrate-version:
      go run cmd/migrate/main.go -action=version

  migrate-create:
      migrate create -ext sql -dir migrations/mysql -seq $(name)
  ```

#### 4.3 CI/CD 整合
- [ ] 添加 migration 檢查到 CI pipeline
- [ ] 建立 migration 測試

### Phase 5：生產環境遷移策略（1天）

#### 5.1 遷移計畫
- [ ] 建立現有生產環境的 schema dump
- [ ] 比對差異並建立補充 migration
- [ ] 撰寫遷移腳本

#### 5.2 驗證與測試
- [ ] 在測試環境完整測試
- [ ] 準備回滾方案
- [ ] 建立 migration 執行 checklist

---

## ⚠️ 風險與挑戰

### 技術風險

1. **現有資料遷移**
   - 風險：生產環境可能已有資料，無法重新建表
   - 緩解：建立增量 migration，只添加缺失的表和欄位

2. **外鍵約束衝突**
   - 風險：現有資料可能違反外鍵約束
   - 緩解：migration 前先檢查並清理無效資料

3. **向下不兼容**
   - 風險：舊版本程式碼無法在新 schema 上運行
   - 緩解：保持向後兼容，使用多階段部署

### 業務風險

1. **部署中斷**
   - 風險：migration 失敗導致服務中斷
   - 緩解：在維護時段執行，準備快速回滾方案

2. **資料遺失**
   - 風險：錯誤的 migration 可能導致資料遺失
   - 緩解：執行前完整備份，測試環境驗證

---

## 📊 成功指標

### 功能指標
- [ ] 所有 entity 都正確建表（包含 ESConnection, Module）
- [ ] 所有外鍵約束正確建立
- [ ] Migration 可在乾淨環境中從頭執行
- [ ] Migration 可正確回滾
- [ ] TimescaleDB 表正確建立

### 品質指標
- [ ] Migration 測試覆蓋率 > 90%
- [ ] 文件完整度 100%
- [ ] 生產環境驗證通過

### 維護指標
- [ ] 新增 migration 的標準流程文件化
- [ ] CI/CD 自動檢查 migration 完整性

---

## 🔗 相關議題

- **Issue #001**: ES 連線管理架構 - 需要 ESConnection 表正確建立
- **未來議題**: 可能需要資料庫分片、讀寫分離等進階功能

---

## 📝 技術決策記錄

### 2025-11-18 - Migration 工具選擇

**參與者**: 開發團隊

**討論選項**:
1. ✅ **golang-migrate/migrate** (選用)
   - 優點：業界標準、功能完整、社群活躍
   - 缺點：需要額外學習
2. ❌ goose
   - 優點：簡單易用
   - 缺點：功能較少
3. ❌ sql-migrate
   - 優點：與 gorp 整合
   - 缺點：不如 golang-migrate 流行

**決議**: 採用 golang-migrate/migrate

### GORM AutoMigrate vs SQL Migration

**保留 GORM AutoMigrate 用於**:
- 開發環境快速迭代
- 單元測試環境

**使用 SQL Migration 用於**:
- 生產環境部署
- 測試環境
- CI/CD pipeline

---

## 📌 備註

- 本議題是架構基礎建設，建議優先處理
- Phase 1-3 應在 Issue #001 完成後立即進行
- 建議建立專門的 database 分支進行開發
- 完成後需要更新部署文件

---

**最後更新**: 2025-11-18
**更新者**: Claude (AI Assistant)
