# 設備群組管理功能

## 概述

本功能新增了獨立的設備群組管理系統，支援以下兩個核心情境：

1. **新增設備時選擇既有群組或同時新增群組**
2. **單獨新增空群組，待未來手動或自動將設備加入**

## 資料庫變更

### 新增表：`device_groups`

```sql
CREATE TABLE IF NOT EXISTS `device_groups` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(50) NOT NULL UNIQUE COMMENT '群組名稱',
    `description` VARCHAR(255) COMMENT '群組描述',
    `created_at` INT UNSIGNED,
    `updated_at` INT UNSIGNED,
    `deleted_at` INT,
    INDEX `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='設備群組表';
```

### Migration 檔案

- **Up**: `migrations/mysql/008_device_groups.up.sql`
- **Down**: `migrations/mysql/008_device_groups.down.sql`

Migration 會自動將現有 `devices` 表中的 `device_group` 遷移到 `device_groups` 表。

## API 端點

### 1. 創建設備群組

**POST** `/api/v1/DeviceGroup/Create`

**Request Body**:
```json
{
  "name": "Web Servers",
  "description": "所有 Web 伺服器"
}
```

**Response**:
```json
{
  "id": 1,
  "name": "Web Servers",
  "description": "所有 Web 伺服器",
  "created_at": 1700000000,
  "updated_at": 1700000000
}
```

### 2. 獲取所有設備群組

**GET** `/api/v1/DeviceGroup/GetAll`

**Response**:
```json
[
  {
    "id": 1,
    "name": "Web Servers",
    "description": "所有 Web 伺服器",
    "device_count": 5,
    "created_at": 1700000000,
    "updated_at": 1700000000
  },
  {
    "id": 2,
    "name": "Database Servers",
    "description": "資料庫伺服器群組",
    "device_count": 0,
    "created_at": 1700000001,
    "updated_at": 1700000001
  }
]
```

### 3. 獲取特定設備群組

**GET** `/api/v1/DeviceGroup/Get/:id`

**Response**:
```json
{
  "id": 1,
  "name": "Web Servers",
  "description": "所有 Web 伺服器",
  "device_count": 5,
  "created_at": 1700000000,
  "updated_at": 1700000000
}
```

### 4. 更新設備群組

**PUT** `/api/v1/DeviceGroup/Update`

**Request Body**:
```json
{
  "id": 1,
  "name": "Web Servers - Production",
  "description": "正式環境的 Web 伺服器"
}
```

### 5. 刪除設備群組

**DELETE** `/api/v1/DeviceGroup/Delete/:id`

**限制**：只能刪除**沒有設備**的群組。如果群組中還有設備，會返回錯誤：

```json
{
  "error": "cannot delete device group: 5 devices still in this group"
}
```

### 6. 獲取設備群組列表（舊端點，向後相容）

**GET** `/api/v1/Device/GetGroup`

此端點已更新為從 `device_groups` 表查詢，並包含設備數量統計：

**Response**:
```json
[
  {
    "id": 1,
    "device_group": "Web Servers",
    "description": "所有 Web 伺服器",
    "device_count": 5
  },
  {
    "id": 2,
    "device_group": "Database Servers",
    "description": "資料庫伺服器群組",
    "device_count": 0
  }
]
```

**注意**：回應欄位 `device_group` 保持原名稱以確保前端相容性。

## 權限控制

所有 DeviceGroup API 使用與 Device 相同的權限：

- **讀取**：需要 `device:read` 權限
- **創建**：需要 `device:create` 權限
- **更新**：需要 `device:update` 權限
- **刪除**：需要 `device:delete` 權限

## 使用情境

### 情境 1：先創建群組，再新增設備

```bash
# 1. 創建空群組
curl -X POST http://localhost:8006/api/v1/DeviceGroup/Create \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production Servers",
    "description": "正式環境伺服器"
  }'

# 2. 稍後新增設備到該群組
curl -X POST http://localhost:8006/api/v1/Device/Create \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '[{
    "device_group": "Production Servers",
    "name": "web-server-01"
  }]'
```

### 情境 2：新增設備時選擇既有群組

```bash
# 1. 查詢既有群組
curl -X GET http://localhost:8006/api/v1/DeviceGroup/GetAll \
  -H "Authorization: Bearer <token>"

