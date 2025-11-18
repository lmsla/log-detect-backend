# Elasticsearch 監控 API 實作狀態報告

## 📊 總覽

- **Phase 1 完成度**: 9/18 API (50%)
- **已實作**: 9 個端點
- **待實作**: 9 個端點（Phase 2）

## ✅ Phase 1 已實作 (9 個)

### 監控配置管理 (7 個)

| 端點 | 方法 | 功能 | 檔案位置 |
|------|------|------|----------|
| `/api/v1/elasticsearch/monitors` | GET | 獲取所有監控配置 | `controller/elasticsearch.go:68` |
| `/api/v1/elasticsearch/monitors` | POST | 新增監控配置 | `controller/elasticsearch.go:17` |
| `/api/v1/elasticsearch/monitors` | PUT | 更新監控配置 | `controller/elasticsearch.go:42` |
| `/api/v1/elasticsearch/monitors/{id}` | GET | 獲取特定配置 | `controller/elasticsearch.go:79` |
| `/api/v1/elasticsearch/monitors/{id}` | DELETE | 刪除監控配置 | `controller/elasticsearch.go:98` |
| `/api/v1/elasticsearch/monitors/test` | POST | 測試連接 | `controller/elasticsearch.go:119` |
| `/api/v1/elasticsearch/monitors/{id}/toggle` | POST | 啟用/停用監控 | `controller/elasticsearch.go:143` |

### 狀態查詢 (2 個)

| 端點 | 方法 | 功能 | 檔案位置 |
|------|------|------|----------|
| `/api/v1/elasticsearch/status` | GET | 獲取所有監控器狀態 | `controller/elasticsearch.go:171` |
| `/api/v1/elasticsearch/statistics` | GET | 獲取統計數據 | `controller/elasticsearch.go:193` |

## ⏳ Phase 2 待實作 (9 個)

### 單個監控器詳細數據 (3 個)

| 端點 | 方法 | 功能 | 說明 |
|------|------|------|------|
| `/api/v1/elasticsearch/status/{id}` | GET | 獲取特定監控器狀態 | 查詢服務已支援 `GetLatestMetrics()` |
| `/api/v1/elasticsearch/status/{id}/history` | GET | 獲取歷史狀態記錄 | 查詢服務已支援 `GetMonitorMetricsByTimeRange()` |
| `/api/v1/elasticsearch/status/{id}/trends` | GET | 獲取趨勢數據 | 查詢服務已支援 `GetPerformanceTrend()` |

### 告警管理 (4 個)

| 端點 | 方法 | 功能 | 說明 |
|------|------|------|------|
| `/api/v1/elasticsearch/alerts` | GET | 獲取告警列表 | 需整合 `es_alert_history` 表查詢 |
| `/api/v1/elasticsearch/alerts/{id}` | GET | 獲取告警詳情 | 需實作告警詳細資訊查詢 |
| `/api/v1/elasticsearch/alerts/{id}/resolve` | POST | 解決告警 | 需實作告警狀態更新邏輯 |
| `/api/v1/elasticsearch/alerts/{id}/acknowledge` | PUT | 確認告警 | 需實作告警確認邏輯 |

### 儀表板整合 (2 個)

| 端點 | 方法 | 功能 | 說明 |
|------|------|------|------|
| `/api/v1/elasticsearch/dashboard` | GET | ES 監控儀表板數據 | 可整合現有 `/statistics` 和 `/status` |
| `/api/v1/elasticsearch/metrics/{id}` | GET | 獲取指標數據 | 查詢服務已支援 `GetMetricsTimeSeries()` |

## 🔧 已修正的文檔不一致

### 1. 更新監控配置路徑
- **修正前**: `PUT /api/v1/elasticsearch/monitors/{id}`
- **修正後**: `PUT /api/v1/elasticsearch/monitors` (ID 從 request body 傳遞)
- **原因**: 後端實作從 body 讀取 ID，符合其他 API 的一致性

### 2. 測試連接路徑
- **修正前**: `POST /api/v1/elasticsearch/monitors/{id}/test`
- **修正後**: `POST /api/v1/elasticsearch/monitors/test`
- **原因**: 測試連接不需要已存在的監控 ID，可直接測試連接參數

### 3. 新增統計 API
- **新增**: `GET /api/v1/elasticsearch/statistics`
- **說明**: 替代原 `/summary` 端點，提供更詳細的統計數據

### 4. 新增 Toggle API
- **新增**: `POST /api/v1/elasticsearch/monitors/{id}/toggle`
- **說明**: 動態啟用/停用監控，不需要完整的 PUT 更新

## 📝 OpenAPI 規範同步狀態

### 已同步到 openapi.yml
✅ 所有 Phase 1 的 9 個 API 已完整定義
✅ 包含完整的 request/response schemas
✅ 包含錯誤碼定義 (400, 401, 403, 404, 500)
✅ 包含 3 個數據模型：
  - `ElasticsearchMonitor`
  - `ESMonitorStatus`
  - `ESStatistics`

### 未同步到 openapi.yml
⏳ Phase 2 的 9 個待實作 API（將在實作時同步）

## 🚀 前端對接建議

### Phase 1 可立即使用的功能

1. **監控配置管理頁面**
   - 列表展示（GET /monitors）
   - 新增配置（POST /monitors）
   - 編輯配置（PUT /monitors）
   - 刪除配置（DELETE /monitors/{id}）
   - 測試連接（POST /monitors/test）
   - 啟用/停用（POST /monitors/{id}/toggle）

2. **監控狀態總覽頁面**
   - 所有監控器狀態卡片（GET /status）
   - 統計數據儀表板（GET /statistics）

### Phase 2 需等待後端實作

1. **單個監控器詳細頁面**
   - 詳細狀態展示
   - 歷史數據圖表
   - 性能趨勢分析

2. **告警管理頁面**
   - 告警列表
   - 告警處理

3. **高級儀表板**
   - 整合儀表板視圖
   - 時序數據圖表

## 📦 實作優先級建議

### 高優先級（前端急需）
1. `GET /elasticsearch/status/{id}` - 單個監控器詳細狀態
2. `GET /elasticsearch/metrics/{id}` - 指標數據（圖表用）
3. `GET /elasticsearch/status/{id}/trends` - 趨勢數據（圖表用）

### 中優先級（功能完整性）
4. `GET /elasticsearch/alerts` - 告警列表
5. `GET /elasticsearch/status/{id}/history` - 歷史記錄
6. `GET /elasticsearch/dashboard` - 整合儀表板

### 低優先級（可選功能）
7. `POST /elasticsearch/alerts/{id}/resolve` - 解決告警
8. `PUT /elasticsearch/alerts/{id}/acknowledge` - 確認告警
9. `GET /elasticsearch/alerts/{id}` - 告警詳情

## 🔗 相關檔案

- **API 文檔**: `docs/elasticsearch-monitoring.md`
- **OpenAPI 規範**: `docs/openapi.yml`
- **Controller**: `controller/elasticsearch.go`
- **Service**: `services/es_monitor_service.go`
- **Query Service**: `services/es_monitor_query.go`
- **Router**: `router/router.go`
- **Entity**: `entities/elasticsearch.go`

---

**更新日期**: 2025-10-06
**版本**: 1.0
