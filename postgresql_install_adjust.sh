#!/bin/bash
# postgresql_install.sh - Ubuntu TimescaleDB 完整部署腳本

echo "🚀 安裝 ES 監控系統 (Ubuntu 版本)..."

# 1. 安裝 TimescaleDB
echo "📦 安裝 TimescaleDB..."
echo "deb https://packagecloud.io/timescale/timescaledb/ubuntu/ $(lsb_release -c -s) main" | sudo tee /etc/apt/sources.list.d/timescaledb.list
wget --quiet -O - https://packagecloud.io/timescale/timescaledb/gpgkey | sudo
apt-key add -
sudo apt update
sudo apt install -y timescaledb-2-postgresql-14

# 2. 優化 TimescaleDB 配置
echo "⚙️ 優化 TimescaleDB 配置..."
sudo timescaledb-tune --quiet --yes
sudo systemctl restart postgresql

# 3. 安裝 MySQL (如果需要)
echo "📦 安裝 MySQL..."
sudo apt install -y mysql-server-8.0
sudo systemctl enable mysql
sudo systemctl start mysql

# 4. 創建數據庫和用戶
echo "💾 設置數據庫..."
sudo -u postgres createdb monitoring
sudo -u postgres psql -d monitoring -c "CREATE EXTENSION IF NOT EXISTS 
timescaledb;"
sudo -u postgres psql -c "CREATE USER logdetect WITH PASSWORD 
'your_secure_password';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE monitoring TO 
logdetect;"
sudo -u postgres psql -c "ALTER DATABASE monitoring OWNER TO logdetect;"

# MySQL 設置
sudo mysql -e "CREATE DATABASE IF NOT EXISTS logdetect CHARACTER SET utf8mb4 
COLLATE utf8mb4_unicode_ci;"
sudo mysql -e "CREATE USER IF NOT EXISTS 'runner'@'localhost' IDENTIFIED BY 
'1qaz2wsx';"
sudo mysql -e "GRANT ALL PRIVILEGES ON logdetect.* TO 'runner'@'localhost';"
sudo mysql -e "FLUSH PRIVILEGES;"

# 5. 初始化時間序列表
echo "📋 初始化 TimescaleDB 表結構..."
sudo -u postgres psql -d monitoring << 'EOF'
-- 啟用 TimescaleDB 擴展
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- ========================================
-- 設備監控歷史表 (核心表 - 替代 MySQL histories)
-- ========================================
CREATE TABLE IF NOT EXISTS device_metrics (
    time TIMESTAMPTZ NOT NULL,
    device_id TEXT NOT NULL,
    device_group TEXT NOT NULL,
    logname TEXT NOT NULL,

    -- 檢查結果
    status TEXT NOT NULL,
    lost BOOLEAN DEFAULT FALSE,
    lost_num INTEGER DEFAULT 0,

    -- 時間信息 (保持與現有 MySQL 結構兼容)
    date VARCHAR(10) NOT NULL,
    hour_time VARCHAR(8) NOT NULL,
    date_time VARCHAR(19) NOT NULL,
    timestamp_unix BIGINT NOT NULL,

    -- 檢查配置
    period VARCHAR(20),
    unit INTEGER,
    target_id INTEGER,
    index_id INTEGER,

    -- 性能指標
    response_time BIGINT DEFAULT 0,
    data_count BIGINT DEFAULT 0,

    -- 錯誤信息
    error_msg TEXT,
    error_code VARCHAR(50),

    -- 額外元數據
    metadata JSONB
);

-- 轉換為時間序列表 (按天分區)
SELECT create_hypertable('device_metrics', 'time',
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- 壓縮設置
ALTER TABLE device_metrics SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'device_id,logname',
    timescaledb.compress_orderby = 'time DESC'
);

-- 自動壓縮策略 (7天後壓縮，節省 90% 空間)
SELECT add_compression_policy('device_metrics', INTERVAL '7 days', if_not_exists
=> TRUE);

-- 自動清理策略 (90天後刪除)
SELECT add_retention_policy('device_metrics', INTERVAL '90 days', if_not_exists
=> TRUE);

