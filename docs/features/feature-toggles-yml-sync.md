# 功能開關與 YML 配置同步機制

## 概述

本功能為 log-detect-backend 新增兩項核心能力：

1. **功能模組開關（Feature Toggles）**：透過 `setting.yml` 控制各功能模組的啟用/停用，支援精簡部署
2. **YML-to-DB 同步機制**：啟動時將 `config.yml` 的配置同步至 MySQL，實現純 YML 配置管理

### 適用場景

| 場景 | config_source | features | 說明 |
|------|--------------|----------|------|
| 完整部署（有前端） | `api` | 全部啟用 | 現有行為，透過 API/前端管理配置 |
| 精簡部署（無前端） | `yml` | 按需啟用 | 客戶只接觸 YML，重啟服務生效 |

---

## 設計原則

```
有前端客戶：  前端 → API → DB → 服務讀取
無前端客戶：  YML → SyncConfigToDB() → DB → 服務讀取
                                        ↑
                                服務內部邏輯完全不變
```

**核心思路**：DB 永遠是 Runtime Source of Truth，YML 只是另一個「輸入介面」。服務層程式碼零修改。

---

## 配置變更

### setting.yml 新增區段

```yaml
# 功能模組開關
features:
  timescaledb: false       # TimescaleDB 連線 + BatchWriter + 偵測結果歷史寫入
  es_monitoring: false     # ES 叢集健康監控排程（獨立，不依賴 TimescaleDB）
  dashboard: true          # 儀表板 API
  auth: true               # JWT/RBAC 認證授權

# 配置來源
config_source: "yml"       # "yml" = 啟動時從 config.yml 同步到 DB
                           # "api" = 由前端/API 管理（預設行為）
```

**注意**：`features` 區段若未定義，所有開關預設為 `false`（Go 零值）。現有部署環境升級時需加入此區段並設定為 `true`。

#### Feature Toggle 依賴關係

| 開關 | 控制範圍 | 依賴 |
|------|---------|------|
| `timescaledb` | TimescaleDB 連線、BatchWriter 初始化、偵測結果 history 寫入 | 無 |
| `es_monitoring` | ES 叢集健康監控排程器（`InitESScheduler`）| 無（獨立） |
| `dashboard` | 儀表板相關 API 路由 | 建議同時開啟 `timescaledb` |
| `auth` | JWT 驗證中介層、RBAC 初始化 | 無 |

> ⚠️ `timescaledb: false` 時 BatchWriter 不會初始化，偵測結果的歷史資料也不會寫入 TimescaleDB。
> `dashboard` 依賴 TimescaleDB 中的歷史資料，若 `timescaledb: false` 儀表板將回傳空資料。

### config.yml 擴充格式

原有格式僅支援 targets，擴充後支援完整的監控配置：

```yaml
# ES 連線配置
es_connections:
  - name: "default"                # 連線名稱（唯一識別鍵）
    host: "10.99.1.213"
    port: 9200
    username: "elastic"
    password: "changeme"
    enable_auth: true
    use_tls: true
    is_default: true               # 標記為預設連線
    description: "主要 ES 叢集"

# 監控目標
targets:
  - subject: "WAF Log Monitor"     # 目標名稱（唯一識別鍵）
    receiver:
      - admin@example.com
      - ops@example.com
    enable: true
    indices:
      - index: "logstash-waf-*"    # ES 索引模式
        logname: "waf"             # 日誌名稱（唯一識別鍵）
        device_group: "waf-group"  # 裝置群組（對應 devices 區段）
        period: "minutes"          # 監控週期：minutes 或 hours
        unit: 5                    # 週期數值
        field: "host.keyword"      # 聚合欄位
        es_connection: "default"   # 引用 es_connections 的 name

# 裝置清單（選填，未定義時由偵測自動發現）
# 注意：devices 格式已升級為物件陣列，支援 ha_group 欄位
devices:
  - device_group: "waf-group"
    names:
      - name: "waf-01"
        ha_group: ""               # 空字串 = 獨立裝置
      - name: "waf-02"
        ha_group: ""
  - device_group: "fw-group"
    names:
      - name: "fw-01-primary"
        ha_group: "fw-cluster-01"  # 相同 ha_group = 同一 HA 叢集
      - name: "fw-01-standby"
        ha_group: "fw-cluster-01"
      - name: "fw-02"
        ha_group: ""
```

> ⚠️ `devices` 區段各 group 的 `device_group` 必須與 `targets.indices.device_group` 對應，否則偵測比對會失效。

**向後相容**：舊版 config.yml（僅含 targets 區段）仍可正常運作。新增的 `es_connections` 和 `devices` 區段為選填。

---

## 程式碼結構

### 新增檔案

1. **Struct 定義**: `structs/features.go`
2. **同步服務**: `services/config_sync.go`

