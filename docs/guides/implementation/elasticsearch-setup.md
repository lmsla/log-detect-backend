# 🔧 Elasticsearch 監控系統實作指南

## 📋 實作概述

本文檔提供 Elasticsearch 監控系統的詳細實作指南，包括資料庫結構、程式碼結構、實作步驟和測試方法。

## 🗄️ 精簡雙層資料庫架構

### **Layer 1: MySQL (配置數據)**

```sql
-- ES 監控配置表 (保留在 MySQL，數據量小，低頻讀寫)
CREATE TABLE elasticsearch_monitors (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INT NOT NULL,
    username VARCHAR(100),
    password VARCHAR(255),
    enable_auth BOOLEAN DEFAULT FALSE,
    check_type VARCHAR(100) NOT NULL,
    interval_seconds INT DEFAULT 60,
    enable_monitor BOOLEAN DEFAULT TRUE,
    receivers JSON,
    subject VARCHAR(200),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_name (name),
    INDEX idx_enable (enable_monitor)
);

-- Cron 任務關聯表
CREATE TABLE elasticsearch_cron_jobs (
    id INT PRIMARY KEY AUTO_INCREMENT,
    monitor_id INT NOT NULL,
    entry_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (monitor_id) REFERENCES elasticsearch_monitors(id) ON DELETE CASCADE,
    UNIQUE KEY unique_monitor (monitor_id)
);
```

### **Layer 2: TimescaleDB (高頻時間序列數據)**

```sql
-- ES 監控指標時間序列表 (高頻寫入，優化查詢)
CREATE TABLE es_metrics (
    time TIMESTAMPTZ NOT NULL,
    monitor_id INTEGER NOT NULL,
    status TEXT NOT NULL,
    cluster_name TEXT,
    cluster_status TEXT,
    response_time BIGINT DEFAULT 0,
    cpu_usage DECIMAL(5,2) DEFAULT 0.00,
    memory_usage DECIMAL(5,2) DEFAULT 0.00,
    disk_usage DECIMAL(5,2) DEFAULT 0.00,
    node_count INTEGER DEFAULT 0,
    data_node_count INTEGER DEFAULT 0,
    query_latency BIGINT DEFAULT 0,
    indexing_rate DECIMAL(10,2) DEFAULT 0.00,
    search_rate DECIMAL(10,2) DEFAULT 0.00,
    total_indices INTEGER DEFAULT 0,
    total_documents BIGINT DEFAULT 0,
    total_size_bytes BIGINT DEFAULT 0,
    active_shards INTEGER DEFAULT 0,
    relocating_shards INTEGER DEFAULT 0,
    unassigned_shards INTEGER DEFAULT 0,
    error_message TEXT,
    warning_message TEXT,
    metadata JSONB
);

-- 轉換為時間序列表
SELECT create_hypertable('es_metrics', 'time', chunk_time_interval => INTERVAL '1 day');

-- 自動壓縮策略 (節省空間)
ALTER TABLE es_metrics SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'monitor_id',
    timescaledb.compress_orderby = 'time DESC'
);
SELECT add_compression_policy('es_metrics', INTERVAL '7 days');

-- 自動清理策略 (保留3個月)
SELECT add_retention_policy('es_metrics', INTERVAL '90 days');

-- 告警歷史表
CREATE TABLE es_alert_history (
    time TIMESTAMPTZ NOT NULL,
    monitor_id INTEGER NOT NULL,
    alert_type TEXT NOT NULL,
    severity TEXT NOT NULL,
    message TEXT NOT NULL,
    status TEXT DEFAULT 'active',
    resolved_at TIMESTAMPTZ,
    resolution_note TEXT
);

SELECT create_hypertable('es_alert_history', 'time', chunk_time_interval => INTERVAL '7 days');
SELECT add_retention_policy('es_alert_history', INTERVAL '90 days');

-- 高性能索引
CREATE INDEX idx_es_metrics_monitor_time ON es_metrics (monitor_id, time DESC);
CREATE INDEX idx_es_metrics_status ON es_metrics (status, time DESC);
CREATE INDEX idx_es_alert_monitor_time ON es_alert_history (monitor_id, time DESC);
CREATE INDEX idx_es_alert_severity ON es_alert_history (severity, time DESC);
```

