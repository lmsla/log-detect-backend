# 📊 TimescaleDB 精簡監控架構設計

## 🎯 架構概述

本文檔詳細說明基於 TimescaleDB + MySQL 的精簡監控系統架構，專為日誌監控和 Elasticsearch 服務監控設計，實現 3 個月數據保留的高性能輕量級解決方案。

## 🏗️ 精簡雙層架構設計

### **主要架構圖**
```
┌─────────────────────────────────────────────────────────────────┐
│                    Application Layer                            │
│  ┌─────────────┬─────────────┬─────────────┬─────────────────┐   │
│  │ 日誌監控     │ ES 監控收集  │  告警引擎   │   Web Dashboard │   │
│  │ Collector   │ ES Collector│Alert Engine │      UI         │   │
│  │             │             │             │                 │   │
│  │ ┌─────────┐ │ ┌─────────┐ │ ┌─────────┐ │ ┌─────────────┐ │   │
│  │ │BatchWrit│ │ │BatchWrit│ │ │QueryOpt │ │ │InMemory     │ │   │
│  │ │er       │ │ │er       │ │ │imizer   │ │ │Cache        │ │   │
│  │ └─────────┘ │ └─────────┘ │ └─────────┘ │ └─────────────┘ │   │
│  └─────────────┴─────────────┴─────────────┴─────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                                    │
┌─────────────────────────────────────────────────────────────────┐
│                  精簡雙層數據存儲                                 │
│  ┌─────────────────────────────┐  ┌─────────────────────────┐   │
│  │       TimescaleDB           │  │        MySQL            │   │
│  │     (時間序列數據)           │  │    (配置+用戶)           │   │
│  │                             │  │                         │   │
│  │ • 所有監控歷史 (3個月)       │  │ • 用戶認證              │   │
│  │ • 告警歷史記錄             │  │ • 監控配置              │   │
│  │ • 高性能時間序列查詢        │  │ • 設備管理              │   │
│  │ • 自動分區/壓縮/清理        │  │ • 權限控制              │   │
│  │ • 亞秒級聚合統計           │  │ • Cron 任務             │   │
│  └─────────────────────────────┘  └─────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘

可選擴展 (高並發時再加入):
┌─────────────────────────────────────────────────────────────────┐
│                    Redis (可選熱數據層)                          │
│               • 毫秒級查詢 • 高併發支援 • 分散式緩存               │
└─────────────────────────────────────────────────────────────────┘
```

## 📊 TimescaleDB 設計詳解

### **為什麼選擇 TimescaleDB**

#### **1. 性能優勢**
```yaml
寫入性能: 500,000+ inserts/sec
查詢性能: 亞秒級時間序列查詢
壓縮率: 90% 數據壓縮 (7天後自動壓縮)
分區: 自動按時間分區，查詢效率極高
```

#### **2. SQL 兼容**
```sql
-- 完全兼容 PostgreSQL SQL 語法
SELECT * FROM device_metrics WHERE device_id = 'server1';

-- 時間序列專用函數
SELECT
    time_bucket('5 minutes', time) as interval,
    device_id,
    avg(response_time) as avg_response_time
FROM device_metrics
WHERE time >= NOW() - INTERVAL '24 hours'
GROUP BY interval, device_id
ORDER BY interval DESC;
```

#### **3. 自動化管理**
```sql
-- 自動數據保留 (3個月後自動刪除)
SELECT add_retention_policy('device_metrics', INTERVAL '90 days');

-- 自動壓縮 (7天後壓縮，節省90%空間)
SELECT add_compression_policy('device_metrics', INTERVAL '7 days');
```

### **表結構設計**

#### **日誌監控表 (device_metrics)**
```sql
CREATE TABLE device_metrics (
    time TIMESTAMPTZ NOT NULL,
    device_id TEXT NOT NULL,
    device_group TEXT NOT NULL,
    logname TEXT NOT NULL,
    status TEXT NOT NULL,
    response_time BIGINT DEFAULT 0,
    lost BOOLEAN DEFAULT FALSE,
    target_id INTEGER,
    index_id INTEGER,
    error_message TEXT,
    metadata JSONB
);

-- 轉換為時間序列表
SELECT create_hypertable('device_metrics', 'time', chunk_time_interval => INTERVAL '1 day');

-- 自動管理策略
ALTER TABLE device_metrics SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'device_id,logname',
    timescaledb.compress_orderby = 'time DESC'
);

SELECT add_compression_policy('device_metrics', INTERVAL '7 days');
SELECT add_retention_policy('device_metrics', INTERVAL '90 days');

-- 高性能索引
CREATE INDEX idx_device_metrics_device_time ON device_metrics (device_id, time DESC);
CREATE INDEX idx_device_metrics_logname_time ON device_metrics (logname, time DESC);
CREATE INDEX idx_device_metrics_status ON device_metrics (status, time DESC);
CREATE INDEX idx_device_metrics_group ON device_metrics (device_group, time DESC);
```