### 修改檔案

1. **Struct**: `structs/env.go` — EnviromentModel 加入 `Features` 和 `ConfigSource`
2. **全域變數**: `global/global.go` — 加入 `YMLConfig`
3. **配置載入**: `utils/utils.go` — 解析 features 和擴充 config.yml
4. **啟動流程**: `main.go` — Feature guards + sync 呼叫
5. **資料庫遷移**: `services/migration.go` — TimescaleDB migration guard
6. **偵測服務**: `services/detect.go` — History 寫入改用 `TimescaleDB` 開關
7. **路由**: `router/router.go` — 條件註冊路由與中介層

---

## 實現細節

### 1. Struct 定義 (`structs/features.go`)

```go
package structs

// FeaturesConfig 功能模組開關
type FeaturesConfig struct {
    TimescaleDB  bool `mapstructure:"timescaledb"`
    ESMonitoring bool `mapstructure:"es_monitoring"`
    Dashboard    bool `mapstructure:"dashboard"`
    Auth         bool `mapstructure:"auth"`
    // 注意：history 已移除，統一由 timescaledb 開關控制
    // 開啟 timescaledb 即同時啟用 BatchWriter 與 history 寫入
}

// YMLConfig config.yml 擴充格式的根結構
type YMLConfig struct {
    ESConnections []YMLESConnection `yaml:"es_connections" mapstructure:"es_connections"`
    Targets       []YMLTarget       `yaml:"targets" mapstructure:"targets"`
    Devices       []YMLDeviceGroup  `yaml:"devices" mapstructure:"devices"`
}

// YMLESConnection ES 連線配置
type YMLESConnection struct {
    Name        string `yaml:"name" mapstructure:"name"`
    Host        string `yaml:"host" mapstructure:"host"`
    Port        int    `yaml:"port" mapstructure:"port"`
    Username    string `yaml:"username" mapstructure:"username"`
    Password    string `yaml:"password" mapstructure:"password"`
    EnableAuth  bool   `yaml:"enable_auth" mapstructure:"enable_auth"`
    UseTLS      bool   `yaml:"use_tls" mapstructure:"use_tls"`
    IsDefault   bool   `yaml:"is_default" mapstructure:"is_default"`
    Description string `yaml:"description" mapstructure:"description"`
}

// YMLTarget 監控目標配置
type YMLTarget struct {
    Subject  string     `yaml:"subject" mapstructure:"subject"`
    Receiver []string   `yaml:"receiver" mapstructure:"receiver"`
    Enable   bool       `yaml:"enable" mapstructure:"enable"`
    Indices  []YMLIndex `yaml:"indices" mapstructure:"indices"`
}

// YMLIndex 索引配置
type YMLIndex struct {
    Index        string `yaml:"index" mapstructure:"index"`
    Logname      string `yaml:"logname" mapstructure:"logname"`
    DeviceGroup  string `yaml:"device_group" mapstructure:"device_group"`
    Period       string `yaml:"period" mapstructure:"period"`
    Unit         int    `yaml:"unit" mapstructure:"unit"`
    Field        string `yaml:"field" mapstructure:"field"`
    ESConnection string `yaml:"es_connection" mapstructure:"es_connection"`
}

// YMLDeviceGroup 裝置群組配置（支援 HA Group）
type YMLDeviceGroup struct {
    DeviceGroup string          `yaml:"device_group" mapstructure:"device_group"`
    Names       []YMLDeviceItem `yaml:"names" mapstructure:"names"`
}

// YMLDeviceItem 單一裝置（含 HA 群組設定）
type YMLDeviceItem struct {
    Name    string `yaml:"name" mapstructure:"name"`
    HAGroup string `yaml:"ha_group" mapstructure:"ha_group"`
}
```

> ⚠️ `YMLDeviceGroup.Names` 已從 `[]string` 改為 `[]YMLDeviceItem` 物件陣列，以支援 `ha_group` 欄位。

### 2. 啟動流程 (`main.go`)

```go
func main() {
    utils.LoadEnvironment()

    clients.LoadDatabase()

    // === Feature Toggle: TimescaleDB ===
    // 控制 TimescaleDB 連線、BatchWriter、偵測結果 history 寫入
    if global.EnvConfig.Features.TimescaleDB {
        clients.LoadTimescaleDB()
        // BatchWriter 初始化
    }

    // Migrations（TimescaleDB 遷移由 timescaledb 開關守衛）
    services.RunMigrations()

    // === YML-to-DB Sync ===
    if global.EnvConfig.ConfigSource == "yml" {
        services.SyncConfigToDB()
    }

    clients.SetElkClient()

    // === Feature Toggle: Auth ===
    if global.EnvConfig.Features.Auth {
        authService := services.NewAuthService()
        authService.CreateDefaultRolesAndPermissions()
        authService.CreateDefaultAdmin()
    }

    services.LoadCrontab()

    // === Feature Toggle: ES Monitoring ===
    if global.EnvConfig.Features.ESMonitoring {
        services.InitESScheduler()
        services.GlobalESScheduler.LoadAllMonitors()
    }

    services.Control_center()
    r := router.LoadRouter()
    r.Run(global.EnvConfig.Server.Port)
}
```

