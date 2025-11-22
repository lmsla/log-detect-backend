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

SELECT create_hypertable('es_alert_history', 'time', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_es_alert_history_monitor_id ON es_alert_history (monitor_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_es_alert_history_status ON es_alert_history (status, time DESC);
CREATE INDEX IF NOT EXISTS idx_es_alert_history_severity ON es_alert_history (severity, time DESC);
CREATE INDEX IF NOT EXISTS idx_es_alert_history_alert_type ON es_alert_history (alert_type, time DESC);
