# 🔧 快速修復：Elasticsearch 監控權限問題

## 問題描述

前端顯示 admin 帳戶沒有 elasticsearch 相關權限，無法訪問 ES 監控功能。

## 根本原因

**services/auth.go** 初始化時缺少 `elasticsearch` 資源的權限定義。

## 🚀 快速修復方案（二選一）

---

### 方案 A: SQL 腳本修復（推薦，最快）

**適用場景**: 資料庫已初始化，只需補充權限

#### 步驟 1: 執行 SQL 腳本

```bash
cd /Users/chen/Downloads/01BiMap/03MyDevs/log-detect/log-detect-backend

mysql -u root -p logdetect < scripts/add_elasticsearch_permissions.sql
```

#### 步驟 2: 重新登入

前端需要重新登入以獲取包含新權限的 JWT token。

```bash
# 測試登入
curl -X POST http://localhost:8006/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

#### 步驟 3: 驗證權限

```bash
# 使用新 token 測試
curl -X GET http://localhost:8006/api/v1/elasticsearch/monitors \
  -H "Authorization: Bearer YOUR_NEW_TOKEN"
```

**預期結果**: 返回 200 OK（即使列表為空）

---

### 方案 B: 重新初始化權限（完整重置）

**適用場景**: 開發環境，可以重建權限表

#### 步驟 1: 代碼已修正

✅ `services/auth.go:218-221` 已添加 elasticsearch 權限定義

#### 步驟 2: 重新初始化

**選項 2.1: 重啟應用**

如果應用啟動時會調用 `CreateDefaultRolesAndPermissions()`：

```bash
# 停止應用
# 重新啟動
go run main.go
```

**選項 2.2: 手動執行初始化函數**

在 Go 代碼中或初始化腳本中：

```go
authService := services.NewAuthService()
err := authService.CreateDefaultRolesAndPermissions()
if err != nil {
    log.Fatal(err)
}
```

#### 步驟 3: 驗證

同方案 A 的步驟 2-3

---

## 📋 檢查清單

完成修復後，請確認以下項目：

- [ ] 資料庫中存在 4 個 elasticsearch 權限
  ```sql
  SELECT * FROM permissions WHERE resource = 'elasticsearch';
  ```

- [ ] admin 角色已分配這些權限
  ```sql
  SELECT COUNT(*) FROM role_permissions rp
  JOIN permissions p ON rp.permission_id = p.id
  JOIN roles r ON rp.role_id = r.id
  WHERE r.name = 'admin' AND p.resource = 'elasticsearch';
  -- 應該返回 4
  ```

- [ ] admin 用戶可以訪問 elasticsearch API
  ```bash
  curl -X GET http://localhost:8006/api/v1/elasticsearch/monitors \
    -H "Authorization: Bearer $TOKEN"
  # 應該返回 200 OK
  ```

- [ ] 前端可以正常顯示 ES 監控頁面

---

## 🔍 故障排查

### 問題 1: 仍然顯示 403 Forbidden

**可能原因**:
1. 使用舊的 JWT token
2. 用戶的 role_id 不正確

**解決方法**:
```bash
# 1. 重新登入獲取新 token
# 2. 檢查用戶角色
mysql -u root -p logdetect -e "
  SELECT u.username, r.name as role
  FROM users u
  JOIN roles r ON u.role_id = r.id
  WHERE u.username = 'admin';
"
```

### 問題 2: SQL 腳本執行失敗

**可能原因**: 權限已存在

**解決方法**:
```sql
-- 檢查是否已存在
SELECT * FROM permissions WHERE resource = 'elasticsearch';

-- 如果已存在但未分配給角色
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'admin' AND p.resource = 'elasticsearch';
```

### 問題 3: Token 有效但仍無權限

**檢查用戶權限鏈**:
```sql
-- 完整檢查用戶 -> 角色 -> 權限鏈
SELECT
  u.id, u.username,
  r.name as role,
  p.resource, p.action
FROM users u
JOIN roles r ON u.role_id = r.id
LEFT JOIN role_permissions rp ON r.id = rp.role_id
LEFT JOIN permissions p ON rp.permission_id = p.id
WHERE u.username = 'admin'
ORDER BY p.resource, p.action;
```

---

## 📚 相關文檔

- **完整指南**: `docs/user-permissions-guide.md`
- **權限驗證邏輯**: `middleware/auth.go:50-76`
- **權限檢查實作**: `services/auth.go:176-192`
- **路由權限配置**: `router/router.go:138-157`

---

## 🆘 需要幫助？

如果以上方法都無法解決問題，請提供以下資訊：

1. 錯誤訊息截圖
2. JWT token 內容（使用 jwt.io 解碼）
3. 以下 SQL 查詢結果：
   ```sql
   -- 1. 用戶資訊
   SELECT * FROM users WHERE username = 'admin';

   -- 2. 角色權限
   SELECT r.name, p.resource, p.action
   FROM roles r
   JOIN role_permissions rp ON r.id = rp.role_id
   JOIN permissions p ON rp.permission_id = p.id
   WHERE r.name = 'admin' AND p.resource = 'elasticsearch';
   ```

---

**更新日期**: 2025-10-07
**狀態**: ✅ 已驗證可用
