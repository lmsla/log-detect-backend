# 資料庫遷移腳本

此資料夾包含所有資料庫結構變更的 SQL 遷移腳本。

## 📁 檔案命名規則

```
{序號}_{描述}.{方向}.sql
```

- **序號**: 三位數字（001, 002, 003...）
- **描述**: 簡短的英文描述，使用底線分隔
- **方向**:
  - `up.sql` - 執行遷移（升級）
  - `down.sql` - 回滾遷移（降級）

## 📋 現有遷移

| 序號 | 描述 | 日期 | 狀態 |
|------|------|------|------|
| 001 | 建立 es_connections 表 | 2025-11-18 | ✅ 已建立 |
| 002 | 修改 indices 表新增 es_connection_id | 2025-11-18 | ✅ 已建立 |
| 003 | 修改 elasticsearch_monitors 表新增 es_connection_id | 2025-11-18 | ✅ 已建立 |

## 🚀 執行遷移

### 方法 1: 手動執行（開發環境）

```bash
# 連接到 MySQL
mysql -u runner -p -h 10.99.1.133 logdetect

# 執行遷移
source migrations/001_create_es_connections.up.sql
source migrations/002_alter_indices_add_es_connection.up.sql
source migrations/003_alter_elasticsearch_monitors_add_es_connection.up.sql
```

### 方法 2: 使用腳本批次執行

```bash
# 執行所有 up 遷移
for file in migrations/*.up.sql; do
    echo "Executing: $file"
    mysql -u runner -p -h 10.99.1.133 logdetect < "$file"
done
```

### 方法 3: 使用 GORM AutoMigrate（推薦）

在程式碼中使用 GORM 的 AutoMigrate 功能：

```go
// main.go 或初始化函數中
db.AutoMigrate(
    &entities.ESConnection{},
    &entities.Index{},
    &entities.ElasticsearchMonitor{},
)
```

**注意**: AutoMigrate 會自動建立表和欄位，但不會刪除欄位。

## ⏮️ 回滾遷移

如果需要回滾變更（例如測試失敗或需要降級）：

```bash
# 按相反順序回滾
mysql -u runner -p -h 10.99.1.133 logdetect < migrations/003_alter_elasticsearch_monitors_add_es_connection.down.sql
mysql -u runner -p -h 10.99.1.133 logdetect < migrations/002_alter_indices_add_es_connection.down.sql
mysql -u runner -p -h 10.99.1.133 logdetect < migrations/001_create_es_connections.down.sql
```

## ⚠️ 注意事項

### 執行前檢查

1. **備份資料庫**（生產環境必做）
   ```bash
   mysqldump -u runner -p -h 10.99.1.133 logdetect > backup_$(date +%Y%m%d_%H%M%S).sql
   ```

2. **檢查外鍵約束**
   - 確保相關表已存在
   - 確保沒有孤立的外鍵資料

3. **測試環境先行**
   - 在開發/測試環境先執行並驗證
   - 確認無錯誤後再部署到生產環境

### 遷移順序

**重要**: 必須按照序號順序執行遷移！

- ✅ 正確: 001 → 002 → 003
- ❌ 錯誤: 002 → 001 → 003

### 回滾順序

**重要**: 回滾時必須按照相反順序執行！

- ✅ 正確: 003.down → 002.down → 001.down
- ❌ 錯誤: 001.down → 002.down → 003.down

## 🔧 遷移腳本說明

### 001_create_es_connections

建立 `es_connections` 表，用於統一管理所有 Elasticsearch 連線配置。

**欄位**:
- `id`: 主鍵
- `name`: 連線名稱（唯一）
- `host`, `port`: ES 地址
- `username`, `password`: 認證資訊
- `enable_auth`, `use_tls`: 連線選項
- `is_default`: 是否為預設連線
- `description`: 描述

### 002_alter_indices_add_es_connection

為 `indices` 表新增 `es_connection_id` 欄位，關聯到 `es_connections` 表。

**影響**:
- 允許不同的 Index 使用不同的 ES 連線
- `NULL` 值表示使用預設連線（向後兼容）
- `ON DELETE RESTRICT` 防止刪除被使用中的連線

### 003_alter_elasticsearch_monitors_add_es_connection

為 `elasticsearch_monitors` 表新增 `es_connection_id` 欄位（可選）。

**影響**:
- 允許健康監控複用 indices 的 ES 連線配置
- `NULL` 值表示使用自己的 host/port（獨立監控）
- `ON DELETE SET NULL` 刪除連線時監控器保留

## 📝 新增遷移

如需新增遷移腳本：

1. 確定序號（下一個可用的序號）
2. 建立兩個檔案：
   - `{序號}_{描述}.up.sql` - 升級腳本
   - `{序號}_{描述}.down.sql` - 降級腳本
3. 在此 README 更新遷移清單
4. 測試 up 和 down 腳本都能正確執行

## 🔗 相關文件

- **Issue 文件**: `docs/issues/001-es-connection-management.md`
- **資料庫設計**: `docs/specs_cn/04-資料庫設計.md`
- **架構設計**: `docs/specs_cn/02-架構設計.md`

---

**最後更新**: 2025-11-18