### **可選 Layer 3: Redis (熱數據緩存 - 高並發時啟用)**

```redis
# Redis 數據結構設計 (可選熱數據層，1小時內)

# 1. ES 監控最新狀態 (可選)
es:latest:{monitor_id} -> JSON (TTL: 1 hour)
{
  "status": "online",
  "cluster_status": "green",
  "response_time": 120,
  "cpu_usage": 45.5,
  "last_check": "2024-09-30T12:00:00Z"
}

# 2. ES 監控群組統計 (可選)
es:stats:summary -> HASH (TTL: 1 hour)
{
  "total_monitors": 5,
  "online_monitors": 4,
  "critical_alerts": 1,
  "last_update": 1696075200
}

# 注意：Redis 層為可選擴展，主要架構僅依賴 MySQL + TimescaleDB
```

### 權限表更新

```sql
-- 新增 ES 監控相關權限
INSERT INTO permissions (name, resource, action, description) VALUES
('elasticsearch:create', 'elasticsearch', 'create', 'Create Elasticsearch monitor'),
('elasticsearch:read', 'elasticsearch', 'read', 'Read Elasticsearch monitor data'),
('elasticsearch:update', 'elasticsearch', 'update', 'Update Elasticsearch monitor'),
('elasticsearch:delete', 'elasticsearch', 'delete', 'Delete Elasticsearch monitor');

-- 為 admin 角色新增權限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 1, id FROM permissions WHERE resource = 'elasticsearch';

-- 為 user 角色新增讀取權限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 2, id FROM permissions WHERE resource = 'elasticsearch' AND action = 'read';
```

## 📁 檔案結構

```
log-detect-backend/
├── entities/
│   └── elasticsearch.go          # ES 監控實體定義
├── models/
│   └── elasticsearch.go          # ES 監控模型
├── controller/
│   └── elasticsearch.go          # ES 監控控制器
├── services/
│   ├── elasticsearch_monitor.go  # ES 監控服務
│   ├── elasticsearch_health.go   # ES 健康檢查
│   ├── elasticsearch_alert.go    # ES 告警服務
│   └── elasticsearch_metrics.go  # ES 指標收集
├── middleware/
│   └── elasticsearch_auth.go     # ES 監控權限中介軟體
└── docs/
    ├── elasticsearch-monitoring.md
    ├── elasticsearch-api-spec.md
    └── elasticsearch-implementation-guide.md
```

## 🏗️ 實體定義 (entities/elasticsearch.go)