### 3. 同步機制 (`services/config_sync.go`)

#### 同步執行順序與設計原則

```
SyncConfigToDB()
├── Step 1: syncESConnectionsUpsert()    — 先建 ES 連線，供後續 Index 參照
├── Step 2: syncDeviceGroups()           — 先建群組，devices 才能正確關聯
├── Step 3: syncTargets()                — Targets + Indices upsert + 刪除
├── Step 4: syncDevices()                — Devices upsert + 刪除（僅處理 YML 有定義的群組）
└── Step 5: syncESConnectionsDelete()    — 必須在 Indices 清理後才執行，避免 FK 牽連
```

> **為什麼 ES 刪除要拆到最後？**
> Index entity 有 `es_connection_id` FK，若先刪除 ES 連線，會造成孤立 FK。
> 先完成 syncTargets（含 Indices 清理）再執行 ES 刪除，確保無殘留引用。

#### 識別鍵策略

| 資料表 | 識別鍵 | 說明 |
|--------|--------|------|
| es_connections | `name` | 唯一名稱 |
| device_groups | `name` | 唯一名稱 |
| targets | `subject` | 郵件主旨即識別鍵 |
| indices | `logname`（同 target 內） | 同一 target 下日誌名稱不重複 |
| devices | `device_group` + `name` | 複合識別鍵 |

#### 同步策略：全量覆蓋

YML 代表完整的期望狀態：

- YML 有、DB 無 → **新增**
- YML 有、DB 有 → **更新**
- YML 無、DB 有 → **刪除**（ES 連線為軟刪除，devices/targets 為硬刪除）

整個同步包在一個 GORM 交易中，任何步驟失敗則全部 Rollback。

> ⚠️ `device_groups` 只做 upsert，**不刪除**。原因：YML 移除的群組可能仍有手動建立的裝置，誤刪會造成資料損失。群組刪除需透過 API 手動確認。

#### `syncDeviceGroups` 特別說明

YML 中 `devices` 區段每個 `device_group` 名稱都會在 `device_groups` 表中自動建立對應記錄。
這確保了 `devices` 表的資料在查詢群組清單時能正確顯示。

#### `syncTargets` — To 欄位序列化注意事項

`Target.To` 欄位有 `gorm:"serializer:json"` tag，更新時**必須使用 struct-based Updates + Select**，
若使用 `map[string]interface{}` 方式，GORM 會繞過序列化器，導致 MySQL Error 3140（Invalid JSON）。

```go
// ✅ 正確寫法
tx.Model(existing).Select("To", "Enable").Updates(entities.Target{
    To:     entities.To(ymlTarget.Receiver),
    Enable: ymlTarget.Enable,
})

// ❌ 錯誤寫法（繞過 serializer:json）
tx.Model(existing).Updates(map[string]interface{}{
    "to":     entities.To(ymlTarget.Receiver),
    "enable": ymlTarget.Enable,
})
```

### 4. History 寫入開關 (`services/detect.go`)

```go
// timescaledb 開關同時控制 BatchWriter 初始化與 history 寫入
if global.EnvConfig.Features.TimescaleDB {
    if global.BatchWriter != nil {
        global.BatchWriter.AddHistory(historyData)
    }
}
```

> history 開關已移除，不再有「history: true 但 timescaledb: false」造成開關看似開啟卻無效的混淆情況。

### 5. 自動發現跨群組防護 (`services/detect.go`)

ES 查詢結果可能包含屬於其他 device_group 的裝置（尤其是測試環境多 target 共用同一 ES index 時）。
系統在自動發現邏輯中加入了全表 name 查詢，**若裝置已存在於任何群組則跳過自動建立**，
避免跨群組告警污染。

```go
// 自動發現前先確認設備不存在於任何群組
var existCount int64
global.Mysql.Model(&entities.Device{}).Where("name = ?", device).Count(&existCount)
if existCount > 0 {
    // 跳過，該裝置已屬於其他 device_group
    continue
}
```

---

## 啟動流程對照

