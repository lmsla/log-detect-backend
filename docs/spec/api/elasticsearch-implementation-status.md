# Elasticsearch 監控 - 實作狀態報告

> **最後更新**: 2025-10-22
> **版本**: v1.3 (Phase 3 修復：告警 API NULL 值處理、PostgreSQL 數組參數綁定)

---

## 📊 整體進度總覽

| 階段 | 狀態 | 完成度 | 說明 |
|------|------|--------|------|
| **Phase 1 - 基礎架構** | ✅ 完成 | 100% | 資料庫、Entity、基礎 API |
| **Phase 2 - 核心功能** | ✅ 完成 | 100% | 健康檢查、指標收集、自動排程 |
| **Phase 3 - 進階功能** | ✅ 完成 | 100% | 告警管理 API、通知功能、閾值配置 |
| **Phase 4 - 優化與完善** | ❌ 未開始 | 0% | 效能優化、進階分析 |

**整體完成度**: **85%**

---

## ✅ Phase 1: 基礎架構 (100%)

### 1.1 資料庫結構 ✅

#### MySQL - 監控配置表
**表名**: `elasticsearch_monitors`
**檔案**: `entities/elasticsearch.go:8-25`

**欄位清單** (14 個):
```go
type ElasticsearchMonitor struct {
    ID             int      // 監控 ID (主鍵)
    Name           string   // 監控名稱
    Host           string   // ES 主機地址
    Port           int      // ES 端口 (預設 9200)
    Username       string   // 認證用戶名
    Password       string   // 認證密碼
    EnableAuth     bool     // 是否啟用認證
    CheckType      string   // 檢查類型 (health,performance,capacity)
    Interval       int      // 檢查間隔（秒，範圍 10-3600）
    EnableMonitor  bool     // 是否啟用監控
    Receivers      []string // 告警接收者陣列 (JSON)
    Subject        string   // 告警主題
    Description    string   // 監控描述
    AlertThreshold string   // 告警閾值配置 (JSON)
}
```

**狀態**: ✅ 完成
- [x] GORM AutoMigrate 註冊
- [x] receivers 改為 `[]string` 類型
- [x] 欄位註釋包含單位說明

---

#### TimescaleDB - 指標資料表
**表名**: `es_metrics`
**檔案**: `postgresql_install.sh:82-124`, `entities/elasticsearch.go:32-57`

**欄位清單** (23 個):
| 分類 | 欄位名 | 類型 | 說明 |
|------|--------|------|------|
| **基礎** | time | TIMESTAMPTZ | 時間戳記 (主鍵) |
| | monitor_id | INTEGER | 監控器 ID |
| | status | TEXT | 狀態 (online/offline/warning/error) |
| | cluster_name | TEXT | 集群名稱 |
| | cluster_status | TEXT | 集群狀態 (green/yellow/red) |
| **性能** | response_time | BIGINT | 響應時間（毫秒） |
| | cpu_usage | DECIMAL(5,2) | CPU 使用率（百分比 0-100） |
| | memory_usage | DECIMAL(5,2) | 記憶體使用率（百分比 0-100） |
| | disk_usage | DECIMAL(5,2) | 磁碟使用率（百分比 0-100） |
| **節點** | node_count | INTEGER | 節點總數 |
| | data_node_count | INTEGER | 數據節點數 |
| **查詢** | query_latency | BIGINT | 查詢延遲（毫秒） |
| | indexing_rate | DECIMAL(10,2) | 索引並發數（瞬時並發操作數，非速率） |
| | search_rate | DECIMAL(10,2) | 搜尋並發數（瞬時並發查詢數，非速率） |
| **容量** | total_indices | INTEGER | 索引總數 |
| | total_documents | BIGINT | 文檔總數 |
| | total_size_bytes | BIGINT | 總大小（bytes） |
| **分片** | active_shards | INTEGER | 活躍分片數 ✅ 已修復 |
| | relocating_shards | INTEGER | 遷移中分片數 ✅ 已修復 |
| | unassigned_shards | INTEGER | 未分配分片數 ✅ 已修復 |
| **其他** | error_message | TEXT | 錯誤訊息 |
| | warning_message | TEXT | 警告訊息 |
| | metadata | JSONB | 額外元數據 |