#### **ES 監控表 (es_metrics)**
```sql
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
    active_shards INTEGER DEFAULT 0,
    unassigned_shards INTEGER DEFAULT 0,
    total_documents BIGINT DEFAULT 0,
    error_message TEXT,
    metadata JSONB
);

SELECT create_hypertable('es_metrics', 'time', chunk_time_interval => INTERVAL '1 day');

ALTER TABLE es_metrics SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'monitor_id',
    timescaledb.compress_orderby = 'time DESC'
);

SELECT add_compression_policy('es_metrics', INTERVAL '7 days');
SELECT add_retention_policy('es_metrics', INTERVAL '90 days');

CREATE INDEX idx_es_metrics_monitor_time ON es_metrics (monitor_id, time DESC);
CREATE INDEX idx_es_metrics_status ON es_metrics (status, time DESC);
CREATE INDEX idx_es_metrics_cluster ON es_metrics (cluster_status, time DESC);
```

#### **告警歷史表 (alert_history)**
```sql
CREATE TABLE alert_history (
    time TIMESTAMPTZ NOT NULL,
    monitor_type TEXT NOT NULL, -- 'device' or 'elasticsearch'
    monitor_id INTEGER NOT NULL,
    device_id TEXT,             -- 只有設備監控時使用
    logname TEXT,               -- 只有設備監控時使用
    alert_type TEXT NOT NULL,
    severity TEXT NOT NULL,     -- 'low', 'medium', 'high', 'critical'
    message TEXT NOT NULL,
    status TEXT DEFAULT 'active', -- 'active', 'resolved'
    resolved_at TIMESTAMPTZ,
    resolution_note TEXT
);

SELECT create_hypertable('alert_history', 'time', chunk_time_interval => INTERVAL '7 days');
SELECT add_retention_policy('alert_history', INTERVAL '90 days');

CREATE INDEX idx_alert_history_type_time ON alert_history (monitor_type, time DESC);
CREATE INDEX idx_alert_history_monitor_time ON alert_history (monitor_id, time DESC);
CREATE INDEX idx_alert_history_severity ON alert_history (severity, time DESC);
CREATE INDEX idx_alert_history_status ON alert_history (status, time DESC);
```

## 🚀 Go 程式碼實作

