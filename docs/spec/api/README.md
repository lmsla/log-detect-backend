# 📋 Log Detect API - OpenAPI 規範

## 📖 概述

此 `openapi.yml` 文件包含完整的 Log Detect API 規範，採用 OpenAPI 3.0.3 標準格式，包含新實現的 JWT 身份驗證和角色-based 訪問控制 (RBAC) 系統。

## 🎯 主要功能

### 🔐 身份驗證系統
- **JWT Bearer Token** 認證
- **角色-based 權限控制**
- **安全的密碼處理**

### 📊 API 端點分類
- **認證端點**: 登錄、註冊、用戶管理
- **設備管理**: CRUD 操作和統計
- **目標管理**: 監控目標配置
- **索引管理**: Elasticsearch 索引配置
- **歷史記錄**: 日誌歷史查詢
- **接收者管理**: 郵件接收者配置
- **Elasticsearch 監控**: ES 集群監控配置和狀態查詢

## 🚀 如何使用

### 1. 在 Swagger UI 中查看
```bash
# 啟動服務後訪問
open http://localhost:8006/swagger/index.html
```

### 2. 使用 API 測試工具
- **Postman**: 導入 `openapi.yml`
- **Insomnia**: 導入 `openapi.yml`
- **Swagger Editor**: 在線編輯和測試

### 3. 代碼生成
```bash
# 生成 TypeScript 客戶端
npx openapi-typescript openapi.yml --output types.ts

# 生成 Python 客戶端
openapi-python-client generate --url openapi.yml

# 生成 Go 客戶端
go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --package=client openapi.yml > client.go
```

## 🔑 認證流程

### 獲取 Access Token
```bash
curl -X POST http://localhost:8006/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

### 使用 Token 訪問受保護端點
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     http://localhost:8006/api/v1/Device/GetAll
```

## 👥 角色和權限

### 內建角色
- **admin**: 完全訪問權限
- **user**: 讀取權限

### 權限格式
`{resource}:{action}`
- `device:create`, `device:read`, `device:update`, `device:delete`
- `target:create`, `target:read`, `target:update`, `target:delete`
- `indices:create`, `indices:read`, `indices:update`, `indices:delete`
- `user:create`, `user:read`, `user:update`, `user:delete`
- `elasticsearch:create`, `elasticsearch:read`, `elasticsearch:update`, `elasticsearch:delete`

## 📋 API 端點總覽

### 認證端點 (公開)
- `POST /auth/login` - 用戶登錄
- `GET /healthcheck` - 健康檢查

### 用戶管理 (需要認證)
- `POST /api/v1/auth/register` - 註冊用戶
- `GET /api/v1/auth/profile` - 獲取個人資料
- `POST /api/v1/auth/refresh` - 刷新token
- `GET /api/v1/auth/users` - 列出用戶
- `GET /api/v1/auth/users/{id}` - 獲取用戶詳情
- `PUT /api/v1/auth/users/{id}` - 更新用戶
- `DELETE /api/v1/auth/users/{id}` - 刪除用戶

### 設備管理 (需要認證)
- `GET /api/v1/Device/GetAll` - 獲取所有設備
- `POST /api/v1/Device/Create` - 創建設備
- `PUT /api/v1/Device/Update` - 更新設備
- `DELETE /api/v1/Device/Delete/{id}` - 刪除設備
- `GET /api/v1/Device/count` - 設備統計
- `GET /api/v1/Device/GetGroup` - 設備分組

### 目標管理 (需要認證)
- `GET /api/v1/Target/GetAll` - 獲取所有目標
- `POST /api/v1/Target/Create` - 創建目標
- `PUT /api/v1/Target/Update` - 更新目標
- `DELETE /api/v1/Target/Delete/{id}` - 刪除目標

### 索引管理 (需要認證)
- `GET /api/v1/Indices/GetAll` - 獲取所有索引
- `POST /api/v1/Indices/Create` - 創建索引
- `PUT /api/v1/Indices/Update` - 更新索引
- `DELETE /api/v1/Indices/Delete/{id}` - 刪除索引
- `GET /api/v1/Indices/GetIndicesByLogname/{logname}` - 按日誌名獲取索引
- `GET /api/v1/Indices/GetIndicesByTargetID/{id}` - 按目標ID獲取索引
- `GET /api/v1/Indices/GetLogname` - 獲取日誌名稱列表

