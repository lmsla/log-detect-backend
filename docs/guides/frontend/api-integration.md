# Elasticsearch 監控 - 前端 API 對接文檔

## 📋 概述

本文檔為前端開發提供完整的 Elasticsearch 監控 API 對接規範。

- **Base URL**: `http://localhost:8006/api/v1/elasticsearch`
- **認證方式**: JWT Bearer Token
- **Content-Type**: `application/json`

## 🔐 認證

所有 API 都需要在 Header 中攜帶 JWT Token：

```javascript
headers: {
  'Authorization': `Bearer ${token}`,
  'Content-Type': 'application/json'
}
```

## ⚠️ 重要說明

### 資料類型和單位

| 欄位 | 類型 | 單位/格式 | 說明 |
|------|------|-----------|------|
| `response_time` | integer | 毫秒 | 響應時間 |
| `cpu_usage` | float | 百分比 (0-100) | CPU 使用率 |
| `memory_usage` | float | 百分比 (0-100) | 記憶體使用率 |
| `disk_usage` | float | 百分比 (0-100) | 磁碟使用率 |
| `interval` | integer | 秒 (10-3600) | 檢查間隔 |
| `last_check_time` | string | ISO 8601 | 時間格式：2024-01-01T12:00:00Z |
| `receivers` | string | JSON 陣列字串 | 格式：'["admin@example.com"]' |

### receivers 欄位處理範例

**發送請求時**（需要序列化）:
```javascript
const receivers = ["admin@example.com", "ops@example.com"];
const body = {
  name: "My Monitor",
  receivers: JSON.stringify(receivers) // 轉成字串
};
```

**接收響應時**（需要反序列化）:
```javascript
const monitor = response.body;
const receivers = JSON.parse(monitor.receivers || '[]'); // 解析成陣列
```

### 時間欄位處理範例

```javascript
// 顯示時間
const lastCheck = new Date(monitor.last_check_time);
console.log(lastCheck.toLocaleString()); // 本地時間格式

// 計算時間差
const now = new Date();
const diff = now - lastCheck;
const minutesAgo = Math.floor(diff / 1000 / 60);
console.log(`${minutesAgo} 分鐘前`);
```

### 百分比欄位顯示範例

```javascript
// CPU 使用率顯示（帶顏色）
const cpuColor = monitor.cpu_usage > 80 ? 'red' :
                 monitor.cpu_usage > 60 ? 'orange' : 'green';

return (
  <Progress
    percent={monitor.cpu_usage}
    status={cpuColor}
    format={percent => `${percent.toFixed(1)}%`}
  />
);
```

## ✅ Phase 1 可用 API (9 個)

### 1. 獲取所有監控配置

```http
GET /api/v1/elasticsearch/monitors
```

**請求示例**:
```javascript
const response = await fetch('http://localhost:8006/api/v1/elasticsearch/monitors', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});
```

**響應示例**:
```json
{
  "success": true,
  "msg": "查詢監控配置成功",
  "body": [
    {
      "id": 1,
      "name": "Production ES",
      "host": "es.example.com",
      "port": 9200,
      "enable_auth": true,
      "username": "monitor",
      "password": "******",
      "check_type": "health,performance",
      "interval": 60,
      "enable_monitor": true,
      "receivers": "[\"admin@example.com\"]",
      "subject": "ES Alert",
      "description": "Production cluster",
      "alert_threshold": "{\"cpu_usage_high\":75.0}",
      "created_at": 1696147200,
      "updated_at": 1696147200
    }
  ]
}
```

### 2. 創建監控配置

```http
POST /api/v1/elasticsearch/monitors
```

**請求 Body**:
```json
{
  "name": "Production ES",
  "host": "es.example.com",
  "port": 9200,
  "enable_auth": true,
  "username": "monitor",
  "password": "secret123",
  "check_type": "health,performance",
  "interval": 60,
  "enable_monitor": true,
  "receivers": "[\"admin@example.com\",\"ops@example.com\"]",
  "subject": "ES Cluster Alert - Production",
  "description": "Production Elasticsearch cluster monitoring"
}
```

