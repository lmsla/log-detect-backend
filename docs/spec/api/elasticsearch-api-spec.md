# 📚 Elasticsearch 監控系統 API 規格

## 🎯 API 概覽

本文檔定義了 Elasticsearch 監控系統的完整 API 規格，包括請求格式、響應結構、錯誤處理和使用範例。

## 🔐 認證說明

所有 API 端點（除了公開端點）都需要 JWT Token 認證：

```http
Authorization: Bearer <your_jwt_token>
```

## 📊 資料模型

### **雙層數據架構**

#### **ElasticsearchMonitor (配置數據 - MySQL)**
```json
{
  "id": 1,
  "name": "Production ES Cluster",
  "host": "https://es-cluster.company.com",
  "port": 9200,
  "username": "monitor_user",
  "password": "********",
  "enable_auth": true,
  "check_type": "health,performance,capacity",
  "interval": 60,
  "enable_monitor": true,
  "receivers": ["admin@company.com", "ops@company.com"],
  "subject": "ES Cluster Alert - Production",
  "description": "生產環境 ES 集群監控",

  // 告警閾值配置（獨立欄位，推薦方式）
  "cpu_usage_high": 75.0,
  "cpu_usage_critical": 85.0,
  "memory_usage_high": 80.0,
  "memory_usage_critical": 90.0,
  "disk_usage_high": 85.0,
  "disk_usage_critical": 95.0,
  "response_time_high": 3000,
  "response_time_critical": 10000,
  "unassigned_shards_threshold": 1,

  // 告警閾值配置（JSON格式，高級選項，向後兼容）
  "alert_threshold": "{\"cpu_usage_high\":75.0,\"cpu_usage_critical\":85.0}",

  "alert_dedupe_window": 300,
  "created_at": "2024-09-30T10:00:00Z",
  "updated_at": "2024-09-30T10:00:00Z"
}
```

**欄位說明**：

- **告警閾值配置（推薦使用獨立欄位）**：
  - `cpu_usage_high` (float64): CPU使用率-高閾值(%)，預設 75.0
  - `cpu_usage_critical` (float64): CPU使用率-危險閾值(%)，預設 85.0
  - `memory_usage_high` (float64): 記憶體使用率-高閾值(%)，預設 80.0
  - `memory_usage_critical` (float64): 記憶體使用率-危險閾值(%)，預設 90.0
  - `disk_usage_high` (float64): 磁碟使用率-高閾值(%)，預設 85.0
  - `disk_usage_critical` (float64): 磁碟使用率-危險閾值(%)，預設 95.0
  - `response_time_high` (int64): 響應時間-高閾值(ms)，預設 3000
  - `response_time_critical` (int64): 響應時間-危險閾值(ms)，預設 10000
  - `unassigned_shards_threshold` (int): 未分配分片閾值，預設 1

- **配置優先級**：
  1. 獨立欄位（最高優先級）
  2. alert_threshold JSON 配置（向後兼容）
  3. 預設值（最低優先級）

- `alert_threshold` (string): 告警閾值配置(JSON格式，高級選項，向後兼容)
  - 如果設置了獨立欄位，此欄位將被忽略

- `alert_dedupe_window` (int): 告警去重時間窗口（秒），預設 300 秒（5 分鐘）
  - 在此時間窗口內，相同監控器、相同類型、相同嚴重性的告警只會記錄和通知一次
  - 建議設置：
    - 高頻檢查（interval < 60s）：60-120 秒
    - 標準檢查（interval = 60s）：180-300 秒
    - 低頻檢查（interval >= 300s）：600-1800 秒

#### **ESMetrics (時間序列數據 - TimescaleDB)**
```json
{
  "time": "2024-09-30T12:00:00Z",
  "monitor_id": 1,
  "status": "online",
  "cluster_name": "production-cluster",
  "cluster_status": "green",
  "response_time": 120,
  "cpu_usage": 45.5,
  "memory_usage": 67.8,
  "disk_usage": 82.3,
  "node_count": 3,
  "data_node_count": 3,
  "query_latency": 25,
  "indexing_rate": 1500.0,
  "search_rate": 300.0,
  "total_indices": 25,
  "total_documents": 10000000,
  "total_size_bytes": 5368709120,
  "active_shards": 75,
  "relocating_shards": 0,
  "unassigned_shards": 0,
  "error_message": "",
  "warning_message": "",
  "metadata": "{\"version\":\"7.10.0\",\"jvm_version\":\"11.0.8\"}"
}
```

**重要指標說明**：

