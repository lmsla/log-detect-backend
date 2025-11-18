# 🔍 Elasticsearch 服務監控系統

## 📋 概述

Elasticsearch 監控系統是 Log Detect 的擴展功能，專門用於監控 Elasticsearch 服務的健康狀況、性能指標和業務數據。該系統與現有的日誌設備監控系統完全整合，共享認證、調度和告警機制。

## 🎯 功能特性

### 1. 多維度監控
- **服務健康檢查**: 連線狀態、集群健康、節點狀態
- **性能監控**: 響應時間、吞吐量、資源使用率
- **業務指標**: 索引狀態、搜尋成功率、文檔統計

### 2. 智能告警
- **多層級告警**: Critical、High、Medium、Low
- **自動通知**: 郵件告警，支援多收件人
- **告警管理**: 告警解決、狀態追蹤

### 3. 系統整合
- **統一認證**: 使用現有 JWT + RBAC 權限系統
- **統一調度**: 整合到現有 Cron 任務系統
- **統一儀表板**: 在現有 Dashboard 中新增 ES 監控區塊

## 🏗️ 系統架構

### 精簡雙層架構 (主要架構)
```
┌─────────────────────────────────────────────────────────────────┐
│                    Load Balancer (Nginx)                        │
└─────────────────────────────────────────────────────────────────┘
                                    │
┌─────────────────────────────────────────────────────────────────┐
│                   監控服務 (Go Application)                      │
│  ┌─────────────┬─────────────┬─────────────┬─────────────────┐   │
│  │ ES 監控收集  │  告警引擎   │  查詢 API   │   Web Dashboard │   │
│  │ Collector   │Alert Engine │Query Service│      UI         │   │
│  │             │             │             │   (內存緩存)     │   │
│  └─────────────┴─────────────┴─────────────┴─────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                                    │
┌─────────────────────────────────────────────────────────────────┐
│                    精簡雙層數據存儲                                │
│  ┌─────────────────────────────┐  ┌─────────────────────────┐   │
│  │       TimescaleDB           │  │        MySQL            │   │
│  │    (時間序列數據)             │  │    (配置+用戶)            │   │
│  │                             │  │                         │   │
│  │ • 所有監控歷史(3個月)          │  │ • 用戶認證               │   │
│  │ • 告警歷史記錄                │  │ • 監控配置               │   │
│  │ • 高性能時間序列查詢           │  │ • 設備管理               │   │
│  │ • 自動分區/壓縮/清理           │  │ • 權限控制               │   │
│  │ • 亞秒級聚合統計              │   │                         │   │
│  └─────────────────────────────┘  └─────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘

可選擴展 (需要時再加入):
┌─────────────────────────────────────────────────────────────────┐
│                    Redis (可選熱數據層)                           │
│               • 毫秒級查詢 • 高併發支援                            │
└─────────────────────────────────────────────────────────────────┘
```

### 監控流程 (精簡版)
```
1. Cron 調度器 → 2. ES 健康檢查 → 3. 數據收集 → 4. 狀態評估
                                    ↓
5. 批量寫入 TimescaleDB ← 4. 告警判斷 → 5. 通知發送
                                    ↓
6. 儀表板展示 (TimescaleDB 直接查詢 + 應用層緩存)
```

### 核心組件
- **ElasticsearchController**: API 控制器
- **ESMonitorService**: 監控服務邏輯 (支援批量處理)
- **ESAlertService**: 告警服務
- **ESHealthChecker**: 健康檢查器
- **ESMetricsCollector**: 指標收集器
- **BatchWriter**: TimescaleDB 批量寫入服務
- **QueryService**: 智能查詢服務 (含內存緩存)
- **InMemoryCache**: 應用層內存緩存 (可選)

## 📊 監控指標

### 健康狀態指標
| 指標 | 說明 | 正常範圍 | 告警閾值 |
|-----|------|---------|----------|
| 連接狀態 | ES 服務連接狀況 | 連通 | 連接失敗 |
| 集群狀態 | 集群整體健康 | Green | Yellow/Red |
| 節點數量 | 活躍節點統計 | 預期值 | 節點減少 |
| 分片狀態 | 分片分佈狀況 | 全部已分配 | 未分配分片 |