**請求示例**:
```javascript
const response = await fetch('http://localhost:8006/api/v1/elasticsearch/monitors', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    name: "Production ES",
    host: "es.example.com",
    port: 9200,
    enable_auth: true,
    username: "monitor",
    password: "secret123",
    interval: 60,
    check_type: "health,performance",
    enable_monitor: true,
    receivers: JSON.stringify(["admin@example.com"]),
    subject: "ES Alert"
  })
});
```

**響應示例**:
```json
{
  "success": true,
  "msg": "創建監控配置成功",
  "body": {
    "id": 1,
    "name": "Production ES",
    "host": "es.example.com",
    "port": 9200,
    "enable_auth": true,
    "username": "monitor",
    "password": "secret123",
    "check_type": "health,performance",
    "interval": 60,
    "enable_monitor": true,
    "receivers": "[\"admin@example.com\"]",
    "subject": "ES Alert",
    "created_at": 1696147200,
    "updated_at": 1696147200
  }
}
```

### 3. 更新監控配置

```http
PUT /api/v1/elasticsearch/monitors
```

**注意**: ID 從 request body 傳遞，不在 URL 中

**請求 Body**:
```json
{
  "id": 1,
  "name": "Production ES Updated",
  "host": "es.example.com",
  "port": 9200,
  "interval": 120
}
```

**請求示例**:
```javascript
const response = await fetch('http://localhost:8006/api/v1/elasticsearch/monitors', {
  method: 'PUT',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    id: 1,
    name: "Production ES Updated",
    host: "es.example.com",
    port: 9200,
    interval: 120
  })
});
```

**響應示例**:
```json
{
  "success": true,
  "msg": "更新監控配置成功",
  "body": {
    "id": 1,
    "name": "Production ES Updated",
    "host": "es.example.com",
    "port": 9200,
    "interval": 120,
    "updated_at": 1696150800
  }
}
```

### 4. 獲取特定監控配置

```http
GET /api/v1/elasticsearch/monitors/{id}
```

**請求示例**:
```javascript
const response = await fetch('http://localhost:8006/api/v1/elasticsearch/monitors/1', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});
```

**響應示例**:
```json
{
  "success": true,
  "msg": "查詢監控配置成功",
  "body": {
    "id": 1,
    "name": "Production ES",
    "host": "es.example.com",
    "port": 9200,
    "enable_auth": true,
    "username": "monitor",
    "check_type": "health,performance",
    "interval": 60,
    "enable_monitor": true
  }
}
```

### 5. 刪除監控配置

```http
DELETE /api/v1/elasticsearch/monitors/{id}
```

**請求示例**:
```javascript
const response = await fetch('http://localhost:8006/api/v1/elasticsearch/monitors/1', {
  method: 'DELETE',
  headers: {
    'Authorization': `Bearer ${token}`
  }
});
```

**響應示例**:
```json
{
  "success": true,
  "msg": "刪除監控配置成功"
}
```

### 6. 測試 ES 連接

```http
POST /api/v1/elasticsearch/monitors/test
```

**注意**: 不需要已存在的監控 ID，可直接測試連接參數

**請求 Body**:
```json
{
  "host": "es.example.com",
  "port": 9200,
  "enable_auth": true,
  "username": "monitor",
  "password": "secret123"
}
```

**請求示例**:
```javascript
const response = await fetch('http://localhost:8006/api/v1/elasticsearch/monitors/test', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    host: "es.example.com",
    port: 9200,
    enable_auth: true,
    username: "monitor",
    password: "secret123"
  })
});
```

**響應示例（成功）**:
```json
{
  "success": true,
  "msg": "連接成功",
  "body": {
    "cluster_name": "production-es",
    "cluster_status": "green",
    "status": "online",
    "response_time": 45
  }
}
```

