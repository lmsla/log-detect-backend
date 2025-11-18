# Issue #001: 整合 ES 連線管理架構

**狀態**: 🚧 進行中 (Phase 1-3 已完成)
**優先級**: 🔴 高
**建立日期**: 2025-11-18
**負責人**: 待指派
**預計時程**: 5-6 天
**目前進度**: Phase 3/5 完成 (60%)

---

## 📌 議題描述

目前系統中存在兩種 Elasticsearch 監控行為：
1. **即時裝置監控**：透過 `setting.yml` 配置單一 ES 連線
2. **Elasticsearch 健康監控**：透過 MySQL 的 `elasticsearch_monitors` 表配置多節點

兩者之間缺乏關聯性，導致以下問題：
- 無法讓不同的設備群組（DeviceGroup）連接到不同的 ES Cluster
- 裝置監控與健康監控無法共享 ES 連線配置
- 配置管理分散（setting.yml vs. 資料庫）

---

## 🔍 背景分析

### 現況架構

#### 即時裝置監控 (Device Monitoring)
- **配置來源**: `setting.yml` 中的 `es:` 區塊
- **資料結構**:
  ```go
  type es struct {
    URL            []string   // ES 連線 URL
    SourceAccount  string     // 認證帳號
    SourcePassword string     // 認證密碼
  }
  ```
- **初始化**: `clients/es.go:14-43` - 全域單一客戶端 `clients.ES`
- **使用流程**:
  ```
  Target (告警規則)
    └─ Index (監控配置)
        ├─ Pattern: "logstash-nginx-*"
        ├─ DeviceGroup: "web-servers"
        └─ Field: "host.keyword"
            ↓
  SearchRequest() → clients.ES (固定單一連線)
  ```
- **限制**: ❌ 只能連接單一 Elasticsearch cluster

#### Elasticsearch 健康監控 (ES Health Monitoring)
- **配置來源**: MySQL 的 `elasticsearch_monitors` 表
- **資料結構**: `entities.ElasticsearchMonitor`
  - Host, Port, Username, Password, EnableAuth
  - CheckType, Interval, EnableMonitor
  - Receivers, Subject, AlertThreshold
- **使用方式**: `services/es_monitor.go` - 為每個監控配置建立獨立 HTTP 客戶端
- **優勢**: ✅ 支援多節點、動態配置、Web UI 管理

### 實際需求場景

```
Index A (nginx logs, DeviceGroup=web-servers)    → ES Cluster 1 (10.99.1.213:9200)
Index B (app logs, DeviceGroup=app-servers)      → ES Cluster 2 (10.99.1.64:9200)
Index C (system logs, DeviceGroup=linux-servers) → ES Cluster 1 (10.99.1.213:9200)

同時需要：
ES Cluster 1 → 健康監控 (ElasticsearchMonitor #1)
ES Cluster 2 → 健康監控 (ElasticsearchMonitor #2)
```

### 核心問題

1. ❌ 所有 Index 共用同一個 ES 連線（`clients.ES`）
2. ❌ 無法讓不同 DeviceGroup 連接到不同的 ES Cluster
3. ⚠️ 即時監控和健康監控無法共享 ES 連線配置
4. ⚠️ 配置管理分散，不易維護

---

## ✅ 解決方案

### 方案概述：ES 連線池架構

建立 **`ESConnection`（ES 連線配置）** 作為基礎實體，讓 `Index` 和 `ElasticsearchMonitor` 都關聯到它。

### 核心設計

#### 1. 新增基礎實體 `es_connections` 表

```sql
CREATE TABLE `es_connections` (
  `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  `name` VARCHAR(100) NOT NULL COMMENT '連線名稱，例如：生產環境-主集群',
  `host` VARCHAR(255) NOT NULL COMMENT 'ES 主機地址',
  `port` INT NOT NULL DEFAULT 9200 COMMENT 'ES 端口',
  `username` VARCHAR(100) COMMENT '認證用戶名',
  `password` VARCHAR(255) COMMENT '認證密碼',
  `enable_auth` BOOLEAN DEFAULT FALSE COMMENT '是否啟用認證',
  `use_tls` BOOLEAN DEFAULT TRUE COMMENT '是否使用 TLS',
  `is_default` BOOLEAN DEFAULT FALSE COMMENT '是否為預設連線',
  `description` TEXT COMMENT '連線描述',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` TIMESTAMP NULL,

  UNIQUE KEY `idx_name` (`name`),
  INDEX `idx_is_default` (`is_default`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Elasticsearch 連線配置表';
```