```go
package entities

import (
    "time"
    "log-detect/models"
)

// ElasticsearchMonitor ES監控配置
type ElasticsearchMonitor struct {
    models.Common
    ID            int       `gorm:"primaryKey;autoIncrement" json:"id"`
    Name          string    `gorm:"type:varchar(100);not null" json:"name" form:"name"`
    Host          string    `gorm:"type:varchar(255);not null" json:"host" form:"host"`
    Port          int       `gorm:"not null" json:"port" form:"port"`
    Username      string    `gorm:"type:varchar(100)" json:"username" form:"username"`
    Password      string    `gorm:"type:varchar(255)" json:"-" form:"password"`
    EnableAuth    bool      `gorm:"default:false" json:"enable_auth" form:"enable_auth"`
    CheckType     string    `gorm:"type:varchar(100);not null" json:"check_type" form:"check_type"`
    IntervalSecs  int       `gorm:"default:60" json:"interval_seconds" form:"interval_seconds"`
    EnableMonitor bool      `gorm:"default:true" json:"enable_monitor" form:"enable_monitor"`
    Receivers     []string  `gorm:"serializer:json" json:"receivers" form:"receivers"`
    Subject       string    `gorm:"type:varchar(200)" json:"subject" form:"subject"`
}

// ESMetrics TimescaleDB 時間序列數據結構
type ESMetrics struct {
    Time             time.Time `json:"time"`
    MonitorID        int       `json:"monitor_id"`
    Status           string    `json:"status"`
    ClusterName      string    `json:"cluster_name"`
    ClusterStatus    string    `json:"cluster_status"`
    ResponseTime     int64     `json:"response_time"`
    CpuUsage         float64   `json:"cpu_usage"`
    MemoryUsage      float64   `json:"memory_usage"`
    DiskUsage        float64   `json:"disk_usage"`
    NodeCount        int       `json:"node_count"`
    DataNodeCount    int       `json:"data_node_count"`
    QueryLatency     int64     `json:"query_latency"`
    IndexingRate     float64   `json:"indexing_rate"`
    SearchRate       float64   `json:"search_rate"`
    TotalIndices     int       `json:"total_indices"`
    TotalDocuments   int64     `json:"total_documents"`
    TotalSizeBytes   int64     `json:"total_size_bytes"`
    ActiveShards     int       `json:"active_shards"`
    RelocatingShards int       `json:"relocating_shards"`
    UnassignedShards int       `json:"unassigned_shards"`
    ErrorMessage     string    `json:"error_message"`
    WarningMessage   string    `json:"warning_message"`
    Metadata         string    `json:"metadata"` // JSONB 格式
}

// ESAlert TimescaleDB 告警歷史
type ESAlert struct {
    Time           time.Time  `json:"time"`
    MonitorID      int        `json:"monitor_id"`
    AlertType      string     `json:"alert_type"`
    Severity       string     `json:"severity"`
    Message        string     `json:"message"`
    Status         string     `json:"status"`
    ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
    ResolutionNote string     `json:"resolution_note"`
}

// ESCacheData 內存緩存數據結構 (可選 Redis 替代)
type ESCacheData struct {
    MonitorID     int     `json:"monitor_id"`
    Status        string  `json:"status"`
    ClusterStatus string  `json:"cluster_status"`
    ResponseTime  int64   `json:"response_time"`
    CpuUsage      float64 `json:"cpu_usage"`
    LastCheck     string  `json:"last_check"`
}

// ElasticsearchCronJob ES Cron任務記錄
type ElasticsearchCronJob struct {
    models.Common
    ID        int `gorm:"primaryKey;autoIncrement" json:"id"`
    MonitorID int `gorm:"not null;uniqueIndex" json:"monitor_id"`
    EntryID   int `gorm:"not null" json:"entry_id"`
}

// 表名設定
func (ElasticsearchMonitor) TableName() string {
    return "elasticsearch_monitors"
}

// 注意：TimescaleDB 的表名在 CREATE TABLE 語句中定義，無需 TableName() 方法

func (ElasticsearchCronJob) TableName() string {
    return "elasticsearch_cron_jobs"
}
```

## 🎛️ 控制器實作 (controller/elasticsearch.go)