**狀態**: ✅ 完成
- [x] Hypertable 設置（按天分區）
- [x] 23 個指標欄位
- [x] 性能索引（monitor_id, status, cluster_status）
- [x] 壓縮策略（7 天後壓縮）
- [x] 保留策略（90 天自動清理）
- [x] **shards 欄位資料提取修復** (2025-10-08)

---

#### TimescaleDB - 告警歷史表
**表名**: `es_alert_history`
**檔案**: `postgresql_install.sh:126-137`, `entities/elasticsearch.go:59-69`

**狀態**: ✅ 完成
- [x] Hypertable 設置（按 7 天分區）
- [x] 保留策略（90 天）
- [x] 告警狀態追蹤（active, resolved, acknowledged）

---

### 1.2 Entity 定義 ✅

**檔案**: `entities/elasticsearch.go`

| Entity | 用途 | 狀態 |
|--------|------|------|
| `ElasticsearchMonitor` | 監控配置 | ✅ |
| `ESMetric` | 指標資料 | ✅ |
| `ESAlertHistory` | 告警歷史 | ✅ |
| `ESHealthCheckResult` | 健康檢查結果 | ✅ (含 ClusterHealth) |
| `ESMonitorStatus` | 監控狀態 | ✅ (含 RelocatingShards) |
| `ESMetricTimeSeries` | 時序資料 | ✅ |
| `ESStatistics` | 統計資料 | ✅ |
| `ESAlertThreshold` | 告警閾值 | ✅ |

---

## ✅ Phase 2: 核心功能 (100%)

### 2.1 API 端點 ✅

**檔案**: `controller/elasticsearch.go`, `router/router.go:138-157`

#### 監控配置管理 (CRUD)

| 端點 | 方法 | 狀態 | 功能 | 權限 | 檔案位置 |
|------|------|------|------|------|----------|
| `/monitors` | GET | ✅ | 獲取所有監控配置 | elasticsearch:read | controller:20-28 |
| `/monitors/{id}` | GET | ✅ | 獲取單一監控配置 | elasticsearch:read | controller:30-50 |
| `/monitors` | POST | ✅ | 創建監控配置 | elasticsearch:create | controller:52-68 |
| `/monitors` | PUT | ✅ | 更新監控配置 | elasticsearch:update | controller:70-91 |
| `/monitors/{id}` | DELETE | ✅ | 刪除監控配置 | elasticsearch:delete | controller:93-114 |

**Service 層**: `services/es_monitor_service.go:12-210`

---

#### 監控操作

| 端點 | 方法 | 狀態 | 功能 | 權限 | 檔案位置 |
|------|------|------|------|------|----------|
| `/monitors/{id}/test` | POST | ✅ | 測試 ES 連線 | elasticsearch:read | controller:135-162 |
| `/monitors/{id}/toggle` | POST | ✅ | 啟用/停用監控 | elasticsearch:update | controller:164-193 |

**特點**:
- ✅ `/test` 端點從路徑取 ID（已統一）
- ✅ 自動整合排程器（create/update/delete/toggle 時自動管理）

---

#### 監控狀態與查詢

| 端點 | 方法 | 狀態 | 功能 | 權限 | 檔案位置 |
|------|------|------|------|------|----------|
| `/status` | GET | ✅ | 獲取所有監控器狀態 | elasticsearch:read | controller:205-222 |
| `/status/{id}/history` | GET | ✅ | 獲取單一監控器歷史 | elasticsearch:read | controller:258-309 |
| `/statistics` | GET | ✅ | 獲取 ES 監控統計 | elasticsearch:read | controller:230-247 |

**Query Service**: `services/es_monitor_query.go`

**特點**:
- ✅ `/status` 輸出包含完整 shards 資料（含 relocating_shards）
- ✅ `/history` 支援時間範圍與間隔參數
- ✅ `/statistics` 提供儀表板統計資料

---

### 2.2 核心服務 ✅

#### 健康檢查服務
**檔案**: `services/es_monitor.go`

