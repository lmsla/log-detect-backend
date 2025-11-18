# [功能模組名稱] API 規格

> **規格版本**: 1.0.0
> **最後更新**: YYYY-MM-DD
> **維護者**: [姓名/團隊]

## 📋 概述

簡要說明此 API 模組的用途、核心功能與設計理念。

**核心功能**:
- 功能點 1
- 功能點 2
- 功能點 3

**設計原則**:
- RESTful 設計
- 統一錯誤處理
- JWT 認證授權

---

## 🔐 認證與授權

### 認證方式
```
Authorization: Bearer <JWT_TOKEN>
```

### 所需權限
| 端點 | 權限 | 說明 |
|------|------|------|
| GET /resource | resource:read | 讀取權限 |
| POST /resource | resource:create | 創建權限 |
| PUT /resource | resource:update | 更新權限 |
| DELETE /resource | resource:delete | 刪除權限 |

---

## 📡 API 端點清單

### 1️⃣ 獲取資源列表

**端點**: `GET /api/v1/[module]/[resource]`

**描述**: 獲取所有 [資源] 的列表，支援分頁與過濾

**請求參數**:

| 參數 | 類型 | 必填 | 說明 | 範例 |
|------|------|------|------|------|
| page | integer | 否 | 頁碼（預設: 1） | 1 |
| page_size | integer | 否 | 每頁筆數（預設: 20，最大: 100） | 20 |
| sort_by | string | 否 | 排序欄位 | created_at |
| order | string | 否 | 排序方向（asc/desc，預設: desc） | desc |
| filter | string | 否 | 過濾條件（JSON 格式） | {"status":"active"} |

**請求範例**:
```bash
curl -X GET "http://localhost:8006/api/v1/module/resource?page=1&page_size=20" \
  -H "Authorization: Bearer eyJhbGc..."
```

**成功回應** (200 OK):
```json
{
  "success": true,
  "msg": "查詢成功",
  "body": {
    "items": [
      {
        "id": 1,
        "name": "資源名稱",
        "status": "active",
        "created_at": "2025-10-08T10:00:00Z",
        "updated_at": "2025-10-08T10:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 100,
      "total_pages": 5
    }
  }
}
```

**錯誤回應**:
- `401 Unauthorized` - Token 無效或過期
- `403 Forbidden` - 權限不足
- `500 Internal Server Error` - 伺服器錯誤

---

### 2️⃣ 獲取單個資源

**端點**: `GET /api/v1/[module]/[resource]/{id}`

**描述**: 根據 ID 獲取特定 [資源] 的詳細資訊

**路徑參數**:

| 參數 | 類型 | 必填 | 說明 |
|------|------|------|------|
| id | integer | 是 | 資源 ID |

**請求範例**:
```bash
curl -X GET "http://localhost:8006/api/v1/module/resource/1" \
  -H "Authorization: Bearer eyJhbGc..."
```

**成功回應** (200 OK):
```json
{
  "success": true,
  "msg": "查詢成功",
  "body": {
    "id": 1,
    "name": "資源名稱",
    "description": "資源描述",
    "status": "active",
    "metadata": {
      "key": "value"
    },
    "created_at": "2025-10-08T10:00:00Z",
    "updated_at": "2025-10-08T10:00:00Z"
  }
}
```

**錯誤回應**:
- `404 Not Found` - 資源不存在
- `401 Unauthorized` - 未授權
- `403 Forbidden` - 權限不足

---

### 3️⃣ 創建資源

**端點**: `POST /api/v1/[module]/[resource]`

**描述**: 創建新的 [資源]

**請求 Body**:
```json
{
  "name": "資源名稱",
  "description": "資源描述",
  "status": "active",
  "config": {
    "option1": "value1",
    "option2": 100
  }
}
```

**欄位說明**:

| 欄位 | 類型 | 必填 | 說明 | 驗證規則 |
|------|------|------|------|---------|
| name | string | 是 | 資源名稱 | 1-100 字元 |
| description | string | 否 | 資源描述 | 最多 500 字元 |
| status | string | 否 | 狀態（預設: active） | active/inactive |
| config | object | 否 | 配置選項 | JSON 物件 |

**請求範例**:
```bash
curl -X POST "http://localhost:8006/api/v1/module/resource" \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{
    "name": "新資源",
    "description": "這是一個新資源",
    "status": "active"
  }'
```

**成功回應** (201 Created):
```json
{
  "success": true,
  "msg": "創建成功",
  "body": {
    "id": 123,
    "name": "新資源",
    "description": "這是一個新資源",
    "status": "active",
    "created_at": "2025-10-08T10:00:00Z",
    "updated_at": "2025-10-08T10:00:00Z"
  }
}
```

**錯誤回應**:
- `400 Bad Request` - 請求參數錯誤
  ```json
  {
    "success": false,
    "msg": "名稱不能為空"
  }
  ```
- `401 Unauthorized` - 未授權
- `403 Forbidden` - 權限不足
- `409 Conflict` - 資源已存在

