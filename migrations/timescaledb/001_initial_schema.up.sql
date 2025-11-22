-- Initial TimescaleDB schema for log-detect-backend
-- Version: 001
-- Created: 2025-11-22
-- Note: Uses DO blocks to handle cases where tables/indexes already exist

-- =============================================
-- ES 監控指標時序表
-- =============================================

DO $$
BEGIN
    -- 建立表（如果不存在）
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

    -- 嘗試轉換為 hypertable（忽略已存在的情況）
    BEGIN
        PERFORM create_hypertable('es_metrics', 'time', if_not_exists => TRUE);
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'es_metrics hypertable already exists or cannot be created: %', SQLERRM;
    END;

    -- 嘗試建立索引（忽略權限錯誤）
    BEGIN
        CREATE INDEX IF NOT EXISTS idx_es_metrics_monitor_id ON es_metrics (monitor_id, time DESC);
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Cannot create idx_es_metrics_monitor_id: %', SQLERRM;
    END;

    BEGIN
        CREATE INDEX IF NOT EXISTS idx_es_metrics_status ON es_metrics (status, time DESC);
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Cannot create idx_es_metrics_status: %', SQLERRM;
    END;
END $$;

-- =============================================
-- ES 告警歷史表
-- =============================================

DO $$
BEGIN
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

    BEGIN
        PERFORM create_hypertable('es_alerts', 'time', if_not_exists => TRUE);
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'es_alerts hypertable already exists or cannot be created: %', SQLERRM;
    END;

    BEGIN
        CREATE INDEX IF NOT EXISTS idx_es_alerts_monitor_id ON es_alerts (monitor_id, time DESC);
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Cannot create idx_es_alerts_monitor_id: %', SQLERRM;
    END;

    BEGIN
        CREATE INDEX IF NOT EXISTS idx_es_alerts_status ON es_alerts (status, time DESC);
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Cannot create idx_es_alerts_status: %', SQLERRM;
    END;

    BEGIN
        CREATE INDEX IF NOT EXISTS idx_es_alerts_severity ON es_alerts (severity, time DESC);
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Cannot create idx_es_alerts_severity: %', SQLERRM;
    END;
END $$;