```go
package controller

import (
    "net/http"
    "strconv"
    "log-detect/entities"
    "log-detect/services"
    "log-detect/models"

    "github.com/gin-gonic/gin"
)

type ElasticsearchController struct {
    esService    *services.ElasticsearchService
    alertService *services.ESAlertService
}

func NewElasticsearchController() *ElasticsearchController {
    return &ElasticsearchController{
        esService:    services.NewElasticsearchService(),
        alertService: services.NewESAlertService(),
    }
}

// @Summary 獲取所有ES監控配置
// @Tags Elasticsearch
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "頁碼"
// @Param limit query int false "每頁筆數"
// @Success 200 {object} models.Response
// @Router /api/v1/elasticsearch/monitors [get]
func (ctrl *ElasticsearchController) GetMonitors(c *gin.Context) {
    // 檢查權限
    if !ctrl.checkPermission(c, "elasticsearch:read") {
        return
    }

    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
    search := c.Query("search")
    enable := c.Query("enable")

    monitors, total, err := ctrl.esService.GetMonitors(page, limit, search, enable)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.Response{
            Code:    500,
            Message: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, models.Response{
        Code:    200,
        Message: "Success",
        Data: gin.H{
            "monitors": monitors,
            "total":    total,
            "page":     page,
            "limit":    limit,
        },
    })
}

// @Summary 新增ES監控配置
// @Tags Elasticsearch
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param monitor body entities.ElasticsearchMonitor true "監控配置"
// @Success 201 {object} models.Response
// @Router /api/v1/elasticsearch/monitors [post]
func (ctrl *ElasticsearchController) CreateMonitor(c *gin.Context) {
    if !ctrl.checkPermission(c, "elasticsearch:create") {
        return
    }

    var monitor entities.ElasticsearchMonitor
    if err := c.ShouldBindJSON(&monitor); err != nil {
        c.JSON(http.StatusBadRequest, models.Response{
            Code:    400,
            Message: "Invalid request data",
        })
        return
    }

    // 驗證數據
    if err := ctrl.validateMonitor(&monitor); err != nil {
        c.JSON(http.StatusBadRequest, models.Response{
            Code:    400,
            Message: err.Error(),
        })
        return
    }

    createdMonitor, err := ctrl.esService.CreateMonitor(&monitor)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.Response{
            Code:    500,
            Message: err.Error(),
        })
        return
    }

    c.JSON(http.StatusCreated, models.Response{
        Code:    201,
        Message: "Monitor created successfully",
        Data:    createdMonitor,
    })
}

// @Summary 測試ES連接
// @Tags Elasticsearch
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "監控ID"
// @Success 200 {object} models.Response
// @Router /api/v1/elasticsearch/monitors/{id}/test [post]
func (ctrl *ElasticsearchController) TestConnection(c *gin.Context) {
    if !ctrl.checkPermission(c, "elasticsearch:read") {
        return
    }

    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, models.Response{
            Code:    400,
            Message: "Invalid monitor ID",
        })
        return
    }

    result, err := ctrl.esService.TestConnection(id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.Response{
            Code:    500,
            Message: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, models.Response{
        Code:    200,
        Message: "Connection test completed",
        Data:    result,
    })
}

// 權限檢查輔助函數
func (ctrl *ElasticsearchController) checkPermission(c *gin.Context, permission string) bool {
    // 從 JWT middleware 獲取用戶信息
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, models.Response{
            Code:    401,
            Message: "Unauthorized",
        })
        return false
    }

    // 檢查權限
    authService := services.NewAuthService()
    hasPermission, err := authService.CheckUserPermission(userID.(uint), permission)
    if err != nil || !hasPermission {
        c.JSON(http.StatusForbidden, models.Response{
            Code:    403,
            Message: "Insufficient permissions",
        })
        return false
    }

    return true
}

// 數據驗證輔助函數
func (ctrl *ElasticsearchController) validateMonitor(monitor *entities.ElasticsearchMonitor) error {
    if monitor.Name == "" {
        return fmt.Errorf("name is required")
    }
    if monitor.Host == "" {
        return fmt.Errorf("host is required")
    }
    if monitor.Port <= 0 || monitor.Port > 65535 {
        return fmt.Errorf("port must be between 1 and 65535")
    }
    if monitor.IntervalSecs < 30 {
        return fmt.Errorf("interval must be at least 30 seconds")
    }
    if len(monitor.Receivers) == 0 {
        return fmt.Errorf("at least one receiver is required")
    }

    return nil
}
```

## 🔧 服務層實作

### ES 監控服務 (services/elasticsearch_monitor.go)

