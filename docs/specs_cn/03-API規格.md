# Log Detect Backend - API 規格文件

## 文件資訊

| 項目 | 內容 |
|------|------|
| 文件版本 | 2.0 |
| 最後更新 | 2025-11-18 |
| 文件類型 | 軟體設計文件 (SDD) - API 規格 |

---

## 1. API 概述

### 1.1 基本資訊

- **基礎 URL**: `http://your-server:8006`
- **協定**: HTTP/HTTPS
- **API 版本**: v1
- **API 前綴**: `/api/v1`
- **認證方式**: JWT Bearer Token
- **回應格式**: JSON
- **字元編碼**: UTF-8

### 1.2 通用回應格式

所有 API 端點統一使用以下回應格式：

```json
{
  "code": 200,
  "message": "success",
  "data": {...}
}
```

#### 狀態碼說明

| HTTP 狀態碼 | code | 說明 |
|------------|------|------|
| 200 | 200 | 成功 |
| 201 | 201 | 建立成功 |
| 400 | 400 | 請求參數錯誤 |
| 401 | 401 | 未授權（令牌無效或過期） |
| 403 | 403 | 禁止訪問（權限不足） |
| 404 | 404 | 資源不存在 |
| 500 | 500 | 伺服器錯誤 |

### 1.3 認證機制

除了 `/auth/login` 端點外，所有 API 請求都需要在 Header 中帶上 JWT 令牌：

```
Authorization: Bearer {your-jwt-token}
```

---

## 2. 認證 API

### 2.1 使用者登入

**端點**: `POST /auth/login`

**描述**: 使用者登入並取得 JWT 令牌

**權限**: 公開

**請求範例**:
```json
{
  "username": "admin",
  "password": "password123"
}
```

**回應範例**:
```json
{
  "code": 200,
  "message": "登入成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com",
      "role": {
        "id": 1,
        "name": "admin",
        "permissions": [...]
      }
    }
  }
}
```

### 2.2 使用者註冊

**端點**: `POST /api/v1/auth/register`

**描述**: 註冊新使用者（需要 user.create 權限）

**權限**: `user.create`

**請求範例**:
```json
{
  "username": "newuser",
  "email": "newuser@example.com",
  "password": "securepass123",
  "role_id": 2
}
```

### 2.3 取得使用者資料

**端點**: `GET /api/v1/auth/profile`

**描述**: 取得當前登入使用者的個人資料

**權限**: 已認證

**回應範例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "username": "admin",
    "email": "admin@example.com",
    "role": {
      "id": 1,
      "name": "admin"
    },
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

### 2.4 刷新令牌

**端點**: `POST /api/v1/auth/refresh`

**描述**: 刷新 JWT 令牌

**權限**: 已認證

---

## 3. 監控目標 API

### 3.1 取得所有監控目標

**端點**: `GET /api/v1/Target/GetAll`

**描述**: 取得所有監控目標列表

**權限**: `target.read`

**回應範例**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "subject": "Production WebServer",
      "to": ["admin@example.com", "ops@example.com"],
      "enable": true,
      "indices": [
        {
          "id": 1,
          "pattern": "logstash-prod-*",
          "logname": "webserver",
          "device_group": "production",
          "period": "minutes",
          "unit": 5,
          "field": "host.keyword"
        }
      ],
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### 3.2 建立監控目標

**端點**: `POST /api/v1/Target/Create`

**描述**: 建立新的監控目標

**權限**: `target.create`

**請求範例**:
```json
{
  "subject": "Production WebServer",
  "to": ["admin@example.com"],
  "enable": true,
  "indices": [1, 2]
}
```

**說明**:
- `subject`: 警報郵件主旨
- `to`: 收件人郵箱列表
- `enable`: 是否啟用監控
- `indices`: 關聯的索引 ID 列表

### 3.3 更新監控目標

**端點**: `PUT /api/v1/Target/Update`

**描述**: 更新現有監控目標

**權限**: `target.update`

**請求範例**:
```json
{
  "id": 1,
  "subject": "Updated Subject",
  "to": ["newadmin@example.com"],
  "enable": false,
  "indices": [1, 3]
}
```

### 3.4 刪除監控目標

**端點**: `DELETE /api/v1/Target/Delete/:id`

**描述**: 刪除指定監控目標

**權限**: `target.delete`

**路徑參數**:
- `id`: 目標 ID

---

## 4. 裝置管理 API

### 4.1 取得所有裝置

**端點**: `GET /api/v1/Device/GetAll`

**描述**: 取得所有監控裝置列表

**權限**: `device.read`

**查詢參數**:
- `device_group`: (可選) 裝置群組名稱

**回應範例**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "device_group": "production",
      "name": "web-server-01",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### 4.2 建立裝置

**端點**: `POST /api/v1/Device/Create`

**描述**: 批次建立裝置

**權限**: `device.create`

