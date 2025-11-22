#!/bin/bash

# ======================================================
# ES 監控自動排程測試腳本
# ======================================================
# 用途：驗證 ES 監控排程器是否正常運作並寫入資料
# ======================================================

set -e

DB_USER="logdetect"
DB_NAME="monitoring"
DB_HOST="localhost"
DB_PORT="5432"

echo "==================================================="
echo "ES 監控自動排程測試"
echo "==================================================="
echo ""

# ============================================
# 1. 檢查 es_metrics 表是否存在
# ============================================
echo "📌 步驟 1: 檢查 es_metrics 表結構..."
psql -U $DB_USER -d $DB_NAME -h $DB_HOST -p $DB_PORT << 'EOF'
SELECT COUNT(*) as column_count
FROM information_schema.columns
WHERE table_name = 'es_metrics';
EOF

echo ""

# ============================================
# 2. 檢查當前資料筆數（執行前）
# ============================================
echo "📌 步驟 2: 檢查當前資料筆數..."
BEFORE_COUNT=$(psql -U $DB_USER -d $DB_NAME -h $DB_HOST -p $DB_PORT -t -c "SELECT COUNT(*) FROM es_metrics;")
echo "執行前資料筆數: $BEFORE_COUNT"
echo ""

# ============================================
# 3. 檢查 MySQL 中的監控配置
# ============================================
echo "📌 步驟 3: 檢查已啟用的監控配置..."
mysql -u root -p <<'MYSQL_EOF'
USE log_detect;
SELECT
    id,
    name,
    host,
    port,
    `interval`,
    enable_monitor,
    created_at
FROM elasticsearch_monitors
WHERE enable_monitor = 1;
MYSQL_EOF

echo ""

# ============================================
# 4. 等待監控執行（根據 interval 設定）
# ============================================
echo "📌 步驟 4: 等待監控執行..."
echo "提示：請確保應用程式正在運行"
echo "等待 70 秒（預設 interval 為 60 秒 + 10 秒緩衝）..."
echo ""

for i in {70..1}; do
    echo -ne "\r倒數: $i 秒   "
    sleep 1
done
echo ""
echo ""

# ============================================
# 5. 檢查資料筆數（執行後）
# ============================================
echo "📌 步驟 5: 檢查執行後資料筆數..."
AFTER_COUNT=$(psql -U $DB_USER -d $DB_NAME -h $DB_HOST -p $DB_PORT -t -c "SELECT COUNT(*) FROM es_metrics;")
echo "執行後資料筆數: $AFTER_COUNT"
echo ""

# ============================================
# 6. 計算新增資料筆數
# ============================================
NEW_RECORDS=$((AFTER_COUNT - BEFORE_COUNT))
echo "📊 新增資料筆數: $NEW_RECORDS"
echo ""

if [ $NEW_RECORDS -gt 0 ]; then
    echo "✅ 測試成功！排程器正常運作並寫入資料"
    echo ""

    # 顯示最新資料
    echo "📌 最新寫入的資料:"
    psql -U $DB_USER -d $DB_NAME -h $DB_HOST -p $DB_PORT << 'EOF'
SELECT
    time,
    monitor_id,
    status,
    cluster_name,
    cluster_status,
    response_time,
    cpu_usage,
    memory_usage,
    disk_usage,
    node_count,
    total_indices,
    total_documents
FROM es_metrics
ORDER BY time DESC
LIMIT 5;
EOF
else
    echo "❌ 測試失敗！排程器未寫入資料"
    echo ""
    echo "🔍 故障排查步驟:"
    echo "1. 檢查應用程式是否正在運行:"
    echo "   ps aux | grep log-detect"
    echo ""
    echo "2. 檢查應用程式日誌:"
    echo "   tail -f logs/app.log | grep -E '(ES Monitor|Starting monitoring)'"
    echo ""
    echo "3. 檢查監控配置是否啟用:"
    echo "   mysql -u root -p -e 'SELECT * FROM log_detect.elasticsearch_monitors;'"
    echo ""
    echo "4. 檢查 ES 連線是否正常:"
    echo "   curl -X POST http://localhost:8006/api/v1/elasticsearch/monitors/1/test \\"
    echo "        -H 'Authorization: Bearer YOUR_TOKEN'"
    echo ""
fi

echo ""
echo "==================================================="
echo "測試完成"
echo "==================================================="