```go
package services

import (
    "fmt"
    "time"
    "log-detect/entities"
    "log-detect/global"

    "gorm.io/gorm"
)

type ElasticsearchService struct {
    db *gorm.DB
}

func NewElasticsearchService() *ElasticsearchService {
    return &ElasticsearchService{
        db: global.Mysql,
    }
}

// 獲取監控配置列表
func (s *ElasticsearchService) GetMonitors(page, limit int, search, enable string) ([]entities.ElasticsearchMonitor, int64, error) {
    var monitors []entities.ElasticsearchMonitor
    var total int64

    query := s.db.Model(&entities.ElasticsearchMonitor{})

    // 搜尋過濾
    if search != "" {
        query = query.Where("name LIKE ?", "%"+search+"%")
    }

    // 狀態過濾
    if enable != "" {
        enableBool := enable == "true"
        query = query.Where("enable_monitor = ?", enableBool)
    }

    // 計算總數
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // 分頁查詢
    offset := (page - 1) * limit
    if err := query.Offset(offset).Limit(limit).Find(&monitors).Error; err != nil {
        return nil, 0, err
    }

    return monitors, total, nil
}

// 新增監控配置
func (s *ElasticsearchService) CreateMonitor(monitor *entities.ElasticsearchMonitor) (*entities.ElasticsearchMonitor, error) {
    // 檢查名稱是否重複
    var count int64
    s.db.Model(&entities.ElasticsearchMonitor{}).Where("name = ?", monitor.Name).Count(&count)
    if count > 0 {
        return nil, fmt.Errorf("monitor name already exists")
    }

    if err := s.db.Create(monitor).Error; err != nil {
        return nil, err
    }

    // 如果啟用監控，立即建立 Cron 任務
    if monitor.EnableMonitor {
        if err := s.CreateCronJob(monitor); err != nil {
            // 記錄錯誤但不回滾，允許手動重啟
            fmt.Printf("Failed to create cron job for monitor %d: %v\n", monitor.ID, err)
        }
    }

    return monitor, nil
}

// 建立 Cron 任務
func (s *ElasticsearchService) CreateCronJob(monitor *entities.ElasticsearchMonitor) error {
    cronExpr := fmt.Sprintf("@every %ds", monitor.IntervalSecs)

    entryID, err := global.Cron.AddFunc(cronExpr, func() {
        s.PerformMonitorCheck(monitor.ID)
    })
    if err != nil {
        return err
    }

    // 記錄 Cron 任務關聯
    cronJob := &entities.ElasticsearchCronJob{
        MonitorID: monitor.ID,
        EntryID:   int(entryID),
    }

    return s.db.Create(cronJob).Error
}

// 執行監控檢查
func (s *ElasticsearchService) PerformMonitorCheck(monitorID int) {
    monitor, err := s.GetMonitorByID(monitorID)
    if err != nil {
        fmt.Printf("Failed to get monitor %d: %v\n", monitorID, err)
        return
    }

    if !monitor.EnableMonitor {
        return
    }

    // 執行健康檢查
    healthChecker := NewESHealthChecker()
    metrics, err := healthChecker.CheckHealth(monitor)
    if err != nil {
        fmt.Printf("Health check failed for monitor %d: %v\n", monitorID, err)
        return
    }

    // 儲存到 TimescaleDB (批量寫入)
    batchWriter := global.TimescaleBatchWriter
    batchWriter.AddESMetric(*metrics)

    // 可選：更新內存緩存
    if cacheAdapter := global.CacheAdapter; cacheAdapter != nil {
        cacheData := ESCacheData{
            MonitorID:     metrics.MonitorID,
            Status:        metrics.Status,
            ClusterStatus: metrics.ClusterStatus,
            ResponseTime:  metrics.ResponseTime,
            CpuUsage:      metrics.CpuUsage,
            LastCheck:     metrics.Time.Format(time.RFC3339),
        }
        cacheAdapter.SetESStatus(monitorID, cacheData)
    }

    // 檢查告警條件
    alertService := NewESAlertService()
    alerts := alertService.CheckAlertConditions(metrics, monitor)

    // 發送告警通知
    if len(alerts) > 0 {
        for _, alert := range alerts {
            if err := alertService.SendAlert(&alert, monitor); err != nil {
                fmt.Printf("Failed to send alert for monitor %d: %v\n", monitorID, err)
            }
        }
    }
}
```

### ES 健康檢查器 (services/elasticsearch_health.go)

