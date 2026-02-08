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
  timescaledb: false       # 停用 TimescaleDB 連線、BatchWriter、時序資料
  es_monitoring: false     # 停用 ES 叢集健康監控排程
  dashboard: false         # 停用儀表板相關 API
  auth: false              # 停用 JWT/RBAC 認證授權
  history: true            # 偵測結果是否寫入 MySQL history 表

# 配置來源
config_source: "yml"       # "yml" = 啟動時從 config.yml 同步到 DB
                           # "api" = 由前端/API 管理（預設行為）
```

**注意**：`features` 區段若未定義，所有開關預設為 `false`。現有部署環境升級時需加入此區段並設定為 `true`。

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
        logname: "waf"             # 日誌名稱
        device_group: "waf-group"  # 裝置群組
        period: "minutes"          # 監控週期：minutes 或 hours
        unit: 5                    # 週期數值
        field: "host.keyword"      # 聚合欄位
        es_connection: "default"   # 引用 es_connections 的 name

  - subject: "Firewall Monitor"
    receiver:
      - ops@example.com
    enable: true
    indices:
      - index: "logstash-fw-*"
        logname: "firewall"
        device_group: "fw-group"
        period: "hours"
        unit: 1
        field: "host.keyword"
        es_connection: "default"

# 裝置清單（選填，未定義時由偵測自動發現）
devices:
  - device_group: "waf-group"
    names:
      - "waf-01"
      - "waf-02"
      - "waf-03"
  - device_group: "fw-group"
    names:
      - "fw-01"
      - "fw-02"
```

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
6. **偵測服務**: `services/detect.go` — History 寫入開關
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
    History      bool `mapstructure:"history"`
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

// YMLDeviceGroup 裝置群組配置
type YMLDeviceGroup struct {
    DeviceGroup string   `yaml:"device_group" mapstructure:"device_group"`
    Names       []string `yaml:"names" mapstructure:"names"`
}
```

### 2. EnviromentModel 擴充 (`structs/env.go`)

```go
type EnviromentModel struct {
    // ... 現有欄位不變 ...
    Features     FeaturesConfig `mapstructure:"features"`       // 新增
    ConfigSource string         `mapstructure:"config_source"`  // 新增
}
```

### 3. 全域變數 (`global/global.go`)

```go
var (
    // ... 現有變數不變 ...
    YMLConfig *structs.YMLConfig  // 新增：擴充格式的 config.yml
)
```

### 4. 配置解析 (`utils/utils.go`)

在 `viperSettingToModel()` 中新增：

```go
// Features
config.Features.TimescaleDB = viper.GetBool("features.timescaledb")
config.Features.ESMonitoring = viper.GetBool("features.es_monitoring")
config.Features.Dashboard = viper.GetBool("features.dashboard")
config.Features.Auth = viper.GetBool("features.auth")
config.Features.History = viper.GetBool("features.history")
config.ConfigSource = viper.GetString("config_source")
```

在 `loadConfigFile()` 中新增擴充格式解析：

```go
// 新增：解析擴充格式到 YMLConfig
var ymlConfig structs.YMLConfig
if err := viper.Unmarshal(&ymlConfig); err != nil {
    fmt.Printf("Warning: could not unmarshal expanded config.yml: %v\n", err)
}
global.YMLConfig = &ymlConfig
```

### 5. 啟動流程 (`main.go`)

```go
func main() {
    utils.LoadEnvironment()

    clients.LoadDatabase()
    mysql, _ := global.Mysql.DB()
    defer mysql.Close()

    // === Feature Toggle: TimescaleDB ===
    if global.EnvConfig.Features.TimescaleDB {
        if err := clients.LoadTimescaleDB(); err != nil {
            log.Fatalf("Failed to initialize TimescaleDB: %v", err)
        }
        defer global.TimescaleDB.Close()

        if global.EnvConfig.BatchWriter.Enabled {
            // ... 現有 BatchWriter 初始化 ...
        }
    } else {
        fmt.Println("TimescaleDB feature disabled, skipping initialization")
    }

    // Migrations
    services.RunMigrations()

    // === YML-to-DB Sync ===
    if global.EnvConfig.ConfigSource == "yml" {
        fmt.Println("Config source is YML, syncing config.yml to database...")
        if err := services.SyncConfigToDB(); err != nil {
            log.Fatalf("Failed to sync config to DB: %v", err)
        }
        fmt.Println("Config sync completed")
    }

    // ES Client（偵測必需）
    clients.SetElkClient()

    // === Feature Toggle: Auth ===
    if global.EnvConfig.Features.Auth {
        authService := services.NewAuthService()
        authService.CreateDefaultRolesAndPermissions()
        authService.CreateDefaultAdmin()
    } else {
        fmt.Println("Auth feature disabled, skipping RBAC initialization")
    }

    services.LoadCrontab()

    // === Feature Toggle: ES Monitoring ===
    if global.EnvConfig.Features.ESMonitoring {
        services.InitESScheduler()
        services.GlobalESScheduler.LoadAllMonitors()
    } else {
        fmt.Println("ES Monitoring feature disabled, skipping scheduler")
    }

    // 核心偵測（永遠執行）
    services.Control_center()

    r := router.LoadRouter()
    r.Run(global.EnvConfig.Server.Port)
}
```

### 6. 同步機制 (`services/config_sync.go`)

#### 同步演算法

```
SyncConfigToDB()
├── 1. 驗證 YMLConfig 是否有效
├── 2. 開始 GORM 交易
├── 3. syncESConnections()
│   ├── 載入 DB 現有 ESConnection
│   ├── 以 name 為識別鍵，新增/更新
│   ├── YML 中不存在的 → soft delete
│   └── 回傳 connMap[name] → ID
├── 4. syncTargets()
│   ├── 載入 DB 現有 Target (Preload Indices)
│   ├── 以 subject 為識別鍵，新增/更新
│   ├── 同步每個 Target 的 Indices
│   │   ├── 以 logname 為識別鍵
│   │   ├── 透過 connMap 解析 es_connection name → ID
│   │   └── YML 中不存在的 index → 刪除
│   └── YML 中不存在的 target → 刪除（含關聯）
├── 5. syncDevices()
│   ├── 以 device_group + name 為識別鍵
│   ├── 新增不存在的裝置
│   └── 刪除 YML 中未定義的裝置
└── 6. Commit 或 Rollback
```

#### 核心程式碼

```go
package services

