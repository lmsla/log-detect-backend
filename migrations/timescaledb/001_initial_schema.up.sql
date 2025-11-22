-- TimescaleDB Initial Schema for log-detect-backend
-- Version: 001
-- Created: 2025-11-22
--
-- 此檔案定義 TimescaleDB 的所有時序表
-- 表格清單：
--   1. device_metrics   - 設備監控指標時序表
--   2. es_metrics       - ES 監控指標時序表
--   3. es_alert_history - ES 告警歷史時序表

-- =============================================
-- 1. device_metrics - 設備監控指標時序表
-- 用途：儲存設備監控的歷史數據
-- 寫入：services/batch_writer.go
-- 讀取：services/timescale_history.go
-- =============================================

CREATE TABLE IF NOT EXISTS device_metrics (
    time TIMESTAMPTZ NOT NULL,
    device_id VARCHAR(100) NOT NULL,
    device_group VARCHAR(50),
    logname VARCHAR(50),
    status VARCHAR(20),
    lost BOOLEAN DEFAULT FALSE,
    lost_num INTEGER DEFAULT 0,
    date VARCHAR(10),
    hour_time VARCHAR(8),
    date_time VARCHAR(19),
    timestamp_unix BIGINT,
    period VARCHAR(20),
    unit INTEGER DEFAULT 1,
    target_id INTEGER,
    index_id INTEGER,
    response_time BIGINT DEFAULT 0,
    data_count BIGINT DEFAULT 0,
    error_msg TEXT,
    error_code VARCHAR(50),
    metadata JSONB
);

-- 轉換為 hypertable（TimescaleDB 特性）
SELECT create_hypertable('device_metrics', 'time', if_not_exists => TRUE);

-- 建立索引
CREATE INDEX IF NOT EXISTS idx_device_metrics_device_id ON device_metrics (device_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_device_metrics_logname ON device_metrics (logname, time DESC);
CREATE INDEX IF NOT EXISTS idx_device_metrics_device_group ON device_metrics (device_group, time DESC);
CREATE INDEX IF NOT EXISTS idx_device_metrics_date ON device_metrics (date, time DESC);
CREATE INDEX IF NOT EXISTS idx_device_metrics_status ON device_metrics (status, time DESC);

-- =============================================
-- 2. es_metrics - ES 監控指標時序表
-- 用途：儲存 Elasticsearch 集群監控指標
-- 寫入：services/batch_writer.go
-- 讀取：services/es_monitor_query.go
-- =============================================

CREATE TABLE IF NOT EXISTS es_metrics (
    time TIMESTAMPTZ NOT NULL,
    monitor_id INTEGER NOT NULL,
    status VARCHAR(20),
    cluster_name VARCHAR(100),
    cluster_status VARCHAR(20),
    response_time BIGINT,
    cpu_usage DOUBLE PRECISION,
    memory_usage DOUBLE PRECISION,
    disk_usage DOUBLE PRECISION,
    node_count INTEGER,
    data_node_count INTEGER,
    query_latency BIGINT,
    indexing_rate DOUBLE PRECISION,
    search_rate DOUBLE PRECISION,
    total_indices INTEGER,
    total_documents BIGINT,
    total_size_bytes BIGINT,
    active_shards INTEGER,
    relocating_shards INTEGER,
    unassigned_shards INTEGER,
    error_message TEXT,
    warning_message TEXT,
    metadata JSONB
);

-- 轉換為 hypertable
SELECT create_hypertable('es_metrics', 'time', if_not_exists => TRUE);

-- 建立索引
CREATE INDEX IF NOT EXISTS idx_es_metrics_monitor_id ON es_metrics (monitor_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_es_metrics_status ON es_metrics (status, time DESC);
CREATE INDEX IF NOT EXISTS idx_es_metrics_cluster_status ON es_metrics (cluster_status, time DESC);

-- =============================================
-- 3. es_alert_history - ES 告警歷史時序表
-- 用途：儲存 ES 監控告警記錄
-- 寫入：services/es_monitor.go
-- 讀取：services/es_alert_service.go, services/es_monitor_query.go
-- =============================================

CREATE TABLE IF NOT EXISTS es_alert_history (
    time TIMESTAMPTZ NOT NULL,
    monitor_id INTEGER NOT NULL,
    alert_type VARCHAR(50),
    severity VARCHAR(20),
    status VARCHAR(20),
    message TEXT,
    cluster_name VARCHAR(100),
    threshold_value DOUBLE PRECISION,
    actual_value DOUBLE PRECISION,
    resolved_at TIMESTAMPTZ,
    resolved_by VARCHAR(100),
    resolution_note TEXT,
    acknowledged_at TIMESTAMPTZ,
    acknowledged_by VARCHAR(100),
    metadata JSONB
);

-- 轉換為 hypertable
SELECT create_hypertable('es_alert_history', 'time', if_not_exists => TRUE);

-- 建立索引
CREATE INDEX IF NOT EXISTS idx_es_alert_history_monitor_id ON es_alert_history (monitor_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_es_alert_history_status ON es_alert_history (status, time DESC);
CREATE INDEX IF NOT EXISTS idx_es_alert_history_severity ON es_alert_history (severity, time DESC);
CREATE INDEX IF NOT EXISTS idx_es_alert_history_alert_type ON es_alert_history (alert_type, time DESC);