| 功能 | 方法 | 狀態 | 說明 |
|------|------|------|------|
| 健康檢查 | `CheckESHealth()` | ✅ | 完整健康檢查流程 |
| 集群健康 | `getClusterHealth()` | ✅ | `/_cluster/health` |
| 節點統計 | `getNodeStats()` | ✅ | `/_nodes/stats/os,jvm,fs,indices` |
| 集群統計 | `getClusterStats()` | ✅ | `/_cluster/stats` |
| 索引統計 | `getIndicesStats()` | ✅ | `/_stats` |
| 指標解析 | `ParseMetricsFromCheckResult()` | ✅ | 23 個指標提取 |
| 告警檢查 | `CheckAlertConditions()` | ✅ | 8 種告警規則 |

**指標提取函數** (全部 ✅):
- `extractNodeCount()` - 節點總數
- `extractDataNodeCount()` - 數據節點數
- `extractCPUUsage()` - CPU 使用率
- `extractMemoryUsage()` - 記憶體使用率
- `extractDiskUsage()` - 磁碟使用率
- `extractActiveShards()` - 活躍分片數 ✅ **已修復**
- `extractRelocatingShards()` - 遷移中分片數 ✅ **已修復**
- `extractUnassignedShards()` - 未分配分片數 ✅ **已修復**
- `extractTotalIndices()` - 索引總數
- `extractTotalDocuments()` - 文檔總數
- `extractTotalSizeBytes()` - 總大小
- `extractQueryLatency()` - 查詢延遲
- `extractIndexingRate()` - 索引並發數（index_current，非速率）
- `extractSearchRate()` - 搜尋並發數（query_current，非速率）

---

#### 自動排程服務 ✅
**檔案**: `services/es_scheduler.go`
**整合**: `main.go:75-79`

| 功能 | 方法 | 狀態 | 說明 |
|------|------|------|------|
| 初始化 | `InitESScheduler()` | ✅ | 單例模式 |
| 載入監控 | `LoadAllMonitors()` | ✅ | 啟動時載入所有啟用的監控 |
| 啟動監控 | `StartMonitor()` | ✅ | 為每個監控創建獨立 ticker |
| 停止監控 | `StopMonitor()` | ✅ | 停止指定監控 |
| 重啟監控 | `RestartMonitor()` | ✅ | 更新配置後重啟 |
| 獲取狀態 | `GetRunningMonitors()` | ✅ | 獲取運行中的監控列表 |
| 停止所有 | `StopAll()` | ✅ | 應用關閉時清理 |

**特點**:
- ✅ 每個監控器獨立運行
- ✅ 根據 `interval` 設定動態調整執行頻率
- ✅ 立即執行 + 定期執行
- ✅ 與 CRUD 操作完全整合

**運作狀態**: 🟢 **已驗證正常運作** (2025-10-08)
```
INFO: Starting ES monitor check for: ES-213
INFO: ES monitor check completed for: ES-213, status: online
INFO: ✅ Successfully flushed 2/2 ES metrics to TimescaleDB
```

---

#### 批量寫入服務 ✅
**檔案**: `services/batch_writer.go`

| 功能 | 狀態 | 說明 |
|------|------|------|
| 設備指標批量寫入 | ✅ | 原有功能 |
| **ES 指標批量寫入** | ✅ | **已整合** |
| 類型切換 | ✅ | 使用 type assertion |
| 自動刷新 | ✅ | 達到 batch size 或間隔時間 |
| Prepared Statement | ✅ | 23 個參數 |

**運作狀態**: 🟢 **已驗證正常運作**

---

#### 查詢服務 ✅
**檔案**: `services/es_monitor_query.go`

| 功能 | 方法 | 狀態 | 說明 |
|------|------|------|------|
| 獲取最新指標 | `GetLatestMetrics()` | ✅ | 最近 1 小時內的最新指標 |
| 獲取所有狀態 | `GetAllMonitorsStatus()` | ✅ | 所有監控器當前狀態 |
| 獲取時序資料 | `GetMetricsTimeSeries()` | ✅ | 時間範圍 + 時間桶聚合 |
| 獲取統計資料 | `GetESStatistics()` | ✅ | 儀表板統計 |

---

### 2.3 權限系統 ✅

**檔案**: `services/auth.go:218-221`

