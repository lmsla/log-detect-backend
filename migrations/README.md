# 資料庫 Migrations

程式啟動時會自動執行未執行過的 migrations。

## 目錄結構

```
migrations/
├── mysql/                              # MySQL migrations
│   ├── 001_initial_schema.up.sql       # 建立所有表
│   └── 001_initial_schema.down.sql     # 回滾用
└── timescaledb/                        # TimescaleDB migrations (一表一檔)
    ├── 001_device_metrics.up.sql       # 設備監控指標表
    ├── 001_device_metrics.down.sql
    ├── 002_es_metrics.up.sql           # ES 監控指標表
    ├── 002_es_metrics.down.sql
    ├── 003_es_alert_history.up.sql     # ES 告警歷史表
    └── 003_es_alert_history.down.sql
```

## TimescaleDB 表格清單

| 表名 | 用途 | 寫入 | 讀取 |
|------|------|------|------|
| `device_metrics` | 設備監控指標時序表 | batch_writer.go | timescale_history.go |
| `es_metrics` | ES 監控指標時序表 | batch_writer.go | es_monitor_query.go |
| `es_alert_history` | ES 告警歷史時序表 | es_monitor.go | es_alert_service.go |
| `schema_migrations` | Migration 版本追蹤 | migration.go | migration.go |

## 運作方式

1. 程式啟動時呼叫 `services.RunMigrations()`
2. 檢查 `schema_migrations` 表（記錄已執行的版本）
3. 掃描 `migrations/` 目錄的 `.up.sql` 檔案
4. 按順序執行未執行過的 migration
5. 記錄已執行的版本

## 檔案命名規則

```
{版本號}_{描述}.{up|down}.sql
```

- **版本號**: 三位數字（001, 002, 003...）
- **描述**: 簡短英文描述
- **up.sql**: 執行變更
- **down.sql**: 回滾變更

## 新增 Migration

使用 Makefile：

```bash
make migrate-create NAME=add_new_feature DB=mysql
```

或手動建立檔案：

```bash
# MySQL
touch migrations/mysql/002_add_new_feature.up.sql
touch migrations/mysql/002_add_new_feature.down.sql

# TimescaleDB
touch migrations/timescaledb/002_add_new_table.up.sql
touch migrations/timescaledb/002_add_new_table.down.sql
```

## 注意事項

1. Migration 按檔名順序執行，確保版本號正確遞增
2. 每個 migration 只執行一次，記錄在 `schema_migrations` 表
3. 回滾需手動執行 `.down.sql` 檔案

---

**最後更新**: 2025-11-22
