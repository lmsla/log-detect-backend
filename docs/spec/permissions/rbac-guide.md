# Log Detect - 用戶權限設定指南

## 📋 權限系統架構

### 1. 權限模型

**架構**: RBAC (Role-Based Access Control)

```
User -> Role -> Permissions
```

- **User**: 用戶帳戶
- **Role**: 角色（admin, user, etc.）
- **Permission**: 權限（resource:action）

### 2. 權限格式

```
{resource}:{action}
```

**範例**:
- `device:read` - 讀取設備
- `device:create` - 創建設備
- `elasticsearch:read` - 讀取 ES 監控配置
- `elasticsearch:create` - 創建 ES 監控配置

---

## 🔍 當前問題診斷

### 問題現象

前端顯示 admin 帳戶沒有 `elasticsearch` 相關權限。

### 根本原因

**services/auth.go:195-217** 的 `CreateDefaultRolesAndPermissions()` 函數中，**缺少 elasticsearch 權限定義**。

當前只有以下權限：
- ✅ `device:*` (create, read, update, delete)
- ✅ `target:*` (create, read, update, delete)
- ✅ `indices:*` (create, read, update, delete)
- ✅ `user:*` (create, read, update, delete)
- ❌ `elasticsearch:*` - **缺少！**

但路由中使用了：
```go
// router/router.go:141, 146-148, 152
esGroup.Use(middleware.PermissionMiddleware("elasticsearch", "read"))
esGroup.POST("/monitors", ...).Use(middleware.PermissionMiddleware("elasticsearch", "create"))
esGroup.PUT("/monitors", ...).Use(middleware.PermissionMiddleware("elasticsearch", "update"))
esGroup.DELETE("/monitors/:id", ...).Use(middleware.PermissionMiddleware("elasticsearch", "delete"))
esGroup.POST("/monitors/:id/toggle", ...).Use(middleware.PermissionMiddleware("elasticsearch", "update"))
```

---

## 🔧 解決方案

### 方案 1: 修改權限初始化（推薦）

在 `services/auth.go` 的 `CreateDefaultRolesAndPermissions()` 函數中添加 elasticsearch 權限。

**需要添加的權限**:
```go
{Name: "elasticsearch:create", Resource: "elasticsearch", Action: "create", Description: "Create ES monitors"},
{Name: "elasticsearch:read", Resource: "elasticsearch", Action: "read", Description: "Read ES monitors"},
{Name: "elasticsearch:update", Resource: "elasticsearch", Action: "update", Description: "Update ES monitors"},
{Name: "elasticsearch:delete", Resource: "elasticsearch", Action: "delete", Description: "Delete ES monitors"},
```

### 方案 2: 手動資料庫更新（臨時方案）

如果資料庫已初始化，可以手動執行 SQL：

```sql
-- 1. 插入 elasticsearch 權限
INSERT INTO permissions (name, resource, action, description, created_at, updated_at) VALUES
('elasticsearch:create', 'elasticsearch', 'create', 'Create ES monitors', NOW(), NOW()),
('elasticsearch:read', 'elasticsearch', 'read', 'Read ES monitors', NOW(), NOW()),
('elasticsearch:update', 'elasticsearch', 'update', 'Update ES monitors', NOW(), NOW()),
('elasticsearch:delete', 'elasticsearch', 'delete', 'Delete ES monitors', NOW(), NOW());

-- 2. 獲取 admin role ID
SELECT id FROM roles WHERE name = 'admin';
-- 假設得到 role_id = 1

-- 3. 獲取新權限的 ID
SELECT id FROM permissions WHERE resource = 'elasticsearch';
-- 假設得到 permission_ids: 17, 18, 19, 20

-- 4. 將權限分配給 admin 角色
INSERT INTO role_permissions (role_id, permission_id) VALUES
(1, 17),
(1, 18),
(1, 19),
(1, 20);
```

---

## 📊 完整權限列表

### 當前系統應有的權限

| Resource | Action | Permission Name | Description |
|----------|--------|-----------------|-------------|
| device | create | device:create | 創建設備 |
| device | read | device:read | 讀取設備 |
| device | update | device:update | 更新設備 |
| device | delete | device:delete | 刪除設備 |
| target | create | target:create | 創建目標 |
| target | read | target:read | 讀取目標 |
| target | update | target:update | 更新目標 |
| target | delete | target:delete | 刪除目標 |
| indices | create | indices:create | 創建索引 |
| indices | read | indices:read | 讀取索引 |
| indices | update | indices:update | 更新索引 |
| indices | delete | indices:delete | 刪除索引 |
| user | create | user:create | 創建用戶 |
| user | read | user:read | 讀取用戶 |
| user | update | user:update | 更新用戶 |
| user | delete | user:delete | 刪除用戶 |
| **elasticsearch** | **create** | **elasticsearch:create** | **創建 ES 監控** |
| **elasticsearch** | **read** | **elasticsearch:read** | **讀取 ES 監控** |
| **elasticsearch** | **update** | **elasticsearch:update** | **更新 ES 監控** |
| **elasticsearch** | **delete** | **elasticsearch:delete** | **刪除 ES 監控** |