import (
    "fmt"
    "log-detect-backend/entities"
    "log-detect-backend/global"
    log "log-detect-backend/log_record"
    "gorm.io/gorm"
)

// SyncConfigToDB 將 config.yml 的配置同步至 MySQL
// 僅在 config_source = "yml" 時呼叫
func SyncConfigToDB() error {
    ymlConfig := global.YMLConfig
    if ymlConfig == nil {
        log.Logrecord_no_rotate("WARNING", "YMLConfig is nil, skipping sync")
        return nil
    }

    tx := global.Mysql.Begin()
    if tx.Error != nil {
        return fmt.Errorf("failed to begin transaction: %w", tx.Error)
    }
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // Step 1: 同步 ES Connections
    connMap, err := syncESConnections(tx, ymlConfig.ESConnections)
    if err != nil {
        tx.Rollback()
        return fmt.Errorf("sync ES connections failed: %w", err)
    }

    // Step 2: 同步 Targets + Indices
    if err := syncTargets(tx, ymlConfig.Targets, connMap); err != nil {
        tx.Rollback()
        return fmt.Errorf("sync targets failed: %w", err)
    }

    // Step 3: 同步 Devices
    if err := syncDevices(tx, ymlConfig.Devices); err != nil {
        tx.Rollback()
        return fmt.Errorf("sync devices failed: %w", err)
    }

    if err := tx.Commit().Error; err != nil {
        return fmt.Errorf("failed to commit sync transaction: %w", err)
    }

    log.Logrecord_no_rotate("INFO", "Config sync completed successfully")
    return nil
}

// syncESConnections 同步 ES 連線配置，回傳 name → DB ID 的映射
func syncESConnections(tx *gorm.DB, connections []structs.YMLESConnection) (map[string]int, error) {
    connMap := make(map[string]int)

    if len(connections) == 0 {
        return connMap, nil
    }

    // 載入現有連線
    var existing []entities.ESConnection
    tx.Where("deleted_at IS NULL").Find(&existing)
    existingByName := make(map[string]*entities.ESConnection)
    for i := range existing {
        existingByName[existing[i].Name] = &existing[i]
    }

    // 建立 YML name 集合
    ymlNames := make(map[string]bool)

    for _, conn := range connections {
        ymlNames[conn.Name] = true
        if dbConn, ok := existingByName[conn.Name]; ok {
            // 更新
            tx.Model(dbConn).Updates(map[string]interface{}{
                "host":        conn.Host,
                "port":        conn.Port,
                "username":    conn.Username,
                "password":    conn.Password,
                "enable_auth": conn.EnableAuth,
                "use_tls":     conn.UseTLS,
                "is_default":  conn.IsDefault,
                "description": conn.Description,
            })
            connMap[conn.Name] = dbConn.ID
        } else {
            // 新增
            newConn := entities.ESConnection{
                Name:        conn.Name,
                Host:        conn.Host,
                Port:        conn.Port,
                Username:    conn.Username,
                Password:    conn.Password,
                EnableAuth:  conn.EnableAuth,
                UseTLS:      conn.UseTLS,
                IsDefault:   conn.IsDefault,
                Description: conn.Description,
            }
            if err := tx.Create(&newConn).Error; err != nil {
                return nil, fmt.Errorf("create ES connection '%s' failed: %w", conn.Name, err)
            }
            connMap[conn.Name] = newConn.ID
        }
    }

    // 刪除 YML 中不存在的
    for name, dbConn := range existingByName {
        if !ymlNames[name] {
            tx.Delete(dbConn)
        }
    }

    return connMap, nil
}

