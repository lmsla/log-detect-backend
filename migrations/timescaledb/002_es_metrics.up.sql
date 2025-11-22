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

SELECT create_hypertable('es_metrics', 'time', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_es_metrics_monitor_id ON es_metrics (monitor_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_es_metrics_status ON es_metrics (status, time DESC);
CREATE INDEX IF NOT EXISTS idx_es_metrics_cluster_status ON es_metrics (cluster_status, time DESC);