### 性能指標
| 指標 | 說明 | 正常範圍 | 告警閾值 |
|-----|------|---------|----------|
| 響應時間 | API 響應延遲 | < 1s | > 5s |
| CPU 使用率 | 處理器使用率 | < 70% | > 80% |
| 記憶體使用率 | 內存使用率 | < 80% | > 85% |
| 磁碟使用率 | 儲存空間使用 | < 85% | > 90% |
| 查詢 QPS | 每秒查詢數 | 正常負載 | 異常負載 |
| 索引 TPS | 每秒索引數 | 正常負載 | 異常負載 |

### 業務指標
| 指標 | 說明 | 監控內容 |
|-----|------|---------|
| 索引總數 | 集群索引統計 | 索引數量變化 |
| 文檔總數 | 總文檔數統計 | 數據增長趨勢 |
| 儲存大小 | 總儲存空間 | 空間使用情況 |
| 搜尋失敗率 | 搜尋錯誤比例 | 搜尋品質 |
| 索引失敗率 | 索引錯誤比例 | 數據寫入品質 |

## 🚨 告警機制

### 告警等級定義

#### Critical (嚴重)
- ES 服務完全無法連接
- 集群狀態為 Red
- 所有節點離線

#### High (高)
- 集群狀態為 Yellow
- 資料節點數量減少
- 磁碟使用率 > 90%
- 存在未分配分片

#### Medium (中)
- 響應時間 > 5秒
- CPU 使用率 > 80%
- 記憶體使用率 > 85%

#### Low (低)
- 響應時間 > 2秒但 < 5秒
- CPU 使用率 > 70%但 < 80%
- 搜尋/索引失敗率上升

### 告警處理流程
1. **檢測**: 定期執行健康檢查
2. **評估**: 根據閾值判斷告警等級
3. **記錄**: 儲存告警記錄到資料庫
4. **通知**: 發送郵件告警給指定收件人
5. **追蹤**: 追蹤告警狀態和解決情況

## 📈 API 設計

### API 實作狀態總覽

| 端點 | 方法 | 功能 | 狀態 | Phase |
|------|------|------|------|-------|
| `/api/v1/elasticsearch/monitors` | GET | 獲取所有監控配置 | ✅ 已實作 | Phase 1 |
| `/api/v1/elasticsearch/monitors` | POST | 新增監控配置 | ✅ 已實作 | Phase 1 |
| `/api/v1/elasticsearch/monitors` | PUT | 更新監控配置 | ✅ 已實作 | Phase 1 |
| `/api/v1/elasticsearch/monitors/{id}` | GET | 獲取特定配置 | ✅ 已實作 | Phase 1 |
| `/api/v1/elasticsearch/monitors/{id}` | DELETE | 刪除監控配置 | ✅ 已實作 | Phase 1 |
| `/api/v1/elasticsearch/monitors/test` | POST | 測試連接 | ✅ 已實作 | Phase 1 |
| `/api/v1/elasticsearch/monitors/{id}/toggle` | POST | 啟用/停用監控 | ✅ 已實作 | Phase 1 |
| `/api/v1/elasticsearch/status` | GET | 獲取所有狀態 | ✅ 已實作 | Phase 1 |
| `/api/v1/elasticsearch/statistics` | GET | 獲取統計數據 | ✅ 已實作 | Phase 1 |
| `/api/v1/elasticsearch/status/{id}` | GET | 獲取特定狀態 | ⏳ 待實作 | Phase 2 |
| `/api/v1/elasticsearch/status/{id}/history` | GET | 獲取歷史記錄 | ⏳ 待實作 | Phase 2 |
| `/api/v1/elasticsearch/status/{id}/trends` | GET | 獲取趨勢數據 | ⏳ 待實作 | Phase 2 |
| `/api/v1/elasticsearch/alerts` | GET | 獲取告警列表 | ⏳ 待實作 | Phase 2 |
| `/api/v1/elasticsearch/alerts/{id}` | GET | 獲取告警詳情 | ⏳ 待實作 | Phase 2 |
| `/api/v1/elasticsearch/alerts/{id}/resolve` | POST | 解決告警 | ⏳ 待實作 | Phase 2 |
| `/api/v1/elasticsearch/alerts/{id}/acknowledge` | PUT | 確認告警 | ⏳ 待實作 | Phase 2 |
| `/api/v1/elasticsearch/dashboard` | GET | 儀表板數據 | ⏳ 待實作 | Phase 2 |
| `/api/v1/elasticsearch/metrics/{id}` | GET | 獲取指標數據 | ⏳ 待實作 | Phase 2 |