// syncTargets 同步監控目標與索引
func syncTargets(tx *gorm.DB, targets []structs.YMLTarget, connMap map[string]int) error {
    // 載入現有 targets
    var existing []entities.Target
    tx.Preload("Indices").Find(&existing)
    existingBySubject := make(map[string]*entities.Target)
    for i := range existing {
        existingBySubject[existing[i].Subject] = &existing[i]
    }

    ymlSubjects := make(map[string]bool)

    for _, t := range targets {
        ymlSubjects[t.Subject] = true

        if dbTarget, ok := existingBySubject[t.Subject]; ok {
            // 更新 target
            tx.Model(dbTarget).Updates(map[string]interface{}{
                "to":     entities.To(t.Receiver),
                "enable": t.Enable,
            })
            // 同步 indices
            if err := syncIndicesForTarget(tx, dbTarget, t.Indices, connMap); err != nil {
                return err
            }
        } else {
            // 新增 target
            newTarget := entities.Target{
                Subject: t.Subject,
                To:      entities.To(t.Receiver),
                Enable:  t.Enable,
            }
            if err := tx.Create(&newTarget).Error; err != nil {
                return fmt.Errorf("create target '%s' failed: %w", t.Subject, err)
            }
            // 新增 indices
            for _, idx := range t.Indices {
                newIndex := buildIndexEntity(idx, connMap)
                if err := tx.Create(&newIndex).Error; err != nil {
                    return fmt.Errorf("create index '%s' failed: %w", idx.Logname, err)
                }
                // 建立關聯
                tx.Exec("INSERT INTO indices_targets (target_id, index_id) VALUES (?, ?)",
                    newTarget.ID, newIndex.ID)
            }
        }
    }

    // 刪除 YML 中不存在的 targets
    for subject, dbTarget := range existingBySubject {
        if !ymlSubjects[subject] {
            // 清除關聯
            tx.Exec("DELETE FROM indices_targets WHERE target_id = ?", dbTarget.ID)
            tx.Exec("DELETE FROM cron_lists WHERE target_id = ?", dbTarget.ID)
            // 刪除 indices
            for _, idx := range dbTarget.Indices {
                tx.Delete(&idx)
            }
            tx.Delete(dbTarget)
        }
    }

    return nil
}

// syncIndicesForTarget 同步單一 Target 下的 Indices
func syncIndicesForTarget(tx *gorm.DB, target *entities.Target, ymlIndices []structs.YMLIndex, connMap map[string]int) error {
    existingByLogname := make(map[string]*entities.Index)
    for i := range target.Indices {
        existingByLogname[target.Indices[i].Logname] = &target.Indices[i]
    }

    ymlLognames := make(map[string]bool)

    for _, idx := range ymlIndices {
        ymlLognames[idx.Logname] = true
        if dbIdx, ok := existingByLogname[idx.Logname]; ok {
            // 更新
            updates := buildIndexUpdates(idx, connMap)
            tx.Model(dbIdx).Updates(updates)
        } else {
            // 新增
            newIndex := buildIndexEntity(idx, connMap)
            if err := tx.Create(&newIndex).Error; err != nil {
                return fmt.Errorf("create index '%s' failed: %w", idx.Logname, err)
            }
            tx.Exec("INSERT INTO indices_targets (target_id, index_id) VALUES (?, ?)",
                target.ID, newIndex.ID)
        }
    }

    // 刪除不存在的
    for logname, dbIdx := range existingByLogname {
        if !ymlLognames[logname] {
            tx.Exec("DELETE FROM indices_targets WHERE index_id = ?", dbIdx.ID)
            tx.Delete(dbIdx)
        }
    }

    return nil
}