### **TimescaleDB 連接管理**
```go
// database/timescale.go
package database

import (
    "database/sql"
    "fmt"
    "time"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

type TimescaleDB struct {
    *gorm.DB
    rawDB *sql.DB
}

func NewTimescaleDB(dsn string) (*TimescaleDB, error) {
    // GORM 連接
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent), // 生產環境關閉日誌
        PrepareStmt: true,                             // 預編譯 SQL
        DisableForeignKeyConstraintWhenMigrating: true,
    })

    if err != nil {
        return nil, fmt.Errorf("failed to connect to TimescaleDB: %v", err)
    }

    // 原生 SQL 連接 (用於批量操作)
    rawDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("failed to get raw database connection: %v", err)
    }

    // 連接池優化
    rawDB.SetMaxOpenConns(100)
    rawDB.SetMaxIdleConns(20)
    rawDB.SetConnMaxLifetime(time.Hour)

    return &TimescaleDB{
        DB:    db,
        rawDB: rawDB,
    }, nil
}

// 初始化時間序列表
func (t *TimescaleDB) InitializeTables() error {
    queries := []string{
        // 創建擴展
        "CREATE EXTENSION IF NOT EXISTS timescaledb;",

        // 設備監控表
        `CREATE TABLE IF NOT EXISTS device_metrics (
            time TIMESTAMPTZ NOT NULL,
            device_id TEXT NOT NULL,
            device_group TEXT NOT NULL,
            logname TEXT NOT NULL,
            status TEXT NOT NULL,
            response_time BIGINT DEFAULT 0,
            lost BOOLEAN DEFAULT FALSE,
            target_id INTEGER,
            index_id INTEGER,
            error_message TEXT,
            metadata JSONB
        );`,

        // ES 監控表
        `CREATE TABLE IF NOT EXISTS es_metrics (
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
            active_shards INTEGER DEFAULT 0,
            unassigned_shards INTEGER DEFAULT 0,
            total_documents BIGINT DEFAULT 0,
            error_message TEXT,
            metadata JSONB
        );`,

        // 告警歷史表
        `CREATE TABLE IF NOT EXISTS alert_history (
            time TIMESTAMPTZ NOT NULL,
            monitor_type TEXT NOT NULL,
            monitor_id INTEGER NOT NULL,
            device_id TEXT,
            logname TEXT,
            alert_type TEXT NOT NULL,
            severity TEXT NOT NULL,
            message TEXT NOT NULL,
            status TEXT DEFAULT 'active',
            resolved_at TIMESTAMPTZ,
            resolution_note TEXT
        );`,
    }

    for _, query := range queries {
        if err := t.rawDB.QueryRow(query).Err(); err != nil && err != sql.ErrNoRows {
            return fmt.Errorf("failed to execute query %s: %v", query, err)
        }
    }

    // 創建時間序列表
    hypertables := []string{
        "SELECT create_hypertable('device_metrics', 'time', chunk_time_interval => INTERVAL '1 day', if_not_exists => TRUE);",
        "SELECT create_hypertable('es_metrics', 'time', chunk_time_interval => INTERVAL '1 day', if_not_exists => TRUE);",
        "SELECT create_hypertable('alert_history', 'time', chunk_time_interval => INTERVAL '7 days', if_not_exists => TRUE);",
    }

    for _, query := range hypertables {
        if _, err := t.rawDB.Exec(query); err != nil {
            return fmt.Errorf("failed to create hypertable: %v", err)
        }
    }

    return t.setupPolicies()
}

// 設置自動管理策略
func (t *TimescaleDB) setupPolicies() error {
    policies := []string{
        // 壓縮策略
        `ALTER TABLE device_metrics SET (
            timescaledb.compress,
            timescaledb.compress_segmentby = 'device_id,logname',
            timescaledb.compress_orderby = 'time DESC'
        );`,
        `ALTER TABLE es_metrics SET (
            timescaledb.compress,
            timescaledb.compress_segmentby = 'monitor_id',
            timescaledb.compress_orderby = 'time DESC'
        );`,

        // 壓縮策略 (7天後壓縮)
        "SELECT add_compression_policy('device_metrics', INTERVAL '7 days');",
        "SELECT add_compression_policy('es_metrics', INTERVAL '7 days');",

        // 保留策略 (3個月後刪除)
        "SELECT add_retention_policy('device_metrics', INTERVAL '90 days');",
        "SELECT add_retention_policy('es_metrics', INTERVAL '90 days');",
        "SELECT add_retention_policy('alert_history', INTERVAL '90 days');",
    }

    for _, policy := range policies {
        if _, err := t.rawDB.Exec(policy); err != nil {
            // 忽略重複設置的錯誤
            if !strings.Contains(err.Error(), "already exists") {
                return fmt.Errorf("failed to setup policy: %v", err)
            }
        }
    }

    return nil
}
```