### 監控配置 API (✅ Phase 1 已實作)
```http
GET    /api/v1/elasticsearch/monitors           # 獲取所有監控配置
POST   /api/v1/elasticsearch/monitors           # 新增監控配置
GET    /api/v1/elasticsearch/monitors/{id}      # 獲取特定監控配置
PUT    /api/v1/elasticsearch/monitors           # 更新監控配置 (ID 從 body 傳遞)
DELETE /api/v1/elasticsearch/monitors/{id}      # 刪除監控配置
POST   /api/v1/elasticsearch/monitors/test      # 測試連接 (不需要已存在的 ID)
POST   /api/v1/elasticsearch/monitors/{id}/toggle  # 啟用/停用監控
```

### 狀態查詢 API (✅ Phase 1 部分已實作)
```http
GET    /api/v1/elasticsearch/status             # ✅ 獲取所有 ES 狀態
GET    /api/v1/elasticsearch/statistics         # ✅ 獲取統計摘要數據
GET    /api/v1/elasticsearch/status/{id}        # ⏳ 獲取特定 ES 狀態 (Phase 2)
GET    /api/v1/elasticsearch/status/{id}/history # ⏳ 獲取歷史狀態記錄 (Phase 2)
GET    /api/v1/elasticsearch/status/{id}/trends  # ⏳ 獲取趨勢數據 (Phase 2)
```

### 告警管理 API (⏳ Phase 2 待實作)
```http
GET    /api/v1/elasticsearch/alerts             # 獲取告警列表
GET    /api/v1/elasticsearch/alerts/{id}        # 獲取告警詳情
POST   /api/v1/elasticsearch/alerts/{id}/resolve # 解決告警
PUT    /api/v1/elasticsearch/alerts/{id}/acknowledge # 確認告警
```

### 儀表板 API (⏳ Phase 2 待實作)
```http
GET    /api/v1/elasticsearch/dashboard          # ES 監控儀表板數據
GET    /api/v1/elasticsearch/summary            # ES 監控摘要 (已有 /statistics 替代)
GET    /api/v1/elasticsearch/metrics/{id}       # 獲取指標數據
```

## 🔐 權限控制

### 新增權限定義
```
elasticsearch:create  - 建立 ES 監控配置
elasticsearch:read    - 查看 ES 監控數據
elasticsearch:update  - 更新 ES 監控配置
elasticsearch:delete  - 刪除 ES 監控配置
```

### 角色權限分配
- **admin**: 所有 elasticsearch 權限
- **user**: 僅 elasticsearch:read 權限

## 📋 配置說明

### 監控配置範例
```json
{
  "name": "Production ES Cluster",
  "host": "https://es-cluster.company.com",
  "port": 9200,
  "username": "monitor_user",
  "password": "secure_password",
  "enable_auth": true,
  "check_type": "health,performance,business",
  "interval": 60,
  "enable": true,
  "receivers": ["admin@company.com", "ops@company.com"],
  "subject": "ES Cluster Alert - Production"
}
```

### 監控間隔建議
- **生產環境**: 60-120 秒
- **測試環境**: 300-600 秒
- **開發環境**: 600-1800 秒

## 🔧 部署和配置

### 三層資料庫架構初始化

#### **TimescaleDB 初始化**
```sql
-- 建立 ES 監控時間序列表
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
    metadata JSONB
);

-- 轉換為時間序列表 (按天分區)
SELECT create_hypertable('es_metrics', 'time', chunk_time_interval => INTERVAL '1 day');

-- 自動清理策略 (保留3個月)
SELECT add_retention_policy('es_metrics', INTERVAL '90 days');

-- 自動壓縮策略 (7天後壓縮)
ALTER TABLE es_metrics SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'monitor_id',
    timescaledb.compress_orderby = 'time DESC'
);
SELECT add_compression_policy('es_metrics', INTERVAL '7 days');

-- 告警歷史表
CREATE TABLE es_alert_history (
    time TIMESTAMPTZ NOT NULL,
    monitor_id INTEGER NOT NULL,
    alert_type TEXT NOT NULL,
    severity TEXT NOT NULL,
    message TEXT NOT NULL,
    status TEXT DEFAULT 'active'
);

SELECT create_hypertable('es_alert_history', 'time', chunk_time_interval => INTERVAL '7 days');
SELECT add_retention_policy('es_alert_history', INTERVAL '90 days');

-- 建立索引
CREATE INDEX idx_es_metrics_monitor_time ON es_metrics (monitor_id, time DESC);
CREATE INDEX idx_es_metrics_status ON es_metrics (status, time DESC);
CREATE INDEX idx_es_alert_monitor_time ON es_alert_history (monitor_id, time DESC);
```