**響應示例（失敗）**:
```json
{
  "success": false,
  "msg": "連接失敗: connection refused",
  "body": {
    "success": false,
    "status": "offline",
    "error_message": "dial tcp: connection refused",
    "response_time": 0
  }
}
```

### 7. 啟用/停用監控

```http
POST /api/v1/elasticsearch/monitors/{id}/toggle
```

**請求 Body**:
```json
{
  "enable": true
}
```

**請求示例**:
```javascript
const response = await fetch('http://localhost:8006/api/v1/elasticsearch/monitors/1/toggle', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    enable: true
  })
});
```

**響應示例**:
```json
{
  "success": true,
  "msg": "監控已啟用"
}
```

### 8. 獲取所有監控器狀態

```http
GET /api/v1/elasticsearch/status
```

**請求示例**:
```javascript
const response = await fetch('http://localhost:8006/api/v1/elasticsearch/status', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});
```

**響應示例**:
```json
{
  "success": true,
  "msg": "查詢成功",
  "body": [
    {
      "monitor_id": 1,
      "monitor_name": "Production ES",
      "host": "es.example.com:9200",
      "status": "online",
      "cluster_status": "green",
      "cluster_name": "production-cluster",
      "response_time": 45,
      "cpu_usage": 35.5,
      "memory_usage": 72.3,
      "disk_usage": 65.8,
      "node_count": 3,
      "active_shards": 120,
      "unassigned_shards": 0,
      "last_check_time": "2024-01-01T12:00:00Z",
      "error_message": "",
      "warning_message": ""
    },
    {
      "monitor_id": 2,
      "monitor_name": "Dev ES",
      "host": "es-dev.example.com:9200",
      "status": "offline",
      "error_message": "Connection timeout"
    }
  ]
}
```

### 9. 獲取統計數據

```http
GET /api/v1/elasticsearch/statistics
```

**請求示例**:
```javascript
const response = await fetch('http://localhost:8006/api/v1/elasticsearch/statistics', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});
```

**響應示例**:
```json
{
  "success": true,
  "msg": "查詢成功",
  "body": {
    "total_monitors": 5,
    "online_monitors": 4,
    "offline_monitors": 1,
    "warning_monitors": 0,
    "total_nodes": 15,
    "total_indices": 250,
    "total_documents": 1000000,
    "total_size_gb": 125.5,
    "avg_response_time": 52.3,
    "avg_cpu_usage": 45.2,
    "avg_memory_usage": 68.7,
    "active_alerts": 2,
    "last_update_time": "2024-01-01 12:00:00"
  }
}
```

## 🎨 前端頁面建議

### 1. 監控配置列表頁面

**建議功能**:
- 表格顯示所有監控配置（API: GET /monitors）
- 新增按鈕（彈窗表單 → POST /monitors）
- 編輯按鈕（彈窗表單 → PUT /monitors）
- 刪除按鈕（確認對話框 → DELETE /monitors/{id}）
- 啟用/停用開關（→ POST /monitors/{id}/toggle）
- 測試連接按鈕（→ POST /monitors/test）

**表格欄位建議**:
- ID
- 名稱
- 主機地址
- 狀態（線上/離線）
- 檢查間隔
- 啟用狀態（開關）
- 操作（編輯/刪除/測試）

### 2. 監控狀態總覽頁面

**建議功能**:
- 統計卡片（API: GET /statistics）
  - 監控器總數
  - 線上/離線數量
  - 告警數量
  - 平均響應時間
- 狀態列表（API: GET /status）
  - 每個監控器的即時狀態卡片
  - 顏色標示（綠色=online, 紅色=offline, 黃色=warning）
  - CPU/Memory/Disk 使用率進度條
  - 集群健康狀態（green/yellow/red）

### 3. 表單欄位說明

**創建/編輯監控配置表單**:

