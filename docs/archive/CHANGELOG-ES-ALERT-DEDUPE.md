# ES 監控告警去重功能更新 (2025-10-08)

## 📝 更新摘要

新增告警去重時間窗口可配置功能，允許用戶針對不同監控器設置不同的去重時間窗口。

## ✨ 新增功能

### 1. 新增配置欄位 `alert_dedupe_window`

**位置**: `entities.ElasticsearchMonitor`

**類型**: `int`

**預設值**: 300 秒（5 分鐘）

**說明**: 告警去重時間窗口（秒），在此時間窗口內，相同監控器、相同類型、相同嚴重性的告警只會記錄和通知一次。

### 2. 去重邏輯更新

**檔案**: `services/es_monitor.go`

**變更**:
- `CreateAlert()` 函數改為接受 `monitor` 參數並返回 `bool`
- `isDuplicateAlert()` 函數使用監控器配置的去重窗口而非固定值
- 只有成功創建新告警時才發送郵件通知

**去重條件**:
```
monitor_id + alert_type + severity + status='active' + 時間窗口
```

## 📋 資料庫變更

### MySQL: elasticsearch_monitors 表

**新增欄位**:
```sql
ALTER TABLE elasticsearch_monitors
  ADD COLUMN alert_dedupe_window INT DEFAULT 300
  COMMENT '告警去重時間窗口(秒,預設300秒=5分鐘)';
```

**執行腳本**: `docs/troubleshooting/add_alert_dedupe_window.sql`

## 📚 文檔更新

### 1. API 規格文檔
**檔案**: `docs/spec/api/elasticsearch-api-spec.md`

**更新內容**:
- ElasticsearchMonitor 模型新增 `alert_dedupe_window` 欄位
- 添加欄位說明和建議設置

### 2. 資料庫 Schema 文檔
**檔案**: `docs/spec/database/schema-validation.md`

**更新內容**:
- ElasticsearchMonitor 實體定義更新
- SQL 表結構預期新增 `alert_dedupe_window` 欄位

### 3. 實作狀態文檔
**檔案**: `docs/spec/api/elasticsearch-implementation-status.md`

**更新內容**:
- Phase 3 進度更新為 100%
- 告警管理 API 狀態更新為 ✅
- 告警通知功能狀態更新為 ✅
- 新增去重時間窗口配置說明

## 🔧 使用方式

### 創建監控時指定去重窗口

```json
POST /api/v1/elasticsearch/monitors
{
  "name": "ES-Production",
  "host": "localhost",
  "port": 9200,
  "interval": 30,
  "alert_dedupe_window": 120,  // 2分鐘去重窗口
  "receivers": ["admin@example.com"]
}
```

### 更新現有監控的去重窗口

```json
PUT /api/v1/elasticsearch/monitors
{
  "id": 1,
  "alert_dedupe_window": 180  // 更新為3分鐘
}
```

### 批量調整建議

```sql
-- 高頻檢查（30秒）設為 2 分鐘去重
UPDATE elasticsearch_monitors
SET alert_dedupe_window = 120
WHERE interval = 30;

-- 標準檢查（60秒）設為 5 分鐘去重
UPDATE elasticsearch_monitors
SET alert_dedupe_window = 300
WHERE interval = 60;

-- 低頻檢查（>=5分鐘）設為 10 分鐘去重
UPDATE elasticsearch_monitors
SET alert_dedupe_window = 600
WHERE interval >= 300;
```

## 💡 建議設置

| 檢查間隔 (interval) | 建議去重窗口 | 說明 |
|-------------------|------------|------|
| 30 秒 | 60-120 秒 | 高頻檢查，短窗口避免漏告警 |
| 60 秒 | 180-300 秒 | 標準設置，平衡告警及時性和騷擾度 |
| 5 分鐘+ | 600-1800 秒 | 低頻檢查，長窗口減少重複通知 |

## 🚀 部署步驟

### 1. 更新資料庫

```bash
# 連接到 MySQL
mysql -u monitor -p config

# 執行 SQL 腳本
source /path/to/docs/troubleshooting/add_alert_dedupe_window.sql
```

### 2. 重啟應用

```bash
# 重啟後端服務
# GORM AutoMigrate 會自動處理新欄位（如果尚未手動添加）
```

### 3. 驗證功能

```bash
# 查看日誌確認去重生效
tail -f log_record/LogDetect-*.log | grep -E "Alert Created|Skipping duplicate"
```

**預期日誌**:
```
WARN  ES Alert Created [high][performance]: Memory usage high: 88.89%
INFO  Alert notification sent to 1 receivers for monitor: ES-93
DEBUG Skipping duplicate alert for monitor 2: Memory usage high: 88.86%
(不會再看到重複的 "Alert notification sent")
```

## ⚠️ 注意事項

1. **向後兼容**: 未設置 `alert_dedupe_window` 的監控器會使用預設值 300 秒
2. **零值處理**: 如果設為 0 或負數，會自動使用預設值 300 秒
3. **最小值建議**: 不建議設置小於 30 秒，可能導致漏告警
4. **與檢查間隔的關係**: 建議去重窗口 >= 檢查間隔的 2 倍

## 🐛 已修復問題

1. ✅ 告警去重邏輯已實作但郵件仍重複發送 → 修復：只有創建新告警才發送通知
2. ✅ `es_alert_history` 表缺少欄位 → 提供修復 SQL: `fix_es_alert_history_columns.sql`
3. ✅ metadata JSON 格式錯誤 → 修復：添加 JSON 驗證和空值處理
4. ✅ 去重時間窗口寫死 → 修復：改為可配置欄位

## 📊 影響範圍

### 程式碼變更
- `entities/elasticsearch.go` - 新增欄位
- `services/es_monitor.go` - 去重邏輯重構
- `controller/elasticsearch.go` - Swagger 自動更新

### 資料庫變更
- `elasticsearch_monitors` 表 - 新增欄位

### 文檔變更
- `docs/spec/api/elasticsearch-api-spec.md`
- `docs/spec/database/schema-validation.md`
- `docs/spec/api/elasticsearch-implementation-status.md`
- `docs/troubleshooting/add_alert_dedupe_window.sql` (新增)

## ✅ 測試驗證

### 功能測試
- [ ] 創建監控時可指定 `alert_dedupe_window`
- [ ] 更新監控時可修改 `alert_dedupe_window`
- [ ] 去重邏輯按配置的窗口生效
- [ ] 重複告警只記錄一次且不發送郵件
- [ ] 超過窗口後可再次創建告警

### 兼容性測試
- [ ] 現有監控器使用預設值 300 秒
- [ ] 零值或負值自動使用預設值
- [ ] Swagger 文檔正確顯示新欄位

## 📞 相關聯絡

- 開發者: Claude
- 日期: 2025-10-08
- 版本: v1.0