| 權限 | 資源 | 操作 | 狀態 |
|------|------|------|------|
| elasticsearch:create | elasticsearch | create | ✅ |
| elasticsearch:read | elasticsearch | read | ✅ |
| elasticsearch:update | elasticsearch | update | ✅ |
| elasticsearch:delete | elasticsearch | delete | ✅ |

**預設授權**: Admin 角色自動擁有所有 ES 權限

---

### 2.4 文件 ✅

| 文件 | 路徑 | 狀態 | 說明 |
|------|------|------|------|
| OpenAPI 規格 | `spec/api/openapi.yml` | ✅ | 完整 API 定義 |
| API 詳細規格 | `spec/api/elasticsearch-api-spec.md` | ✅ | 端點詳細說明 |
| 前端對接指南 | `guides/frontend/api-integration.md` | ✅ | 前端開發參考 |
| 實作指南 | `guides/implementation/elasticsearch-setup.md` | ✅ | 後端實作說明 |
| 故障排除 | `troubleshooting/` | ✅ | 4 份故障排除文件 |

---

## ✅ Phase 3: 進階功能 (100%)

### 3.1 告警管理 API ✅

**已完整實作且已修復**

| 端點 | 方法 | 狀態 | 功能 | 檔案位置 |
|------|------|------|------|----------|
| `/alerts` | GET | ✅ | 獲取告警列表（支援過濾、分頁） | `controller/elasticsearch.go:329` |
| `/alerts/{monitor_id}` | GET | ✅ | 獲取單一告警 | `controller/elasticsearch.go:392` |
| `/alerts/{monitor_id}/resolve` | POST | ✅ | 標記告警為已解決 | `controller/elasticsearch.go:437` |
| `/alerts/{monitor_id}/acknowledge` | PUT | ✅ | 確認告警 | `controller/elasticsearch.go:496` |

**已實作**:
- ✅ Controller: `controller/elasticsearch.go` (4 個端點函數)
- ✅ Service: `services/es_alert_service.go` (完整告警服務)
- ✅ Routes: `router/router.go:160-163` (路由註冊)
- ✅ Models: `models/response.go` (ESAlertQueryParams, ESAlertStatistics)
- ✅ 查詢參數支援: status[], severity[], alert_type[], monitor_id, start_time, end_time, page, page_size
- ✅ 權限控制: 使用 elasticsearch:read/update 權限

**已修復問題** (2025-10-22):
1. ✅ **NULL 值掃描錯誤**
   - **問題**: 資料庫 NULL 值無法直接掃描到 string 類型
   - **修復**: 使用 `sql.NullString` 處理可空欄位 (cluster_name, resolved_by, resolution_note, acknowledged_by, metadata)
   - **影響**: `GetAlerts()` 和 `GetAlertByID()` 從 500 錯誤恢復正常

2. ✅ **PostgreSQL 數組參數錯誤**
   - **問題**: `[]string` 無法直接作為 PostgreSQL ANY() 參數
   - **修復**: 使用 `pq.Array()` 包裝數組參數 (status[], severity[], alert_type[])
   - **影響**: 帶過濾條件的查詢從 500 錯誤恢復正常

---

### 3.2 告警通知功能 ✅

**檔案**: `services/es_monitor.go:690-738`

**已完整實作**:
```go
func (s *ESMonitorService) SendAlertNotification(monitor entities.ElasticsearchMonitor, alert entities.ESAlert) {
    // 構建告警郵件主題和內容
    // 發送郵件給所有收件人
    Mail4(monitor.Receivers, nil, nil, subject, monitor.Name, details)
}
```

**已實作功能**:
| 功能 | 狀態 | 說明 |
|------|------|------|
| Email 通知 | ✅ | 整合現有 Mail4 服務，支援 HTML 格式 |
| 告警去重 | ✅ | 可配置時間窗口（預設 5 分鐘） |
| 去重時間窗口配置 | ✅ | 新增 `alert_dedupe_window` 欄位，每個監控器可獨立配置 |
| 告警類型過濾 | ✅ | 根據 `check_type` 配置動態啟用告警類型 |
| 告警詳情 | ✅ | 包含監控名稱、集群名稱、告警類型、嚴重程度、閾值/實際值 |
| 自定義主題 | ✅ | 支援 `subject` 欄位自定義告警主題 |