---

## 🔐 權限驗證流程

### 1. 用戶登入
```
POST /auth/login
↓
AuthService.Login()
↓
生成 JWT (包含 user_id, role_id)
```

### 2. API 請求驗證
```
Request with Bearer Token
↓
AuthMiddleware() - 驗證 JWT
↓
提取 user_id, role_id 放入 context
↓
PermissionMiddleware(resource, action) - 檢查權限
↓
AuthService.CheckPermission(user_id, resource, action)
↓
查詢 User -> Role -> Permissions
↓
比對 permission.Resource == resource && permission.Action == action
```

### 3. 權限檢查邏輯

**檔案**: `services/auth.go:176-192`

```go
func (s *AuthService) CheckPermission(userID uint, resource, action string) (bool, error) {
    var user entities.User
    err := global.Mysql.Preload("Role.Permissions").
           Where("id = ? AND is_active = ?", userID, true).
           First(&user).Error
    if err != nil {
        return false, err
    }

    // 檢查用戶角色的所有權限
    for _, permission := range user.Role.Permissions {
        if permission.Resource == resource && permission.Action == action {
            return true, nil
        }
    }

    return false, nil
}
```

---

## 🛠️ 測試權限設定

### 1. 查看當前用戶權限

```bash
# 登入後獲取 token
TOKEN="your_jwt_token_here"

# 查看當前用戶資訊
curl -X GET http://localhost:8006/api/v1/auth/profile \
  -H "Authorization: Bearer $TOKEN"
```

### 2. 測試 elasticsearch 權限

```bash
# 測試讀取權限
curl -X GET http://localhost:8006/api/v1/elasticsearch/monitors \
  -H "Authorization: Bearer $TOKEN"

# 如果返回 403 Forbidden，表示缺少權限
# 如果返回 200 OK，表示權限正常
```

### 3. 查詢資料庫確認

```sql
-- 查看用戶的角色和權限
SELECT
    u.id, u.username, u.email,
    r.name as role_name,
    p.name as permission_name,
    p.resource, p.action
FROM users u
JOIN roles r ON u.role_id = r.id
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id
WHERE u.username = 'admin';

-- 檢查 elasticsearch 權限是否存在
SELECT * FROM permissions WHERE resource = 'elasticsearch';
```

---

## 🚀 快速修復步驟

### 步驟 1: 更新權限定義代碼
修改 `services/auth.go:195-217`，添加 elasticsearch 權限。

### 步驟 2: 重新初始化權限
```go
// 在應用啟動時或專門的初始化腳本中
authService := services.NewAuthService()
err := authService.CreateDefaultRolesAndPermissions()
if err != nil {
    log.Fatal(err)
}
```

### 步驟 3: 驗證
```bash
# 重新登入獲取新 token
curl -X POST http://localhost:8006/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 測試權限
curl -X GET http://localhost:8006/api/v1/elasticsearch/monitors \
  -H "Authorization: Bearer $NEW_TOKEN"
```

---

## 📝 預設角色配置

### Admin 角色
- **所有權限**: ✅
- 包含所有 resource 的 create, read, update, delete 權限

### User 角色
- **唯讀權限**: ✅
- 只包含所有 resource 的 read 權限

### 預設帳戶
- **Username**: admin
- **Password**: admin123
- **Email**: admin@logdetect.com
- **Role**: admin

---

## 🔄 動態權限管理

### 添加新權限

```go
newPermission := entities.Permission{
    Name: "resource:action",
    Resource: "resource",
    Action: "action",
    Description: "Description",
}
global.Mysql.Create(&newPermission)
```

### 分配權限給角色

```go
var role entities.Role
global.Mysql.Where("name = ?", "admin").First(&role)

var permission entities.Permission
global.Mysql.Where("name = ?", "elasticsearch:read").First(&permission)

global.Mysql.Model(&role).Association("Permissions").Append(&permission)
```

---

## ⚠️ 注意事項

1. **權限檢查順序**
   - 先檢查 JWT 有效性 (AuthMiddleware)
   - 再檢查用戶權限 (PermissionMiddleware)

2. **權限緩存**
   - 當前實作每次請求都查詢資料庫
   - 可考慮添加 Redis 緩存提升性能

3. **權限更新**
   - 用戶權限變更後，需要重新登入獲取新 token
   - 或實作 token 刷新機制

4. **安全建議**
   - 定期審計權限分配
   - 遵循最小權限原則
   - 記錄權限變更日誌

---

**最後更新**: 2025-10-07
