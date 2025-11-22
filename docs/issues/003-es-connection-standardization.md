# Issue #003: ES 連線配置標準化

**狀態**: 📋 待討論
**優先級**: 🟡 中
**建立日期**: 2025-11-22
**負責人**: 待指派

---

## 議題描述

目前 ES 連線配置存在重複：
1. **es_connections** 表：統一的 ES 連線管理（用於日誌監控）
2. **elasticsearch_monitors** 表：有自己的 host/port/auth 欄位（用於 ES 健康監控）

這導致：
- 同一個 ES 叢集可能需要在兩個地方重複配置
- 兩套連線測試 API（`/api/v1/ESConnection/Test` 和 `/api/v1/ESMonitor/Test/:id`）
- 配置管理分散，不易維護

---

## 現有架構

```
┌─────────────────────────────────────────────────────────────────┐
│                     es_connections                              │
│                   (ES 連線配置表)                                │
└─────────────────┬───────────────────────────┬───────────────────┘
                  │                           │
                  ▼                           ▼
┌─────────────────────────────┐   ┌───────────────────────────────┐
│        Index (索引)          │   │  elasticsearch_monitors       │
│  es_connection_id (外鍵)     │   │  es_connection_id (可選外鍵)   │
│  Pattern (索引模式)          │   │  ──────────────────────        │
│  DeviceGroup (設備群組名)    │   │  Host/Port/Auth (獨立配置) ❌  │
│  Field (欄位名)              │   │                               │
└──────────────┬──────────────┘   └───────────────────────────────┘
               │
               │ DeviceGroup 字串匹配
               ▼
┌─────────────────────────────┐
│        Device (設備)         │
│  DeviceGroup (設備群組名)    │
│  Name (設備名稱)             │
└─────────────────────────────┘
```

### 日誌監控 ES 連線流程

```
1. 排程觸發 → Target + Index
2. Index.ESConnectionID → ESConnectionManager.GetClient()
3. 用 Index.Pattern 查 ES (如 "logstash-nginx-*")
4. 用 Index.Field 取得文件中的設備欄位 (如 "host.keyword")
5. 比對 Device 表中 DeviceGroup 相同的設備清單
6. 回報：哪些設備有日誌、哪些沒有
```

### ES 健康監控連線流程（現況）

```
1. 排程觸發 → ElasticsearchMonitor
2. 如果 es_connection_id 有值 → 用 ESConnectionManager
   如果 es_connection_id 為空 → 用自己的 host/port/auth 建立連線
3. 呼叫 ES API 取得健康狀態
4. 記錄指標到 TimescaleDB
```

---

## 解決方案

### 目標

讓 `es_connections` 成為**唯一的 ES 連線配置來源**。

### 整合後架構

```
┌─────────────────────────────────────────────────────────────────┐
│                     es_connections                              │
│                 (唯一 ES 連線配置來源)                           │
└─────────────────┬───────────────────────────┬───────────────────┘
                  │                           │
                  ▼                           ▼
┌─────────────────────────────┐   ┌───────────────────────────────┐
│        Index (索引)          │   │  elasticsearch_monitors       │
│  es_connection_id (外鍵)     │   │  es_connection_id (必填外鍵)  │
│  Pattern                     │   │  CheckType, Interval...       │
│  DeviceGroup, Field          │   │  AlertThreshold...            │
└──────────────┬──────────────┘   └───────────────────────────────┘
               │
               ▼
┌─────────────────────────────┐
│        Device (設備)         │
└─────────────────────────────┘
```

---

## 實作計畫

### Phase 1：資料庫異動

```sql
-- elasticsearch_monitors 移除冗餘欄位，es_connection_id 改為必填
ALTER TABLE elasticsearch_monitors
  DROP COLUMN host,
  DROP COLUMN port,
  DROP COLUMN username,
  DROP COLUMN password,
  DROP COLUMN enable_auth,
  MODIFY es_connection_id INT UNSIGNED NOT NULL;

-- 新增外鍵約束
ALTER TABLE elasticsearch_monitors
  ADD CONSTRAINT fk_es_monitors_connection
  FOREIGN KEY (es_connection_id) REFERENCES es_connections(id)
  ON UPDATE CASCADE ON DELETE RESTRICT;
```