// syncDevices 同步裝置清單
func syncDevices(tx *gorm.DB, deviceGroups []structs.YMLDeviceGroup) error {
    if len(deviceGroups) == 0 {
        return nil // 不定義 devices 時跳過，由偵測自動發現
    }

    for _, group := range deviceGroups {
        // 載入該群組現有裝置
        var existing []entities.Device
        tx.Where("device_group = ?", group.DeviceGroup).Find(&existing)
        existingNames := make(map[string]bool)
        for _, d := range existing {
            existingNames[d.Name] = true
        }

        ymlNames := make(map[string]bool)
        for _, name := range group.Names {
            ymlNames[name] = true
        }

        // 新增不存在的
        for _, name := range group.Names {
            if !existingNames[name] {
                tx.Create(&entities.Device{
                    DeviceGroup: group.DeviceGroup,
                    Name:        name,
                })
            }
        }

        // 刪除 YML 中沒有的
        for _, d := range existing {
            if !ymlNames[d.Name] {
                tx.Delete(&d)
            }
        }
    }

    return nil
}

// 輔助函式
func buildIndexEntity(idx structs.YMLIndex, connMap map[string]int) entities.Index {
    index := entities.Index{
        Pattern:     idx.Index,
        Logname:     idx.Logname,
        DeviceGroup: idx.DeviceGroup,
        Period:      idx.Period,
        Unit:        idx.Unit,
        Field:       idx.Field,
    }
    if idx.ESConnection != "" {
        if connID, ok := connMap[idx.ESConnection]; ok {
            index.ESConnectionID = &connID
        }
    }
    return index
}

func buildIndexUpdates(idx structs.YMLIndex, connMap map[string]int) map[string]interface{} {
    updates := map[string]interface{}{
        "pattern":      idx.Index,
        "device_group": idx.DeviceGroup,
        "period":       idx.Period,
        "unit":         idx.Unit,
        "field":        idx.Field,
    }
    if idx.ESConnection != "" {
        if connID, ok := connMap[idx.ESConnection]; ok {
            updates["es_connection_id"] = connID
        }
    }
    return updates
}
```

### 7. History 開關 (`services/detect.go`)

在 `Detect()` 函式中，包裹 history 寫入邏輯：

```go
historyEnabled := global.EnvConfig.Features.History

// 在線上裝置迴圈中
if historyEnabled {
    Insert_HistoryData(historyData)
}
if global.BatchWriter != nil {
    global.BatchWriter.AddHistory(historyData)
}