---

### 4️⃣ 更新資源

**端點**: `PUT /api/v1/[module]/[resource]/{id}`

**描述**: 更新現有 [資源] 的資訊

**路徑參數**:

| 參數 | 類型 | 必填 | 說明 |
|------|------|------|------|
| id | integer | 是 | 資源 ID |

**請求 Body**:
```json
{
  "name": "更新後的名稱",
  "description": "更新後的描述",
  "status": "inactive"
}
```

**請求範例**:
```bash
curl -X PUT "http://localhost:8006/api/v1/module/resource/123" \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{
    "name": "更新後的名稱",
    "status": "inactive"
  }'
```

**成功回應** (200 OK):
```json
{
  "success": true,
  "msg": "更新成功",
  "body": {
    "id": 123,
    "name": "更新後的名稱",
    "status": "inactive",
    "updated_at": "2025-10-08T11:00:00Z"
  }
}
```

**錯誤回應**:
- `404 Not Found` - 資源不存在
- `400 Bad Request` - 請求參數錯誤
- `401 Unauthorized` - 未授權
- `403 Forbidden` - 權限不足

---

### 5️⃣ 刪除資源

**端點**: `DELETE /api/v1/[module]/[resource]/{id}`

**描述**: 刪除指定的 [資源]

**路徑參數**:

| 參數 | 類型 | 必填 | 說明 |
|------|------|------|------|
| id | integer | 是 | 資源 ID |

**請求範例**:
```bash
curl -X DELETE "http://localhost:8006/api/v1/module/resource/123" \
  -H "Authorization: Bearer eyJhbGc..."
```

**成功回應** (200 OK):
```json
{
  "success": true,
  "msg": "刪除成功"
}
```

**錯誤回應**:
- `404 Not Found` - 資源不存在
- `401 Unauthorized` - 未授權
- `403 Forbidden` - 權限不足
- `409 Conflict` - 資源正在使用中，無法刪除

---

## 📊 資料模型

### Resource 物件

```json
{
  "id": 1,                           // 唯一識別碼
  "name": "string",                  // 資源名稱（1-100 字元）
  "description": "string",           // 資源描述（可選，最多 500 字元）
  "status": "active|inactive",       // 狀態（預設: active）
  "config": {                        // 配置選項（JSON 物件）
    "option1": "value1",
    "option2": 100
  },
  "metadata": {},                    // 額外元數據（JSONB）
  "created_at": "2025-10-08T10:00:00Z",  // ISO 8601 格式
  "updated_at": "2025-10-08T10:00:00Z"   // ISO 8601 格式
}
```

### 欄位詳細說明

| 欄位 | 類型 | 必填 | 說明 | 約束 |
|------|------|------|------|------|
| id | integer | 是 | 主鍵 | 自動遞增 |
| name | string | 是 | 資源名稱 | 1-100 字元，唯一 |
| description | string | 否 | 資源描述 | 最多 500 字元 |
| status | string | 是 | 狀態 | active, inactive |
| config | object | 否 | 配置選項 | JSON 格式 |
| metadata | object | 否 | 額外元數據 | JSONB 格式 |
| created_at | timestamp | 是 | 創建時間 | ISO 8601 |
| updated_at | timestamp | 是 | 更新時間 | ISO 8601 |

---

## ⚠️ 錯誤處理

### 統一錯誤格式

```json
{
  "success": false,
  "msg": "錯誤訊息描述",
  "error_code": "ERROR_CODE",      // 可選
  "details": {}                     // 可選，詳細錯誤資訊
}
```

### 常見錯誤碼

| HTTP Status | error_code | 說明 |
|-------------|------------|------|
| 400 | INVALID_REQUEST | 請求參數錯誤 |
| 401 | UNAUTHORIZED | Token 無效或過期 |
| 403 | FORBIDDEN | 權限不足 |
| 404 | NOT_FOUND | 資源不存在 |
| 409 | CONFLICT | 資源衝突（如重複） |
| 422 | VALIDATION_ERROR | 資料驗證失敗 |
| 500 | INTERNAL_ERROR | 伺服器內部錯誤 |

---

## 💡 使用注意事項

1. **認證**: 所有 API 都需要有效的 JWT Token
2. **分頁**: 列表查詢建議使用分頁，避免一次查詢過多資料
3. **時間格式**: 所有時間欄位使用 ISO 8601 格式（UTC）
4. **JSON 格式**: Content-Type 必須為 `application/json`
5. **冪等性**: PUT 和 DELETE 操作是冪等的
6. **速率限制**: API 有速率限制（每分鐘 60 次）

---

## 🔗 相關資源

- [OpenAPI 完整規格](./openapi.yml)
- [前端對接指南](../../guides/frontend/api-integration.md)
- [故障排除](../../troubleshooting/api/)

---

**變更歷史**:
- v1.0.0 (YYYY-MM-DD) - 初始版本
