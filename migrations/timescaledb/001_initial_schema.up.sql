-- Initial TimescaleDB schema for log-detect-backend
-- Version: 001
-- Created: 2025-11-22

-- =============================================
-- ES 監控指標時序表
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

-- 轉換為 hypertable（TimescaleDB 特性）
SELECT create_hypertable('es_metrics', 'time', if_not_exists => TRUE);

-- 建立索引
CREATE INDEX IF NOT EXISTS idx_es_metrics_monitor_id ON es_metrics (monitor_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_es_metrics_status ON es_metrics (status, time DESC);

-- =============================================
-- ES 告警歷史表
-- =============================================

CREATE TABLE IF NOT EXISTS es_alerts (
    time TIMESTAMPTZ NOT NULL,
    monitor_id INTEGER NOT NULL,
    alert_type VARCHAR(50),
    severity VARCHAR(20),
    message TEXT,
    status VARCHAR(20),
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
SELECT create_hypertable('es_alerts', 'time', if_not_exists => TRUE);

-- 建立索引
CREATE INDEX IF NOT EXISTS idx_es_alerts_monitor_id ON es_alerts (monitor_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_es_alerts_status ON es_alerts (status, time DESC);
CREATE INDEX IF NOT EXISTS idx_es_alerts_severity ON es_alerts (severity, time DESC);
