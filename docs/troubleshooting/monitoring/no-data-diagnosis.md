# 🔍 ES 監控無資料問題 - 完整診斷與解決方案

## 問題描述

已在前端設置 ES 監控配置，但 `es_metrics` 表中一直沒有資料進來。

---

## 🎯 根本原因

**ES 監控的 Cron 自動排程功能尚未實作！**

### 當前狀態分析

#### ✅ 已實作的部分
1. **監控配置 CRUD** - 可以新增/編輯/刪除監控配置 ✅
2. **健康檢查邏輯** - `MonitorESCluster()` 函數已實作 ✅
3. **資料寫入邏輯** - BatchWriter 支援 ES 指標 ✅
4. **資料庫表結構** - es_metrics 表已建立 ✅

#### ❌ 缺少的關鍵部分
**自動排程系統** - 沒有定時執行 `MonitorESCluster()` 的機制 ❌

### 現有的 Cron 系統

專案中已有針對 **log 檢測** 的 Cron 系統：
- `services/center.go:LoadCrontab()` - Cron 初始化
- `services/center.go:ExecuteCrontab()` - 執行 log 檢測任務
- `main.go:73` - 應用啟動時載入 Crontab

但 **ES 監控沒有整合到這個系統中**。

---

## 🚀 解決方案

### 方案 A: 手動觸發測試（臨時方案）

在有自動排程之前，可以手動觸發監控來驗證功能：

#### 步驟 1: 創建測試腳本

創建文件：`scripts/test_es_monitor.go`

```go
package main

import (
    "fmt"
    "log"
    "log-detect/entities"
    "log-detect/global"
    "log-detect/services"
)

func main() {
    fmt.Println("=== ES 監控手動測試腳本 ===")

    // 1. 初始化資料庫連接
    // TODO: 調用實際的資料庫初始化
    // 參考 main.go 中的初始化代碼

    // 2. 從資料庫載入所有啟用的監控配置
    var monitors []entities.ElasticsearchMonitor
    result := global.Mysql.Where("enable_monitor = ?", true).Find(&monitors)

    if result.Error != nil {
        log.Fatalf("❌ 無法載入監控配置: %v", result.Error)
    }

    if len(monitors) == 0 {
        fmt.Println("⚠️  沒有找到啟用的監控配置")
        fmt.Println("請先在前端創建並啟用 ES 監控")
        return
    }

    fmt.Printf("✅ 找到 %d 個啟用的監控配置\n\n", len(monitors))

    // 3. 逐個執行監控檢查
    esService := services.NewESMonitorService()

    for _, monitor := range monitors {
        fmt.Printf("📊 開始檢查: %s (%s:%d)\n", monitor.Name, monitor.Host, monitor.Port)

        // 執行監控
        esService.MonitorESCluster(monitor)

        fmt.Printf("✅ 檢查完成: %s\n\n", monitor.Name)
    }

    fmt.Println("🎉 所有監控檢查完成！")
    fmt.Println("請檢查 TimescaleDB es_metrics 表是否有資料")
}
```

#### 步驟 2: 執行測試腳本

```bash
cd /Users/chen/Downloads/01BiMap/03MyDevs/log-detect/log-detect-backend

# 需要先實作資料庫初始化部分
go run scripts/test_es_monitor.go
```

#### 步驟 3: 驗證資料

```sql
-- 檢查是否有資料寫入
psql -U logdetect -d monitoring -c "
    SELECT
        time,
        monitor_id,
        status,
        cluster_name,
        response_time,
        cpu_usage
    FROM es_metrics
    ORDER BY time DESC
    LIMIT 10;
"
```

---

### 方案 B: 實作 Cron 自動排程（正式方案）

#### 實作步驟

**1. 創建 ES 監控排程服務**

創建文件：`services/es_scheduler.go`

