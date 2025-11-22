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

SELECT create_hypertable('device_metrics', 'time', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_device_metrics_device_id ON device_metrics (device_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_device_metrics_logname ON device_metrics (logname, time DESC);
CREATE INDEX IF NOT EXISTS idx_device_metrics_device_group ON device_metrics (device_group, time DESC);
CREATE INDEX IF NOT EXISTS idx_device_metrics_date ON device_metrics (date, time DESC);
CREATE INDEX IF NOT EXISTS idx_device_metrics_status ON device_metrics (status, time DESC);
