# 📚 Log Detect Backend 文件中心

> **規格驅動開發 (Specification-Driven Development)** - 以清晰的規格為核心，確保實作與需求一致

## 📖 文件導航

### 🎯 核心規格 (specs_cn/)

基於 **IEEE SDD 標準**編寫的完整系統設計文件（繁體中文）

- **[📋 規格文件導覽](specs_cn/README.md)** - 完整的文件索引與角色導讀
- **[01-系統概述](specs_cn/01-系統概述.md)** - 專案簡介、功能特色、架構總覽
- **[02-架構設計](specs_cn/02-架構設計.md)** - 系統架構、技術棧、模組設計
- **[03-API規格](specs_cn/03-API規格.md)** - 完整的 REST API 規格說明
- **[04-資料庫設計](specs_cn/04-資料庫設計.md)** - MySQL、TimescaleDB、Elasticsearch 設計
- **[05-安全與部署](specs_cn/05-安全與部署.md)** - 安全設計、部署指南、維運建議

**輔助文件**：
- **[快速參考](specs_cn/快速參考.md)** - 常用指令與配置速查
- **[架構圖表](specs_cn/架構圖表.md)** - 系統架構圖說明
- **[程式碼分析](specs_cn/程式碼分析.md)** - 程式碼結構分析

### 📐 API 規格文件

- **[OpenAPI 3.0 規格](openapi.yml)** - 完整的 RESTful API 定義（適用於前後端協作）
- **Swagger UI**: http://localhost:8006/swagger/index.html

### 📝 SDD 模板套件 (templates/sdd-template/)

可重用的 SDD 文件模板，適用於其他專案

- **[模板使用指南](templates/sdd-template/模板使用指南.md)** - 完整使用說明、撰寫技巧、品質檢查表
- **系統文件模板**：01-系統概述、02-架構設計、03-API規格、04-資料庫設計、05-安全與部署
- **README 模板** - 專案首頁模板

---

### 📘 實作指南 (guides/)

開發與整合的實用指南

#### 前端整合
- **[前端 API 對接指南](guides/frontend/api-integration.md)** - 前端開發者必讀

---

### 🔧 故障排除 (troubleshooting/)

常見問題的診斷與解決方案

#### 資料庫問題
- **[PostgreSQL 權限錯誤修復](troubleshooting/database/permission-errors.md)** - 解決 "must be owner of table" 錯誤
- **[ES Metrics 表結構修復](troubleshooting/database/es-metrics-table-fix.md)** - 修復缺少欄位的問題

#### 監控問題
- **[ES 監控無資料診斷](troubleshooting/monitoring/no-data-diagnosis.md)** - 診斷 es_metrics 表無資料問題
- **[ES 權限問題修復](troubleshooting/monitoring/es-permissions-fix.md)** - 修復監控權限錯誤
- **[告警 API 修復記錄](troubleshooting/monitoring/alert-api-fixes-20251022.md)** - 2025-10-22 告警功能修復

#### SQL 腳本
- `add_alert_dedupe_window.sql` - 新增告警去重視窗欄位
- `add_threshold_fields.sql` - 新增閾值欄位
- `fix_es_alert_history_columns.sql` - 修復 ES 告警歷史表結構

---

### 📦 歷史歸檔 (archive/)

已完成或過時的文件記錄

- **[前端調整記錄](archive/adjust-records/)** - adjust.md 系列調整文件
- **[實作狀態快照](archive/status-snapshots/)** - 各階段實作狀態記錄
- **[ES 告警去重變更日誌](archive/CHANGELOG-ES-ALERT-DEDUPE.md)** - ES 告警去重功能開發記錄
- **[Code Review 清單](archive/code-review-todo.md)** - 歷史 code review 項目
- **[專案舊版說明](archive/project-legacy.md)** - 早期專案文件

---

## 🚀 快速開始

### 新進開發者
1. 閱讀 **[系統概述](specs_cn/01-系統概述.md)** 了解專案背景與架構
2. 參考 **[架構設計](specs_cn/02-架構設計.md)** 了解技術棧與模組設計
3. 查看 **[資料庫設計](specs_cn/04-資料庫設計.md)** 了解資料結構
4. 查看 **[安全與部署](specs_cn/05-安全與部署.md)** 了解 JWT、RBAC 與部署流程

### 前端開發者
1. 查看 **[前端 API 對接指南](guides/frontend/api-integration.md)**
2. 參考 **[OpenAPI 規格](openapi.yml)** 或使用 Swagger UI 了解 API 端點
3. 閱讀 **[API規格](specs_cn/03-API規格.md)** 了解完整 API 設計與範例

### 後端開發者
1. 遵循 **[API規格](specs_cn/03-API規格.md)** 實作 API
2. 參考 **[資料庫設計](specs_cn/04-資料庫設計.md)** 進行資料表操作
3. 遇到問題查閱 **[故障排除](troubleshooting/)** 文件