**去重機制**:
- 時間窗口: 用戶可配置（預設 300 秒）
- 去重條件: monitor_id + alert_type + severity + status='active'
- 實作位置: `services/es_monitor.go:673-703`
- 配置欄位: `entities.ElasticsearchMonitor.AlertDedupeWindow`

**當前行為**:
- ✅ 告警條件檢測正常
- ✅ 告警類型過濾（health/performance/capacity）
- ✅ 告警寫入 `es_alert_history` 表（帶去重檢查）
- ✅ 實際發送 Email 通知
- ✅ 重複告警自動跳過（可配置時間窗口內）

**日誌範例**:
```
WARN  2025/10/08 15:03:05 ES Alert Created [high][performance]: Memory usage high: 88.89%
INFO  2025/10/08 15:03:07 Alert notification sent to 1 receivers for monitor: ES-93
DEBUG 2025/10/08 15:03:35 Skipping duplicate alert for monitor 2: Memory usage high: 88.86%
```

---

### 3.3 告警規則引擎 ✅

**檔案**: `services/es_monitor.go:493-664`

**已實作規則** (10 種，按三大類劃分):

#### 🟢 健康類告警 (Health) - 2 種

| 規則 | 嚴重性 | 告警類型 | 觸發條件 | 狀態 | 代碼行數 |
|------|--------|----------|---------|------|----------|
| 集群狀態 RED | Critical | health | cluster_status == "red" | ✅ | 650-661 |
| 未分配分片 | High | health | unassigned_shards >= threshold | ✅ | 633-648 |

**說明**:
- 只有啟用 `check_type: "health"` 才會檢測
- 集群 RED 狀態表示部分主分片不可用，數據可能丟失
- 未分配分片可能導致數據不完整或無法訪問

---

#### 🔵 性能類告警 (Performance) - 6 種

| 規則 | 嚴重性 | 告警類型 | 觸發條件 | 狀態 | 代碼行數 |
|------|--------|----------|---------|------|----------|
| CPU 使用率危險 | Critical | performance | cpu_usage >= critical_threshold (預設 85%) | ✅ | 510-523 |
| CPU 使用率高 | High | performance | cpu_usage >= high_threshold (預設 75%) | ✅ | 524-538 |
| 記憶體使用率危險 | Critical | performance | memory_usage >= critical_threshold (預設 90%) | ✅ | 541-554 |
| 記憶體使用率高 | High | performance | memory_usage >= high_threshold (預設 80%) | ✅ | 555-569 |
| 響應時間危險 | Critical | performance | response_time >= critical_threshold (預設 10000ms) | ✅ | 603-616 |
| 響應時間高 | High | performance | response_time >= high_threshold (預設 3000ms) | ✅ | 617-631 |

**說明**:
- 只有啟用 `check_type: "performance"` 才會檢測
- CPU/記憶體使用率高可能導致查詢變慢、索引速度下降
- 響應時間過長會影響用戶體驗

---

#### 🟡 容量類告警 (Capacity) - 2 種

| 規則 | 嚴重性 | 告警類型 | 觸發條件 | 狀態 | 代碼行數 |
|------|--------|----------|---------|------|----------|
| 磁碟使用率危險 | Critical | capacity | disk_usage >= critical_threshold (預設 95%) | ✅ | 572-585 |
| 磁碟使用率高 | High | capacity | disk_usage >= high_threshold (預設 85%) | ✅ | 586-600 |

**說明**:
- 只有啟用 `check_type: "capacity"` 才會檢測
- 磁碟空間不足會導致無法寫入新數據
- 建議在達到 85% 時開始清理或擴容

---

### 3.4 告警閾值配置 ✅

**檔案**: `entities/elasticsearch.go:26-39`, `entities/elasticsearch.go:47-89`

**配置方式** (兩種，優先級由高到低):

#### 方式 1: 獨立欄位配置（推薦）✅

