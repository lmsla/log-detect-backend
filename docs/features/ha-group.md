# HA Group 高可用裝置群組支援

## 概述

本功能為 log-detect-backend 新增 HA（High Availability）裝置群組支援，讓具備主備機制的設備（如 FortiGate 防火牆）在偵測時以「同組全部失聯」作為真正失聯的判定條件。

### 適用場景

| 場景 | 說明 |
|------|------|
| 防火牆 HA | FortiGate 主備切換，只有 active 節點產生日誌 |
| 叢集設備 | 多台設備互為備援，任一台在線即服務正常 |
| 獨立設備 | 無 HA 設定，維持現有偵測行為 |

---

## 設計原則

1. **動態角色推導**：不需預先標記主備，由偵測結果即時判定
2. **向後相容**：`ha_group` 為空 = 獨立裝置，行為完全不變
3. **最小改動**：DB 僅加一個欄位，History 表不改結構

---

## 狀態判定邏輯

```
獨立裝置（ha_group 為空）：
  有日誌 → online
  沒日誌 → offline（觸發告警）

HA 裝置（ha_group 非空）：
  有日誌 → online（當前 active 節點）
  沒日誌，同組有其他成員 online → standby（正常待命，不告警）
  沒日誌，同組全部都沒日誌 → offline（觸發告警）
```

### History 紀錄對照表

| 裝置類型 | ES 有日誌 | 同組狀態 | status | 觸發告警 |
|----------|-----------|----------|--------|----------|
| 獨立裝置 | 有 | N/A | `online` | 否 |
| 獨立裝置 | 無 | N/A | `offline` | 是 |
| HA 裝置 | 有 | N/A | `online` | 否 |
| HA 裝置 | 無 | 同組有人 online | `standby` | 否 |
| HA 裝置 | 無 | 同組全部無日誌 | `offline` | 是 |

---

## 資料結構

### Device Entity

```go
type Device struct {
    models.Common
    ID          int    `gorm:"primaryKey;index" json:"id" form:"id"`
    DeviceGroup string `gorm:"type:varchar(50)" json:"device_group" form:"device_group"`
    HAGroup     string `gorm:"type:varchar(50);default:''" json:"ha_group" form:"ha_group"`
    Name        string `gorm:"type:varchar(50)" json:"name" form:"name"`
}
```

- `HAGroup` 為空字串 → 獨立裝置
- `HAGroup` 相同的裝置視為同一組 HA 叢集
- GORM AutoMigrate 自動加欄位

### config.yml 格式

```yaml
devices:
  - device_group: "forti"
    names:
      - name: "fw-01-primary"
        ha_group: "fw-cluster-01"
      - name: "fw-01-standby"
        ha_group: "fw-cluster-01"
      - name: "fw-02"           # 無 ha_group = 獨立裝置
        ha_group: ""
```

---

## 偵測流程

### 原始流程

```
ES 查詢 → result_list
DB 查詢 → device_list
ListCompare(device_list, result_list)
  → added（新發現）, removed（失聯）, intersection（在線）
removed 不為空 → 寄信 + 記錄 offline history
```

### HA 增強流程

```
ES 查詢 → result_list
DB 查詢 → device_list
ListCompare(device_list, result_list)
  → added, removed, intersection

filterHAGroups(removed, intersection)
  → trulyRemoved（真正失聯）, standbyDevices（正常待命）

trulyRemoved 不為空 → 寄信（含 HA 群組資訊）
History 紀錄：
  intersection   → status: "online"
  trulyRemoved   → status: "offline"
  standbyDevices → status: "standby"
```

### filterHAGroups 演算法

```
輸入：removed[], intersection[]
輸出：trulyRemoved[], standbyDevices[]

1. 查詢所有 removed 裝置的 ha_group（一次 DB 查詢）
2. 查詢所有 intersection 裝置的 ha_group（一次 DB 查詢）
3. 建立 onlineHAGroups = intersection 中有的 ha_group 集合
4. 遍歷 removed：
   - ha_group 為空 → trulyRemoved
   - ha_group 在 onlineHAGroups 中 → standbyDevices
   - ha_group 不在 onlineHAGroups 中 → trulyRemoved
```

---

## 郵件格式

### HA 全組失聯時的告警郵件

```
遺失設備清單：
+----+------------------+----------------------------+
| #  | Host             | HA Group                   |
+----+------------------+----------------------------+
| 1  | device-A         | -                          |
| 2  | fw-01-primary    | fw-cluster-01 (全組失聯)    |
| 3  | fw-01-standby    | fw-cluster-01 (全組失聯)    |
+----+------------------+----------------------------+
```

---

## 儀表板呈現（前端）

### 收合檢視（預設）

```
forti 群組
├── fw-cluster-01 (HA)  ● 正常    [展開]
├── fw-02               ● 正常
└── fw-03               ✗ 失聯
```

### 展開 HA 群組

```
fw-cluster-01 (HA)      ● 正常
  ├── fw-01-primary      ● online (active)
  └── fw-01-standby      ○ standby
```

### 聚合邏輯

```
HA 群組狀態 = 任一成員 online ? "正常" : "失聯"
```

---

## 影響範圍

| 項目 | 變更 |
|------|------|
| `devices` 表 | 加 `ha_group` 欄位（AutoMigrate） |
| `history` 表 | 不改，`status` 多一個 `standby` 值 |
| `entities/targets.go` | Device struct 加一行 |
| `structs/features.go` | YMLDeviceGroup 結構改為物件陣列 |
| `services/detect.go` | 偵測邏輯加 HA 過濾 |
| `services/mail.go` | 郵件格式支援 HA 群組 |
| `services/config_sync.go` | 同步 ha_group 欄位 |
| `config.yml` | devices 格式擴充 |
| 前端 API | Device CRUD 自動支援（GORM binding） |

---

## 驗證場景

1. **獨立裝置**：ha_group 為空，行為與改動前完全一致
2. **HA 正常**：主機 online、備機 standby → 不觸發告警
3. **HA 全組失聯**：主備都離線 → 觸發告警，郵件標示 HA 群組
4. **HA Failover**：主備切換後 history 自動反映角色變化
5. **向後相容**：現有 devices 表資料（無 ha_group）預設為空字串 = 獨立裝置