### **批量寫入服務**
```go
// services/batch_writer.go
package services

import (
    "context"
    "database/sql"
    "fmt"
    "sync"
    "time"
)

type BatchWriter struct {
    timescaleDB    *sql.DB
    deviceBatch    []DeviceMetric
    esBatch        []ESMetric
    alertBatch     []AlertRecord
    batchSize      int
    flushInterval  time.Duration
    mutex          sync.Mutex
    ticker         *time.Ticker
    stopChan       chan struct{}
}

type DeviceMetric struct {
    Time         time.Time `json:"time"`
    DeviceID     string    `json:"device_id"`
    DeviceGroup  string    `json:"device_group"`
    Logname      string    `json:"logname"`
    Status       string    `json:"status"`
    ResponseTime int64     `json:"response_time"`
    Lost         bool      `json:"lost"`
    TargetID     int       `json:"target_id"`
    IndexID      int       `json:"index_id"`
    ErrorMessage string    `json:"error_message"`
    Metadata     string    `json:"metadata"`
}

type ESMetric struct {
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
    ActiveShards     int       `json:"active_shards"`
    UnassignedShards int       `json:"unassigned_shards"`
    TotalDocuments   int64     `json:"total_documents"`
    ErrorMessage     string    `json:"error_message"`
    Metadata         string    `json:"metadata"`
}

type AlertRecord struct {
    Time           time.Time  `json:"time"`
    MonitorType    string     `json:"monitor_type"`
    MonitorID      int        `json:"monitor_id"`
    DeviceID       string     `json:"device_id,omitempty"`
    Logname        string     `json:"logname,omitempty"`
    AlertType      string     `json:"alert_type"`
    Severity       string     `json:"severity"`
    Message        string     `json:"message"`
    Status         string     `json:"status"`
    ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
    ResolutionNote string     `json:"resolution_note,omitempty"`
}

func NewBatchWriter(db *sql.DB, batchSize int, flushInterval time.Duration) *BatchWriter {
    bw := &BatchWriter{
        timescaleDB:   db,
        deviceBatch:   make([]DeviceMetric, 0, batchSize),
        esBatch:       make([]ESMetric, 0, batchSize),
        alertBatch:    make([]AlertRecord, 0, batchSize),
        batchSize:     batchSize,
        flushInterval: flushInterval,
        ticker:        time.NewTicker(flushInterval),
        stopChan:      make(chan struct{}),
    }

    go bw.startFlushRoutine()
    return bw
}

func (bw *BatchWriter) startFlushRoutine() {
    for {
        select {
        case <-bw.ticker.C:
            bw.flushAllBatches()
        case <-bw.stopChan:
            bw.flushAllBatches()
            return
        }
    }
}

// 添加設備監控數據
func (bw *BatchWriter) AddDeviceMetric(metric DeviceMetric) {
    bw.mutex.Lock()
    defer bw.mutex.Unlock()

    bw.deviceBatch = append(bw.deviceBatch, metric)

    if len(bw.deviceBatch) >= bw.batchSize {
        bw.flushDeviceMetrics()
    }
}

// 添加 ES 監控數據
func (bw *BatchWriter) AddESMetric(metric ESMetric) {
    bw.mutex.Lock()
    defer bw.mutex.Unlock()

    bw.esBatch = append(bw.esBatch, metric)

    if len(bw.esBatch) >= bw.batchSize {
        bw.flushESMetrics()
    }
}

// 添加告警數據
func (bw *BatchWriter) AddAlert(alert AlertRecord) {
    bw.mutex.Lock()
    defer bw.mutex.Unlock()

    bw.alertBatch = append(bw.alertBatch, alert)

    if len(bw.alertBatch) >= bw.batchSize {
        bw.flushAlerts()
    }
}

func (bw *BatchWriter) flushAllBatches() {
    bw.mutex.Lock()
    defer bw.mutex.Unlock()

    bw.flushDeviceMetrics()
    bw.flushESMetrics()
    bw.flushAlerts()
}

func (bw *BatchWriter) flushDeviceMetrics() {
    if len(bw.deviceBatch) == 0 {
        return
    }

    tx, err := bw.timescaleDB.Begin()
    if err != nil {
        fmt.Printf("Failed to begin transaction for device metrics: %v\n", err)
        return
    }
    defer tx.Rollback()

    stmt, err := tx.Prepare(`
        INSERT INTO device_metrics
        (time, device_id, device_group, logname, status, response_time, lost, target_id, index_id, error_message, metadata)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `)
    if err != nil {
        fmt.Printf("Failed to prepare device metrics statement: %v\n", err)
        return
    }
    defer stmt.Close()

    for _, metric := range bw.deviceBatch {
        _, err := stmt.Exec(
            metric.Time, metric.DeviceID, metric.DeviceGroup, metric.Logname,
            metric.Status, metric.ResponseTime, metric.Lost, metric.TargetID,
            metric.IndexID, metric.ErrorMessage, metric.Metadata,
        )
        if err != nil {
            fmt.Printf("Failed to insert device metric: %v\n", err)
            continue
        }
    }

    if err := tx.Commit(); err != nil {
        fmt.Printf("Failed to commit device metrics: %v\n", err)
        return
    }

    fmt.Printf("Successfully flushed %d device metrics\n", len(bw.deviceBatch))
    bw.deviceBatch = bw.deviceBatch[:0] // 清空批次
}

func (bw *BatchWriter) flushESMetrics() {
    if len(bw.esBatch) == 0 {
        return
    }

    tx, err := bw.timescaleDB.Begin()
    if err != nil {
        fmt.Printf("Failed to begin transaction for ES metrics: %v\n", err)
        return
    }
    defer tx.Rollback()

    stmt, err := tx.Prepare(`
        INSERT INTO es_metrics
        (time, monitor_id, status, cluster_name, cluster_status, response_time,
         cpu_usage, memory_usage, disk_usage, node_count, active_shards,
         unassigned_shards, total_documents, error_message, metadata)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
    `)
    if err != nil {
        fmt.Printf("Failed to prepare ES metrics statement: %v\n", err)
        return
    }
    defer stmt.Close()

    for _, metric := range bw.esBatch {
        _, err := stmt.Exec(
            metric.Time, metric.MonitorID, metric.Status, metric.ClusterName,
            metric.ClusterStatus, metric.ResponseTime, metric.CpuUsage,
            metric.MemoryUsage, metric.DiskUsage, metric.NodeCount,
            metric.ActiveShards, metric.UnassignedShards, metric.TotalDocuments,
            metric.ErrorMessage, metric.Metadata,
        )
        if err != nil {
            fmt.Printf("Failed to insert ES metric: %v\n", err)
            continue
        }
    }

    if err := tx.Commit(); err != nil {
        fmt.Printf("Failed to commit ES metrics: %v\n", err)
        return
    }

    fmt.Printf("Successfully flushed %d ES metrics\n", len(bw.esBatch))
    bw.esBatch = bw.esBatch[:0] // 清空批次
}

func (bw *BatchWriter) flushAlerts() {
    if len(bw.alertBatch) == 0 {
        return
    }

    tx, err := bw.timescaleDB.Begin()
    if err != nil {
        fmt.Printf("Failed to begin transaction for alerts: %v\n", err)
        return
    }
    defer tx.Rollback()

    stmt, err := tx.Prepare(`
        INSERT INTO alert_history
        (time, monitor_type, monitor_id, device_id, logname, alert_type,
         severity, message, status, resolved_at, resolution_note)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `)
    if err != nil {
        fmt.Printf("Failed to prepare alert statement: %v\n", err)
        return
    }
    defer stmt.Close()

    for _, alert := range bw.alertBatch {
        _, err := stmt.Exec(
            alert.Time, alert.MonitorType, alert.MonitorID, alert.DeviceID,
            alert.Logname, alert.AlertType, alert.Severity, alert.Message,
            alert.Status, alert.ResolvedAt, alert.ResolutionNote,
        )
        if err != nil {
            fmt.Printf("Failed to insert alert: %v\n", err)
            continue
        }
    }

    if err := tx.Commit(); err != nil {
        fmt.Printf("Failed to commit alerts: %v\n", err)
        return
    }

    fmt.Printf("Successfully flushed %d alerts\n", len(bw.alertBatch))
    bw.alertBatch = bw.alertBatch[:0] // 清空批次
}

func (bw *BatchWriter) Stop() {
    bw.ticker.Stop()
    close(bw.stopChan)
}
```