# 2. 新增設備並選擇群組
curl -X POST http://localhost:8006/api/v1/Device/Create \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '[{
    "device_group": "Web Servers",
    "name": "web-server-02"
  }]'
```

### 情境 3：新增設備時同時創建新群組

前端需要先檢查群組是否存在：

```javascript
// 前端邏輯範例
async function createDeviceWithGroup(deviceData) {
  // 1. 檢查群組是否存在
  const groups = await fetch('/api/v1/DeviceGroup/GetAll');
  const groupExists = groups.some(g => g.name === deviceData.device_group);

  // 2. 如果群組不存在，先創建群組
  if (!groupExists) {
    await fetch('/api/v1/DeviceGroup/Create', {
      method: 'POST',
      body: JSON.stringify({
        name: deviceData.device_group,
        description: ''
      })
    });
  }

  // 3. 創建設備
  await fetch('/api/v1/Device/Create', {
    method: 'POST',
    body: JSON.stringify([deviceData])
  });
}
```

## 程式碼結構

### 新增檔案

1. **Entity**: `entities/targets.go` (新增 `DeviceGroup` 結構)
2. **Service**: `services/device_group.go`
3. **Controller**: `controller/device_group.go`
4. **Migration**: `migrations/mysql/008_device_groups.up.sql`
5. **Migration Down**: `migrations/mysql/008_device_groups.down.sql`

### 修改檔案

1. **Router**: `router/router.go` (新增 `/DeviceGroup` 路由群組)
2. **Device Service**: `services/device.go` (`GetDeviceGroup` 改為從 `device_groups` 表查詢)

## 資料一致性

- `devices.device_group` 欄位仍然存在，保持原有資料結構
- `device_groups.name` 與 `devices.device_group` 透過名稱關聯
- Migration 會自動將現有設備群組遷移到新表
- 未來可考慮新增外鍵約束確保資料一致性

## 測試建議

### 單元測試

1. 創建群組成功
2. 創建重複名稱群組失敗
3. 更新群組資訊
4. 刪除空群組成功
5. 刪除有設備的群組失敗
6. 查詢群組列表含設備數量

### 整合測試

1. 創建群組 → 新增設備 → 查詢設備數量正確
2. 刪除群組內所有設備 → 刪除群組成功
3. 舊端點 `/Device/GetGroup` 向後相容性

## 前端整合建議

### 新增設備表單

```jsx
// 設備群組選擇器（支援選擇既有或新增）
<Select
  options={existingGroups}
  allowCreate={true}
  placeholder="選擇或輸入新群組名稱"
  onChange={(value) => {
    if (!existingGroups.includes(value)) {
      // 提示使用者將創建新群組
      showNotification('將創建新群組: ' + value);
    }
  }}
/>
```

### 群組管理頁面

建議新增獨立的群組管理頁面：

- 列表顯示所有群組及設備數量
- 支援創建、編輯、刪除群組
- 顯示空群組（`device_count = 0`）
- 群組詳情頁顯示該群組的所有設備

## 未來增強

1. **批次操作**：支援批次將設備移動到不同群組
2. **群組層級**：支援群組的父子關係（樹狀結構）
3. **外鍵約束**：在 `devices` 表新增外鍵指向 `device_groups`
4. **群組模板**：預設群組模板快速創建
5. **群組權限**：基於群組的細粒度權限控制

## 注意事項

⚠️ **重要**：執行 Migration 前請先備份資料庫！

```bash
# 備份資料庫
mysqldump -u bimap -p logdetect > backup_before_device_groups.sql

# 執行 Migration（假設使用 golang-migrate）
migrate -path ./migrations/mysql -database "mysql://bimap:1qaz2wsx@tcp(10.99.1.213:3306)/logdetect" up
```

## 回滾

如果需要回滾此功能：

```bash
migrate -path ./migrations/mysql -database "mysql://bimap:1qaz2wsx@tcp(10.99.1.213:3306)/logdetect" down 1
```

**警告**：回滾會刪除 `device_groups` 表，但 `devices.device_group` 欄位仍保留資料。