```go
// ElasticsearchMonitor 結構
CPUUsageHigh            *float64  // CPU使用率-高閾值(%)
CPUUsageCritical        *float64  // CPU使用率-危險閾值(%)
MemoryUsageHigh         *float64  // 記憶體使用率-高閾值(%)
MemoryUsageCritical     *float64  // 記憶體使用率-危險閾值(%)
DiskUsageHigh           *float64  // 磁碟使用率-高閾值(%)
DiskUsageCritical       *float64  // 磁碟使用率-危險閾值(%)
ResponseTimeHigh        *int64    // 響應時間-高閾值(ms)
ResponseTimeCritical    *int64    // 響應時間-危險閾值(ms)
UnassignedShardsThreshold *int    // 未分配分片閾值
```

**優點**:
- ✅ 前端友好，可使用數字輸入框、滑桿等控件
- ✅ 每個欄位可獨立驗證
- ✅ 不需要了解 JSON 格式
- ✅ 支援 NULL（使用預設值）

---

#### 方式 2: JSON 配置（高級選項，向後兼容）✅

```go
AlertThreshold string  // JSON 格式: {"cpu_usage_high":75.0,...}
```

**用途**:
- 向後兼容舊版本
- 批量配置腳本
- 如果獨立欄位已設置，此欄位將被忽略

---

**預設閾值**: `entities/elasticsearch.go:221-234`
```go
func DefaultESAlertThreshold() ESAlertThreshold {
    return ESAlertThreshold{
        CPUUsageHigh:         75.0,   // 75%
        CPUUsageCritical:     85.0,   // 85%
        MemoryUsageHigh:      80.0,   // 80%
        MemoryUsageCritical:  90.0,   // 90%
        DiskUsageHigh:        85.0,   // 85%
        DiskUsageCritical:    95.0,   // 95%
        ResponseTimeHigh:     3000,   // 3000ms
        ResponseTimeCritical: 10000,  // 10000ms
        UnassignedShards:     1,      // 1個
    }
}
```

**配置優先級**:
1. 獨立欄位（最高優先級） - 如果設置了 `cpu_usage_high` 等欄位
2. JSON 配置（向後兼容） - 如果獨立欄位未設置，解析 `alert_threshold`
3. 預設值（最低優先級） - 以上都沒有時使用

**實作方法**: `entities/elasticsearch.go:47-89`
```go
func (m *ElasticsearchMonitor) GetAlertThreshold() ESAlertThreshold {
    // 優先使用獨立欄位 → JSON 配置 → 預設值
}
```

---

## ❌ Phase 4: 優化與完善 (0%)

### 4.1 效能優化 ❌

| 功能 | 狀態 | 說明 |
|------|------|------|
| Redis 快取 | ❌ | 快取最新指標減少資料庫查詢 |
| 連線池優化 | ❌ | ES HTTP 連線復用 |
| 查詢效能優化 | ❌ | 優化時序資料查詢 |
| 指標聚合優化 | ❌ | Continuous Aggregates |

---

### 4.2 進階分析 ❌

| 功能 | 狀態 | 說明 |
|------|------|------|
| 趨勢分析 | ❌ | CPU/Memory/Disk 使用趨勢預測 |
| 異常檢測 | ❌ | 自動檢測異常指標波動 |
| 容量規劃 | ❌ | 根據歷史資料預測容量需求 |
| 效能基線 | ❌ | 建立正常運作基線 |

---

### 4.3 整合與擴展 ❌

| 功能 | 狀態 | 說明 |
|------|------|------|
| Grafana 整合 | ❌ | 提供 Grafana 資料源 |
| Prometheus 匯出 | ❌ | Prometheus Exporter |
| Slack 整合 | ❌ | 告警發送到 Slack |
| Teams 整合 | ❌ | 告警發送到 Microsoft Teams |

---

## 🎯 待辦事項優先級

### ✅ 已完成 (Phase 3)

1. ✅ **告警管理 API** (4 個端點)
   - GetESAlerts, GetESAlertByID, ResolveESAlert, AcknowledgeESAlert
   - 完成時間: 2025-10-08

2. ✅ **Email 告警通知**
   - 整合 Mail4 服務，包含集群名稱、閾值、實際值
   - 完成時間: 2025-10-08

3. ✅ **告警去重與可配置時間窗口**
   - 避免告警風暴，每個監控器可獨立配置去重時間
   - 完成時間: 2025-10-08