#### 2. 修改 `indices` 表 - 關聯 ES 連線

```sql
ALTER TABLE `indices`
ADD COLUMN `es_connection_id` INT UNSIGNED COMMENT 'ES 連線 ID（外鍵到 es_connections）',
ADD CONSTRAINT `fk_indices_es_connection`
  FOREIGN KEY (`es_connection_id`)
  REFERENCES `es_connections`(`id`)
  ON UPDATE CASCADE
  ON DELETE RESTRICT;
```

**設計說明**:
- `es_connection_id` 可為 NULL（向後兼容，使用預設連線）
- 不同的 Index 可指向不同的 ESConnection
- 同一個 ESConnection 可被多個 Index 共用

#### 3. 修改 `elasticsearch_monitors` 表 - 支援複用連線

```sql
ALTER TABLE `elasticsearch_monitors`
ADD COLUMN `es_connection_id` INT UNSIGNED COMMENT 'ES 連線 ID（複用 indices 的 ES 連線）',
ADD CONSTRAINT `fk_es_monitors_connection`
  FOREIGN KEY (`es_connection_id`)
  REFERENCES `es_connections`(`id`)
  ON UPDATE CASCADE
  ON DELETE SET NULL;
```

**設計說明**:
- `es_connection_id` 為 NULL：使用自己的 host/port 配置（獨立監控目標）
- `es_connection_id` 有值：監控的是 indices 使用的 ES（複用連線配置）

#### 4. ES 連線管理器 (Singleton Pattern)

```go
// services/es_connection_manager.go

type ESConnectionManager struct {
    mu              sync.RWMutex
    clients         map[int]*elasticsearch.Client  // connectionID -> client
    defaultClient   *elasticsearch.Client
    defaultConnID   int
}

// 核心方法
- Initialize()                                    // 從資料庫載入所有連線
- GetClient(connectionID int)                     // 根據 ID 取得客戶端
- GetClientForIndex(indexID int)                  // 為 Index 取得對應客戶端
- GetDefaultClient()                              // 取得預設客戶端
- ReloadConnection(connectionID int)              // 重新載入連線
```

### 資料關聯範例

```
es_connections:
  #1: {name: "生產-主集群", host: "10.99.1.213", port: 9200, is_default: true}
  #2: {name: "測試-集群", host: "10.99.1.64", port: 9200, is_default: false}

indices:
  #101: {pattern: "nginx-*", device_group: "web", es_connection_id: 1}
  #102: {pattern: "app-*", device_group: "app", es_connection_id: 2}
  #103: {pattern: "system-*", device_group: "linux", es_connection_id: 1}
  #104: {pattern: "old-*", device_group: "legacy", es_connection_id: NULL}  // 使用預設

elasticsearch_monitors:
  #201: {name: "主集群健康", es_connection_id: 1}           // 監控 #101, #103 用的 ES
  #202: {name: "測試集群健康", es_connection_id: 2}         // 監控 #102 用的 ES
  #203: {name: "外部ES", host: "10.99.2.100", es_connection_id: NULL}  // 獨立目標
```

---

## 🎯 方案優勢

| 需求場景 | 解決方案 |
|---------|---------|
| **不同 DeviceGroup 連不同 ES** | Index.es_connection_id 指定連線 |
| **同一 ES 被多個 Index 使用** | 多個 Index 指向同一 ESConnection |
| **健康監控與裝置監控共享連線** | ElasticsearchMonitor 可關聯 ESConnection |
| **獨立的健康監控目標** | ElasticsearchMonitor 使用自己的 host/port |
| **向後兼容 setting.yml** | es_connection_id 為 NULL 時使用預設連線 |
| **動態新增/修改連線** | Web UI 管理 es_connections 表 |
| **連線池管理** | ESConnectionManager 統一管理所有客戶端 |

---

## 📋 實作計畫

### Phase 1：基礎建設（2天）✅ 已完成
**目標**: 建立 ES 連線管理核心架構
**完成日期**: 2025-11-18
**Commit**: `633ad98` - feat: Phase 1 - ES 連線管理架構基礎建設

#### 資料庫層
- [x] 建立 `es_connections` 表（含索引、約束）
- [x] 修改 `indices` 表，新增 `es_connection_id` 欄位與外鍵
- [x] 修改 `elasticsearch_monitors` 表，新增 `es_connection_id` 欄位（可選）
- [x] 撰寫資料庫遷移 SQL 腳本