```go
package services

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "log-detect/entities"

    "github.com/elastic/go-elasticsearch/v8"
)

type ESHealthChecker struct{}

func NewESHealthChecker() *ESHealthChecker {
    return &ESHealthChecker{}
}

// 執行 ES 健康檢查
func (hc *ESHealthChecker) CheckHealth(monitor *entities.ElasticsearchMonitor) (*entities.ESMetrics, error) {
    metrics := &entities.ESMetrics{
        Time:      time.Now(),
        MonitorID: monitor.ID,
    }

    // 建立 ES 客戶端
    client, err := hc.createESClient(monitor)
    if err != nil {
        metrics.Status = "error"
        metrics.ErrorMessage = fmt.Sprintf("Failed to create ES client: %v", err)
        return metrics, nil
    }

    // 檢查連接
    start := time.Now()
    if err := hc.checkConnection(client, metrics); err != nil {
        metrics.Status = "offline"
        metrics.ErrorMessage = err.Error()
        metrics.ResponseTime = time.Since(start).Milliseconds()
        return metrics, nil
    }
    metrics.ResponseTime = time.Since(start).Milliseconds()

    // 檢查集群健康
    if err := hc.checkClusterHealth(client, metrics); err != nil {
        metrics.WarningMessage += fmt.Sprintf("Cluster health check failed: %v; ", err)
    }

    // 檢查節點統計
    if err := hc.checkNodeStats(client, metrics); err != nil {
        metrics.WarningMessage += fmt.Sprintf("Node stats check failed: %v; ", err)
    }

    // 檢查索引統計
    if err := hc.checkIndexStats(client, metrics); err != nil {
        metrics.WarningMessage += fmt.Sprintf("Index stats check failed: %v; ", err)
    }

    // 判斷整體狀態
    hc.determineOverallStatus(metrics)

    return metrics, nil
}

// 建立 ES 客戶端
func (hc *ESHealthChecker) createESClient(monitor *entities.ElasticsearchMonitor) (*elasticsearch.Client, error) {
    cfg := elasticsearch.Config{
        Addresses: []string{fmt.Sprintf("%s:%d", monitor.Host, monitor.Port)},
    }

    if monitor.EnableAuth {
        cfg.Username = monitor.Username
        cfg.Password = monitor.Password
    }

    return elasticsearch.NewClient(cfg)
}

// 檢查連接
func (hc *ESHealthChecker) checkConnection(client *elasticsearch.Client, metrics *entities.ESMetrics) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    res, err := client.Info(client.Info.WithContext(ctx))
    if err != nil {
        return fmt.Errorf("connection failed: %v", err)
    }
    defer res.Body.Close()

    if res.IsError() {
        return fmt.Errorf("ES returned error: %s", res.Status())
    }

    return nil
}

// 檢查集群健康
func (hc *ESHealthChecker) checkClusterHealth(client *elasticsearch.Client, metrics *entities.ESMetrics) error {
    res, err := client.Cluster.Health()
    if err != nil {
        return err
    }
    defer res.Body.Close()

    var health map[string]interface{}
    if err := json.NewDecoder(res.Body).Decode(&health); err != nil {
        return err
    }

    if clusterName, ok := health["cluster_name"].(string); ok {
        metrics.ClusterName = clusterName
    }

    if clusterStatus, ok := health["status"].(string); ok {
        metrics.ClusterStatus = clusterStatus
    }

    if nodeCount, ok := health["number_of_nodes"].(float64); ok {
        metrics.NodeCount = int(nodeCount)
    }

    if dataNodeCount, ok := health["number_of_data_nodes"].(float64); ok {
        metrics.DataNodeCount = int(dataNodeCount)
    }

    if activeShards, ok := health["active_shards"].(float64); ok {
        metrics.ActiveShards = int(activeShards)
    }

    if relocatingShards, ok := health["relocating_shards"].(float64); ok {
        metrics.RelocatingShards = int(relocatingShards)
    }

    if unassignedShards, ok := health["unassigned_shards"].(float64); ok {
        metrics.UnassignedShards = int(unassignedShards)
    }

    return nil
}

// 判斷整體狀態
func (hc *ESHealthChecker) determineOverallStatus(metrics *entities.ESMetrics) {
    // 預設為 online
    metrics.Status = "online"

    // 檢查嚴重問題
    if metrics.ClusterStatus == "red" {
        metrics.Status = "error"
        return
    }

    if metrics.UnassignedShards > 0 || metrics.ClusterStatus == "yellow" {
        metrics.Status = "warning"
        return
    }

    // 檢查效能問題
    if metrics.ResponseTime > 5000 || metrics.CpuUsage > 80 || metrics.MemoryUsage > 85 {
        metrics.Status = "warning"
        return
    }
}
```