```go
package services

import (
    "fmt"
    "log-detect/entities"
    "log-detect/global"
    "log-detect/log"
    "time"
)

// ESMonitorScheduler ES 監控排程器
type ESMonitorScheduler struct {
    monitors map[int]*time.Ticker // monitor_id -> ticker
    stopChan map[int]chan bool    // monitor_id -> stop channel
}

var GlobalESScheduler *ESMonitorScheduler

// InitESScheduler 初始化 ES 監控排程器
func InitESScheduler() {
    GlobalESScheduler = &ESMonitorScheduler{
        monitors: make(map[int]*time.Ticker),
        stopChan: make(map[int]chan bool),
    }

    log.Logrecord_no_rotate("INFO", "ES Monitor Scheduler initialized")
}

// LoadAllMonitors 載入所有啟用的監控配置並啟動排程
func (s *ESMonitorScheduler) LoadAllMonitors() {
    var monitors []entities.ElasticsearchMonitor

    // 查詢所有啟用的監控
    result := global.Mysql.Where("enable_monitor = ?", true).Find(&monitors)

    if result.Error != nil {
        log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to load ES monitors: %s", result.Error.Error()))
        return
    }

    log.Logrecord_no_rotate("INFO", fmt.Sprintf("Loaded %d enabled ES monitors", len(monitors)))

    // 為每個監控創建排程
    for _, monitor := range monitors {
        s.StartMonitor(monitor)
    }
}

// StartMonitor 啟動單一監控的排程
func (s *ESMonitorScheduler) StartMonitor(monitor entities.ElasticsearchMonitor) {
    // 如果已存在，先停止舊的
    s.StopMonitor(monitor.ID)

    // 創建 Ticker (interval 單位是秒)
    interval := time.Duration(monitor.Interval) * time.Second
    ticker := time.NewTicker(interval)
    stopChan := make(chan bool)

    s.monitors[monitor.ID] = ticker
    s.stopChan[monitor.ID] = stopChan

    log.Logrecord_no_rotate("INFO", fmt.Sprintf(
        "Started ES monitor: %s (ID: %d, Interval: %ds)",
        monitor.Name, monitor.ID, monitor.Interval,
    ))

    // 立即執行一次
    go func() {
        esService := NewESMonitorService()
        esService.MonitorESCluster(monitor)
    }()

    // 啟動定時任務
    go func() {
        esService := NewESMonitorService()

        for {
            select {
            case <-ticker.C:
                // 執行監控檢查
                esService.MonitorESCluster(monitor)

            case <-stopChan:
                ticker.Stop()
                log.Logrecord_no_rotate("INFO", fmt.Sprintf(
                    "Stopped ES monitor: %s (ID: %d)",
                    monitor.Name, monitor.ID,
                ))
                return
            }
        }
    }()
}

// StopMonitor 停止單一監控的排程
func (s *ESMonitorScheduler) StopMonitor(monitorID int) {
    if stopChan, exists := s.stopChan[monitorID]; exists {
        close(stopChan)
        delete(s.monitors, monitorID)
        delete(s.stopChan, monitorID)
    }
}

// RestartMonitor 重啟監控排程（用於更新 interval 後）
func (s *ESMonitorScheduler) RestartMonitor(monitorID int) {
    var monitor entities.ElasticsearchMonitor

    result := global.Mysql.First(&monitor, monitorID)
    if result.Error != nil {
        log.Logrecord_no_rotate("ERROR", fmt.Sprintf(
            "Failed to load monitor %d: %s",
            monitorID, result.Error.Error(),
        ))
        return
    }

    if monitor.EnableMonitor {
        s.StartMonitor(monitor)
    } else {
        s.StopMonitor(monitorID)
    }
}

// StopAll 停止所有監控
func (s *ESMonitorScheduler) StopAll() {
    for monitorID := range s.monitors {
        s.StopMonitor(monitorID)
    }
    log.Logrecord_no_rotate("INFO", "All ES monitors stopped")
}
```

**2. 修改 main.go 啟動流程**

