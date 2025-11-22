# 🔧 快速修復：es_metrics 表結構錯誤

## 問題描述

API 請求 `/api/v1/elasticsearch/statistics` 時出現錯誤：

```json
{
  "msg": "pq: column \"total_indices\" does not exist",
  "success": false
}
```

## 根本原因

TimescaleDB 中的 `es_metrics` 表使用舊版腳本創建，缺少以下欄位：
- `total_indices`
- `total_documents`
- `total_size_bytes`
- `active_shards`
- `relocating_shards`
- `unassigned_shards`
- `query_latency`
- `indexing_rate`
- `search_rate`

標準表應該有 **23 個欄位**，舊版本可能只有 **14 個**。

---

## 🚀 快速修復方案

### 方案 A: 執行自動修復腳本（推薦）

```bash
cd /Users/chen/Downloads/01BiMap/03MyDevs/log-detect/log-detect-backend

# 執行檢查與修復腳本
psql -U logdetect -d monitoring -f scripts/check_and_fix_es_metrics_table.sql
```

**腳本會自動**:
1. ✅ 檢查當前表結構
2. ✅ 列出缺少的欄位
3. ✅ 安全地添加缺少的欄位（使用 IF NOT EXISTS）
4. ✅ 驗證最終結構

---

### 方案 B: 手動添加欄位

如果無法執行腳本，手動執行以下 SQL：

```sql
-- 連接到 TimescaleDB
psql -U logdetect -d monitoring

-- 添加缺少的欄位
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS total_indices INTEGER DEFAULT 0;
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS total_documents BIGINT DEFAULT 0;
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS total_size_bytes BIGINT DEFAULT 0;
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS active_shards INTEGER DEFAULT 0;
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS relocating_shards INTEGER DEFAULT 0;
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS unassigned_shards INTEGER DEFAULT 0;
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS query_latency BIGINT DEFAULT 0;
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS indexing_rate DECIMAL(10,2) DEFAULT 0.00;
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS search_rate DECIMAL(10,2) DEFAULT 0.00;

-- 驗證欄位數量（應該是 23）
SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'es_metrics';
```

---

### 方案 C: 重建表（如果資料可以丟失）

⚠️ **警告**: 此方案會刪除所有現有資料

```bash
# 重新執行完整的初始化腳本
psql -U logdetect -d monitoring

# 刪除舊表
DROP TABLE IF EXISTS es_metrics CASCADE;
DROP TABLE IF EXISTS es_alert_history CASCADE;

# 重新創建（執行 postgresql_install.sh 中的相關部分）
\i postgresql_install.sh
```

---

## 📋 驗證步驟

### 1. 檢查表結構

```sql
-- 連接到資料庫
psql -U logdetect -d monitoring

-- 查看所有欄位
SELECT column_name, data_type, column_default
FROM information_schema.columns
WHERE table_name = 'es_metrics'
ORDER BY ordinal_position;

-- 檢查欄位數量（應該是 23）
SELECT COUNT(*) as column_count
FROM information_schema.columns
WHERE table_name = 'es_metrics';
```

**預期結果**: 23 個欄位

### 2. 檢查必要欄位

```sql
-- 檢查關鍵欄位是否存在
SELECT
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'es_metrics' AND column_name = 'total_indices') THEN '✅' ELSE '❌' END as total_indices,
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'es_metrics' AND column_name = 'total_documents') THEN '✅' ELSE '❌' END as total_documents,
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'es_metrics' AND column_name = 'total_size_bytes') THEN '✅' ELSE '❌' END as total_size_bytes,
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'es_metrics' AND column_name = 'active_shards') THEN '✅' ELSE '❌' END as active_shards,
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'es_metrics' AND column_name = 'query_latency') THEN '✅' ELSE '❌' END as query_latency;
```

**預期結果**: 所有欄位都顯示 ✅

### 3. 測試 API