#### **MySQL 配置表**
```sql
-- ES 監控配置表 (保留在 MySQL)
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
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### Docker 部署配置 (精簡版)
```yaml
# docker-compose.yml - 極簡雙層架構
version: '3.8'
services:
  # 監控應用
  log-detect:
    build: .
    ports:
      - "8006:8006"
    environment:
      - TIMESCALE_URL=postgresql://monitor:password@timescaledb:5432/monitoring
      - MYSQL_URL=mysql://root:password@mysql:3306/config
      - BATCH_SIZE=100
      - BATCH_TIMEOUT=30s
      - ENABLE_MEMORY_CACHE=true
    depends_on:
      - timescaledb
      - mysql

  # TimescaleDB (時間序列數據)
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

  # MySQL (配置數據)
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

volumes:
  timescale_data:
  mysql_data:

# 可選: Redis 擴展 (需要時取消註釋)
# redis:
#   image: redis:7-alpine
#   volumes:
#     - redis_data:/data
#   ports:
#     - "6379:6379"
#
# volumes:
#   redis_data: (add to volumes section)
```

### 環境變數配置 (精簡版)
```bash
# ES 監控相關配置
ES_MONITOR_ENABLED=true
ES_MONITOR_DEFAULT_INTERVAL=60
ES_MONITOR_BATCH_SIZE=100
ES_MONITOR_BATCH_TIMEOUT=30s

# 數據庫配置 (僅需2個)
TIMESCALE_URL=postgresql://monitor:password@localhost:5432/monitoring
MYSQL_URL=mysql://root:password@localhost:3306/config

# 緩存配置
ENABLE_MEMORY_CACHE=true
MEMORY_CACHE_TTL=300s
MEMORY_CACHE_SIZE=1000

# 數據保留配置
DATA_RETENTION=90d  # 統一3個月保留策略

# 可選: Redis 配置 (需要時啟用)
# REDIS_URL=redis://localhost:6379
# REDIS_ENABLED=false
```

## 🚀 部署方案選擇

### 方案比較

| 特性 | Docker 部署 | Native 安裝 (deb/rpm) |
|------|-------------|------------------------|
| **部署速度** | ⭐⭐⭐⭐⭐ 極快 | ⭐⭐⭐ 中等 |
| **性能** | ⭐⭐⭐⭐ 略有開銷 | ⭐⭐⭐⭐⭐ 最佳 |
| **資源使用** | ⭐⭐⭐ 容器開銷 | ⭐⭐⭐⭐⭐ 最優 |
| **維護難度** | ⭐⭐⭐⭐⭐ 簡單 | ⭐⭐⭐ 中等 |
| **擴展性** | ⭐⭐⭐⭐ 容易 | ⭐⭐⭐⭐⭐ 靈活 |
| **生產就緒** | ⭐⭐⭐⭐ 適合中型 | ⭐⭐⭐⭐⭐ 適合大型 |

### 推薦使用場景

**Docker 部署** (推薦大多數情況)：
- 開發測試環境
- 快速概念驗證
- 中小型生產環境 (< 10萬筆/分)
- 需要環境隔離的場景

**Native 安裝** (推薦高性能場景)：
- 大型生產環境 (> 10萬筆/分)
- 對性能要求極高
- 需要與現有系統深度整合
- 傳統運維管理方式

## 🐳 方案一：Docker 部署 (推薦)

### 自動化部署腳本
```bash
#!/bin/bash
# deploy-docker.sh - Docker 精簡雙層架構部署

echo "部署 ES 監控系統 (Docker 版)..."

# 1. 啟動基礎設施 (僅需2個組件)
echo "啟動 TimescaleDB 和 MySQL..."
docker-compose up -d timescaledb mysql

# 2. 等待資料庫啟動
echo "等待資料庫啟動..."
sleep 30

# 3. 檢查 TimescaleDB 狀態
until docker exec timescaledb pg_isready -U monitor; do
    echo "等待 TimescaleDB 就緒..."
    sleep 5