```go
// main.go

func main() {
    // ... 現有初始化代碼 ...

    services.LoadCrontab() // 現有的 log 檢測 cron

    // 新增: 初始化並啟動 ES 監控排程
    services.InitESScheduler()
    services.GlobalESScheduler.LoadAllMonitors()

    services.Control_center()

    r := router.LoadRouter()
    r.Run(global.EnvConfig.Server.Port)
}
```

**3. 修改 ES 監控 Service，支援動態更新**

在 `services/es_monitor_service.go` 中添加：

```go
// CreateESMonitor 創建 ES 監控配置
func CreateESMonitor(monitor entities.ElasticsearchMonitor) models.Response {
    // ... 現有創建邏輯 ...

    // 如果啟用監控，立即啟動排程
    if monitor.EnableMonitor && GlobalESScheduler != nil {
        GlobalESScheduler.StartMonitor(monitor)
    }

    return models.Response{
        Success: true,
        Msg:     "創建監控配置成功",
        Body:    monitor,
    }
}

// UpdateESMonitor 更新 ES 監控配置
func UpdateESMonitor(monitor entities.ElasticsearchMonitor) models.Response {
    // ... 現有更新邏輯 ...

    // 重啟排程（以應用新的 interval）
    if GlobalESScheduler != nil {
        GlobalESScheduler.RestartMonitor(monitor.ID)
    }

    return models.Response{
        Success: true,
        Msg:     "更新監控配置成功",
        Body:    monitor,
    }
}

// DeleteESMonitor 刪除 ES 監控配置
func DeleteESMonitor(id int) models.Response {
    // 先停止排程
    if GlobalESScheduler != nil {
        GlobalESScheduler.StopMonitor(id)
    }

    // ... 現有刪除邏輯 ...
}

// ToggleESMonitor 啟用/停用 ES 監控
func ToggleESMonitor(id int, enable bool) models.Response {
    // ... 現有切換邏輯 ...

    // 更新排程狀態
    if GlobalESScheduler != nil {
        GlobalESScheduler.RestartMonitor(id)
    }

    return models.Response{ /* ... */ }
}
```

---

### 方案 C: 使用現有 Cron 系統（整合方案）

將 ES 監控整合到現有的 `CronList` 系統：

**優點**: 重用現有基礎設施
**缺點**: 需要較多改動，且 CronList 設計主要針對 log 檢測

**不推薦**，因為 ES 監控和 log 檢測的排程需求不同。

---

## 📋 診斷檢查清單

在實作自動排程之前，先確認以下項目：

### 1. 監控配置是否正確

```bash
# 檢查 MySQL 監控配置
mysql -u root -p logdetect -e "
    SELECT
        id,
        name,
        host,
        port,
        enable_monitor,
        \`interval\`
    FROM elasticsearch_monitors;
"
```

**預期結果**:
- 至少有一筆記錄
- `enable_monitor` = 1
- `interval` 在 10-3600 之間

### 2. ES 連接是否正常

```bash
# 使用測試端點
curl -X POST http://localhost:8006/api/v1/elasticsearch/monitors/1/test \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**預期結果**: `{ "success": true, "msg": "連接成功", ... }`

### 3. BatchWriter 是否初始化

檢查 `main.go` 或初始化代碼中是否有：

```go
global.BatchWriter = services.NewBatchWriter(global.TimescaleDB, batchSize, flushInterval)
```

### 4. TimescaleDB 連接是否正常

```bash
psql -U logdetect -d monitoring -c "SELECT version();"
```

### 5. 應用日誌檢查

```bash
# 檢查是否有 ES 監控相關的日誌
tail -f logs/app.log | grep -i "ES\|elasticsearch"
```

---

## 🔧 快速驗證流程

### 步驟 1: 確認配置存在

```sql
-- MySQL
SELECT * FROM elasticsearch_monitors WHERE enable_monitor = 1;
```

### 步驟 2: 手動觸發一次監控（Go 代碼）

在應用中添加測試端點（臨時用）：

```go
// controller/elasticsearch.go