### **內存緩存管理 (可選 Redis 集成)**
```go
// services/cache_manager.go (可選實作)
package services

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    "time"
)

// 簡單的內存緩存實現
type MemoryCache struct {
    data map[string]CacheItem
    mutex sync.RWMutex
}

type CacheItem struct {
    Value      interface{}
    Expiration time.Time
}

func NewMemoryCache() *MemoryCache {
    cache := &MemoryCache{
        data: make(map[string]CacheItem),
    }
    // 定期清理過期數據
    go cache.cleanup()
    return cache
}

func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    c.data[key] = CacheItem{
        Value:      value,
        Expiration: time.Now().Add(ttl),
    }
}

func (c *MemoryCache) Get(key string) (interface{}, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()

    item, exists := c.data[key]
    if !exists || time.Now().After(item.Expiration) {
        return nil, false
    }
    return item.Value, true
}

func (c *MemoryCache) cleanup() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        c.mutex.Lock()
        now := time.Now()
        for key, item := range c.data {
            if now.After(item.Expiration) {
                delete(c.data, key)
            }
        }
        c.mutex.Unlock()
    }
}

// 適配器模式：支援 Redis 或內存緩存
type CacheAdapter interface {
    SetDeviceStatus(logname, deviceID string, data interface{}) error
    GetDeviceStatus(logname, deviceID string) (interface{}, error)
    SetESStatus(monitorID int, data interface{}) error
    GetESStatus(monitorID int) (interface{}, error)
}

// 內存緩存適配器
type MemoryCacheAdapter struct {
    cache *MemoryCache
}

func NewMemoryCacheAdapter() *MemoryCacheAdapter {
    return &MemoryCacheAdapter{
        cache: NewMemoryCache(),
    }
}

func (m *MemoryCacheAdapter) SetDeviceStatus(logname, deviceID string, data interface{}) error {
    key := fmt.Sprintf("device:%s:%s", logname, deviceID)
    m.cache.Set(key, data, time.Hour)
    return nil
}

func (m *MemoryCacheAdapter) GetDeviceStatus(logname, deviceID string) (interface{}, error) {
    key := fmt.Sprintf("device:%s:%s", logname, deviceID)
    if value, exists := m.cache.Get(key); exists {
        return value, nil
    }
    return nil, fmt.Errorf("not found")
}

func (m *MemoryCacheAdapter) SetESStatus(monitorID int, data interface{}) error {
    key := fmt.Sprintf("es:%d", monitorID)
    m.cache.Set(key, data, time.Hour)
    return nil
}

func (m *MemoryCacheAdapter) GetESStatus(monitorID int) (interface{}, error) {
    key := fmt.Sprintf("es:%d", monitorID)
    if value, exists := m.cache.Get(key); exists {
        return value, nil
    }
    return nil, fmt.Errorf("not found")
}

// Redis 適配器 (當需要時才實作)
// type RedisCacheAdapter struct {
//     redis *redis.Client
// }
```