done

# 4. 初始化 TimescaleDB
echo "初始化 TimescaleDB..."
docker exec timescaledb psql -U monitor -d monitoring -f /docker-entrypoint-initdb.d/init.sql

# 5. 啟動應用
echo "啟動應用服務..."
docker-compose up -d log-detect

echo "========================================="
echo "部署完成！精簡雙層架構已啟動"
echo "========================================="
echo "監控面板: http://localhost:8006"
echo "TimescaleDB: localhost:5432"
echo "MySQL: localhost:3306"
echo ""
echo "組件狀態檢查:"
docker-compose ps
echo ""
echo "如需 Redis 擴展，請修改 docker-compose.yml 並重新部署"
```

## 📦 方案二：Native 安裝 (高性能)

### Ubuntu/Debian 安裝腳本
```bash
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
sudo -u postgres psql -c "CREATE USER monitor WITH PASSWORD 'password';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE monitoring TO monitor;"

# MySQL 設置
sudo mysql -e "CREATE DATABASE config CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
sudo mysql -e "CREATE USER 'monitor'@'localhost' IDENTIFIED BY 'password';"
sudo mysql -e "GRANT ALL PRIVILEGES ON config.* TO 'monitor'@'localhost';"
sudo mysql -e "FLUSH PRIVILEGES;"

# 5. 初始化時間序列表
echo "初始化 TimescaleDB 表結構..."
sudo -u postgres psql -d monitoring << 'EOF'
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
    active_shards INTEGER DEFAULT 0,
    unassigned_shards INTEGER DEFAULT 0,
    total_documents BIGINT DEFAULT 0,
    error_message TEXT,
    metadata JSONB
);

-- 創建時間序列表
SELECT create_hypertable('es_metrics', 'time', chunk_time_interval => INTERVAL '1 day', if_not_exists => TRUE);

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
EOF

echo "========================================="
echo "Native 安裝完成！"
echo "========================================="
echo "PostgreSQL (TimescaleDB): localhost:5432"
echo "MySQL: localhost:3306"
echo ""
echo "下一步："
echo "1. 編譯並部署 log-detect-backend 應用"
echo "2. 配置環境變數:"
echo "   TIMESCALE_URL=postgresql://monitor:password@localhost:5432/monitoring"
echo "   MYSQL_URL=mysql://monitor:password@localhost:3306/config"
echo "3. 啟動應用服務"
```

### CentOS/RHEL 安裝腳本
```bash
#!/bin/bash
# deploy-native-centos.sh - Native CentOS/RHEL 部署

echo "安裝 ES 監控系統 (Native CentOS 版)..."

# 1. 添加 TimescaleDB 源
sudo tee /etc/yum.repos.d/timescale_timescaledb.repo <<EOL
[timescale_timescaledb]
name=timescale_timescaledb
baseurl=https://packagecloud.io/timescale/timescaledb/el/$(rpm -E %{rhel})/\$basearch
repo_gpgcheck=1
gpgcheck=0
enabled=1
gpgkey=https://packagecloud.io/timescale/timescaledb/gpgkey
EOL

# 2. 安裝軟體包
sudo yum update -y
sudo yum install -y postgresql14-server timescaledb-2-postgresql-14 mysql-server

# 3. 初始化 PostgreSQL
sudo postgresql-14-setup initdb
sudo systemctl enable postgresql-14
sudo systemctl start postgresql-14

# 4. 配置 TimescaleDB
sudo timescaledb-tune --quiet --yes
sudo systemctl restart postgresql-14

# 5. 啟動 MySQL
sudo systemctl enable mysqld
sudo systemctl start mysqld

# 後續設置同 Ubuntu 版本...
echo "基礎安裝完成，請繼續數據庫設置..."
```

### 系統服務配置
```bash
# 創建應用服務文件
sudo tee /etc/systemd/system/log-detect.service <<EOF
[Unit]
Description=Log Detect ES Monitoring Backend
After=network.target postgresql.service mysql.service
Requires=postgresql.service mysql.service

[Service]
Type=simple
User=logdetect
Group=logdetect
WorkingDirectory=/opt/log-detect
ExecStart=/opt/log-detect/log-detect-backend
Environment=TIMESCALE_URL=postgresql://monitor:password@localhost:5432/monitoring
Environment=MYSQL_URL=mysql://monitor:password@localhost:3306/config
Environment=BATCH_SIZE=100
Environment=BATCH_TIMEOUT=30s
Environment=ENABLE_MEMORY_CACHE=true
Restart=always
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

