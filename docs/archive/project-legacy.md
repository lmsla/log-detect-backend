⏺ Log Detect Backend 專案架構與功能分析

  🏗️ 主要架構

  技術棧：
  - 語言: Go 1.24
  - Web框架: Gin
  - ORM: GORM
  - 資料庫: MySQL
  - 搜尋引擎: Elasticsearch
  - 任務調度: Cron (robfig/cron/v3)
  - 身份認證: JWT
  - API文檔: Swagger/OpenAPI

  目錄結構：
  log-detect-backend/
  ├── clients/          # 資料庫與Elasticsearch客戶端
  ├── controller/       # API控制器（8個控制器）
  ├── entities/         # 資料實體定義
  ├── middleware/       # 中介軟體（身份驗證、權限控制）
  ├── models/          # 資料模型
  ├── services/        # 業務邏輯服務層
  ├── router/          # 路由配置
  ├── utils/           # 工具函數
  ├── global/          # 全域配置
  └── docs/           # API文檔

  🎯 核心功能

  1. 日誌設備監控系統

  - 目標 (Target): 監控任務配置，包含接收者、主題、啟用狀態
  - 索引 (Index): Elasticsearch索引配置，定義監控模式、時間週期、欄位
  - 設備 (Device): 被監控的設備清單，按群組管理
  - 歷史記錄 (History): 設備狀態記錄（線上/離線、響應時間、錯誤信息）

  2. 自動化檢測邏輯

  // services/detect.go:13
  func Detect(execute_time, index, field, period, unit, receiver, subject, logname, device_group)
  - 定期從 Elasticsearch 查詢設備日誌
  - 比對資料庫中的設備清單與實際日誌中的設備
  - 自動發現新設備並加入管理
  - 偵測離線設備並發送告警郵件
  - 記錄所有檢查結果到歷史表

  3. 任務調度系統

  // services/center.go:56
  func Control_center()
  - 使用 Cron 調度器執行定期檢測
  - 支援分鐘級和小時級監控週期
  - 動態管理監控任務的啟用/停用
  - 系統重啟時自動重新載入所有監控任務

  4. 完整的身份認證系統

  - JWT Token 認證: 24小時有效期
  - RBAC 權限控制: 基於角色的存取控制
  - 預設角色: admin（全權限）、user（唯讀權限）
  - 權限粒度: 針對 device、target、indices、user 的 CRUD 操作
  - 中介軟體保護: 所有 API 端點的權限檢查

  5. RESTful API 設計

  - 公開端點: 登入、健康檢查、Swagger文檔
  - 認證端點: 用戶管理、個人資料、Token刷新
  - 業務端點: 設備管理、目標管理、索引管理、歷史查詢
  - 儀表板端點: 統計數據、趨勢分析、告警管理

  🔄 工作流程

  1. 初始化階段 (main.go:28)
    - 載入環境配置
    - 連接 MySQL 和 Elasticsearch
    - 建立資料表結構
    - 初始化預設管理員帳戶
    - 載入 Cron 調度器
    - 啟動所有監控任務
  2. 監控執行階段
    - Cron 任務按設定週期執行
    - 查詢 Elasticsearch 獲取最新設備日誌
    - 與資料庫設備清單比對
    - 記錄線上設備狀態，發現離線設備
    - 自動新增未知設備到管理清單
    - 發送告警郵件給指定接收者
  3. 資料管理階段
    - 提供完整的 CRUD API
    - 支援歷史記錄查詢和統計
    - 資料歸檔和清理功能
    - 即時儀表板數據展示

  📊 資料模型關係

  - Target (1) ↔ (N) Index: 一個監控目標可包含多個索引配置
  - Index (1) ↔ (N) Device: 一個索引對應一個設備群組
  - Target + Index → History: 每次檢測結果記錄到歷史表
  - User ↔ Role ↔ Permission: 完整的 RBAC 權限體系

  這是一個企業級的日誌監控與告警系統，具備自動化設備發現、即時狀態監控、權限管理和可視化儀表板等完整功能
  。