### **雙層查詢服務**
```go
// services/query_service.go
package services

import (
    "database/sql"
    "fmt"
    "time"
)

type QueryService struct {
    timescaleDB   *sql.DB
    cacheAdapter  CacheAdapter
}

type TimeRange struct {
    Start time.Time
    End   time.Time
}

func (tr TimeRange) IsRecent() bool {
    return time.Since(tr.Start) <= time.Hour
}

func NewQueryService(timescaleDB *sql.DB, cacheAdapter CacheAdapter) *QueryService {
    return &QueryService{
        timescaleDB:  timescaleDB,
        cacheAdapter: cacheAdapter,
    }
}

// 智能查詢：根據時間範圍選擇最佳數據源
func (q *QueryService) GetDeviceHistory(logname, deviceID string, timeRange TimeRange) ([]DeviceMetric, error) {
    // 1. 如果查詢最近數據，優先從緩存獲取
    if timeRange.IsRecent() && q.cacheAdapter != nil {
        if data, err := q.cacheAdapter.GetDeviceStatus(logname, deviceID); err == nil {
            // 轉換為標準格式
            if metric, ok := data.(DeviceMetric); ok {
                return []DeviceMetric{metric}, nil
            }
        }
    }

    // 2. 從 TimescaleDB 查詢
    return q.getDeviceHistoryFromTimescale(logname, deviceID, timeRange)
}

func (q *QueryService) getDeviceHistoryFromTimescale(logname, deviceID string, timeRange TimeRange) ([]DeviceMetric, error) {
    query := `
        SELECT time, device_id, device_group, logname, status,
               response_time, lost, target_id, index_id, error_message
        FROM device_metrics
        WHERE logname = $1
            AND device_id = $2
            AND time >= $3
            AND time <= $4
        ORDER BY time DESC
        LIMIT 1000
    `

    rows, err := q.timescaleDB.Query(query, logname, deviceID, timeRange.Start, timeRange.End)
    if err != nil {
        return nil, fmt.Errorf("failed to query device history: %v", err)
    }
    defer rows.Close()

    var results []DeviceMetric
    for rows.Next() {
        var metric DeviceMetric
        var targetID, indexID sql.NullInt64
        var errorMessage sql.NullString

        err := rows.Scan(
            &metric.Time, &metric.DeviceID, &metric.DeviceGroup, &metric.Logname,
            &metric.Status, &metric.ResponseTime, &metric.Lost, &targetID,
            &indexID, &errorMessage,
        )
        if err != nil {
            continue
        }

        if targetID.Valid {
            metric.TargetID = int(targetID.Int64)
        }
        if indexID.Valid {
            metric.IndexID = int(indexID.Int64)
        }
        if errorMessage.Valid {
            metric.ErrorMessage = errorMessage.String
        }

        results = append(results, metric)
    }

    return results, nil
}

// 聚合查詢：獲取設備統計數據
func (q *QueryService) GetDeviceStats(logname string, timeRange TimeRange) (*DeviceStats, error) {
    // 直接從 TimescaleDB 聚合查詢
    query := `
        SELECT
            COUNT(DISTINCT device_id) as total_devices,
            COUNT(*) as total_checks,
            COUNT(*) FILTER (WHERE status = 'online' AND NOT lost) as online_checks,
            COUNT(*) FILTER (WHERE status = 'offline' OR lost) as offline_checks,
            AVG(response_time) as avg_response_time,
            MAX(response_time) as max_response_time
        FROM device_metrics
        WHERE logname = $1
            AND time >= $2
            AND time <= $3
    `

    row := q.timescaleDB.QueryRow(query, logname, timeRange.Start, timeRange.End)

    var stats DeviceStats
    err := row.Scan(
        &stats.TotalDevices, &stats.TotalChecks, &stats.OnlineChecks,
        &stats.OfflineChecks, &stats.AvgResponseTime, &stats.MaxResponseTime,
    )

    if err != nil {
        return nil, fmt.Errorf("failed to get device stats: %v", err)
    }

    // 計算在線率
    if stats.TotalChecks > 0 {
        stats.UptimeRate = float64(stats.OnlineChecks) / float64(stats.TotalChecks) * 100
    }

    return &stats, nil
}

type DeviceStats struct {
    TotalDevices    int     `json:"total_devices"`
    TotalChecks     int     `json:"total_checks"`
    OnlineChecks    int     `json:"online_checks"`
    OfflineChecks   int     `json:"offline_checks"`
    UptimeRate      float64 `json:"uptime_rate"`
    AvgResponseTime float64 `json:"avg_response_time"`
    MaxResponseTime float64 `json:"max_response_time"`
}

func parseGroupStats(stats map[string]string) *DeviceStats {
    // 從 Redis hash 解析統計數據
    // 實作省略...
    return &DeviceStats{}
}

// 時間序列聚合查詢
func (q *QueryService) GetDeviceTimeSeries(logname, deviceID string, timeRange TimeRange, interval string) ([]TimeSeriesPoint, error) {
    query := fmt.Sprintf(`
        SELECT
            time_bucket('%s', time) as bucket,
            AVG(response_time) as avg_response_time,
            COUNT(*) as data_points,
            COUNT(*) FILTER (WHERE status = 'online' AND NOT lost) as online_count
        FROM device_metrics
        WHERE logname = $1
            AND device_id = $2
            AND time >= $3
            AND time <= $4
        GROUP BY bucket
        ORDER BY bucket DESC
    `, interval)

    rows, err := q.timescaleDB.Query(query, logname, deviceID, timeRange.Start, timeRange.End)
    if err != nil {
        return nil, fmt.Errorf("failed to query time series: %v", err)
    }
    defer rows.Close()

    var results []TimeSeriesPoint
    for rows.Next() {
        var point TimeSeriesPoint
        err := rows.Scan(&point.Time, &point.AvgResponseTime, &point.DataPoints, &point.OnlineCount)
        if err != nil {
            continue
        }
        results = append(results, point)
    }

    return results, nil
}

type TimeSeriesPoint struct {
    Time            time.Time `json:"time"`
    AvgResponseTime float64   `json:"avg_response_time"`
    DataPoints      int       `json:"data_points"`
    OnlineCount     int       `json:"online_count"`
}
```