-- ========================================
-- 高性能索引 (查詢加速，這是關鍵！)
-- ========================================
-- 按設備 + 時間查詢 (用於設備時間線)
CREATE INDEX IF NOT EXISTS idx_device_metrics_device_time
    ON device_metrics (device_id, time DESC);

-- 按日誌名稱 + 時間查詢 (用於統計查詢)
CREATE INDEX IF NOT EXISTS idx_device_metrics_logname_time
    ON device_metrics (logname, time DESC);

-- 按狀態查詢 (部分索引，只索引離線設備)
CREATE INDEX IF NOT EXISTS idx_device_metrics_status
    ON device_metrics (status, time DESC) WHERE lost = TRUE;

-- 按設備群組查詢 (用於群組統計)
CREATE INDEX IF NOT EXISTS idx_device_metrics_group
    ON device_metrics (device_group, time DESC);

-- 按日期查詢 (用於日期範圍過濾)
CREATE INDEX IF NOT EXISTS idx_device_metrics_date
    ON device_metrics (date);

-- 複合索引：設備群組 + 日誌名稱 + 時間 (用於儀表板)
CREATE INDEX IF NOT EXISTS idx_device_metrics_group_logname_time
    ON device_metrics (device_group, logname, time DESC);

-- ========================================
-- ES 監控指標表 (未來功能)
-- ========================================
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
    active_shards INTEGER DEFAULT 0,
    unassigned_shards INTEGER DEFAULT 0,
    total_documents BIGINT DEFAULT 0,
    error_message TEXT,
    metadata JSONB
);

SELECT create_hypertable('es_metrics', 'time',
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

ALTER TABLE es_metrics SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'monitor_id',
    timescaledb.compress_orderby = 'time DESC'
);

SELECT add_compression_policy('es_metrics', INTERVAL '7 days', if_not_exists =>
TRUE);
SELECT add_retention_policy('es_metrics', INTERVAL '90 days', if_not_exists =>
TRUE);

CREATE INDEX IF NOT EXISTS idx_es_metrics_monitor_time
    ON es_metrics (monitor_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_es_metrics_status
    ON es_metrics (status, time DESC);
CREATE INDEX IF NOT EXISTS idx_es_metrics_cluster
    ON es_metrics (cluster_status, time DESC);

-- ========================================
-- 告警歷史表
-- ========================================
CREATE TABLE IF NOT EXISTS alert_history (
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
);

SELECT create_hypertable('alert_history', 'time',
    chunk_time_interval => INTERVAL '7 days',
    if_not_exists => TRUE
);

SELECT add_retention_policy('alert_history', INTERVAL '90 days', if_not_exists =>
TRUE);

CREATE INDEX IF NOT EXISTS idx_alert_history_monitor_time
    ON alert_history (monitor_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_alert_history_severity
    ON alert_history (severity, time DESC);
CREATE INDEX IF NOT EXISTS idx_alert_history_status
    ON alert_history (status, time DESC);
CREATE INDEX IF NOT EXISTS idx_alert_history_device
    ON alert_history (device_id, time DESC) WHERE device_id IS NOT NULL;

EOF

echo ""
echo "✅ ========================================"
echo "✅ Ubuntu TimescaleDB 部署完成！"
echo "✅ ========================================"
echo ""
echo "📊 數據庫信息:"
echo "   PostgreSQL (TimescaleDB): localhost:5432"
echo "   MySQL: localhost:3306"
echo ""
echo "🔗 TimescaleDB 連接字串:"
echo "   postgresql://logdetect:your_secure_password@localhost:5432/monitoring"
echo ""
echo "📝 下一步:"
echo "   1. 更新應用配置文件 setting.yml"
echo "   2. 實作 TimescaleDB 連接代碼"
echo "   3. 實作批量寫入服務"
echo "   4. 修改 detect.go 切換到 TimescaleDB"
echo ""
echo "🔍 驗證安裝:"
echo "   sudo -u postgres psql -d monitoring -c '\dt'"
echo "   sudo -u postgres psql -d monitoring -c '\di'"
echo ""