**請求範例**:
```json
{
  "devices": [
    {
      "device_group": "production",
      "name": "web-server-01"
    },
    {
      "device_group": "production",
      "name": "web-server-02"
    }
  ]
}
```

### 4.3 取得裝置群組

**端點**: `GET /api/v1/Device/GetGroup`

**描述**: 取得所有裝置群組及每個群組的裝置數量

**權限**: `device.read`

**回應範例**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "device_group": "production",
      "count": 25
    },
    {
      "device_group": "staging",
      "count": 10
    }
  ]
}
```

---

## 5. 索引管理 API

### 5.1 取得所有索引

**端點**: `GET /api/v1/Indices/GetAll`

**描述**: 取得所有 Elasticsearch 索引配置

**權限**: `indices.read`

**回應範例**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "pattern": "logstash-prod-*",
      "device_group": "production",
      "logname": "webserver",
      "period": "minutes",
      "unit": 5,
      "field": "host.keyword",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### 5.2 建立索引

**端點**: `POST /api/v1/Indices/Create`

**描述**: 建立新的 ES 索引配置

**權限**: `indices.create`

**請求範例**:
```json
{
  "pattern": "logstash-prod-*",
  "device_group": "production",
  "logname": "webserver",
  "period": "minutes",
  "unit": 5,
  "field": "host.keyword"
}
```

**欄位說明**:
- `pattern`: Elasticsearch 索引模式（支援萬用字元）
- `device_group`: 裝置群組名稱
- `logname`: 日誌名稱（用於識別）
- `period`: 監控週期類型（"minutes" 或 "hours"）
- `unit`: 監控頻率（與 period 配合，如 5 分鐘、2 小時）
- `field`: 用於聚合的欄位（通常是裝置名稱欄位）

---

## 6. Elasticsearch 監控 API

### 6.1 取得所有 ES 監控器

**端點**: `GET /api/v1/elasticsearch/monitors`

**描述**: 取得所有 Elasticsearch 叢集監控器

**權限**: `elasticsearch.read`

**回應範例**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "Production ES Cluster",
      "host": "10.99.1.213",
      "port": 9200,
      "enable_monitor": true,
      "check_type": "health,performance",
      "interval": 60,
      "cpu_usage_high": 75.0,
      "cpu_usage_critical": 85.0,
      "memory_usage_high": 80.0,
      "memory_usage_critical": 90.0,
      "receivers": ["admin@example.com"]
    }
  ]
}
```

### 6.2 建立 ES 監控器

**端點**: `POST /api/v1/elasticsearch/monitors`

**描述**: 建立新的 ES 監控器

**權限**: `elasticsearch.create`

**請求範例**:
```json
{
  "name": "Production ES Cluster",
  "host": "10.99.1.213",
  "port": 9200,
  "username": "elastic",
  "password": "password",
  "enable_auth": true,
  "enable_monitor": true,
  "check_type": "health,performance",
  "interval": 60,
  "cpu_usage_high": 75.0,
  "cpu_usage_critical": 85.0,
  "memory_usage_high": 80.0,
  "memory_usage_critical": 90.0,
  "disk_usage_high": 85.0,
  "disk_usage_critical": 95.0,
  "receivers": ["admin@example.com"],
  "subject": "ES Cluster Alert"
}
```

### 6.3 測試 ES 連線

**端點**: `POST /api/v1/elasticsearch/monitors/:id/test`

**描述**: 測試 ES 監控器連線

**權限**: `elasticsearch.read`

**回應範例**:
```json
{
  "code": 200,
  "message": "連線成功",
  "data": {
    "cluster_name": "production-cluster",
    "cluster_status": "green",
    "response_time": 45
  }
}
```

### 6.4 啟用/停用監控器

**端點**: `POST /api/v1/elasticsearch/monitors/:id/toggle`

**描述**: 啟用或停用 ES 監控器

**權限**: `elasticsearch.update`

**請求範例**:
```json
{
  "enable": true
}
```

### 6.5 取得 ES 警報

**端點**: `GET /api/v1/elasticsearch/alerts`

**描述**: 取得 ES 監控警報列表

**權限**: `elasticsearch.read`

**查詢參數**:
- `monitor_id`: (可選) 監控器 ID
- `status`: (可選) 警報狀態（active, resolved, acknowledged）
- `severity`: (可選) 嚴重程度（low, medium, high, critical）
- `start_time`: (可選) 開始時間（ISO 8601 格式）
- `end_time`: (可選) 結束時間