- **indexing_rate** (float64): 索引並發數（非吞吐率）
  - 來源：ES API `_stats` 的 `index_current` 欄位
  - 含義：當前正在執行的索引操作數量（瞬時並發數）
  - 範圍：0-N，表示同一時刻有幾個索引操作正在進行
  - 常見值：通常為 0-10 之間的小數（包含 0），因為操作完成速度很快
  - 注意：這**不是**每秒索引文檔數（docs/sec），而是並發操作計數

- **search_rate** (float64): 查詢並發數（非吞吐率）
  - 來源：ES API `_stats` 的 `query_current` 欄位
  - 含義：當前正在執行的查詢數量（瞬時並發數）
  - 範圍：0-N，表示同一時刻有幾個查詢正在執行
  - 常見值：通常為 0-10 之間的小數（包含 0），因為查詢完成速度很快
  - 注意：這**不是**每秒查詢數（queries/sec），而是並發查詢計數

**為什麼是並發數而非速率？**
- ES 原生 API 提供的是瞬時並發數（current），不是速率（rate）
- 計算真實吞吐率需要兩個時間點的累計值差：`(total_t2 - total_t1) / (t2 - t1)`
- 但累計值（`index_total` / `query_total`）從 ES 啟動開始累積，可達數十億，容易造成數據溢出
- 因此當前實作採用並發數，用於判斷系統是否正在處理請求

#### **ESCacheData (內存緩存數據 - 可選 Redis)**
```json
{
  "monitor_id": 1,
  "status": "online",
  "cluster_status": "green",
  "response_time": 120,
  "cpu_usage": 45.5,
  "last_check": "2024-09-30T12:00:00Z"
}
```

#### **ESAlert (告警記錄 - TimescaleDB)**
```json
{
  "time": "2024-09-30T12:05:00Z",
  "monitor_id": 1,
  "alert_type": "performance",
  "severity": "medium",
  "message": "CPU usage exceeded 80%: current 85.2%",
  "status": "active",
  "resolved_at": null,
  "resolution_note": ""
}
```

## 🛠️ API 端點詳細規格

### 1. 監控配置管理

#### 1.1 獲取所有監控配置
```http
GET /api/v1/elasticsearch/monitors
```

**權限**: `elasticsearch:read`

**Query Parameters**:
- `page` (int, optional): 頁碼，預設 1
- `limit` (int, optional): 每頁筆數，預設 10
- `search` (string, optional): 搜尋關鍵字
- `enable` (bool, optional): 過濾啟用狀態

**Response**:
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "monitors": [...],
    "total": 25,
    "page": 1,
    "limit": 10
  }
}
```

#### 1.2 新增監控配置
```http
POST /api/v1/elasticsearch/monitors
```

**權限**: `elasticsearch:create`

**Request Body**:
```json
{
  "name": "Production ES Cluster",
  "host": "https://es-cluster.company.com",
  "port": 9200,
  "username": "monitor_user",
  "password": "secure_password",
  "enable_auth": true,
  "check_type": "health,performance,capacity",
  "interval": 60,
  "enable": true,
  "receivers": ["admin@company.com"],
  "subject": "ES Cluster Alert"
}
```

**Validation Rules**:
- `name`: 必填，長度 1-100 字元
- `host`: 必填，有效的 URL 格式
- `port`: 必填，範圍 1-65535
- `interval`: 必填，最小值 30 秒
- `check_type`: 必填，可選值: health, performance, capacity
- `receivers`: 必填，有效郵件地址陣列

**Response**:
```json
{
  "code": 201,
  "message": "Monitor created successfully",
  "data": { ... }
}
```

#### 1.3 獲取特定監控配置
```http
GET /api/v1/elasticsearch/monitors/{id}
```

**權限**: `elasticsearch:read`

**Response**:
```json
{
  "code": 200,
  "message": "Success",
  "data": { ... }
}
```

#### 1.4 更新監控配置
```http
PUT /api/v1/elasticsearch/monitors/{id}
```

**權限**: `elasticsearch:update`

**Request Body**: 同新增監控配置，所有欄位可選

#### 1.5 刪除監控配置
```http
DELETE /api/v1/elasticsearch/monitors/{id}
```

**權限**: `elasticsearch:delete`

**Response**:
```json
{
  "code": 200,
  "message": "Monitor deleted successfully"
}
```

#### 1.6 測試連接
```http
POST /api/v1/elasticsearch/monitors/{id}/test
```

**權限**: `elasticsearch:read`

**Response**:
```json
{
  "code": 200,
  "message": "Connection test completed",
  "data": {
    "status": "success",
    "response_time": 150,
    "cluster_name": "production-cluster",
    "cluster_status": "green",
    "node_count": 3,
    "error_message": ""
  }
}
```

### 2. 狀態查詢 (雙層數據源)

#### 2.1 獲取即時狀態 (內存緩存 + TimescaleDB)
```http
GET /api/v1/elasticsearch/status/realtime
```

**權限**: `elasticsearch:read`

**Query Parameters**:
- `monitor_ids` (string, optional): 監控ID列表，逗號分隔

**Response** (快速響應):
```json
{
  "code": 200,
  "message": "Success",
  "data": [
    {
      "monitor_id": 1,
      "status": "online",
      "cluster_status": "green",
      "response_time": 120,
      "cpu_usage": 45.5,
      "last_check": "2024-09-30T12:00:00Z",
      "data_source": "cache" // 或 "timescale"
    }
  ]
}
```

#### 2.2 獲取歷史狀態 (TimescaleDB 時間序列)
```http
GET /api/v1/elasticsearch/status/history
```

**權限**: `elasticsearch:read`

**Query Parameters**:
- `monitor_id` (int, required): 監控ID
- `from_time` (string, required): 開始時間 (ISO 8601)
- `to_time` (string, required): 結束時間 (ISO 8601)
- `interval` (string, optional): 聚合間隔 (1m/5m/1h/1d)，預設 5m

**Response**:
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "monitor_id": 1,
    "interval": "5m",
    "metrics": [
      {
        "time": "2024-09-30T12:00:00Z",
        "status": "online",
        "cluster_status": "green",
        "avg_response_time": 120,
        "avg_cpu_usage": 45.5,
        "avg_memory_usage": 67.8,
        "data_points": 5
      }
    ],
    "statistics": {
      "total_points": 288,
      "uptime_rate": 99.3,
      "avg_response_time": 125,
      "max_response_time": 250
    }
  }
}
```

