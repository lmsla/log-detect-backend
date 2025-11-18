# 📊 歷史數據管理指南

## 🎯 概述

歷史記錄存儲在 SQL 確實會面臨數據量持續增長的問題。本指南提供了完整的數據管理策略，幫助您有效管理歷史數據。

## 📈 數據量分析

### 當前數據增長速度

以 100 個設備、每 5 分鐘檢查一次為例：

| 時間週期 | 單日記錄數 | 每月記錄數 | 每年記錄數 | 估計大小 |
|---------|-----------|-----------|-----------|---------|
| 100設備×12檢查 | 28,800 條 | 864,000 條 | 10,512,000 條 | ~500MB/年 |

**實際增長可能更快**：
- 設備數量會增加
- 檢查頻率可能提高
- 需要存儲更多元數據

## 🛠️ 數據管理策略

### 1. **數據清理 (Data Cleaning)**

#### 自動清理舊數據
```bash
# 清理 90 天前的歷史記錄
DELETE FROM histories WHERE date < DATE_SUB(NOW(), INTERVAL 90 DAY);
```

#### API 調用
```bash
# 清理 180 天前的數據
curl -X DELETE "http://localhost:8006/api/v1/admin/data/clean-history?days=180" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

### 2. **數據歸檔 (Data Archiving)**

#### 歸檔策略
```sql
-- 創建歸檔表
CREATE TABLE history_archives LIKE histories;

-- 將舊數據移動到歸檔表
INSERT INTO history_archives SELECT * FROM histories WHERE date < '2024-01-01';
DELETE FROM histories WHERE date < '2024-01-01';
```

#### API 調用
```bash
# 將 365 天前的數據歸檔
curl -X POST "http://localhost:8006/api/v1/admin/data/archive-history?days=365" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

### 3. **數據聚合 (Data Aggregation)**

#### 創建每日統計表
```sql
CREATE TABLE history_daily_stats (
    date DATE PRIMARY KEY,
    logname VARCHAR(50),
    device_group VARCHAR(50),
    total_checks INT DEFAULT 0,
    online_count INT DEFAULT 0,
    offline_count INT DEFAULT 0,
    uptime_rate DECIMAL(5,2) DEFAULT 0.00,
    avg_response_time DECIMAL(10,2) DEFAULT 0.00
);
```

#### 生成聚合數據
```bash
# 生成昨天的統計數據
curl -X POST "http://localhost:8006/api/v1/admin/data/create-aggregates" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"

# 生成指定日期的統計
curl -X POST "http://localhost:8006/api/v1/admin/data/create-aggregates?date=2024-01-15" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

### 4. **存儲優化 (Storage Optimization)**

#### 表分區 (Table Partitioning)
```sql
-- 按月份分區
ALTER TABLE histories PARTITION BY RANGE (YEAR(date) * 100 + MONTH(date)) (
    PARTITION p202401 VALUES LESS THAN (202402),
    PARTITION p202402 VALUES LESS THAN (202403),
    PARTITION p202403 VALUES LESS THAN (202404)
);
```

#### 數據壓縮
```sql
-- 啟用壓縮
ALTER TABLE history_archives ROW_FORMAT=COMPRESSED;
```

## 📊 監控數據大小

### 檢查存儲統計
```bash
# 查看各表存儲統計
curl -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
     http://localhost:8006/api/v1/admin/data/storage-stats
```

響應示例：
```json
[
  {
    "table_name": "histories",
    "record_count": 864000,
    "data_size_mb": 245.67,
    "index_size_mb": 89.23,
    "total_size_mb": 334.90,
    "oldest_record": "2024-01-01",
    "newest_record": "2024-03-31"
  },
  {
    "table_name": "history_archives",
    "record_count": 2100000,
    "data_size_mb": 589.45,
    "index_size_mb": 156.78,
    "total_size_mb": 746.23,
    "oldest_record": "2023-01-01",
    "newest_record": "2023-12-31"
  }
]
```

## 🔄 自動化數據管理

### 定時任務配置

#### 每日統計生成
```bash
# crontab 配置 - 每天凌晨 1 點生成前一天統計
0 1 * * * curl -X POST "http://localhost:8006/api/v1/admin/data/create-aggregates"
```

#### 每週數據清理
```bash
# 每週日凌晨 2 點清理 180 天前的數據
0 2 * * 0 curl -X DELETE "http://localhost:8006/api/v1/admin/data/clean-history?days=180"
```

#### 每月數據歸檔
```bash
# 每月 1 日凌晨 3 點歸檔 365 天前的數據
0 3 1 * * curl -X POST "http://localhost:8006/api/v1/admin/data/archive-history?days=365"
```

### 智能清理策略

#### 基於業務需求的清理
```go
// 保留策略
keepDays := map[string]int{
    "critical_devices": 365,  // 關鍵設備保留1年
    "normal_devices": 180,    // 普通設備保留6個月
    "test_devices": 30,       // 測試設備保留1個月
}
```

#### 基於存儲空間的清理
```go
// 當存儲空間超過閾值時自動清理
maxStorageGB := 100.0
if currentStorageGB > maxStorageGB {
    // 清理最舊的數據
    cleanOldestData(30) // 清理 30 天前數據
}
```

## 🎨 查詢優化策略

### 1. **索引優化**
```sql
-- 主要查詢索引
CREATE INDEX idx_histories_device_date ON histories (name, date);
CREATE INDEX idx_histories_status ON histories (status);
CREATE INDEX idx_histories_timestamp ON histories (timestamp);