## 🚨 告警服務實作 (services/elasticsearch_alert.go)

```go
package services

import (
    "fmt"
    "time"
    "log-detect/entities"
    "log-detect/global"
)

type ESAlertService struct {
    db *gorm.DB
}

func NewESAlertService() *ESAlertService {
    return &ESAlertService{
        db: global.Mysql,
    }
}

// 檢查告警條件
func (s *ESAlertService) CheckAlertConditions(metrics *entities.ESMetrics, monitor *entities.ElasticsearchMonitor) []entities.ESAlert {
    var alerts []entities.ESAlert
    now := time.Now()

    // 連接失敗告警
    if metrics.Status == "offline" {
        alerts = append(alerts, entities.ESAlert{
            Time:      now,
            MonitorID: monitor.ID,
            AlertType: "connection",
            Severity:  "critical",
            Message:   fmt.Sprintf("Elasticsearch connection failed: %s", metrics.ErrorMessage),
            Status:    "active",
        })
    }

    // 集群狀態告警
    if metrics.ClusterStatus == "red" {
        alerts = append(alerts, entities.ESAlert{
            Time:      now,
            MonitorID: monitor.ID,
            AlertType: "cluster",
            Severity:  "high",
            Message:   "Cluster status is RED - data may be unavailable",
            Status:    "active",
        })
    } else if metrics.ClusterStatus == "yellow" {
        alerts = append(alerts, entities.ESAlert{
            Time:      now,
            MonitorID: monitor.ID,
            AlertType: "cluster",
            Severity:  "medium",
            Message:   "Cluster status is YELLOW - some replicas are unallocated",
            Status:    "active",
        })
    }

    // 性能告警
    if metrics.ResponseTime > 5000 {
        alerts = append(alerts, entities.ESAlert{
            Time:      now,
            MonitorID: monitor.ID,
            AlertType: "performance",
            Severity:  "medium",
            Message:   fmt.Sprintf("High response time: %dms", metrics.ResponseTime),
            Status:    "active",
        })
    }

    if metrics.CpuUsage > 80 {
        alerts = append(alerts, entities.ESAlert{
            Time:      now,
            MonitorID: monitor.ID,
            AlertType: "performance",
            Severity:  "medium",
            Message:   fmt.Sprintf("High CPU usage: %.2f%%", metrics.CpuUsage),
            Status:    "active",
        })
    }

    if metrics.DiskUsage > 90 {
        alerts = append(alerts, entities.ESAlert{
            Time:      now,
            MonitorID: monitor.ID,
            AlertType: "disk",
            Severity:  "high",
            Message:   fmt.Sprintf("High disk usage: %.2f%%", metrics.DiskUsage),
            Status:    "active",
        })
    }

    // 分片告警
    if metrics.UnassignedShards > 0 {
        alerts = append(alerts, entities.ESAlert{
            Time:      now,
            MonitorID: monitor.ID,
            AlertType: "shards",
            Severity:  "medium",
            Message:   fmt.Sprintf("%d unassigned shards detected", metrics.UnassignedShards),
            Status:    "active",
        })
    }

    return alerts
}

// 發送告警通知
func (s *ESAlertService) SendAlert(alert *entities.ESAlert, monitor *entities.ElasticsearchMonitor) error {
    // 檢查是否已經發送過相同告警（避免重複發送）
    if s.isDuplicateAlert(alert) {
        return nil
    }

    // 儲存告警記錄到 TimescaleDB
    batchWriter := global.TimescaleBatchWriter
    batchWriter.AddESAlert(*alert)

    // 準備郵件內容
    subject := fmt.Sprintf("[%s] %s", alert.Severity, monitor.Subject)
    body := s.buildAlertEmailBody(alert, monitor)

    // 發送郵件
    return SendMail(monitor.Receivers, subject, body)
}

// 檢查是否為重複告警 (查詢 TimescaleDB)
func (s *ESAlertService) isDuplicateAlert(alert *entities.ESAlert) bool {
    // 查詢 TimescaleDB 檢查重複告警
    query := `
        SELECT COUNT(*)
        FROM es_alert_history
        WHERE monitor_id = $1
            AND alert_type = $2
            AND status = 'active'
            AND time > $3
    `

    var count int64
    global.TimescaleDB.QueryRow(query,
        alert.MonitorID,
        alert.AlertType,
        time.Now().Add(-time.Hour)).Scan(&count)

    return count > 0
}

// 建立告警郵件內容
func (s *ESAlertService) buildAlertEmailBody(alert *entities.ESAlert, monitor *entities.ElasticsearchMonitor) string {
    return fmt.Sprintf(`