#### 2.3 獲取特定監控最新狀態 (智能路由)
```http
GET /api/v1/elasticsearch/status/{monitor_id}/latest
```

**權限**: `elasticsearch:read`

**智能數據源選擇**:
- 優先從內存緩存獲取 (最近數據)
- 回退到 TimescaleDB (3個月內數據)

**Response**:
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "monitor_id": 1,
    "status": "online",
    "cluster_name": "production-cluster",
    "cluster_status": "green",
    "response_time": 120,
    "cpu_usage": 45.5,
    "memory_usage": 67.8,
    "disk_usage": 82.3,
    "node_count": 3,
    "active_shards": 75,
    "unassigned_shards": 0,
    "last_check": "2024-09-30T12:00:00Z",
    "data_source": "cache" // 或 "timescale"
  }
}
```

**權限**: `elasticsearch:read`

**Response**: 返回最新狀態記錄

#### 2.3 獲取歷史狀態記錄
```http
GET /api/v1/elasticsearch/status/{monitor_id}/history
```

**權限**: `elasticsearch:read`

**Query Parameters**:
- `from_time` (string, required): 開始時間
- `to_time` (string, required): 結束時間
- `interval` (string, optional): 聚合間隔 (1m/5m/1h/1d)

#### 2.4 獲取趨勢數據
```http
GET /api/v1/elasticsearch/status/{monitor_id}/trends
```

**權限**: `elasticsearch:read`

**Query Parameters**:
- `metric` (string, required): 指標名稱 (cpu_usage/memory_usage/response_time)
- `period` (string, optional): 時間週期 (1h/6h/24h/7d/30d)

**Response**:
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "metric": "cpu_usage",
    "period": "24h",
    "data_points": [
      {
        "timestamp": 1696075200,
        "value": 45.5,
        "time": "2024-09-30T12:00:00Z"
      }
    ],
    "statistics": {
      "min": 35.2,
      "max": 78.9,
      "avg": 52.3,
      "current": 45.5
    }
  }
}
```

### 3. 告警管理

#### 3.1 獲取告警列表
```http
GET /api/v1/elasticsearch/alerts
```

**權限**: `elasticsearch:read`

**Query Parameters**:
- `monitor_id` (int, optional): 過濾特定監控
- `status` (string, optional): 告警狀態 (active/resolved)
- `severity` (string, optional): 嚴重程度 (low/medium/high/critical)
- `alert_type` (string, optional): 告警類型

#### 3.2 獲取告警詳情
```http
GET /api/v1/elasticsearch/alerts/{id}
```

**權限**: `elasticsearch:read`

#### 3.3 解決告警
```http
POST /api/v1/elasticsearch/alerts/{id}/resolve
```

**權限**: `elasticsearch:update`