#### 實體層
- [x] 建立 `entities/es_connection.go`
  - [x] ESConnection 結構體
  - [x] TableName() 方法
  - [x] GetURL() 方法
- [x] 修改 `entities/targets.go` 的 Index 結構體
  - [x] 新增 ESConnectionID 欄位
  - [x] 新增 ESConnection 關聯
- [x] 修改 `entities/elasticsearch.go` 的 ElasticsearchMonitor（可選）
  - [x] 新增 ESConnectionID 欄位

#### 服務層
- [x] 建立 `services/es_connection_manager.go`
  - [x] ESConnectionManager 結構體（單例模式）
  - [x] Initialize() - 從資料庫載入所有連線
  - [x] GetClient(connectionID) - 根據 ID 取得客戶端
  - [x] GetClientForIndex(indexID) - 為 Index 取得客戶端
  - [x] GetDefaultClient() - 取得預設客戶端
  - [x] ReloadConnection(connectionID) - 重新載入連線
  - [x] createClient() - 建立 ES 客戶端（私有方法）
  - [x] loadFromConfig() - 從 setting.yml 載入（向後兼容）
- [x] 修改 `clients/es.go`
  - [x] 使用 ESConnectionManager 初始化
  - [x] Fallback 到 setting.yml（向後兼容）

#### 工具層
- [x] 建立 `utils/migrate_es_config.go`
  - [x] MigrateESConfigToDB() - 將 setting.yml 的 ES 配置遷移到資料庫
  - [x] 自動檢測並提示遷移

#### 成果
- 新增 10 個檔案，修改 3 個檔案
- 新增 ~850 行程式碼
- 完整的連線管理器實作（單例模式、執行緒安全）
- 向後兼容策略完整實作

### Phase 2：整合裝置監控（1.5天）✅ 已完成
**目標**: 修改裝置監控使用新的連線管理架構
**完成日期**: 2025-11-18
**Commit**: `d9fce8b` - feat: Phase 2 - 整合裝置監控使用多連線架構

#### 服務層修改
- [x] 修改 `services/detect.go`
  - [x] 修改 Detect() 函數簽名（傳入 indexID）
  - [x] 使用 GetClientForIndex() 取得對應客戶端
  - [x] 錯誤處理（連線不存在時的處理）
- [x] 修改 `services/es_query.go`
  - [x] 新增 SearchRequestWithClient() - 支援自訂客戶端
  - [x] 保留原 SearchRequest() - 向後兼容
- [x] 修改 `services/es_insert.go`（如果有寫入操作）
  - [x] 新增支援自訂客戶端的寫入函數

#### 排程層修改
- [x] 修改 `services/center.go`（或相關排程邏輯）
  - [x] 傳遞 indexID 給 Detect() 函數
  - [x] 確保排程任務使用正確的 ES 連線

#### 測試
- [x] 檢查所有調用 Detect 函數的地方
- [x] 驗證向後兼容性（保留原有函數）
- [ ] 單元測試：ESConnectionManager（待整合測試時驗證）
- [ ] 整合測試：不同 Index 使用不同 ES 連線（待整合測試時驗證）
- [ ] 向後兼容測試：es_connection_id 為 NULL 時的行為（待整合測試時驗證）
- [ ] 錯誤處理測試：連線失效時的降級策略（待整合測試時驗證）

#### 成果
- 修改 4 個檔案
- 新增智能連線路由機制
- 完整的 Fallback 與錯誤處理
- 100% 向後兼容（保留所有原有函數）

### Phase 3：API 與前端（1.5天）✅ 已完成
**目標**: 提供 Web UI 管理 ES 連線
**完成日期**: 2025-11-18
**Commit**: 待提交 - feat: Phase 3 - ES 連線管理 API 層實作

#### API 層
- [x] 建立 `services/es_connection_service.go`
  - [x] GetAllESConnections - 取得所有連線
  - [x] GetESConnection - 取得單一連線
  - [x] CreateESConnection - 建立連線（含名稱唯一性驗證、預設連線管理）
  - [x] UpdateESConnection - 更新連線（含名稱衝突檢查、預設連線管理）
  - [x] DeleteESConnection - 刪除連線（含依賴檢查：Index、Monitor）
  - [x] TestESConnection - 測試連線（不儲存到資料庫）
  - [x] SetDefaultESConnection - 設定預設連線
  - [x] ReloadESConnection - 重新載入指定連線
  - [x] ReloadAllESConnections - 重新載入所有連線