**注意**：需要先處理現有資料的遷移（將 host/port/auth 轉移到 es_connections）

### Phase 2：實體層修改

**檔案**: `entities/elasticsearch.go`

```go
// 移除這些欄位
// Host              string
// Port              int
// Username          string
// Password          string
// EnableAuth        bool

// es_connection_id 改為必填
ESConnectionID int           `gorm:"not null;index" json:"es_connection_id"`
ESConnection   *ESConnection `gorm:"foreignKey:ESConnectionID" json:"es_connection,omitempty"`
```

### Phase 3：服務層修改

**檔案**: `services/es_monitor.go`

- 移除獨立建立 ES 客戶端的邏輯
- 統一使用 `ESConnectionManager.GetClient(monitor.ESConnectionID)`

```go
// 現況
func CheckESHealth(monitor entities.ElasticsearchMonitor) {
    // 如果有 es_connection_id 用 manager，否則用自己的 host/port
}

// 改為
func CheckESHealth(monitor entities.ElasticsearchMonitor) {
    client := services.GetESConnectionManager().GetClient(monitor.ESConnectionID)
    // 使用 client 檢查健康狀態
}
```

### Phase 4：API 層修改

**檔案**: `controller/es_monitor.go`

- 移除 `/api/v1/ESMonitor/Test/:id` 端點
- Create/Update API 驗證 es_connection_id 必填

**檔案**: `router/router.go`

- 移除 Test 路由

### Phase 5：前端調整

- ES Monitor 設定頁面：
  - 移除 Host/Port/Username/Password/EnableAuth 輸入欄位
  - 新增「ES 連線」下拉選單（從 es_connections 取得）
- 新增 ES 監控前，必須先建立 ES 連線

---

## 影響範圍

| 檔案 | 改動內容 |
|------|---------|
| `entities/elasticsearch.go` | 移除 Host/Port/Auth 欄位，ESConnectionID 改必填 |
| `services/es_monitor.go` | 移除獨立連線邏輯，統一用 ESConnectionManager |
| `controller/es_monitor.go` | 移除 Test API，驗證 es_connection_id 必填 |
| `router/router.go` | 移除 ESMonitor Test 路由 |
| `migrations/mysql/006_elasticsearch_monitors.up.sql` | 移除冗餘欄位 |
| `docs/openapi.yml` | 更新 API 文件 |

---

## 資料遷移策略

對於現有的 `elasticsearch_monitors` 資料：

1. **有 es_connection_id**：不需處理
2. **沒有 es_connection_id（使用自己的 host/port）**：
   - 自動在 es_connections 建立對應記錄
   - 更新 es_connection_id 指向新記錄

```sql
-- 遷移腳本範例
INSERT INTO es_connections (name, host, port, username, password, enable_auth, use_tls)
SELECT
  CONCAT('ES Monitor - ', name),
  host,
  port,
  username,
  password,
  enable_auth,
  1  -- 預設啟用 TLS
FROM elasticsearch_monitors
WHERE es_connection_id IS NULL;

-- 更新外鍵
UPDATE elasticsearch_monitors em
JOIN es_connections ec ON ec.name = CONCAT('ES Monitor - ', em.name)
SET em.es_connection_id = ec.id
WHERE em.es_connection_id IS NULL;
```

---

## 優點

| 面向 | 改善 |
|------|------|
| **配置管理** | 統一入口，避免重複配置 |
| **程式碼** | 移除重複的連線邏輯 |
| **API** | 減少一個測試端點 |
| **維護性** | 單一來源，修改密碼只需改一處 |

## 風險

| 風險 | 緩解措施 |
|------|---------|
| 現有資料遷移失敗 | 先在測試環境驗證遷移腳本 |
| 前端配合時程 | 後端可先完成，前端逐步調整 |
| 破壞性變更 | 提供回滾腳本 |

---

## 討論事項

- [ ] 確認是否執行此整合
- [ ] 確認前端開發時程
- [ ] 確認資料遷移策略

---

**最後更新**: 2025-11-22
**更新者**: Claude (AI Assistant)