**Request Body**:
```json
{
  "resolution_note": "Issue has been resolved by restarting the service"
}
```

#### 3.4 確認告警
```http
POST /api/v1/elasticsearch/alerts/{id}/acknowledge
```

**權限**: `elasticsearch:update`

### 4. 儀表板和統計

#### 4.1 ES 監控儀表板 (高性能混合查詢)
```http
GET /api/v1/elasticsearch/dashboard
```

**權限**: `elasticsearch:read`

**智能數據整合**: 結合內存緩存與 TimescaleDB 歷史統計

**Response** (快速響應):
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "summary": {
      "total_monitors": 5,
      "active_monitors": 4,
      "online_clusters": 3,
      "offline_clusters": 1,
      "active_alerts": 2,
      "resolved_alerts_today": 8,
      "data_freshness": "real-time" // 緩存數據
    },
    "realtime_status": [
      {
        "monitor_id": 1,
        "name": "Production ES",
        "status": "online",
        "cluster_status": "green",
        "response_time": 120,
        "cpu_usage": 45.5,
        "last_check": "2024-09-30T12:00:00Z",
        "source": "cache"
      }
    ],
    "performance_trends": {
      "period": "24h",
      "source": "timescale",
      "metrics": [
        {
          "monitor_id": 1,
          "avg_response_time": 125,
          "uptime_rate": 99.3,
          "peak_cpu": 78.5,
          "trend": "stable"
        }
      ]
    },
    "recent_alerts": [
      {
        "id": 123,
        "monitor_id": 2,
        "severity": "medium",
        "message": "CPU usage high: 85%",
        "time": "2024-09-30T11:45:00Z",
        "status": "active"
      }
    ],
    "system_health": {
      "avg_response_time": 150,
      "total_documents": 50000000,
      "total_storage_gb": 500,
      "compression_ratio": 85.5,
      "last_updated": "2024-09-30T12:00:00Z"
    }
  }
}
```

#### 4.2 ES 監控摘要
```http
GET /api/v1/elasticsearch/summary
```

**權限**: `elasticsearch:read`

**Query Parameters**:
- `period` (string, optional): 統計週期 (24h/7d/30d)

#### 4.3 獲取指標數據
```http
GET /api/v1/elasticsearch/metrics/{monitor_id}
```

**權限**: `elasticsearch:read`

**Query Parameters**:
- `metrics` (string, required): 指標列表，逗號分隔
- `from_time` (string, required): 開始時間
- `to_time` (string, required): 結束時間
- `interval` (string, optional): 數據間隔

## ⚠️ 錯誤處理

### 標準錯誤響應格式
```json
{
  "code": 400,
  "message": "Invalid request parameters",
  "errors": [
    {
      "field": "interval",
      "message": "Interval must be at least 30 seconds"
    }
  ]
}
```

### 常見錯誤碼
- `400` - 請求參數錯誤
- `401` - 未認證
- `403` - 權限不足
- `404` - 資源不存在
- `409` - 資源衝突 (如重複名稱)
- `422` - 數據驗證失敗
- `500` - 伺服器內部錯誤

## 📝 使用範例

### 完整流程範例

```bash
# 1. 登入獲取 Token
TOKEN=$(curl -s -X POST http://localhost:8006/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.token')

# 2. 新增 ES 監控配置
curl -X POST http://localhost:8006/api/v1/elasticsearch/monitors \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test ES Cluster",
    "host": "http://localhost:9200",
    "port": 9200,
    "enable_auth": false,
    "check_type": "health,performance",
    "interval": 60,
    "enable": true,
    "receivers": ["admin@test.com"],
    "subject": "ES Test Alert"
  }'

# 3. 測試連接
curl -X POST http://localhost:8006/api/v1/elasticsearch/monitors/1/test \
  -H "Authorization: Bearer $TOKEN"

# 4. 查看監控狀態
curl -X GET http://localhost:8006/api/v1/elasticsearch/status/1 \
  -H "Authorization: Bearer $TOKEN"

# 5. 查看儀表板
curl -X GET http://localhost:8006/api/v1/elasticsearch/dashboard \
  -H "Authorization: Bearer $TOKEN"
```

## 🔄 WebSocket 即時更新 (未來功能)

```javascript
// 連接 WebSocket 獲取即時狀態更新
const ws = new WebSocket('ws://localhost:8006/ws/elasticsearch/status');

ws.onmessage = function(event) {
  const statusUpdate = JSON.parse(event.data);
  console.log('ES Status Update:', statusUpdate);
};
```

---

**版本**: 1.0
**最後更新**: 2024-09-30
**作者**: Log Detect 開發團隊