// @Summary Manual Trigger ES Monitor (for testing)
// @Tags Elasticsearch
// @Param id path int true "Monitor ID"
// @Router /api/v1/elasticsearch/monitors/{id}/trigger [post]
func TriggerESMonitor(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))

    // 獲取監控配置
    var monitor entities.ElasticsearchMonitor
    if err := global.Mysql.First(&monitor, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Monitor not found"})
        return
    }

    // 執行監控
    esService := services.NewESMonitorService()
    go esService.MonitorESCluster(monitor)

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "msg": "監控已觸發，請稍後檢查 es_metrics 表",
    })
}
```

註冊路由：

```go
// router/router.go
esGroup.POST("/monitors/:id/trigger", controller.TriggerESMonitor)
```

### 步驟 3: 觸發並驗證

```bash
# 觸發監控
curl -X POST http://localhost:8006/api/v1/elasticsearch/monitors/1/trigger \
  -H "Authorization: Bearer YOUR_TOKEN"

# 等待 5 秒

# 檢查資料
psql -U logdetect -d monitoring -c "
    SELECT COUNT(*) FROM es_metrics WHERE monitor_id = 1;
"
```

---

## 💡 建議實作順序

### 短期（立即可做）
1. ✅ 實作手動觸發端點（用於測試）
2. ✅ 驗證監控邏輯和資料寫入是否正常
3. ✅ 測試告警條件檢查

### 中期（本週內）
1. ⏳ 實作 `services/es_scheduler.go`（方案 B）
2. ⏳ 修改 `main.go` 啟動流程
3. ⏳ 修改 CRUD Service 支援動態排程
4. ⏳ 測試自動排程功能

### 長期（優化）
1. ⏳ 添加排程狀態監控端點
2. ⏳ 實作錯誤重試機制
3. ⏳ 添加排程管理界面（前端）

---

## 🆘 故障排查

### 問題 1: 手動觸發後仍無資料

**檢查點**:
```bash
# 1. 檢查應用日誌
tail -f logs/app.log | grep -E "ES monitor|ESMetric"

# 2. 檢查 BatchWriter 是否正常
# 在代碼中添加日誌：
log.Logrecord_no_rotate("INFO", fmt.Sprintf("Adding ES metric to batch: %+v", metric))
```

### 問題 2: ES 連接失敗

**檢查點**:
```bash
# 測試 ES 連接
curl http://YOUR_ES_HOST:9200/_cluster/health

# 檢查認證
curl -u username:password http://YOUR_ES_HOST:9200/_cluster/health
```

### 問題 3: TimescaleDB 寫入失敗

**檢查點**:
```sql
-- 檢查表結構
\d es_metrics

-- 檢查權限
SELECT grantee, privilege_type
FROM information_schema.role_table_grants
WHERE table_name = 'es_metrics';

-- 測試手動插入
INSERT INTO es_metrics (time, monitor_id, status, cluster_name, cluster_status, response_time)
VALUES (NOW(), 1, 'online', 'test', 'green', 100);
```

---

## 📊 總結

### 當前狀態
- ✅ 監控配置管理完整
- ✅ 健康檢查邏輯完整
- ✅ 資料寫入邏輯完整
- ❌ **缺少自動排程系統**

### 解決方案
**推薦方案 B**: 實作獨立的 ES 監控排程器

### 預期工作量
- 編碼: 2-3 小時
- 測試: 1 小時
- 總計: **3-4 小時**

---

**下一步**: 實作 `services/es_scheduler.go` 並整合到應用啟動流程。

**相關檔案**:
- 新增: `services/es_scheduler.go` - 排程服務
- 修改: `main.go` - 啟動流程
- 修改: `services/es_monitor_service.go` - CRUD 整合
- 測試: `scripts/test_es_monitor.go` - 手動測試腳本

**更新日期**: 2025-10-07