| 欄位名 | 類型 | 必填 | 說明 | 預設值 |
|--------|------|------|------|--------|
| name | string | 是 | 監控名稱 | - |
| host | string | 是 | ES 主機地址 | - |
| port | number | 否 | ES 端口 | 9200 |
| enable_auth | boolean | 否 | 啟用認證 | false |
| username | string | 條件 | 用戶名（enable_auth=true 時必填） | - |
| password | string | 條件 | 密碼（enable_auth=true 時必填） | - |
| check_type | string | 否 | 檢查類型 | "health,performance" |
| interval | number | 否 | 檢查間隔（秒） | 60 |
| enable_monitor | boolean | 否 | 啟用監控 | true |
| receivers | string | 否 | 告警收件人（JSON 陣列字串） | "[]" |
| subject | string | 否 | 告警主題 | - |
| description | string | 否 | 描述 | - |

**欄位驗證規則**:
```javascript
const validationRules = {
  name: {
    required: true,
    minLength: 1,
    maxLength: 100
  },
  host: {
    required: true,
    pattern: /^[a-zA-Z0-9.-]+$/
  },
  port: {
    min: 1,
    max: 65535
  },
  interval: {
    min: 10,
    max: 3600
  },
  receivers: {
    validate: (value) => {
      try {
        const emails = JSON.parse(value);
        return Array.isArray(emails) && emails.every(e => /\S+@\S+\.\S+/.test(e));
      } catch {
        return false;
      }
    }
  }
};
```

## 🔄 輪詢更新建議

對於狀態頁面，建議使用輪詢定期更新：

```javascript
// 每 30 秒更新一次狀態
const updateInterval = 30000;

const fetchStatus = async () => {
  const response = await fetch('/api/v1/elasticsearch/status', {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  const data = await response.json();
  updateUI(data.body);
};

// 初始加載
fetchStatus();

// 定時更新
const intervalId = setInterval(fetchStatus, updateInterval);

// 組件卸載時清除
onUnmount(() => clearInterval(intervalId));
```

## ⚠️ 錯誤處理

### 常見錯誤碼

| 狀態碼 | 說明 | 處理建議 |
|--------|------|----------|
| 400 | 請求參數錯誤 | 檢查表單驗證，顯示錯誤訊息 |
| 401 | 未授權 | Token 過期，跳轉登入頁 |
| 403 | 權限不足 | 顯示權限錯誤訊息 |
| 404 | 資源不存在 | 刷新列表，移除不存在的項目 |
| 500 | 伺服器錯誤 | 顯示通用錯誤訊息，建議重試 |

### 錯誤處理範例

```javascript
try {
  const response = await fetch('/api/v1/elasticsearch/monitors', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(monitorConfig)
  });

  if (!response.ok) {
    if (response.status === 401) {
      // Token 過期
      router.push('/login');
      return;
    }

    const error = await response.json();
    throw new Error(error.msg || '操作失敗');
  }

  const data = await response.json();

  if (!data.success) {
    throw new Error(data.msg || '操作失敗');
  }

  // 成功處理
  showSuccessMessage(data.msg);
  refreshList();

} catch (error) {
  console.error('Error:', error);
  showErrorMessage(error.message);
}
```

## 📚 相關文檔

- **Swagger UI**: `http://localhost:8006/swagger/index.html`
- **OpenAPI 規範**: `docs/openapi.yml`
- **API 實作狀態**: `docs/elasticsearch-api-status.md`
- **完整文檔**: `docs/elasticsearch-monitoring.md`

## 🚀 快速測試

### 使用 curl 測試

```bash
# 1. 登入獲取 token
TOKEN=$(curl -X POST http://localhost:8006/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  | jq -r '.access_token')

# 2. 獲取所有監控配置
curl http://localhost:8006/api/v1/elasticsearch/monitors \
  -H "Authorization: Bearer $TOKEN"

# 3. 創建監控配置
curl -X POST http://localhost:8006/api/v1/elasticsearch/monitors \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test ES",
    "host": "localhost",
    "port": 9200,
    "interval": 60
  }'
```

---

**文檔版本**: 1.0
**最後更新**: 2025-10-06
**維護者**: Log Detect 開發團隊