- [x] 建立 `controller/es_connection.go`
  - [x] 9 個端點完整實作（GetAll, Get, Create, Update, Delete, Test, SetDefault, Reload, ReloadAll）
  - [x] Swagger 文件註解完整
  - [x] 統一的錯誤處理與返回格式
- [x] 註冊路由到 `router/router.go`
  - [x] 添加 `/api/v1/ESConnection` 路由組
  - [x] 整合 AuthMiddleware（JWT 認證）
  - [x] 整合 PermissionMiddleware（使用 indices 權限）
  - [x] RESTful 風格路由設計

#### 修改現有 API
- [x] 修改 Index Service 層
  - [x] GetAllIndices - Preload("ESConnection") 返回關聯資訊
  - [x] GetIndicesByTargetID - Preload("Indices.ESConnection")
  - [x] CreateIndex/UpdateIndex - 自動支援 es_connection_id（透過 Gin Bind）

#### 文件更新
- [x] 更新 `docs/openapi.yml`
  - [x] 新增 ES Connection Management 端點（9 個完整端點）
  - [x] 新增 ESConnection schema 定義
  - [x] 新增 ESConnectionSummary schema 定義
  - [x] 更新 Index schema（添加 es_connection_id 和 es_connection 欄位）

#### 前端開發（待前端團隊配合）
- [ ] ES 連線管理頁面
  - [ ] 連線列表（表格顯示）
  - [ ] 新增連線表單
  - [ ] 編輯連線表單
  - [ ] 刪除連線確認對話框
  - [ ] 測試連線按鈕
  - [ ] 設為預設連線開關
- [ ] 修改 Index 管理頁面
  - [ ] 新增「ES 連線」下拉選單
  - [ ] 顯示當前使用的 ES 連線資訊

#### 成果
- 新增 2 個檔案（services/es_connection_service.go, controller/es_connection.go）
- 修改 3 個檔案（router/router.go, services/indices.go, docs/openapi.yml）
- 新增 ~700 行程式碼
- 9 個完整的 RESTful API 端點
- 完整的錯誤處理與驗證邏輯
- 安全設計：返回 ESConnectionSummary（不含密碼）

### Phase 4：整合健康監控（可選，1天）
**目標**: 讓健康監控支援複用連線配置

#### 資料庫層（已完成於 Phase 1）
- [x] 已在 Phase 1 新增 `elasticsearch_monitors.es_connection_id`

#### 服務層
- [ ] 修改 `services/es_monitor.go`
  - [ ] CheckESHealth() - 支援從 ESConnection 讀取配置
  - [ ] 優先使用 es_connection_id，回退到 host/port
- [ ] 修改監控排程邏輯
  - [ ] 確保監控任務使用正確的連線配置

#### API 層
- [ ] 修改 ElasticsearchMonitor CRUD API
  - [ ] CreateMonitor - 支援選擇「複用連線」或「獨立配置」
  - [ ] UpdateMonitor - 支援切換連線方式
  - [ ] GetMonitor - 返回關聯的 ESConnection 資訊

#### 前端調整
- [ ] 健康監控表單
  - [ ] 新增「連線方式」選擇器（複用/獨立）
  - [ ] 複用模式：顯示 ES 連線下拉選單
  - [ ] 獨立模式：顯示 host/port/auth 欄位

### Phase 5：文件與部署（0.5天）
**目標**: 更新文件並準備上線

#### 文件更新
- [ ] 更新 `docs/specs_cn/02-架構設計.md`
  - [ ] 新增 ES 連線管理架構說明
  - [ ] 更新架構圖
- [ ] 更新 `docs/specs_cn/04-資料庫設計.md`
  - [ ] 新增 es_connections 表結構
  - [ ] 更新 ER 圖（indices, elasticsearch_monitors 的新關聯）
- [ ] 更新 `docs/specs_cn/03-API規格.md`
  - [ ] 新增 ES 連線管理 API 規格
- [ ] 更新 `docs/openapi.yml`
  - [ ] 新增 ES 連線相關端點定義
- [ ] 撰寫遷移指南
  - [ ] `docs/guides/migration/es-connection-migration.md`
  - [ ] setting.yml → 資料庫的遷移步驟
  - [ ] 降級回滾步驟

#### 部署準備
- [ ] 撰寫資料庫遷移腳本
  - [ ] `migrations/xxx_create_es_connections.sql`
  - [ ] `migrations/xxx_alter_indices_add_es_connection.sql`
  - [ ] `migrations/xxx_alter_es_monitors_add_es_connection.sql`