## 📋 部署和運維

### **Docker 部署配置**
```yaml
# docker-compose.yml
version: '3.8'
services:
  # 應用服務
  log-detect:
    build: .
    ports:
      - "8006:8006"
    environment:
      - TIMESCALE_URL=postgresql://monitor:password@timescaledb:5432/monitoring
      - MYSQL_URL=mysql://root:password@mysql:3306/config
      - BATCH_SIZE=100
      - BATCH_TIMEOUT=30s
      # Redis 可選配置 (需要時啟用)
      # - REDIS_URL=redis://redis:6379
    depends_on:
      - timescaledb
      - mysql

  # TimescaleDB (主要時間序列數據存儲)
  timescaledb:
    image: timescale/timescaledb:latest-pg14
    environment:
      POSTGRES_DB: monitoring
      POSTGRES_USER: monitor
      POSTGRES_PASSWORD: password
      POSTGRES_INITDB_ARGS: "--encoding=UTF-8 --lc-collate=C --lc-ctype=C"
    volumes:
      - timescale_data:/var/lib/postgresql/data
      - ./init-timescale.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"
    command: >
      postgres
      -c shared_preload_libraries=timescaledb
      -c max_connections=200
      -c work_mem=256MB
      -c maintenance_work_mem=512MB
      -c effective_cache_size=2GB

  # MySQL (配置和用戶數據)
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: config
      MYSQL_COLLATION_SERVER: utf8mb4_unicode_ci
    volumes:
      - mysql_data:/var/lib/mysql
    ports:
      - "3306:3306"

  # Redis (可選 - 高並發時再啟用)
  # redis:
  #   image: redis:7-alpine
  #   volumes:
  #     - redis_data:/data
  #     - ./redis.conf:/usr/local/etc/redis/redis.conf
  #   ports:
  #     - "6379:6379"
  #   command: redis-server /usr/local/etc/redis/redis.conf

volumes:
  timescale_data:
  mysql_data:
  # redis_data: # Redis 相關時再啟用
```

