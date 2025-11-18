# 🔐 Log Detect 身份驗證系統

## 📋 概述

Log Detect 現在包含完整的 JWT 身份驗證和角色-based 訪問控制 (RBAC) 系統。

## 🚀 快速開始

### 1. 數據庫設置

首次運行時，系統會自動創建所需的所有表和默認數據。如果需要手動創建，可以運行：

```bash
go run create_tables.go
```

### 2. 默認用戶

系統會在首次啟動時自動創建以下默認用戶：

- **管理員用戶**
  - 用戶名: `admin`
  - 密碼: `admin123`
  - 郵箱: `admin@logdetect.com`
  - 角色: 管理員 (所有權限)

### 2. 登錄

```bash
curl -X POST http://localhost:8006/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

響應示例：
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "admin",
    "email": "admin@logdetect.com",
    "role": {
      "id": 1,
      "name": "admin",
      "description": "Administrator with full access"
    }
  }
}
```

## 🔑 API 認證

### 使用 JWT Token

在請求頭中包含 Authorization:

```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     http://localhost:8006/api/v1/Device/GetAll
```

## 👥 角色和權限

### 內建角色

1. **admin** - 管理員
   - 擁有所有權限
   - 可以管理用戶、角色和權限

2. **user** - 普通用戶
   - 只能讀取數據
   - 無法創建、修改或刪除資源

### 權限系統

權限格式: `{resource}:{action}`

| 資源 | 操作 | 說明 |
|-----|-----|-----|
| device | create, read, update, delete | 設備管理 |
| target | create, read, update, delete | 目標管理 |
| indices | create, read, update, delete | 索引管理 |
| user | create, read, update, delete | 用戶管理 |

## 📚 API 端點

### 公開端點 (無需認證)

- `POST /auth/login` - 用戶登錄
- `GET /healthcheck` - 健康檢查
- `GET /swagger/*` - API 文檔

### 認證端點

- `POST /api/v1/auth/register` - 註冊新用戶 (需要權限)
- `GET /api/v1/auth/profile` - 獲取個人資料
- `POST /api/v1/auth/refresh` - 刷新token
- `GET /api/v1/auth/users` - 列出所有用戶 (管理員)
- `GET /api/v1/auth/users/{id}` - 獲取用戶詳情
- `PUT /api/v1/auth/users/{id}` - 更新用戶
- `DELETE /api/v1/auth/users/{id}` - 刪除用戶

### 受保護的業務端點

所有原有的 API 端點現在都需要認證，並根據用戶權限進行授權：

- `/api/v1/Device/*` - 設備相關操作
- `/api/v1/Target/*` - 目標相關操作
- `/api/v1/Indices/*` - 索引相關操作
- `/api/v1/Receiver/*` - 接收者相關操作
- `/api/v1/History/*` - 歷史記錄相關操作

## 🔧 配置

### JWT 配置

在 `services/auth.go` 中：

```go
const (
	JWTSecretKey = "your-secret-key-change-in-production" // TODO: 移至環境變數
	JWTExpireHours = 24
)
```

⚠️ **重要**: 在生產環境中，請將 `JWTSecretKey` 移至環境變數！

### 環境變數

添加以下環境變數：

```bash
export JWT_SECRET="your-super-secret-key-here"
```

然後修改代碼從環境變數讀取密鑰。

## 🧪 測試認證

### 1. 測試登錄

```bash
# 登錄獲取token
TOKEN=$(curl -s -X POST http://localhost:8006/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.token')

echo "JWT Token: $TOKEN"
```

### 2. 測試受保護的端點

```bash
# 使用token訪問受保護的端點
curl -H "Authorization: Bearer $TOKEN" \
     http://localhost:8006/api/v1/Device/GetAll
```

### 3. 測試權限不足

```bash
# 創建一個普通用戶token (需要先創建普通用戶)
curl -H "Authorization: Bearer $USER_TOKEN" \
     -X POST http://localhost:8006/api/v1/Device/Create \
     -H "Content-Type: application/json" \
     -d '{"name":"test-device","device_group":"test"}'
# 應該返回 403 Forbidden
```

## 🔒 安全注意事項

1. **密碼存儲**: 使用 bcrypt 加密
2. **JWT 過期**: Token 24小時後過期
3. **權限檢查**: 每個請求都進行權限驗證
4. **SQL 注入防護**: 使用參數化查詢
5. **輸入驗證**: 所有輸入都進行驗證

## 🚨 遷移指南

如果您有現有的系統，需要遷移到新的認證系統：

1. **數據庫遷移**: 運行 `CreateTable()` 來創建新的認證表
2. **用戶遷移**: 手動創建用戶或提供遷移腳本
3. **API 更新**: 更新客戶端代碼以包含認證頭
4. **測試**: 徹底測試所有受保護的端點

## 📞 支持

如果您在使用認證系統時遇到問題，請檢查：

1. JWT token 是否正確
2. 用戶權限是否正確設置
3. API 端點是否正確保護
4. 日誌中的錯誤信息

---

**最後更新**: 2025年9月23日