# 啟用服務
sudo systemctl daemon-reload
sudo systemctl enable log-detect
sudo systemctl start log-detect
```

### 性能調優配置
```bash
# TimescaleDB 高性能配置
sudo tee -a /var/lib/pgsql/14/data/postgresql.conf <<EOF
# TimescaleDB 高性能配置
shared_preload_libraries = 'timescaledb'
max_connections = 300
shared_buffers = 2GB
effective_cache_size = 6GB
maintenance_work_mem = 1GB
checkpoint_completion_target = 0.9
wal_buffers = 32MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
work_mem = 512MB

# TimescaleDB 專用設置
timescaledb.max_background_workers = 16
EOF

sudo systemctl restart postgresql-14
```

## 📚 使用指南

### 1. 新增 ES 監控
1. 登入系統 (需要 elasticsearch:create 權限)
2. 進入 ES 監控配置頁面
3. 填寫 ES 服務連接資訊
4. 設定監控間隔和告警接收者
5. 測試連接並啟用監控

### 2. 查看監控狀態
1. 進入 ES 監控儀表板
2. 查看即時狀態和歷史趨勢
3. 檢視告警記錄和處理狀態

### 3. 處理告警
1. 收到告警郵件後登入系統
2. 查看具體告警詳情
3. 處理問題後標記告警為已解決

## 🚀 未來擴展

### Phase 2 功能
- **叢集深度監控**: 節點級別詳細監控
- **自動修復**: 簡單問題的自動修復機制
- **預測告警**: 基於歷史數據的預測性告警

### Phase 3 功能
- **多集群監控**: 支援多個 ES 集群統一監控
- **可視化增強**: 更豐富的圖表和分析功能
- **API 監控**: 監控特定 ES API 的使用情況

---

## 🛠️ 實作狀態 (Phase 1)

### 已完成功能

#### ✅ 數據結構設計
- **entities/elasticsearch.go**: 完整的 ES 監控實體定義
  - `ElasticsearchMonitor`: MySQL 配置表結構
  - `ESMetric`: TimescaleDB 時間序列指標（23個欄位）
  - `ESAlert`: 告警記錄結構
  - `ESMonitorStatus`: 監控狀態摘要（前端 Dashboard 用）
  - `ESMetricTimeSeries`: 時間序列數據（前端圖表用）
  - `ESStatistics`: 統計數據（前端儀表板用）
  - `ESAlertThreshold`: 告警閾值配置（含默認值）

#### ✅ 監控收集服務
- **services/es_monitor.go**: ES 健康檢查與指標收集
  - `CheckESHealth()`: 集群健康檢查（HTTP API）
  - `getClusterHealth()`: 集群狀態查詢
  - `getNodesStats()`: 節點統計資訊
  - `getClusterStats()`: 集群統計資訊
  - `ParseMetricsFromCheckResult()`: 指標數據提取
  - `CheckAlertConditions()`: 告警條件檢查
  - `MonitorESCluster()`: 完整監控流程整合
  - 支援 CPU、Memory、Disk、Response Time、Shards 監控

#### ✅ 批量寫入優化
- **services/batch_writer.go**: 擴展支援 ES 監控指標
  - 新增 `esBatch []entities.ESMetric` 批次緩存
  - 新增 `esStmt *sql.Stmt` 預編譯語句
  - 修改 `AddHistory()` 支援類型切換（History / ESMetric）
  - 新增 `flushESMetrics()` ES 指標專用刷新
  - 維持原有 `flushDeviceMetrics()` 設備監控刷新
  - 使用相同的批量寫入策略（batch size + 定時刷新）

#### ✅ 數據庫架構
- **MySQL 配置表**: services/sqltable.go
  - 已添加 `elasticsearch_monitors` 表到 AutoMigrate
  - 儲存監控配置、認證資訊、告警設定

- **TimescaleDB 時間序列表**: postgresql_install.sh
  - `es_metrics` 表（完整 23 欄位）包含：
    - 基礎資訊: time, monitor_id, status, cluster_name, cluster_status
    - 性能指標: response_time, cpu_usage, memory_usage, disk_usage
    - 節點資訊: node_count, data_node_count
    - 查詢性能: query_latency, indexing_rate, search_rate
    - 索引資訊: total_indices, total_documents, total_size_bytes
    - 分片狀態: active_shards, relocating_shards, unassigned_shards
    - 訊息: error_message, warning_message, metadata
  - `es_alert_history` 告警歷史表
  - 自動分區策略：按天分區
  - 壓縮策略：7 天後自動壓縮
  - 保留策略：90 天自動清理
  - 性能索引：monitor_id + time, status + time, cluster_status + time

#### ✅ 查詢服務
- **services/es_monitor_query.go**: ES 監控數據查詢服務
  - `GetLatestMetrics()`: 獲取最新監控指標
  - `GetMetricsTimeSeries()`: 時間序列數據查詢（支援自動聚合間隔）
  - `GetAllMonitorsStatus()`: 所有監控器當前狀態
  - `GetESStatistics()`: 統計摘要數據（Dashboard 用）
  - `GetMonitorMetricsByTimeRange()`: 時間範圍原始數據查詢
  - `GetClusterHealthHistory()`: 集群健康狀態歷史統計
  - `GetPerformanceTrend()`: 性能趨勢分析（支援多種指標）
  - `ExportMetricsToJSON()`: JSON 格式數據導出
  - 使用 TimescaleDB `time_bucket()` 函數實現高效聚合
  - 整合 MySQL 配置與 TimescaleDB 指標查詢

### 待實作功能 (Phase 2)

#### ⏳ API 控制器層
- **controllers/elasticsearch_controller.go**: REST API 實作
  - CRUD API: 監控配置管理
  - 狀態查詢 API: 即時狀態與歷史數據
  - 告警管理 API: 告警查詢與處理
  - Dashboard API: 儀表板數據整合

#### ⏳ Cron 定時任務
- **整合現有 Cron 系統**: 自動化監控收集
  - 根據配置的 interval 自動調度
  - 支援動態新增/移除監控任務
  - 錯誤重試機制

#### ⏳ 告警通知
- **整合現有郵件系統**: 告警通知發送
  - 支援多收件人
  - 自定義告警主題與內容
  - 告警等級分類通知

#### ⏳ 前端整合
- **Dashboard 視覺化**:
  - 監控狀態總覽
  - 即時性能圖表（CPU、Memory、Disk、Response Time）
  - 集群健康趨勢圖
  - 告警歷史列表
  - 監控配置管理界面

### 技術實作亮點

1. **前端視覺化考量**
   - 所有數據結構設計都考慮前端圖表需求
   - `ESMetricTimeSeries` 專為時間序列圖表設計
   - `ESStatistics` 提供 Dashboard 摘要數據
   - 支援自動時間間隔調整（1分鐘/10分鐘/1小時）

2. **高性能批量寫入**
   - 複用現有 BatchWriter 架構，降低系統複雜度
   - 使用預編譯語句減少 SQL 解析開銷
   - 事務批量提交，確保數據一致性
   - 異步刷新策略，不阻塞主流程

3. **TimescaleDB 優化**
   - 使用 Hypertable 自動分區管理
   - 時間序列壓縮節省 90% 儲存空間
   - 智能保留策略自動清理舊數據
   - `time_bucket()` 函數實現毫秒級聚合查詢

4. **彈性擴展設計**
   - 支援多種檢查類型（health, performance, business）
   - 可配置的告警閾值
   - 元數據欄位支援未來擴展
   - 模組化服務架構便於功能迭代

### 部署驗證

#### 測試 TimescaleDB 表結構
```bash
# 執行更新後的部署腳本
bash postgresql_install.sh

# 驗證表結構
psql -U logdetect -d monitoring -c "\d es_metrics"
psql -U logdetect -d monitoring -c "\d es_alert_history"

# 確認 Hypertable 設置
psql -U logdetect -d monitoring -c "SELECT * FROM timescaledb_information.hypertables WHERE hypertable_name='es_metrics';"
```

#### 測試 MySQL 配置表
```bash
# 驗證 MySQL 表創建
mysql -u root -p logdetect -e "SHOW TABLES LIKE 'elasticsearch_monitors';"
mysql -u root -p logdetect -e "DESCRIBE elasticsearch_monitors;"
```

---

**版本**: 1.1 (Phase 1 完成)
**最後更新**: 2025-10-06
**作者**: Log Detect 開發團隊