4. ✅ **告警類型過濾**
   - 根據 check_type 配置動態啟用告警類型（health/performance/capacity）
   - 完成時間: 2025-10-08

5. ✅ **用戶友好的閾值配置**
   - 9 個獨立欄位 + JSON 向後兼容 + 預設值
   - 完成時間: 2025-10-10

---

### 🟡 中優先級 (使用體驗提升)

6. **Webhook 通知**
   - 整合外部系統（PagerDuty, OpsGenie 等）
   - 預估時間: 2 小時

7. **告警模板自定義**
   - 允許自定義告警訊息格式
   - 預估時間: 2 小時

8. **Redis 快取**
   - 提升查詢效能
   - 預估時間: 4 小時

9. **Cluster Yellow 告警**
   - 補充黃色狀態告警（Medium 級別）
   - 預估時間: 0.5 小時

---

### 🟢 低優先級 (長期優化)

10. **進階分析功能**
    - 趨勢分析、異常檢測
    - 預估時間: 10+ 小時

11. **第三方整合**
    - Grafana, Prometheus, Slack
    - 預估時間: 8+ 小時

---

## 📈 進度里程碑

### ✅ 已完成 (2025-10-10)

- [x] **Phase 1: 基礎架構** (100%)
  - [x] MySQL 監控配置表
  - [x] TimescaleDB 指標與告警表
  - [x] Entity 定義完整
  - [x] GORM AutoMigrate

- [x] **Phase 2: 核心功能** (100%)
  - [x] 自動排程系統
  - [x] 健康檢查邏輯
  - [x] 指標收集與儲存 (23 個指標)
  - [x] 批量寫入整合
  - [x] 查詢 API
  - [x] shards 欄位資料修復

- [x] **Phase 3: 進階功能** (100%)
  - [x] 告警規則引擎 (10 種規則，三大類)
  - [x] 告警寫入資料庫（帶去重）
  - [x] 告警管理 API (4 個端點)
  - [x] Email 告警通知
  - [x] 可配置去重時間窗口
  - [x] 告警類型過濾 (health/performance/capacity)
  - [x] 用戶友好的閾值配置（9 個獨立欄位）
  - [x] 配置優先級（獨立欄位 → JSON → 預設值）

### 📅 規劃中

- [ ] **Phase 4: 優化與完善** (0%)
  - [ ] Redis 快取
  - [ ] Webhook 通知
  - [ ] 進階分析功能
  - [ ] 第三方整合

---

## 🔗 相關文件

### API 規格與文檔
- [OpenAPI 規格](./openapi.yml)
- [Elasticsearch API 詳細規格](./elasticsearch-api-spec.md)
- [資料庫 Schema 驗證](../database/schema-validation.md)

### 實作指南
- [前端對接指南](../../guides/frontend/api-integration.md)
- [後端實作指南](../../guides/implementation/elasticsearch-setup.md)
- [告警閾值配置指南](../../guides/implementation/alert-threshold-configuration.md) ⭐ **新增**

### 故障排除
- [無資料問題診斷](../../troubleshooting/monitoring/no-data-diagnosis.md)
- [權限問題修復](../../troubleshooting/monitoring/es-permissions-fix.md)
- [告警去重功能更新](../../CHANGELOG-ES-ALERT-DEDUPE.md) ⭐ **新增**

### SQL 腳本
- [添加告警去重時間窗口欄位](../../troubleshooting/add_alert_dedupe_window.sql)
- [添加告警閾值獨立欄位](../../troubleshooting/add_threshold_fields.sql) ⭐ **新增**

---

**版本歷史**:
- **v1.3 (2025-10-22)** - 🔧 修復：告警 API NULL 值處理、PostgreSQL 數組參數綁定（pq.Array）
- v1.2 (2025-10-10) - ✅ Phase 3 完成：告警閾值配置優化（9 個獨立欄位）、告警規則按三大類劃分（10 種規則）
- v1.1 (2025-10-08) - 告警管理 API、Email 通知、告警去重、自動排程、修復 shards 資料
- v1.0 (2025-10-07) - 初始版本：基礎架構與核心功能

**下次更新預估**: Phase 4 啟動時