- [ ] 撰寫自動遷移工具使用說明
- [ ] 準備回滾腳本（萬一需要降級）

---

## 🔄 向後兼容策略

### 1. 配置讀取優先順序
```
1. Index.es_connection_id 指定的連線（優先）
2. 資料庫中 is_default=true 的連線
3. setting.yml 的 ES 配置（Fallback）
```

### 2. 遷移策略
- **自動遷移**: 系統啟動時檢測 setting.yml，自動建立預設 ESConnection 記錄
- **手動遷移**: 提供 CLI 工具或 API 端點手動觸發遷移
- **混合模式**: Phase 1-2 期間同時支援資料庫和 setting.yml

### 3. 降級方案
- 保留 `clients.ES` 全域變數，指向預設客戶端
- 保留原有的 SearchRequest() 函數（向後兼容）
- 出錯時自動 Fallback 到預設連線

---

## ⚠️ 風險與挑戰

### 技術風險
1. **連線池管理複雜度**
   - 風險：多連線併發管理可能導致資源耗盡
   - 緩解：設定連線數上限、健康檢查機制

2. **資料庫遷移失敗**
   - 風險：外鍵約束可能導致遷移失敗
   - 緩解：先在測試環境驗證、準備回滾腳本

3. **效能影響**
   - 風險：每次查詢都需要查資料庫取得 ESConnection
   - 緩解：連線管理器使用記憶體快取、定期更新

### 業務風險
1. **向後兼容問題**
   - 風險：現有排程任務可能中斷
   - 緩解：保留 Fallback 機制、充分測試

2. **前後端協作**
   - 風險：前端開發進度影響上線時程
   - 緩解：Phase 1-2 可獨立完成並上線（後端先行）

### 緩解措施
- ✅ 分階段實作，每個 Phase 可獨立測試
- ✅ 充分的單元測試與整合測試
- ✅ 在測試環境完整驗證後再上生產
- ✅ 準備快速回滾方案

---

## 📊 成功指標

### 功能指標
- [ ] 支援至少 5 個不同的 ES 連線同時運作
- [ ] 不同 Index 可成功連接到各自指定的 ES Cluster
- [ ] 健康監控可複用裝置監控的 ES 連線配置
- [ ] 透過 Web UI 可完成所有連線管理操作
- [ ] 向後兼容：未指定連線的 Index 仍可正常運作

### 效能指標
- [ ] ES 連線初始化時間 < 5 秒
- [ ] GetClientForIndex() 查詢時間 < 10ms（含快取）
- [ ] 連線切換不影響現有查詢（平滑切換）

### 品質指標
- [ ] 單元測試覆蓋率 > 80%
- [ ] 整合測試通過率 100%
- [ ] 無資料遷移錯誤
- [ ] 無現有功能退化

---

## 🔗 相關文件

- **架構設計**: `docs/specs_cn/02-架構設計.md`
- **資料庫設計**: `docs/specs_cn/04-資料庫設計.md`
- **API 規格**: `docs/specs_cn/03-API規格.md`
- **程式碼位置**:
  - ES 客戶端初始化: `clients/es.go`
  - 裝置監控邏輯: `services/detect.go`, `services/es_query.go`
  - 健康監控邏輯: `services/es_monitor.go`
  - 資料實體: `entities/targets.go`, `entities/elasticsearch.go`

---

## 📝 討論記錄

### 2025-11-18 - 初始討論
**參與者**: 系統架構師、開發團隊

**討論要點**:
1. ✅ 確認需求：不同 DeviceGroup 需要連接不同的 ES Cluster
2. ✅ 確認關聯性：即時監控與健康監控應共享 ES 連線配置
3. ✅ 方案選擇：採用「ES 連線池架構」方案
4. ✅ 實作策略：分 5 個 Phase 段階式實作
5. ⏳ 待確認：前端開發時程安排

**決議事項**:
- 採用 `es_connections` 表作為基礎實體
- Phase 1-2 優先完成（後端先行）
- Phase 3 需前端配合
- Phase 4 可視需求選擇性實作

**下一步行動**:
- [ ] 評估並確認實作方案
- [ ] 確認前端配合時程
- [ ] 開始 Phase 1 開發

---

## 📌 備註

- 本議題涉及核心架構變更，建議優先處理
- 實作過程中需注意資料庫遷移的安全性
- 建議在開發環境充分測試後再部署到生產環境
- 可考慮使用 Feature Flag 控制新功能的啟用

---

**最後更新**: 2025-11-18
**更新者**: Claude (AI Assistant)