// 離線裝置迴圈中同理
if historyEnabled {
    Insert_HistoryData(historyData)
}
if global.BatchWriter != nil {
    global.BatchWriter.AddHistory(historyData)
}
```

### 8. 路由條件註冊 (`router/router.go`)

```go
func LoadRouter() *gin.Engine {
    features := global.EnvConfig.Features

    // 動態選擇中介層
    var authMW gin.HandlerFunc
    if features.Auth {
        authMW = middleware.AuthMiddleware()
    } else {
        authMW = func(c *gin.Context) { c.Next() }
    }

    // Auth 路由
    if features.Auth {
        auth := router.Group("/auth")
        { /* 現有 auth 路由 */ }
    }

    // Dashboard 路由
    if features.Dashboard {
        dashboard := apiv1.Group("/dashboard")
        dashboard.Use(authMW)
        { /* 現有 dashboard 路由 */ }
    }

    // ES Monitoring 路由
    if features.ESMonitoring {
        esGroup := apiv1.Group("/elasticsearch")
        esGroup.Use(authMW)
        { /* 現有 ES monitoring 路由 */ }
    }

    // 核心路由（Target, Device, Index 等）永遠註冊
    // ...
}
```

### 9. TimescaleDB Migration Guard (`services/migration.go`)

```go
func RunMigrations() error {
    if err := runMySQLMigrations(); err != nil {
        return fmt.Errorf("MySQL migration failed: %w", err)
    }

    if global.EnvConfig.Features.TimescaleDB {
        if err := runTimescaleDBMigrations(); err != nil {
            return fmt.Errorf("TimescaleDB migration failed: %w", err)
        }
    } else {
        fmt.Println("TimescaleDB feature disabled, skipping TimescaleDB migrations")
    }

    return nil
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

## 識別鍵策略

同步時以業務欄位判斷「同一筆資料」：

| 資料表 | 識別鍵 | 說明 |
|--------|--------|------|
| es_connections | `name` | 已有 unique index |
| targets | `subject` | 已有重複檢查邏輯 |
| indices | `logname`（同 target 內） | 同一 target 下日誌名稱不重複 |
| devices | `device_group` + `name` | 已有 unique key |

---

## 同步策略：全量覆蓋

YML 代表完整的期望狀態：

- YML 有、DB 無 → **新增**
- YML 有、DB 有 → **更新**
- YML 無、DB 有 → **刪除**

整個同步包在一個 GORM 交易中，任何步驟失敗則全部 Rollback。

⚠️ **重要**：空的 config.yml（無 targets）會導致 DB 中所有 targets 被刪除。這是預期行為，但啟動時會輸出醒目的警告訊息。

---

## 邊界情況處理

| 情況 | 處理方式 |
|------|---------|
| config.yml 無 `es_connections` 區段 | 跳過 ES 連線同步，indices 使用預設 ES 客戶端 |
| config.yml 無 `devices` 區段 | 跳過裝置同步，由 Detect() 自動發現 |
| Index 引用不存在的 `es_connection` name | 同步失敗，Rollback，服務拒絕啟動 |
| `features.timescaledb=false` 但 `batch_writer.enabled=true` | BatchWriter 嵌套在 TimescaleDB 開關內，隱式停用 |
| `features.history=false` 但 dashboard 查詢 history | Dashboard 也應停用；返回空資料 |
| `setting.yml` 無 `features` 區段 | Go 零值，所有開關為 `false`（安全的簡化模式） |
| `config_source` 未定義 | 預設空字串，同步不執行，現有行為不變 |

---

## 使用情境

### 情境 1：精簡部署（無前端客戶）

**setting.yml**：
```yaml
database:
  host: "127.0.0.1"
  port: "3306"
  user: "logdetect"
  password: "password"
  name: "logdetect"
  # ...

email:
  user: "alert@customer.com"
  host: "smtp.customer.com"
  port: "587"
  # ...

features:
  timescaledb: false
  es_monitoring: false
  dashboard: false
  auth: false
  history: true

config_source: "yml"

server:
  mode: "release"
  port: ":8006"
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
      - "waf-node-01"
      - "waf-node-02"
```

**客戶操作**：編輯上述兩個 YML 檔案，重啟服務即生效。

### 情境 2：完整部署（有前端，現有行為不變）

**setting.yml** 加入：
```yaml
features:
  timescaledb: true
  es_monitoring: true
  dashboard: true
  auth: true
  history: true

config_source: "api"
```

所有功能照常運作，同步機制不執行。

---

## 測試建議

### 單元測試

1. `TestSyncESConnections_CreateNew` — 空 DB + YML 有連線 → 驗證正確建立
2. `TestSyncESConnections_UpdateExisting` — DB 有連線，YML 修改 host → 驗證更新
3. `TestSyncESConnections_DeleteRemoved` — DB 有兩筆，YML 只有一筆 → 驗證另一筆被刪
4. `TestSyncTargets_CreateWithIndices` — 空 DB + YML 有 target → 驗證 target + indices + 關聯表
5. `TestSyncTargets_ResolveESConnection` — index 引用 es_connection name → 驗證 ID 正確
6. `TestSyncTargets_InvalidESConnectionRef` — 引用不存在的 name → 驗證回傳錯誤
7. `TestSyncDevices_CreateAndPrune` — 驗證新增/刪除裝置
8. `TestSyncFullTransaction_Rollback` — 注入錯誤 → 驗證 DB 未變更
9. `TestFeatureToggles_TimescaleDBDisabled` — 驗證 TimescaleDB 未初始化
10. `TestFeatureToggles_AuthDisabled` — 驗證路由不需 JWT

### 整合測試

1. **精簡部署啟動**：features 全 false + config_source=yml → 啟動無 TimescaleDB 錯誤、DB 有正確的 targets、cron 正常
2. **完整部署啟動**：features 全 true + config_source=api → 現有行為完全不變
3. **路由驗證**：auth=false 時 API 不需 JWT token；dashboard=false 時回 404

---

## 修改記錄

- **日期**: 待定
- **新增文件**: `structs/features.go`, `services/config_sync.go`
- **修改文件**: `structs/env.go`, `global/global.go`, `utils/utils.go`, `main.go`, `services/migration.go`, `services/detect.go`, `router/router.go`, `setting.yml`, `config.yml`
- **影響函數**: `main()`, `LoadEnvironment()`, `viperSettingToModel()`, `loadConfigFile()`, `RunMigrations()`, `Detect()`, `LoadRouter()`