### 歷史記錄 (需要認證)
- `GET /api/v1/History/GetData/{logname}` - 獲取歷史數據
- `GET /api/v1/History/GetLognameData` - 獲取日誌名稱數據

### 接收者管理 (需要認證)
- `GET /api/v1/Receiver/GetAll` - 獲取所有接收者
- `POST /api/v1/Receiver/Create` - 創建接收者
- `PUT /api/v1/Receiver/Update` - 更新接收者
- `DELETE /api/v1/Receiver/Delete/{id}` - 刪除接收者

### Elasticsearch 監控 (需要認證)
- `GET /api/v1/elasticsearch/monitors` - 獲取所有 ES 監控配置
- `POST /api/v1/elasticsearch/monitors` - 創建 ES 監控配置
- `PUT /api/v1/elasticsearch/monitors` - 更新 ES 監控配置
- `GET /api/v1/elasticsearch/monitors/{id}` - 獲取特定 ES 監控配置
- `DELETE /api/v1/elasticsearch/monitors/{id}` - 刪除 ES 監控配置
- `POST /api/v1/elasticsearch/monitors/test` - 測試 ES 連接
- `POST /api/v1/elasticsearch/monitors/{id}/toggle` - 啟用/停用監控
- `GET /api/v1/elasticsearch/status` - 獲取所有監控器狀態
- `GET /api/v1/elasticsearch/statistics` - 獲取 ES 監控統計數據

## 📊 數據模型

### 核心實體
- **User**: 用戶信息
- **Role**: 角色定義
- **Permission**: 權限定義
- **Device**: 設備信息
- **Target**: 監控目標
- **Index**: Elasticsearch 索引配置
- **Receiver**: 郵件接收者
- **History**: 歷史記錄
- **ElasticsearchMonitor**: ES 監控配置
- **ESMonitorStatus**: ES 監控狀態
- **ESStatistics**: ES 監控統計

### 認證相關
- **LoginRequest**: 登錄請求
- **LoginResponse**: 登錄響應
- **ErrorResponse**: 錯誤響應
- **SuccessResponse**: 成功響應

## 🛠️ 開發工具支持

### API 客戶端生成
```bash
# TypeScript
npm install -g openapi-typescript
openapi-typescript openapi.yml -o api-types.ts

# Python
pip install openapi-python-client
openapi-python-client generate --url openapi.yml --output python-client/

# Go
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest
oapi-codegen -package client openapi.yml > client.go

# JavaScript/Node.js
npm install -g swagger-js-codegen
swagger-js-codegen -l javascript -i openapi.yml -o js-client.js
```

### API 測試工具
```bash
# Newman (Postman collection runner)
npm install -g newman
newman run collection.json

# REST Client (VS Code extension)
# 直接在 .http 文件中使用
```

### 文檔生成
```bash
# 生成 HTML 文檔
npm install -g redoc-cli
redoc-cli bundle openapi.yml -o api-docs.html

# 生成 PDF
npm install -g openapi-to-postman
openapi-to-postman convert -i openapi.yml -o collection.json
```

## 🔍 驗證規範

### 使用官方驗證器
```bash
# 在線驗證
curl -X POST "https://validator.swagger.io/validator/debug" \
  -H "accept: application/json" \
  -H "Content-Type: application/yaml" \
  --data-binary @openapi.yml
```

### 本地驗證
```bash
# 使用 swagger-cli
npm install -g swagger-cli
swagger-cli validate openapi.yml
```

## 📝 更新規範

當 API 發生變化時：

1. **更新 Controller 註釋**: 修改 `controller/*.go` 中的 swagger 註釋
2. **重新生成文檔**: 運行 `swag init`
3. **手動更新 openapi.yml**: 將新的 swagger.yaml 轉換為 OpenAPI 3.0 格式
4. **驗證更改**: 確保所有端點和模型都正確定義

## 🚨 注意事項

1. **安全性**: 所有敏感端點都需要 JWT 認證
2. **權限控制**: 根據用戶角色控制資源訪問
3. **數據驗證**: 所有請求數據都會被驗證
4. **錯誤處理**: 統一的錯誤響應格式
5. **版本控制**: API 可能會演進，請注意版本兼容性

## 📞 支持

如有問題或需要協助，請參考：
- `README_AUTH.md` - 認證系統詳細說明
- `TROUBLESHOOTING.md` - 故障排除指南
- `test_auth.sh` - API 測試腳本
