# Log Detect Backend - 完整規格說明文件

歡迎來到 Log Detect Backend 程式碼庫的全面分析文件。本目錄包含透過深入分析整個程式碼庫所產生的完整文件。

## 文件清單

### 1. CODEBASE_ANALYSIS.md (40 KB, 1334 行)
**整個程式碼庫的全面技術分析**

這是最詳盡的文件,涵蓋:
- 執行摘要與整體架構
- 完整專案結構與技術堆疊
- 所有 22 個服務與詳細函數說明
- 9 個控制器與所有端點
- 資料庫架構 (MySQL 與 TimescaleDB 資料表)
- 完整 API 端點參考 (60+ 路由)
- 身份驗證與授權機制
- 外部整合 (Elasticsearch、電子郵件、SSO)
- 業務邏輯工作流程
- 使用的設計模式
- 安全性考量
- 可擴展性與效能注意事項
- 設定檔格式

**適用於**: 深入了解系統運作方式、架構決策、資料模型和整合點。

---

### 2. ARCHITECTURE_DIAGRAMS.md (38 KB, 703 行)
**視覺化架構與資料流程圖**

本文件包含:
- 系統架構總覽圖
- 分層架構視覺化
- 裝置監控偵測流程
- Elasticsearch 監控流程
- 身份驗證與授權流程
- 批次處理與資料一致性流程
- 元件相依性圖
- 設定結構圖
- 請求/回應流程範例
- 部署架構
- 資料庫連線拓樸

**適用於**: 了解元件如何互動、視覺化資料流程,以及觀察系統設計的全貌。

---

### 3. QUICK_REFERENCE.md (13 KB, 517 行)
**開發人員與維運人員的快速查詢指南**

此參考指南包含:
- 專案概覽
- 關鍵功能一覽
- 核心資料模型
- 重要檔案位置
- API 端點快速清單
- 工作流程範例 (裝置監控、ES 監控、批次寫入)
- 關鍵服務與職責
- 常見操作
- 資料庫連線詳情
- JWT 權杖詳情
- 權限模型
- 啟動順序
- 除錯技巧
- 檔案統計
- 技術版本

**適用於**: 快速查詢、API 整合、除錯,以及新團隊成員上手。

---

## 快速導覽

### 針對不同角色

**系統架構師**
- 從這裡開始: ARCHITECTURE_DIAGRAMS.md
- 接著閱讀: CODEBASE_ANALYSIS.md 第 1-3、12 章節

**後端開發人員**
- 從這裡開始: QUICK_REFERENCE.md
- 接著閱讀: CODEBASE_ANALYSIS.md 第 2-6、8 章節
- 參考: ARCHITECTURE_DIAGRAMS.md 流程圖

**DevOps/維運人員**
- 從這裡開始: QUICK_REFERENCE.md (啟動順序、資料庫連線詳情)
- 接著閱讀: CODEBASE_ANALYSIS.md 第 1、7 章節
- 參考: ARCHITECTURE_DIAGRAMS.md (部署架構)

**API 使用者**
- 從這裡開始: QUICK_REFERENCE.md (API 端點快速清單)
- 接著閱讀: CODEBASE_ANALYSIS.md 第 5 章節 (API 端點與路由)
- 參考: ARCHITECTURE_DIAGRAMS.md (請求/回應流程範例)

**QA/測試人員**
- 從這裡開始: QUICK_REFERENCE.md (工作流程範例、常見操作)
- 接著閱讀: CODEBASE_ANALYSIS.md 第 3、6、8 章節
- 參考: ARCHITECTURE_DIAGRAMS.md (資料流程圖)

---

## 關鍵統計

| 指標 | 數值 |
|--------|-------|
| 文件總行數 | 2,554 |
| 文件總大小 | 91 KB |
| 已記錄服務數 | 22 |
| 已記錄控制器數 | 9 |
| 資料模型 | 20+ |
| API 端點 | 60+ |
| 資料庫資料表 | 16+ |
| 外部整合 | 3 (ES, MySQL, TimescaleDB) |

