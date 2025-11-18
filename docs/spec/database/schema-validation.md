# Elasticsearch 監控資料庫表結構檢查

## 📋 檢查結果

### ✅ MySQL - elasticsearch_monitors 表

**狀態**: ✅ **不需要更新**

**實體定義** (`entities/elasticsearch.go`):
```go
type ElasticsearchMonitor struct {
    models.Common
    ID                int      `gorm:"primaryKey;index" json:"id" form:"id"`
    Name              string   `json:"name" gorm:"type:varchar(100);not null;comment:監控名稱"`
    Host              string   `json:"host" gorm:"type:varchar(255);not null;comment:ES 主機地址"`
    Port              int      `json:"port" gorm:"type:int;not null;default:9200;comment:ES 端口"`
    Username          string   `json:"username" gorm:"type:varchar(100);comment:認證用戶名"`
    Password          string   `json:"password" gorm:"type:varchar(255);comment:認證密碼"`
    EnableAuth        bool     `json:"enable_auth" gorm:"type:tinyint(1);default:0;comment:是否啟用認證"`
    CheckType         string   `json:"check_type" gorm:"type:varchar(100);default:'health,performance';comment:檢查類型(逗號分隔)"`
    Interval          int      `json:"interval" gorm:"type:int;not null;default:60;comment:檢查間隔(秒,範圍:10-3600)"`
    EnableMonitor     bool     `json:"enable_monitor" gorm:"type:tinyint(1);default:1;comment:是否啟用監控"`
    Receivers         []string `json:"receivers" gorm:"type:json;serializer:json;comment:告警收件人陣列"`
    Subject           string   `json:"subject" gorm:"type:varchar(255);comment:告警主題"`
    Description       string   `json:"description" gorm:"type:text;comment:監控描述"`

    // 告警閾值配置（獨立欄位，前端友好）
    CPUUsageHigh            *float64 `json:"cpu_usage_high" gorm:"type:decimal(5,2);comment:CPU使用率-高閾值(%)"`
    CPUUsageCritical        *float64 `json:"cpu_usage_critical" gorm:"type:decimal(5,2);comment:CPU使用率-危險閾值(%)"`
    MemoryUsageHigh         *float64 `json:"memory_usage_high" gorm:"type:decimal(5,2);comment:記憶體使用率-高閾值(%)"`
    MemoryUsageCritical     *float64 `json:"memory_usage_critical" gorm:"type:decimal(5,2);comment:記憶體使用率-危險閾值(%)"`
    DiskUsageHigh           *float64 `json:"disk_usage_high" gorm:"type:decimal(5,2);comment:磁碟使用率-高閾值(%)"`
    DiskUsageCritical       *float64 `json:"disk_usage_critical" gorm:"type:decimal(5,2);comment:磁碟使用率-危險閾值(%)"`
    ResponseTimeHigh        *int64   `json:"response_time_high" gorm:"type:bigint;comment:響應時間-高閾值(ms)"`
    ResponseTimeCritical    *int64   `json:"response_time_critical" gorm:"type:bigint;comment:響應時間-危險閾值(ms)"`
    UnassignedShardsThreshold *int   `json:"unassigned_shards_threshold" gorm:"type:int;comment:未分配分片閾值"`

    // 保留 JSON 欄位作為高級配置選項（向後兼容）
    AlertThreshold    string   `json:"alert_threshold" gorm:"type:json;comment:告警閾值配置(JSON,高級選項)"`
    AlertDedupeWindow int      `json:"alert_dedupe_window" gorm:"type:int;default:300;comment:告警去重時間窗口(秒,預設300秒=5分鐘)"`
}
```

**AutoMigrate 狀態**: ✅ 已註冊
- 檔案: `services/sqltable.go:25`
- 代碼: `&entities.ElasticsearchMonitor{}, // ES 監控配置表`

**說明**:
- ✅ 使用 GORM AutoMigrate，會自動創建/更新表結構
- ✅ 包含 models.Common（created_at, updated_at, deleted_at）
- ✅ 所有欄位定義完整
- ✅ 欄位註釋已更新（包含單位說明）
- ✅ 不需要手動 SQL 腳本