Dear Administrator,

An alert has been triggered for Elasticsearch monitor: %s

Alert Details:
- Type: %s
- Severity: %s
- Message: %s
- Time: %s
- Monitor: %s (%s:%d)

Please check your Elasticsearch cluster and take appropriate action.

Best regards,
Log Detect Monitoring System
`,
        monitor.Name,
        alert.AlertType,
        alert.Severity,
        alert.Message,
        alert.Time.Format("2006-01-02 15:04:05"),
        monitor.Name,
        monitor.Host,
        monitor.Port,
    )
}
```

## 📈 實作步驟

### Phase 1: 基礎建設 (雙層架構)
1. **資料庫建立**:
   - MySQL: 執行監控配置表 SQL
   - TimescaleDB: 建立時間序列表和索引
2. **實體定義**: 建立 `entities/elasticsearch.go`
3. **基本 API**: 實作監控配置的 CRUD API
4. **權限整合**: 新增權限定義和中介軟體

### Phase 2: 監控核心 (TimescaleDB 集成)
1. **健康檢查器**: 實作 ES 連接和健康檢查
2. **批量寫入**: 整合 TimescaleDB 批量寫入機制
3. **監控服務**: 實作定期監控邏輯，支援高頻數據寫入
4. **Cron 整合**: 將 ES 監控整合到現有調度系統

### Phase 3: 告警系統 (TimescaleDB 存儲)
1. **告警規則**: 實作告警條件判斷
2. **告警存儲**: 告警歷史存入 TimescaleDB
3. **通知服務**: 整合現有郵件系統發送告警
4. **重複檢查**: 基於 TimescaleDB 查詢避免重複告警

### Phase 4: 儀表板和優化
1. **查詢優化**: 實作高性能時間序列查詢
2. **API 完善**: 支援雙層架構的查詢和統計 API
3. **儀表板整合**: 在現有 Dashboard 中新增 ES 監控
4. **可選擴展**: 評估是否需要加入 Redis 緩存層

## 🧪 測試方法

### 單元測試
```go
func TestElasticsearchHealthCheck(t *testing.T) {
    // 測試 ES 健康檢查邏輯
}

func TestAlertConditions(t *testing.T) {
    // 測試告警條件判斷
}
```

### 整合測試
```bash
# 測試完整監控流程
./test_es_monitoring.sh
```

### 手動測試
```bash
# 建立測試監控配置
curl -X POST http://localhost:8006/api/v1/elasticsearch/monitors \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Test ES","host":"localhost","port":9200,...}'

# 測試連接
curl -X POST http://localhost:8006/api/v1/elasticsearch/monitors/1/test \
  -H "Authorization: Bearer $TOKEN"
```

---

**版本**: 1.0
**最後更新**: 2024-09-30
**作者**: Log Detect 開發團隊