### 架構師 / 技術主管
1. 審閱 **[specs_cn/](specs_cn/)** 完整的 SDD 規格文件
2. 使用 **[templates/sdd-template/](templates/sdd-template/)** 於其他專案建立規格文件

### QA 測試人員
1. 參考 **[API規格](specs_cn/03-API規格.md)** 第6章測試指南
2. 使用 **[OpenAPI 規格](openapi.yml)** 進行 API 測試
3. 查看 **[故障排除](troubleshooting/)** 了解已知問題

---

## 📁 目錄結構

```
docs/
├── README.md                       # 📚 文件導覽（本文件）
├── openapi.yml                     # 🎯 OpenAPI 3.0 規格（前後端協作）
├── docs.go                         # 🤖 Swagger 自動生成
├── swagger.json                    # 🤖 Swagger JSON 定義
├── swagger.yaml                    # 🤖 Swagger YAML 定義
│
├── specs_cn/                       # 📋 SDD 規格文件（繁體中文）
│   ├── README.md                   #     文件索引與角色導讀
│   ├── 01-系統概述.md              #     專案簡介、功能、架構
│   ├── 02-架構設計.md              #     系統架構、技術棧
│   ├── 03-API規格.md               #     REST API 完整規格
│   ├── 04-資料庫設計.md            #     資料庫架構設計
│   ├── 05-安全與部署.md            #     安全機制與部署指南
│   ├── 快速參考.md                 #     常用指令速查
│   ├── 架構圖表.md                 #     架構圖說明
│   └── 程式碼分析.md               #     程式碼結構分析
│
├── templates/                      # 📝 SDD 模板套件
│   └── sdd-template/               #     可重用的 SDD 文件模板
│       ├── 模板使用指南.md         #     使用說明、撰寫技巧
│       ├── 01-系統概述-模板.md
│       ├── 02-架構設計-模板.md
│       ├── 03-API規格-模板.md
│       ├── 04-資料庫設計-模板.md
│       ├── 05-安全與部署-模板.md
│       └── README-模板.md
│
├── guides/                         # 📘 實作指南
│   └── frontend/                   #     前端整合
│       └── api-integration.md      #     API 對接指南
│
├── troubleshooting/                # 🔧 故障排除
│   ├── database/                   #     資料庫問題
│   │   ├── es-metrics-table-fix.md
│   │   └── permission-errors.md
│   ├── monitoring/                 #     監控系統問題
│   │   ├── alert-api-fixes-20251022.md
│   │   ├── es-permissions-fix.md
│   │   └── no-data-diagnosis.md
│   ├── *.sql                       #     修復用 SQL 腳本
│   └── ...
│
└── archive/                        # 📦 歷史歸檔
    ├── adjust-records/             #     前端調整記錄
    ├── status-snapshots/           #     實作狀態快照
    ├── CHANGELOG-ES-ALERT-DEDUPE.md
    └── ...
```

---

## 📝 文件更新原則

### 規格驅動開發流程
1. **規格優先** - 任何新功能必須先在 `specs_cn/` 定義規格
2. **文件同步** - 實作完成後立即更新相關文件
3. **問題記錄** - 故障排除方案必須記錄在 `troubleshooting/`
4. **定期歸檔** - 過時文件移至 `archive/` 保留歷史記錄

### 文件維護者
- **SDD 規格文件** (`specs_cn/`): 需經過架構師或技術負責人審核
- **API 規格** (`openapi.yml`): 前後端共同維護，需經過技術負責人審核
- **實作指南** (`guides/`): 由實作開發者編寫並維護
- **故障排除** (`troubleshooting/`): 遇到問題的開發者負責記錄解決方案

### 文件品質標準
- 遵循 **IEEE SDD 標準**（參考 `templates/sdd-template/模板使用指南.md`）
- 使用**繁體中文**編寫
- 提供**完整範例**與**實際程式碼片段**
- 包含**ER 圖、架構圖**等視覺化說明

---

## 🔗 相關資源

### 開發環境
- **Swagger UI**: http://localhost:8006/swagger/index.html
- **API Base URL**: http://localhost:8006/api/v1
- **Frontend**: http://localhost:5173

### 技術文件
- **Go Gin Framework**: https://gin-gonic.com/
- **TimescaleDB**: https://docs.timescale.com/
- **Elasticsearch**: https://www.elastic.co/guide/
- **OpenAPI 3.0**: https://swagger.io/specification/

### 標準參考
- **IEEE SDD 標準**: IEEE Std 1016-2009
- **RESTful API 設計**: https://restfulapi.net/

---

**最後更新**: 2025-11-18
**維護團隊**: Log Detect Development Team
**文件版本**: v2.0 (SDD 重構版)