**回應範例**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "monitor_id": 1,
      "alert_type": "performance",
      "severity": "high",
      "message": "CPU 使用率超過高閾值：78.5%",
      "status": "active",
      "cluster_name": "production-cluster",
      "threshold_value": 75.0,
      "actual_value": 78.5,
      "time": "2024-01-01T12:00:00Z"
    }
  ]
}
```

### 6.6 解決警報

**端點**: `POST /api/v1/elasticsearch/alerts/:monitor_id/resolve`

**描述**: 標記警報為已解決

**權限**: `elasticsearch.update`

**請求範例**:
```json
{
  "alert_id": 1,
  "resolution_note": "已手動擴展 ES 叢集容量"
}
```

---

## 7. 儀表板 API

### 7.1 系統概覽

**端點**: `GET /api/v1/dashboard/overview`

**描述**: 取得系統整體概覽資料

**權限**: 已認證

**回應範例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total_targets": 10,
    "active_targets": 8,
    "total_devices": 150,
    "online_devices": 145,
    "offline_devices": 5,
    "uptime_rate": 96.67,
    "active_alerts": 3,
    "last_update_time": "2024-01-01T12:00:00Z"
  }
}
```

### 7.2 取得統計資料

**端點**: `GET /api/v1/dashboard/statistics`

**描述**: 取得歷史統計資料

**權限**: 已認證

**查詢參數**:
- `logname`: (可選) 日誌名稱
- `device_group`: (可選) 裝置群組
- `start_date`: (必填) 開始日期（YYYY-MM-DD）
- `end_date`: (必填) 結束日期（YYYY-MM-DD）

**回應範例**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "date": "2024-01-01",
      "logname": "webserver",
      "device_group": "production",
      "total_checks": 288,
      "online_count": 280,
      "offline_count": 8,
      "uptime_rate": 97.22,
      "avg_response_time": 45
    }
  ]
}
```

### 7.3 取得趨勢資料

**端點**: `GET /api/v1/dashboard/trends`

**描述**: 取得監控趨勢資料

**權限**: 已認證

**查詢參數**: 同統計資料

### 7.4 取得裝置時間軸

**端點**: `GET /api/v1/dashboard/devices/:device_name/timeline`

**描述**: 取得特定裝置的狀態時間軸

**權限**: 已認證

**查詢參數**:
- `start_time`: 開始時間
- `end_time`: 結束時間

**回應範例**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "timestamp": "2024-01-01T12:00:00Z",
      "status": "online",
      "response_time": 45,
      "data_count": 1000
    },
    {
      "timestamp": "2024-01-01T12:05:00Z",
      "status": "offline",
      "response_time": 0,
      "data_count": 0,
      "error_msg": "連線逾時"
    }
  ]
}
```

---

## 8. 歷史資料 API

### 8.1 取得歷史資料

**端點**: `GET /api/v1/History/GetData/:logname`

**描述**: 依日誌名稱取得歷史監控資料

**權限**: 已認證

**路徑參數**:
- `logname`: 日誌名稱

**查詢參數**:
- `device_name`: (可選) 裝置名稱
- `start_time`: (可選) 開始時間
- `end_time`: (可選) 結束時間
- `limit`: (可選) 限制筆數，預設 100

---

## 9. 錯誤處理

### 9.1 錯誤回應格式

```json
{
  "code": 400,
  "message": "參數錯誤：缺少必填欄位 'username'",
  "data": null
}
```

### 9.2 常見錯誤碼

| code | 說明 | 解決方式 |
|------|------|---------|
| 401 | JWT 令牌無效或過期 | 重新登入取得新令牌 |
| 403 | 權限不足 | 聯繫管理員授予相應權限 |
| 404 | 資源不存在 | 檢查請求的 ID 是否正確 |
| 422 | 資料驗證失敗 | 檢查請求參數格式 |
| 500 | 伺服器內部錯誤 | 聯繫技術支援 |

---

## 10. API 使用範例

### 10.1 完整流程範例（使用 curl）

```bash
# 1. 登入取得令牌
TOKEN=$(curl -X POST http://localhost:8006/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  | jq -r '.data.token')

# 2. 建立監控索引
curl -X POST http://localhost:8006/api/v1/Indices/Create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "pattern": "logstash-prod-*",
    "device_group": "production",
    "logname": "webserver",
    "period": "minutes",
    "unit": 5,
    "field": "host.keyword"
  }'

# 3. 建立監控目標
curl -X POST http://localhost:8006/api/v1/Target/Create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "subject": "Production WebServer Alert",
    "to": ["admin@example.com"],
    "enable": true,
    "indices": [1]
  }'

# 4. 查看儀表板概覽
curl -X GET http://localhost:8006/api/v1/dashboard/overview \
  -H "Authorization: Bearer $TOKEN"
```

---

## 11. Swagger 文件

完整的互動式 API 文件可透過 Swagger UI 存取：

**URL**: `http://your-server:8006/swagger/index.html`

Swagger 文件包含：
- 所有端點的詳細說明
- 請求/回應範例
- 互動式測試介面
- 資料模型定義

---

## 12. API 版本控制

當前版本：**v1**

未來版本更新規則：
- **向下相容變更**：在 v1 中直接更新
- **破壞性變更**：發布新版本（v2, v3...）
- **舊版本支援期**：至少 6 個月

---

**文件結束**