---

## 系統概覽

**Log Detect Backend** 是一個使用 Go 語言建構的全面性監控系統,功能包括:

1. **監控裝置健康狀態**: 查詢 Elasticsearch 取得裝置狀態,與資料庫比對,並在發生變更時發出警報
2. **監控 Elasticsearch 健康狀態**: 追蹤 ES 叢集健康狀態 (CPU、記憶體、磁碟、分片)
3. **管理使用者與存取權限**: 採用 JWT 身份驗證的 RBAC (角色型存取控制)
4. **儲存時間序列資料**: 高效批次寫入至 TimescaleDB
5. **傳送警報**: 透過 SMTP 發送 HTML 格式的電子郵件通知
6. **提供分析功能**: 儀表板包含統計資料、趨勢和時間軸視覺化
7. **提供 REST API**: 60+ 個端點用於管理與監控

**技術堆疊**:
- 程式語言: Go 1.24
- Web 框架: Gin
- 資料庫: MySQL (設定) + TimescaleDB (指標) + Elasticsearch (日誌)
- 排程器: robfig/cron
- 身份驗證: JWT + bcrypt + RBAC
- 電子郵件: SMTP
- 文件: Swagger

---

## 核心工作流程

### 裝置監控工作流程
1. 管理員建立 Target (電子郵件收件人) 與 Index (ES 模式)
2. 系統自動註冊 cron 作業
3. 依排程執行: Detect() 查詢 ES,與資料庫比對
4. 建立 History 記錄並發送電子郵件警報
5. 透過 BatchWriter 將指標儲存至 TimescaleDB

### Elasticsearch 健康狀態監控工作流程
1. 管理員建立 ElasticsearchMonitor 並設定檢查間隔
2. 系統開始週期性健康檢查
3. 收集 CPU、記憶體、磁碟、分片資料
4. 與可設定的閾值比對
5. 建立警報並儲存指標

### 批次寫入工作流程
1. 服務產生 History 或 ESMetric 記錄
2. BatchWriter 將它們排入記憶體佇列
3. 在以下情況執行清空: 大小 >= 50 筆記錄 或 經過 5 秒
4. 透過預備語句以原子方式插入至 TimescaleDB
5. 記錄成功數量並繼續

---

## 資料庫架構

### MySQL (logdetect 資料庫)
- **設定**: Users、Roles、Permissions、Devices、Targets、Indices
- **歷史記錄**: 監控結果、警報、郵件日誌、cron 作業追蹤
- **用途**: 系統設定與運作狀態

### TimescaleDB (monitoring 資料庫)
- **device_metrics**: 時間序列裝置監控資料 (超表)
- **es_metrics**: 時間序列 Elasticsearch 指標 (超表)
- **alert_history**: 帶有時間戳記的警報記錄
- **用途**: 高效儲存與查詢時間序列資料

### Elasticsearch
- **索引**: logstash-*、自訂模式
- **用途**: 日誌儲存與裝置探索
- **整合**: 裝置監控查詢日誌以取得主機資訊

---

## API 結構