### **可選 Redis 配置 (高並發時啟用)**
```conf
# redis.conf (可選配置文件)
# 內存優化
maxmemory 2gb
maxmemory-policy allkeys-lru

# 持久化配置
save 900 1
save 300 10
save 60 10000

# 網絡優化
tcp-keepalive 300
timeout 0

# 日誌配置
loglevel notice
logfile ""
```

### **性能監控腳本**
```bash
#!/bin/bash
# monitor.sh - 系統性能監控

echo "=== TimescaleDB 性能監控 ==="

# 1. 檢查表大小和壓縮率
docker exec timescaledb psql -U monitor -d monitoring -c "
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size,
    pg_total_relation_size(schemaname||'.'||tablename) as size_bytes
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY size_bytes DESC;
"

# 2. 檢查壓縮效率
docker exec timescaledb psql -U monitor -d monitoring -c "
SELECT
    chunk_schema,
    chunk_name,
    pg_size_pretty(before_compression_bytes) as before,
    pg_size_pretty(after_compression_bytes) as after,
    ROUND((before_compression_bytes - after_compression_bytes) * 100.0 / before_compression_bytes, 1) as compression_ratio
FROM timescaledb_information.compression_settings
WHERE before_compression_bytes > 0;
"

# 3. 檢查分區數量
docker exec timescaledb psql -U monitor -d monitoring -c "
SELECT
    hypertable_name,
    COUNT(*) as chunk_count,
    pg_size_pretty(SUM(total_bytes)) as total_size
FROM timescaledb_information.chunks
GROUP BY hypertable_name;
"

# Redis 性能監控 (可選 - 啟用 Redis 時才執行)
# echo "=== Redis 性能監控 ==="
# docker exec redis redis-cli info memory | grep -E "used_memory_human|maxmemory_human"
# docker exec redis redis-cli info clients
# docker exec redis redis-cli --hotkeys

echo "=== 應用性能監控 ==="

# 7. 檢查批量寫入性能
docker logs log-detect_log-detect_1 | grep "Successfully flushed" | tail -10
```

## 🎯 性能優化建議

### **TimescaleDB 調優**
```sql
-- 1. 查詢性能優化
ANALYZE device_metrics;
ANALYZE es_metrics;
ANALYZE alert_history;

-- 2. 檢查慢查詢
SELECT query, mean_time, calls
FROM pg_stat_statements
WHERE mean_time > 1000
ORDER BY mean_time DESC;

-- 3. 索引使用情況
SELECT
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;
```

### **應用層優化**
```go
// 1. TimescaleDB 連接池優化
db.SetMaxOpenConns(100)          // 最大連接數
db.SetMaxIdleConns(20)           // 最大空閒連接
db.SetConnMaxLifetime(time.Hour) // 連接最大生命週期

// 2. 批量寫入優化
batchSize := 100                 // 批次大小
flushInterval := 30 * time.Second // 刷新間隔

// 3. 內存緩存優化
cacheCleanupInterval := 5 * time.Minute // 清理間隔
cacheTTL := time.Hour                   // 緩存生存時間

// 4. Redis 連接池 (可選啟用)
// redisOptions := &redis.Options{
//     PoolSize:     100,
//     MinIdleConns: 20,
//     MaxRetries:   3,
// }
```

## 📊 總結

**精簡雙層架構**提供了：

### **核心優勢**
1. **極高性能**: TimescaleDB 50萬+ writes/sec，亞秒級查詢
2. **自動管理**: 分區、壓縮、清理全自動
3. **SQL 兼容**: 無學習成本，直接使用現有技能
4. **成本控制**: 90% 壓縮率，3個月數據生命週期
5. **部署簡化**: 僅需 2 個數據庫服務，維護成本低

### **架構特點**
- **TimescaleDB**: 處理高頻時間序列數據，自動優化性能
- **MySQL**: 繼續處理配置和用戶數據，保持現有邏輯
- **內存緩存**: 應用內緩存提供基本性能優化
- **Redis 可選**: 高並發需求時再加入，無強制依賴

### **擴展彈性**
- 起步簡單：僅需兩層架構即可滿足大部分需求
- 按需擴展：當並發增加時，可輕鬆加入 Redis 層
- 向後兼容：現有 MySQL 邏輯完全保留

這個精簡架構既保證了監控系統的高性能需求，又降低了部署和維護的複雜度！

---

**版本**: 1.0
**最後更新**: 2024-09-30
**作者**: Log Detect 開發團隊