```bash
# 測試統計端點
curl -X GET http://localhost:8006/api/v1/elasticsearch/statistics \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**預期結果**: 返回 200 OK 和統計資料

---

## 🔍 完整的表結構定義

標準的 `es_metrics` 表應該包含以下 23 個欄位：

| # | 欄位名 | 資料類型 | 預設值 | 說明 |
|---|--------|----------|--------|------|
| 1 | time | TIMESTAMPTZ | - | 時間戳記（主鍵） |
| 2 | monitor_id | INTEGER | - | 監控器 ID |
| 3 | status | TEXT | - | 狀態 (online/offline/warning/error) |
| 4 | cluster_name | TEXT | NULL | 集群名稱 |
| 5 | cluster_status | TEXT | NULL | 集群狀態 (green/yellow/red) |
| 6 | response_time | BIGINT | 0 | 響應時間（毫秒） |
| 7 | cpu_usage | DECIMAL(5,2) | 0.00 | CPU 使用率（%） |
| 8 | memory_usage | DECIMAL(5,2) | 0.00 | 記憶體使用率（%） |
| 9 | disk_usage | DECIMAL(5,2) | 0.00 | 磁碟使用率（%） |
| 10 | node_count | INTEGER | 0 | 節點數量 |
| 11 | data_node_count | INTEGER | 0 | 數據節點數量 |
| 12 | query_latency | BIGINT | 0 | 查詢延遲（毫秒） |
| 13 | indexing_rate | DECIMAL(10,2) | 0.00 | 索引速率（docs/s） |
| 14 | search_rate | DECIMAL(10,2) | 0.00 | 搜尋速率（queries/s） |
| 15 | **total_indices** | INTEGER | 0 | 索引總數 ⚠️ |
| 16 | **total_documents** | BIGINT | 0 | 文檔總數 ⚠️ |
| 17 | **total_size_bytes** | BIGINT | 0 | 總大小（bytes） ⚠️ |
| 18 | **active_shards** | INTEGER | 0 | 活躍分片數 ⚠️ |
| 19 | **relocating_shards** | INTEGER | 0 | 遷移中分片數 ⚠️ |
| 20 | **unassigned_shards** | INTEGER | 0 | 未分配分片數 ⚠️ |
| 21 | error_message | TEXT | NULL | 錯誤訊息 |
| 22 | warning_message | TEXT | NULL | 警告訊息 |
| 23 | metadata | JSONB | NULL | 額外元數據 |

⚠️ 標記的欄位是錯誤訊息中提到可能缺少的欄位

---

## 🛠️ 故障排查

### 問題 1: 腳本執行後仍報錯

**可能原因**: 應用緩存或連接池

**解決方法**:
```bash
# 重啟應用
# 或刷新資料庫連接池
```

### 問題 2: 權限不足

**錯誤**: `ERROR: permission denied for table es_metrics`

**解決方法**:
```sql
-- 授予權限
GRANT ALL PRIVILEGES ON TABLE es_metrics TO logdetect;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO logdetect;
```

### 問題 3: Hypertable 限制

**錯誤**: `ERROR: cannot add column to hypertable`

**解決方法**:
```sql
-- TimescaleDB 2.0+ 支援添加欄位到 hypertable
-- 如果版本太舊，需要升級 TimescaleDB
SELECT extversion FROM pg_extension WHERE extname = 'timescaledb';

-- 或者暫時禁用壓縮後添加
SELECT decompress_chunk(chunk) FROM show_chunks('es_metrics') chunk;
-- 添加欄位
ALTER TABLE es_metrics ADD COLUMN ...;
-- 重新啟用壓縮
SELECT compress_chunk(chunk) FROM show_chunks('es_metrics') chunk;
```

---

## 📊 驗證清單

完成修復後，請確認以下項目：

- [ ] es_metrics 表有 23 個欄位
  ```sql
  SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'es_metrics';
  ```

- [ ] total_indices 欄位存在
  ```sql
  SELECT column_name FROM information_schema.columns
  WHERE table_name = 'es_metrics' AND column_name = 'total_indices';
  ```

- [ ] API 請求成功
  ```bash
  curl http://localhost:8006/api/v1/elasticsearch/statistics \
    -H "Authorization: Bearer $TOKEN"
  ```

- [ ] 返回正確的統計資料結構
  ```json
  {
    "success": true,
    "msg": "查詢成功",
    "body": {
      "total_monitors": 0,
      "online_monitors": 0,
      ...
    }
  }
  ```

---

## 📝 預防措施

### 1. 使用最新的初始化腳本

確保使用 `postgresql_install.sh` 的最新版本（包含所有 23 個欄位）。

### 2. 版本控制

將資料庫 schema 版本記錄在版本控制中：

```sql
-- 創建 schema_version 表
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at TIMESTAMPTZ DEFAULT NOW(),
    description TEXT
);

-- 記錄當前版本
INSERT INTO schema_version (version, description)
VALUES (2, 'Added missing columns to es_metrics table');
```

### 3. 遷移腳本

為未來的 schema 變更創建遷移腳本，放在 `migrations/` 目錄。

---

## 🆘 需要幫助？

如果以上方法都無法解決問題，請提供：

1. 當前表結構：
   ```sql
   \d+ es_metrics
   ```

2. TimescaleDB 版本：
   ```sql
   SELECT extversion FROM pg_extension WHERE extname = 'timescaledb';
   ```

3. 完整錯誤訊息（包含 stack trace）

---

**更新日期**: 2025-10-07
**相關檔案**:
- `postgresql_install.sh:82-124` - 完整表定義
- `services/es_monitor_query.go:196-229` - 使用 total_indices 的查詢
- `scripts/check_and_fix_es_metrics_table.sql` - 自動修復腳本