-- 統計查詢索引
CREATE INDEX idx_daily_stats_date ON history_daily_stats (date, logname, device_group);
```

### 2. **查詢重寫**
```sql
-- 原始查詢 (慢)
SELECT * FROM histories WHERE date >= '2024-01-01' ORDER BY timestamp DESC LIMIT 1000;

-- 優化後 (快)
SELECT * FROM histories
WHERE date >= '2024-01-01'
    AND timestamp >= UNIX_TIMESTAMP('2024-01-01')
ORDER BY timestamp DESC LIMIT 1000;
```

### 3. **分頁查詢**
```go
// 使用游標分頁替代 OFFSET
func GetHistoryPage(cursor int64, limit int) []History {
    return db.Where("timestamp < ?", cursor).
        Order("timestamp DESC").
        Limit(limit).
        Find(&histories)
}
```

## 📈 擴展策略

### 1. **冷熱數據分離**
```
熱數據 (Hot Data): histories 表
├── 最近 90 天的數據
├── 高頻查詢
└── 需要即時分析

冷數據 (Cold Data): history_archives 表
├── 90 天前的數據
├── 低頻查詢
├── 長期保存
└── 壓縮存儲
```

### 2. **多級存儲**
```
內存 (Memory): Redis 緩存最近 1 小時數據
SSD (Fast Storage): 最近 30 天數據
HDD (Slow Storage): 30 天前數據
磁帶/雲存儲 (Archive): 1 年前數據
```

### 3. **數據湖集成**
```go
// 將歷史數據同步到數據湖
func ExportToDataLake() {
    // 將舊數據導出到 Parquet 格式
    // 上傳到 S3/MinIO
    // 在 ClickHouse/Presto 中創建外部表
}
```

## 🚨 告警和監控

### 數據大小告警
```bash
# 當表大小超過閾值時發送告警
TABLE_SIZE=$(mysql -e "SELECT data_length+index_length FROM information_schema.tables WHERE table_name='histories'" logdetect)
if [ $TABLE_SIZE -gt 1073741824 ]; then  # 1GB
    send_alert "History table size exceeded 1GB"
fi
```

### 清理任務監控
```bash
# 檢查清理任務是否正常運行
LAST_CLEAN=$(mysql -e "SELECT MAX(created_at) FROM clean_logs" logdetect)
if [ $(($(date +%s) - $(date -d "$LAST_CLEAN" +%s))) -gt 86400 ]; then
    send_alert "Data cleanup job hasn't run in 24 hours"
fi
```

## 📋 最佳實踐

### 1. **定期維護**
- 每天生成統計數據
- 每週清理舊數據
- 每月歸檔歷史數據
- 每季度檢查存儲使用情況

### 2. **監控指標**
- 表大小趨勢
- 查詢性能
- 清理任務狀態
- 數據完整性

### 3. **備份策略**
- 熱數據：每日備份
- 冷數據：每週備份
- 歸檔數據：每月備份

### 4. **災難恢復**
- 定義 RTO/RPO
- 測試恢復流程
- 準備降級方案

## 🎯 總結

通過實施這套數據管理策略，您可以：

1. ✅ **控制數據增長** - 定期清理和歸檔
2. ✅ **優化查詢性能** - 索引和聚合優化
3. ✅ **降低存儲成本** - 冷熱數據分離
4. ✅ **保持系統穩定** - 自動化管理和監控
5. ✅ **支持業務需求** - 平衡數據保留和性能

**關鍵原則**：數據應該按照業務價值和查詢頻率進行分層管理，而不是一股腦兒存儲在單一表中。