**預期生成的表結構**:
```sql
CREATE TABLE `elasticsearch_monitors` (
  `id` int NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  `name` varchar(100) NOT NULL COMMENT '監控名稱',
  `host` varchar(255) NOT NULL COMMENT 'ES 主機地址',
  `port` int NOT NULL DEFAULT 9200 COMMENT 'ES 端口',
  `username` varchar(100) DEFAULT NULL COMMENT '認證用戶名',
  `password` varchar(255) DEFAULT NULL COMMENT '認證密碼',
  `enable_auth` tinyint(1) DEFAULT 0 COMMENT '是否啟用認證',
  `check_type` varchar(100) DEFAULT 'health,performance' COMMENT '檢查類型(逗號分隔)',
  `interval` int NOT NULL DEFAULT 60 COMMENT '檢查間隔(秒,範圍:10-3600)',
  `enable_monitor` tinyint(1) DEFAULT 1 COMMENT '是否啟用監控',
  `receivers` json DEFAULT NULL COMMENT '告警收件人陣列',
  `subject` varchar(255) DEFAULT NULL COMMENT '告警主題',
  `description` text COMMENT '監控描述',

  -- 告警閾值配置（獨立欄位，前端友好）
  `cpu_usage_high` decimal(5,2) DEFAULT NULL COMMENT 'CPU使用率-高閾值(%)',
  `cpu_usage_critical` decimal(5,2) DEFAULT NULL COMMENT 'CPU使用率-危險閾值(%)',
  `memory_usage_high` decimal(5,2) DEFAULT NULL COMMENT '記憶體使用率-高閾值(%)',
  `memory_usage_critical` decimal(5,2) DEFAULT NULL COMMENT '記憶體使用率-危險閾值(%)',
  `disk_usage_high` decimal(5,2) DEFAULT NULL COMMENT '磁碟使用率-高閾值(%)',
  `disk_usage_critical` decimal(5,2) DEFAULT NULL COMMENT '磁碟使用率-危險閾值(%)',
  `response_time_high` bigint DEFAULT NULL COMMENT '響應時間-高閾值(ms)',
  `response_time_critical` bigint DEFAULT NULL COMMENT '響應時間-危險閾值(ms)',
  `unassigned_shards_threshold` int DEFAULT NULL COMMENT '未分配分片閾值',

  -- 保留 JSON 欄位（向後兼容）
  `alert_threshold` json DEFAULT NULL COMMENT '告警閾值配置(JSON,高級選項)',
  `alert_dedupe_window` int DEFAULT 300 COMMENT '告警去重時間窗口(秒,預設300秒=5分鐘)',

  KEY `idx_elasticsearch_monitors_deleted_at` (`deleted_at`),
  KEY `idx_id` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

---

### ✅ TimescaleDB - es_metrics 表

**狀態**: ✅ **不需要更新**

**實體定義** (`entities/elasticsearch.go`):
```go
type ESMetric struct {
    Time               time.Time `json:"time"`
    MonitorID          int       `json:"monitor_id"`
    Status             string    `json:"status"` // online, offline, warning, error
    ClusterName        string    `json:"cluster_name"`
    ClusterStatus      string    `json:"cluster_status"` // green, yellow, red
    ResponseTime       int64     `json:"response_time"` // 響應時間（單位：毫秒）
    CPUUsage           float64   `json:"cpu_usage"` // CPU 使用率（單位：百分比 0-100）
    MemoryUsage        float64   `json:"memory_usage"` // 記憶體使用率（單位：百分比 0-100）
    DiskUsage          float64   `json:"disk_usage"` // 磁碟使用率（單位：百分比 0-100）
    NodeCount          int       `json:"node_count"`
    DataNodeCount      int       `json:"data_node_count"`
    QueryLatency       int64     `json:"query_latency"` // 毫秒
    IndexingRate       float64   `json:"indexing_rate"` // 索引並發數（index_current，非速率）
    SearchRate         float64   `json:"search_rate"` // 搜尋並發數（query_current，非速率）
    TotalIndices       int       `json:"total_indices"` // 索引總數
    TotalDocuments     int64     `json:"total_documents"` // 文檔總數
    TotalSizeBytes     int64     `json:"total_size_bytes"` // 總大小(字節)
    ActiveShards       int       `json:"active_shards"` // 活躍分片數
    RelocatingShards   int       `json:"relocating_shards"` // 遷移中分片數
    UnassignedShards   int       `json:"unassigned_shards"` // 未分配分片數
    ErrorMessage       string    `json:"error_message"`
    WarningMessage     string    `json:"warning_message"`
    Metadata           string    `json:"metadata"` // JSON 格式的額外元數據
}
```

**SQL 腳本狀態**: ✅ 已完整
- 檔案: `postgresql_install.sh:82-124`
- 欄位: 23 個欄位全部包含
- 索引: 3 個性能索引已創建
- Hypertable: 已設置，按天分區
- 壓縮策略: 7 天後壓縮
- 保留策略: 90 天自動清理

**實際 SQL 定義**:
```sql
CREATE TABLE IF NOT EXISTS es_metrics (
    time TIMESTAMPTZ NOT NULL,
    monitor_id INTEGER NOT NULL,
    status TEXT NOT NULL,
    cluster_name TEXT,
    cluster_status TEXT,
    response_time BIGINT DEFAULT 0,
    cpu_usage DECIMAL(5,2) DEFAULT 0.00,
    memory_usage DECIMAL(5,2) DEFAULT 0.00,
    disk_usage DECIMAL(5,2) DEFAULT 0.00,
    node_count INTEGER DEFAULT 0,
    data_node_count INTEGER DEFAULT 0,
    query_latency BIGINT DEFAULT 0,
    indexing_rate DECIMAL(10,2) DEFAULT 0.00,
    search_rate DECIMAL(10,2) DEFAULT 0.00,
    total_indices INTEGER DEFAULT 0,
    total_documents BIGINT DEFAULT 0,
    total_size_bytes BIGINT DEFAULT 0,
    active_shards INTEGER DEFAULT 0,
    relocating_shards INTEGER DEFAULT 0,
    unassigned_shards INTEGER DEFAULT 0,
    error_message TEXT,
    warning_message TEXT,
    metadata JSONB
);
```

**對比結果**: ✅ **完全一致**

---

### ✅ TimescaleDB - es_alert_history 表

**狀態**: ✅ **不需要更新**

**實體定義** (`entities/elasticsearch.go`):
```go
type ESAlert struct {
    Time           time.Time  `json:"time"`
    MonitorID      int        `json:"monitor_id"`
    AlertType      string     `json:"alert_type"` // health, performance, capacity, availability
    Severity       string     `json:"severity"` // critical, high, medium, low
    Message        string     `json:"message"`
    Status         string     `json:"status"` // active, resolved
    ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
    ResolutionNote string     `json:"resolution_note,omitempty"`
}
```

**SQL 腳本狀態**: ✅ 已完整
- 檔案: `postgresql_install.sh:126-137`
- 欄位: 所有欄位包含
- Hypertable: 已設置，按 7 天分區
- 保留策略: 90 天自動清理

**實際 SQL 定義**:
```sql
CREATE TABLE IF NOT EXISTS es_alert_history (
    time TIMESTAMPTZ NOT NULL,
    monitor_id INTEGER NOT NULL,
    alert_type TEXT NOT NULL,
    severity TEXT NOT NULL,
    message TEXT NOT NULL,
    status TEXT DEFAULT 'active',
    resolved_at TIMESTAMPTZ,
    resolution_note TEXT
);
```

**對比結果**: ✅ **完全一致**

---

## 📊 總結

| 資料庫 | 表名 | 狀態 | 需要更新 |
|--------|------|------|----------|
| MySQL | `elasticsearch_monitors` | ✅ 完整 | ❌ 不需要 |
| TimescaleDB | `es_metrics` | ✅ 完整 | ❌ 不需要 |
| TimescaleDB | `es_alert_history` | ✅ 完整 | ❌ 不需要 |

### ✅ 確認項目

1. **MySQL 表**
   - ✅ 實體定義完整
   - ✅ 已註冊到 AutoMigrate
   - ✅ 欄位註釋已更新（包含單位說明）
   - ✅ 會在應用啟動時自動創建/更新

2. **TimescaleDB 表**
   - ✅ es_metrics 表結構完整（23 欄位）
   - ✅ es_alert_history 表結構完整
   - ✅ Hypertable 設置正確
   - ✅ 索引已創建
   - ✅ 壓縮和保留策略已配置

3. **註釋更新**
   - ✅ entities/elasticsearch.go 中的註釋已更新
   - ✅ 包含單位說明（毫秒、百分比、秒）
   - ✅ MySQL GORM comment 會自動同步到資料庫

---

## 🚀 部署檢查清單

### 首次部署（新環境）

1. **執行 TimescaleDB 初始化腳本**
   ```bash
   bash postgresql_install.sh
   ```

2. **啟動應用**（MySQL 表會自動創建）
   ```bash
   go run main.go
   ```

3. **驗證表結構**
   ```bash
   # 檢查 MySQL
   mysql -u root -p logdetect -e "DESCRIBE elasticsearch_monitors;"

   # 檢查 TimescaleDB
   psql -U logdetect -d monitoring -c "\d es_metrics"
   psql -U logdetect -d monitoring -c "\d es_alert_history"
   ```

### 現有環境升級

**情況 1**: 如果 elasticsearch_monitors 表已存在
```sql
-- MySQL 不需要手動更新
-- GORM AutoMigrate 會自動添加缺少的欄位
-- 只需重啟應用即可
```

**情況 2**: 如果 TimescaleDB 表已存在
```sql
-- 檢查是否缺少欄位
SELECT column_name, data_type
FROM information_schema.columns
WHERE table_name = 'es_metrics'
ORDER BY ordinal_position;

