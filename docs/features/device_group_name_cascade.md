# Device Group Name Cascade Update

## 問題描述

當設備群組名稱（`device_groups.name`）被更新時，相關聯的 `devices.device_group` 和 `indices.device_group` 欄位不會自動更新，導致數據不一致。

## 解決方案

### 應用層級同步更新

在 `services/device_group.go` 的 `UpdateDeviceGroup` 函數中實現事務性的級聯更新：

**核心邏輯：**

1. **獲取舊群組名稱**：在更新前先查詢並保存原有的群組名稱
2. **事務處理**：使用資料庫事務確保所有更新操作的原子性
3. **級聯更新**：當群組名稱變更時，同步更新 `devices` 和 `indices` 表
4. **錯誤回滾**：任何步驟失敗都會回滾整個事務

### 實現細節

```go
func UpdateDeviceGroup(group entities.DeviceGroup) models.Response {
    // 1. 獲取舊群組信息
    var oldGroup entities.DeviceGroup
    result := global.Mysql.Where("id = ?", group.ID).First(&oldGroup)

    // 2. 開啟事務
    tx := global.Mysql.Begin()

    // 3. 更新 device_groups 表
    err := tx.Select("*").Where("id = ?", group.ID).Updates(&group).Error

    // 4. 如果名稱有變更，級聯更新相關表
    if oldGroup.Name != group.Name {
        // 更新 devices.device_group
        tx.Model(&entities.Device{}).
            Where("device_group = ?", oldGroup.Name).
            Update("device_group", group.Name)

        // 更新 indices.device_group
        tx.Model(&entities.Index{}).
            Where("device_group = ?", oldGroup.Name).
            Update("device_group", group.Name)
    }

    // 5. 提交事務
    tx.Commit()
}
```

### 影響範圍

**更新的表：**
- `device_groups` - 主要更新目標
- `devices` - 當群組名稱變更時自動同步
- `indices` - 當群組名稱變更時自動同步

**事務保證：**
- 所有更新操作在同一事務中執行
- 任何步驟失敗都會回滾全部更改
- 確保數據一致性

## 使用示例

### API 調用

```bash
# 更新設備群組名稱
PUT /api/v1/DeviceGroup/Update
{
  "id": 1,
  "name": "新群組名稱",
  "description": "更新後的描述"
}
```

### 數據變更追蹤

更新操作會在日誌中記錄：

```
INFO: Device group name updated from '舊群組名稱' to '新群組名稱', cascaded to devices and indices tables
```

## 與外鍵約束的比較

### 應用層級更新（當前方案）

**優點：**
- 保持設計彈性，無需資料庫層級的強制約束
- 可以在更新時加入額外的業務邏輯
- 更容易進行錯誤處理和自定義回應
- 符合現有架構設計原則

**缺點：**
- 需要手動維護一致性邏輯
- 依賴應用層的正確實現

### 資料庫外鍵約束（替代方案）

**優點：**
- 資料庫層級自動保證一致性
- 無需應用程式碼維護

**缺點：**
- 減少靈活性
- 無法動態創建群組
- 與現有設計理念不符

## 測試建議

### 1. 正常更新測試

```sql
-- 準備測試數據
INSERT INTO device_groups (id, name) VALUES (99, 'test-group-old');
INSERT INTO devices (id, device_group, name) VALUES (999, 'test-group-old', 'test-device');
INSERT INTO indices (id, device_group, logname) VALUES (999, 'test-group-old', 'test-log');

-- 執行更新
-- API: PUT /api/v1/DeviceGroup/Update {"id": 99, "name": "test-group-new"}

-- 驗證結果
SELECT name FROM device_groups WHERE id = 99;  -- 應為 'test-group-new'
SELECT device_group FROM devices WHERE id = 999;  -- 應為 'test-group-new'
SELECT device_group FROM indices WHERE id = 999;  -- 應為 'test-group-new'
```

### 2. 事務回滾測試

模擬更新失敗情況，確保所有表都正確回滾到原始狀態。

### 3. 僅更新描述測試

```sql
-- 測試不改變名稱的更新
-- API: PUT /api/v1/DeviceGroup/Update {"id": 99, "name": "test-group", "description": "new desc"}

-- 驗證：devices 和 indices 表不應被觸及
```

## 性能考量

### 索引支持

相關表已有適當索引支持快速查詢和更新：

```sql
-- devices 表
INDEX `idx_devices_device_group` (`device_group`)

-- device_groups 表
INDEX `idx_name` (`name`)
```

### 批量更新

使用 GORM 的批量更新功能，單次 SQL 語句更新所有相關記錄：

```sql
UPDATE devices SET device_group = 'new-name' WHERE device_group = 'old-name';
UPDATE indices SET device_group = 'new-name' WHERE device_group = 'old-name';
```

## 後續優化建議

1. **添加審計日誌**：記錄所有群組名稱變更歷史
2. **監控影響範圍**：在更新前統計並返回受影響的設備和索引數量
3. **非同步處理**：對於大量設備的情況，考慮使用後台任務處理

## 修改記錄

- **2025-11-25**: 實現應用層級的級聯更新機制
- **修改文件**: `services/device_group.go`
- **影響函數**: `UpdateDeviceGroup()`