| 步驟 | 現有 | 新增邏輯 |
|------|------|---------|
| 1. LoadEnvironment() | 載入 setting.yml + config.yml | 額外解析 features、config_source、YMLConfig |
| 2. LoadDatabase() | MySQL 連線 | 不變 |
| 3. LoadTimescaleDB() | TimescaleDB 連線 | `if features.timescaledb` |
| 4. InitBatchWriter() | 批次寫入初始化 | 嵌套在 TimescaleDB 開關內 |
| 5. RunMigrations() | MySQL + TimescaleDB 遷移 | TimescaleDB 遷移加 guard |
| **6. SyncConfigToDB()** | **不存在** | **新增：`if config_source == "yml"`** |
| 7. SetElkClient() | ES 客戶端初始化 | 不變 |
| 8. Auth Setup | 建立預設角色 | `if features.auth` |
| 9. LoadCrontab() | 初始化 cron | 不變 |
| 10. ES Scheduler | 初始化 ES 監控 | `if features.es_monitoring` |
| 11. Control_center() | 從 DB 載入 targets，註冊 cron | 不變（已從 DB 讀取同步後的資料） |
| 12. HTTP Server | 啟動 API 服務 | 不變 |

---

## 邊界情況處理

| 情況 | 處理方式 |
|------|---------|
| config.yml 無 `es_connections` 區段 | 跳過 ES 連線同步，indices 使用預設 ES 客戶端 |
| config.yml 無 `devices` 區段 | 跳過裝置同步，由 Detect() 自動發現 |
| config.yml 無 `targets` 區段 | 輸出 WARNING 並刪除 DB 中所有 targets |
| `features.timescaledb=false` | BatchWriter 不初始化，偵測結果 history 不寫入，dashboard 回傳空資料 |
| `features.timescaledb=false` 但 `batch_writer.enabled=true` | BatchWriter 嵌套在 TimescaleDB 開關內，隱式停用 |
| `setting.yml` 無 `features` 區段 | Go 零值，所有開關為 `false`（安全的精簡模式） |
| `config_source` 未定義 | 預設空字串，同步不執行，現有行為不變 |
| 不同 target 使用同一 ES index | 自動發現時跨群組設備會被過濾，不寫入錯誤群組 |
| devices.yml 與 config.yml 同時存在 | `devices.yml` 覆蓋 `config.yml` 的 `devices` 區段 |

---

## 使用情境

### 情境 1：精簡部署（無前端客戶）

**setting.yml**：
```yaml
features:
  timescaledb: false
  es_monitoring: false
  dashboard: false
  auth: false

config_source: "yml"
```

**config.yml**：
```yaml
es_connections:
  - name: "default"
    host: "es-cluster.local"
    port: 9200
    enable_auth: true
    username: "elastic"
    password: "secret"
    use_tls: true
    is_default: true

targets:
  - subject: "WAF Log Detect"
    receiver:
      - ops@customer.com
    enable: true
    indices:
      - index: "logstash-waf-*"
        logname: "waf"
        device_group: "waf"
        period: "minutes"
        unit: 5
        field: "host.keyword"
        es_connection: "default"

devices:
  - device_group: "waf"
    names:
      - name: "waf-node-01"
        ha_group: ""
      - name: "waf-node-02"
        ha_group: ""
```

**客戶操作**：編輯上述兩個 YML 檔案，重啟服務即生效。無需操作前端或資料庫。

### 情境 2：完整部署（有前端，現有行為不變）

**setting.yml**：
```yaml
features:
  timescaledb: true
  es_monitoring: true
  dashboard: true
  auth: true

config_source: "api"
```

所有功能照常運作，同步機制不執行。配置透過前端 API 管理。

### 情境 3：有前端但同時啟用 YML 同步（混合模式）

不建議。`config_source: "yml"` 每次重啟都會以 YML 覆蓋 DB，
若前端同時修改了配置，重啟後會被 YML 蓋回去。請擇一使用。

---

## devices.yml 與 config.yml 的關係

當專案根目錄存在 `devices.yml` 時，其 `devices` 區段會**覆蓋** `config.yml` 的 `devices` 區段。

```
載入順序：
1. loadConfigFile()    → 讀取 config.yml（含 devices 區段）
2. loadDevicesFile()   → 讀取 devices.yml → 覆蓋 global.YMLConfig.Devices
3. SyncConfigToDB()    → 以最終的 YMLConfig（已覆蓋）同步至 DB
```

適用場景：裝置清單頻繁更動，希望獨立維護而不動 config.yml。

---

## 修改記錄

| 日期 | 版本 | 變更內容 |
|------|------|---------|
| 2026-03-30 | 1.1 | 移除 `history` 開關（合併至 `timescaledb`）；新增 `syncDeviceGroups` Step 2；拆分 `syncESConnections` 為 upsert/delete 兩階段；新增跨群組污染防護說明；更新 `YMLDeviceGroup.Names` 為物件陣列格式 |
| 待定 | 1.0 | 初始版本，新增 Feature Toggles 與 YML 同步機制 |