-- 如果缺少欄位，手動添加（不太可能，因為之前的版本已經是完整的）
-- ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS xxx ...;
```

**結論**: ✅ **不需要任何資料庫更新**

---

## 💡 註釋更新的影響

### GORM 註釋更新

**更新的註釋**:
```go
// 舊註釋
Interval int `gorm:"type:int;not null;default:60;comment:檢查間隔(秒)"`

// 新註釋
Interval int `gorm:"type:int;not null;default:60;comment:檢查間隔(秒,範圍:10-3600)"`
```

**影響**:
- ✅ 只影響欄位說明文字
- ✅ 不改變欄位類型或結構
- ✅ MySQL AutoMigrate 會更新 COMMENT
- ✅ 不影響現有資料

**驗證方式**:
```sql
-- 查看更新後的註釋
SHOW FULL COLUMNS FROM elasticsearch_monitors LIKE 'interval';
```

---

## 📝 建議

### ✅ 不需要做的事
- ❌ 不需要手動修改 MySQL 表結構
- ❌ 不需要重新執行 postgresql_install.sh
- ❌ 不需要資料遷移
- ❌ 不需要更新索引

### ✅ 建議做的事（可選）
1. **重啟應用**
   - 讓 GORM 更新 MySQL 註釋
   - 確保最新的程式碼生效

2. **驗證表結構**
   ```bash
   # 驗證 MySQL 表
   mysql -u root -p logdetect -e "SHOW CREATE TABLE elasticsearch_monitors\G"

   # 驗證 TimescaleDB 表
   psql -U logdetect -d monitoring -c "\d+ es_metrics"
   ```

3. **測試資料插入**
   ```bash
   # 使用 API 創建一個測試監控配置
   curl -X POST http://localhost:8006/api/v1/elasticsearch/monitors \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"name":"Test","host":"localhost","port":9200}'
   ```

---

**檢查日期**: 2025-10-06
**結論**: ✅ **資料庫表結構完整，不需要更新**
