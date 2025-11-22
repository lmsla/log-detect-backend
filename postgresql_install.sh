#!/bin/bash
# deploy-native-ubuntu.sh - Native 精簡雙層架構部署

echo "安裝 ES 監控系統 (Native Ubuntu 版)..."

# 1. 安裝 TimescaleDB
echo "安裝 TimescaleDB..."
echo "deb https://packagecloud.io/timescale/timescaledb/ubuntu/ $(lsb_release -c -s) main" | sudo tee /etc/apt/sources.list.d/timescaledb.list
wget --quiet -O - https://packagecloud.io/timescale/timescaledb/gpgkey | sudo apt-key add -
sudo apt update
sudo apt install -y timescaledb-2-postgresql-14

# 2. 優化 TimescaleDB 配置
echo "優化 TimescaleDB 配置..."
sudo timescaledb-tune --quiet --yes
sudo systemctl restart postgresql

# 3. 安裝 MySQL
echo "安裝 MySQL..."
sudo apt install -y mysql-server-8.0
sudo systemctl enable mysql
sudo systemctl start mysql

# 4. 創建數據庫和用戶
echo "設置數據庫..."
sudo -u postgres createdb monitoring
sudo -u postgres psql -d monitoring -c "CREATE EXTENSION IF NOT EXISTS timescaledb;"
sudo -u postgres psql -c "CREATE USER logdetect WITH PASSWORD 'your_secure_password';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE monitoring TO logdetect;"

# MySQL 設置
sudo mysql -e "CREATE DATABASE config CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
sudo mysql -e "CREATE USER 'monitor'@'localhost' IDENTIFIED BY 'password';"
sudo mysql -e "GRANT ALL PRIVILEGES ON config.* TO 'monitor'@'localhost';"
sudo mysql -e "FLUSH PRIVILEGES;"

# 5. 初始化時間序列表
echo "初始化 TimescaleDB 表結構..."
sudo -u postgres psql -d monitoring << 'EOF'
-- 設備監控指標表 (核心表)
CREATE TABLE IF NOT EXISTS device_metrics (
    time TIMESTAMPTZ NOT NULL,
    device_id TEXT NOT NULL,
    device_group TEXT,
    logname TEXT NOT NULL,
    status TEXT NOT NULL,
    lost BOOLEAN DEFAULT false,
    lost_num INTEGER DEFAULT 0,
    date DATE NOT NULL,
    hour_time TEXT,
    date_time TEXT,
    timestamp_unix BIGINT,
    period TEXT,
    unit INTEGER,
    target_id INTEGER,
    index_id INTEGER,
    response_time INTEGER DEFAULT 0,
    data_count INTEGER DEFAULT 0,
    error_msg TEXT,
    error_code TEXT,
    metadata JSONB
);

-- 創建時間序列表
SELECT create_hypertable('device_metrics', 'time', chunk_time_interval => INTERVAL '1 day', if_not_exists => TRUE);

-- 創建索引
CREATE INDEX IF NOT EXISTS idx_device_metrics_device_time ON device_metrics (device_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_device_metrics_logname_date ON device_metrics (logname, date);
CREATE INDEX IF NOT EXISTS idx_device_metrics_status ON device_metrics (status, time DESC);

-- 設置自動壓縮和清理
ALTER TABLE device_metrics SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'device_id, logname',
    timescaledb.compress_orderby = 'time DESC'
);

SELECT add_compression_policy('device_metrics', INTERVAL '7 days');
SELECT add_retention_policy('device_metrics', INTERVAL '90 days');

-- ES 監控指標表
CREATE TABLE IF NOT EXISTS es_metrics (
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

SELECT create_hypertable('es_metrics', 'time', chunk_time_interval => INTERVAL '1 day', if_not_exists => TRUE);

-- 創建性能索引
CREATE INDEX IF NOT EXISTS idx_es_metrics_monitor_time ON es_metrics (monitor_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_es_metrics_status ON es_metrics (status, time DESC);
CREATE INDEX IF NOT EXISTS idx_es_metrics_cluster_status ON es_metrics (cluster_status, time DESC);

-- 設置自動壓縮和清理
ALTER TABLE es_metrics SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'monitor_id',
    timescaledb.compress_orderby = 'time DESC'
);

SELECT add_compression_policy('es_metrics', INTERVAL '7 days');
SELECT add_retention_policy('es_metrics', INTERVAL '90 days');

-- 創建告警歷史表
CREATE TABLE IF NOT EXISTS es_alert_history (
    time TIMESTAMPTZ NOT NULL,
    monitor_id INTEGER NOT NULL,
    alert_type TEXT NOT NULL,
    severity TEXT NOT NULL,
    message TEXT NOT NULL,
    status TEXT DEFAULT 'active',
    resolved_at TIMESTAMPTZ,
    resolution_note TEXT
);

SELECT create_hypertable('es_alert_history', 'time', chunk_time_interval => INTERVAL '7 days', if_not_exists => TRUE);
SELECT add_retention_policy('es_alert_history', INTERVAL '90 days');

-- 授予用戶完整權限
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO logdetect;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO logdetect;
EOF

echo "========================================="
echo "安裝完成！"
echo "========================================="
echo "PostgreSQL (TimescaleDB): localhost:5432"
echo "  - Database: monitoring"
echo "  - User: logdetect"
echo "  - Password: your_secure_password"
echo ""
echo "MySQL: localhost:3306"
echo "  - Database: config"
echo "  - User: monitor"
echo "  - Password: password"
echo ""
echo "下一步："
echo "1. 更新 setting.yml 配置檔："
echo "   timescale:"
echo "     host: \"localhost\""
echo "     port: \"5432\""
echo "     user: \"logdetect\""
echo "     password: \"your_secure_password\""
echo "     name: \"monitoring\""
echo ""
echo "2. 編譯並啟動應用："
echo "   go build && ./log-detect"