# 📚 Log Detect Backend 文件中心

> **規格驅動開發 (Specification-Driven Development)** - 以清晰的規格為核心，確保實作與需求一致

## 📖 文件導航

### 🎯 核心規格 (Specifications)

所有系統設計與 API 規格的權威來源

#### API 規格
- **[OpenAPI 3.0 規格](spec/api/openapi.yml)** - 完整的 RESTful API 定義
- **[Elasticsearch API 規格](spec/api/elasticsearch-api-spec.md)** - ES 監控 API 詳細說明
- **[API 使用說明](spec/api/README.md)** - OpenAPI 文件使用指南

#### 資料庫規格
- **[TimescaleDB 架構設計](spec/database/timescaledb-architecture.md)** - 時序資料庫架構
- **[Schema 驗證規範](spec/database/schema-validation.md)** - 資料表結構驗證

#### 權限規格
- **[RBAC 權限指南](spec/permissions/rbac-guide.md)** - 角色權限系統設計

---

### 📘 實作指南 (Implementation Guides)

開發與部署的實用指南

#### 功能實作
- **[Elasticsearch 監控設置](guides/implementation/elasticsearch-setup.md)** - ES 監控功能實作
- **[Elasticsearch 總覽](guides/implementation/elasticsearch-overview.md)** - ES 監控系統概述
- **[TimescaleDB 遷移指南](guides/implementation/timescaledb-migration-guide.md)** - 從舊系統遷移到 TimescaleDB

#### 前端整合
- **[前端 API 對接指南](guides/frontend/api-integration.md)** - 前端開發者必讀

---

### 🔧 故障排除 (Troubleshooting)

常見問題的診斷與解決方案

#### 資料庫問題
- **[PostgreSQL 權限錯誤修復](troubleshooting/database/permission-errors.md)** - 解決 "must be owner of table" 錯誤
- **[ES Metrics 表結構修復](troubleshooting/database/es-metrics-table-fix.md)** - 修復缺少欄位的問題

#### 監控問題
- **[ES 監控無資料診斷](troubleshooting/monitoring/no-data-diagnosis.md)** - 診斷 es_metrics 表無資料問題
- **[ES 權限問題修復](troubleshooting/monitoring/es-permissions-fix.md)** - 修復監控權限錯誤

---

### 📦 歷史歸檔 (Archive)

已完成或過時的文件記錄

- **[前端調整記錄](archive/adjust-records/)** - adjust.md 系列調整文件
- **[實作狀態快照](archive/status-snapshots/)** - 各階段實作狀態記錄
- **[Code Review 清單](archive/code-review-todo.md)** - 歷史 code review 項目
- **[專案舊版說明](archive/project-legacy.md)** - 早期專案文件

---

## 🚀 快速開始

### 新進開發者
1. 閱讀 [OpenAPI 規格](spec/api/openapi.yml) 了解 API 設計
2. 參考 [TimescaleDB 架構設計](spec/database/timescaledb-architecture.md) 了解資料結構
3. 查看 [RBAC 權限指南](spec/permissions/rbac-guide.md) 了解權限系統

### 前端開發者
1. 查看 [前端 API 對接指南](guides/frontend/api-integration.md)
2. 參考 [OpenAPI 規格](spec/api/openapi.yml) 了解端點定義
3. 使用 [Elasticsearch API 規格](spec/api/elasticsearch-api-spec.md) 實作監控頁面

### 後端開發者
1. 遵循 [OpenAPI 規格](spec/api/openapi.yml) 實作 API
2. 參考 [實作指南](guides/implementation/) 進行功能開發
3. 遇到問題查閱 [故障排除](troubleshooting/) 文件

---

## 📁 目錄結構

```
docs/
├── README.md (本文件)
├── spec/                          # 🎯 核心規格
│   ├── api/                       # API 規格定義
│   ├── database/                  # 資料庫架構設計
│   └── permissions/               # 權限系統規格
├── guides/                        # 📘 實作指南
│   ├── implementation/            # 功能實作指南
│   └── frontend/                  # 前端整合指南
├── troubleshooting/               # 🔧 故障排除
│   ├── database/                  # 資料庫相關問題
│   └── monitoring/                # 監控系統問題
└── archive/                       # 📦 歷史歸檔
    ├── adjust-records/            # 前端調整記錄
    └── status-snapshots/          # 實作狀態快照
```

---

## 📝 文件更新原則

### 規格驅動開發流程
1. **規格優先** - 任何新功能必須先在 `spec/` 目錄定義規格
2. **文件同步** - 實作完成後立即更新相關文件
3. **問題記錄** - 故障排除方案必須記錄在 `troubleshooting/`
4. **定期歸檔** - 過時文件移至 `archive/` 保留歷史記錄

### 文件維護者
- **API 規格**: 需經過技術負責人審核
- **實作指南**: 由實作開發者編寫並維護
- **故障排除**: 遇到問題的開發者負責記錄解決方案

---

## 🔗 相關資源

- **Swagger UI**: http://localhost:8006/swagger/index.html
- **API Base URL**: http://localhost:8006/api/v1
- **TimescaleDB**: PostgreSQL extension for time-series data
- **Go Gin Framework**: https://gin-gonic.com/

---

**最後更新**: 2025-10-08
**維護團隊**: Log Detect Development Team