- **公開**: POST /auth/login
- **受保護**: 所有 /api/v1/* 路由需要 JWT + 選擇性權限檢查
- **中介軟體鏈**: AuthMiddleware → PermissionMiddleware
- **文件**: Swagger 可在 /swagger/ 取得

**路由群組**:
- /auth - 使用者身份驗證 (8 個端點)
- /Target - 監控目標 (4 個端點)
- /Device - 裝置管理 (6 個端點)
- /Indices - 索引模式 (7 個端點)
- /Receiver - 電子郵件收件人 (4 個端點)
- /History - 歷史資料 (2 個端點)
- /dashboard - 分析 (10+ 個端點)
- /elasticsearch - ES 監控 (13 個端點)

---

## 身份驗證與授權

**JWT 權杖流程**:
1. 使用者使用憑證登入
2. 系統針對 MySQL 中的 bcrypt 雜湊值進行驗證
3. 產生 HS256 簽署的 JWT (24 小時到期)
4. 用戶端在 Authorization 標頭中包含 Bearer 權杖
5. 中介軟體驗證簽章與到期時間
6. PermissionMiddleware 檢查 resource.action 權限

**RBAC 模型**:
- User → Role (一對一)
- Role → Permissions (多對多)
- Permission = Resource + Action (例如: "device.create")

**預設角色**:
- admin: 所有權限
- user: 唯讀
- operator: 指派資源的 CRUD 操作

---

## 設定檔

### config.yml
主要設定檔,包含以下章節:
- database (MySQL 連線)
- timescale (TimescaleDB 連線)
- batch_writer (大小: 50,間隔: 5秒)
- es (Elasticsearch 連線)
- email (SMTP 設定)
- server (連接埠: 8006,模式: debug/release)
- cors (允許的來源)
- sso (Keycloak 整合)

### setting.yml
監控設定,包含:
- targets: 監控設定清單
  - receiver: 電子郵件收件人
  - subject: 警報主旨
  - indices: ES 索引模式與頻率

---

## 入門指南

### 了解系統
1. 閱讀: QUICK_REFERENCE.md (專案概覽、關鍵功能)
2. 研究: ARCHITECTURE_DIAGRAMS.md (系統架構、資料流程)
3. 深入探討: CODEBASE_ANALYSIS.md (完整詳情)

### 設定監控
1. 在 config.yml 中設定資料庫連線
2. 透過 API 或 UI 建立 Indices (ES 模式)
3. 透過 API 或 UI 建立 Targets (電子郵件收件人)
4. 將 Indices 連結至 Targets
5. 啟用 targets 並等待 cron 排程

### 疑難排解
1. 檢查: QUICK_REFERENCE.md (除錯技巧、常見問題)
2. 監控: MySQL (history、cron_lists) 與日誌
3. 驗證: Elasticsearch 連線、SMTP 設定

---

## 重要概念

**裝置監控**: 系統查詢 Elasticsearch 以取得符合模式的裝置,與資料庫 Device 清單比對,偵測變更 (新增/離線/上線),並傳送警報

**批次寫入**: 在記憶體中累積記錄,批次清空至資料庫以減少資料庫負擔並提升效能

**Cron 作業**: 動態註冊週期性監控任務,儲存於 cron_lists 資料表,由 robfig/cron 排程器管理

**時間序列資料**: 使用 TimescaleDB 的超表高效儲存裝置指標與 ES 指標,並自動依時間分區

**RBAC**: 多層級存取控制,採用 User → Role → Permissions 模型,在 API 端點層級強制執行

---

## 文件維護

這些文件產生於全面性靜態程式碼分析: **2024 年 10 月 29 日**

分析涵蓋:
- 所有 22 個服務檔案
- 所有 9 個控制器檔案
- 所有實體與模型定義
- 路由器與中介軟體設定
- 用戶端/資料庫連線
- 設定檔格式
- 資料庫架構定義
- API 端點路由

---

## 文件位置

所有文件儲存於: `/specs/`

- CODEBASE_ANALYSIS.md
- ARCHITECTURE_DIAGRAMS.md
- QUICK_REFERENCE.md
- README.md (本檔案)

---

## 相關文件

另請參閱:
- `/docs/` - Swagger API 文件 (自動產生)
- README_AUTH.md - 身份驗證詳情
- DATA_MANAGEMENT_GUIDE.md - 資料封存與管理
- TROUBLESHOOTING.md - 常見問題與解決方案
- `go.mod` - 相依套件版本

---

本文件提供了解 Log Detect Backend 系統架構、元件與功能的完整參考。請使用上方的目錄來尋找與您需求最相關的資訊。
