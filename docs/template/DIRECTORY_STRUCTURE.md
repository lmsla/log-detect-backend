# 📂 文件目錄結構模板

> 基於規格驅動開發 (Specification-Driven Development) 理念設計的標準文件架構

## 🎯 設計原則

1. **規格為核心** - spec/ 目錄存放所有權威規格，是開發的唯一真相來源
2. **用途分類** - 按使用目的分類（規格、指南、故障排除、歷史）
3. **層級清晰** - 最多 3 層目錄，避免過深結構
4. **命名一致** - 使用小寫、連字符命名（kebab-case）
5. **可擴展性** - 模組化設計，易於新增功能文件

## 📁 標準目錄結構

```
docs/
│
├── README.md                      # 總導航索引（必須）
│
├── spec/                          # 🎯 核心規格（必須）
│   │
│   ├── api/                       # API 規格定義
│   │   ├── README.md             # API 文件說明
│   │   ├── openapi.yml           # OpenAPI 3.0 規格（主要）
│   │   ├── [module-a]-api.md     # 模組 A 的 API 詳細規格
│   │   └── [module-b]-api.md     # 模組 B 的 API 詳細規格
│   │
│   ├── database/                  # 資料庫架構設計
│   │   ├── architecture.md       # 資料庫整體架構
│   │   ├── schema.md             # 完整 Schema 定義
│   │   ├── schema-validation.md  # Schema 驗證規範
│   │   └── erd.md                # Entity Relationship Diagram
│   │
│   ├── permissions/               # 權限系統規格
│   │   ├── rbac-guide.md         # RBAC 權限設計
│   │   └── permission-matrix.md  # 權限矩陣表
│   │
│   └── business/                  # 業務規格（可選）
│       ├── workflows.md          # 業務流程定義
│       └── data-models.md        # 業務資料模型
│
├── guides/                        # 📘 實作指南（必須）
│   │
│   ├── setup/                     # 環境設置
│   │   ├── development.md        # 開發環境設置
│   │   ├── deployment.md         # 部署指南
│   │   └── docker.md             # Docker 容器化指南
│   │
│   ├── implementation/            # 功能實作指南
│   │   ├── [feature-a]-setup.md  # 功能 A 實作
│   │   ├── [feature-b]-setup.md  # 功能 B 實作
│   │   └── overview.md           # 功能總覽
│   │
│   ├── frontend/                  # 前端整合指南
│   │   ├── api-integration.md    # API 對接指南
│   │   ├── data-formats.md       # 資料格式說明
│   │   └── state-management.md   # 狀態管理（可選）
│   │
│   └── testing/                   # 測試指南（可選）
│       ├── unit-testing.md       # 單元測試
│       ├── integration-testing.md # 整合測試
│       └── e2e-testing.md        # E2E 測試
│
├── troubleshooting/               # 🔧 故障排除（推薦）
│   │
│   ├── database/                  # 資料庫相關問題
│   │   ├── connection-issues.md  # 連線問題
│   │   ├── permission-errors.md  # 權限錯誤
│   │   └── migration-issues.md   # 遷移問題
│   │
│   ├── api/                       # API 相關問題
│   │   ├── auth-issues.md        # 認證問題
│   │   └── performance-issues.md # 效能問題
│   │
│   ├── deployment/                # 部署相關問題
│   │   ├── docker-issues.md      # Docker 問題
│   │   └── env-config.md         # 環境變數配置
│   │
│   └── [module-name]/             # 特定模組問題
│       └── [specific-issue].md
│
├── archive/                       # 📦 歷史歸檔（推薦）
│   │
│   ├── requirement-changes/       # 需求變更記錄
│   │   └── YYYY-MM-DD-[change].md
│   │
│   ├── status-snapshots/          # 實作狀態快照
│   │   └── YYYY-MM-DD-status.md
│   │
│   ├── adr/                       # Architecture Decision Records
│   │   ├── 0001-use-postgresql.md
│   │   └── 0002-adopt-microservices.md
│   │
│   └── legacy/                    # 舊版文件
│       └── [outdated-docs].md
│
└── template/                      # 📝 文件模板（可選）
    ├── README_TEMPLATE.md        # README 模板
    ├── DIRECTORY_STRUCTURE.md    # 本文件
    ├── api-spec-template.md      # API 規格模板
    ├── guide-template.md         # 指南模板
    └── troubleshooting-template.md # 故障排除模板
```

## 📋 各目錄說明

### spec/ - 核心規格
**用途**: 存放所有系統設計與規格的權威文件
**特點**:
- 必須在實作前完成
- 任何變更需經過審核
- 是開發的唯一真相來源

**子目錄**:
- `api/` - RESTful API 定義（OpenAPI 規格）
- `database/` - 資料庫架構與 Schema
- `permissions/` - 權限系統設計
- `business/` - 業務邏輯與流程（可選）

### guides/ - 實作指南
**用途**: 提供開發、部署、測試的實用指南
**特點**:
- 面向實際操作
- 包含完整範例
- 定期更新

**子目錄**:
- `setup/` - 環境設置與部署
- `implementation/` - 功能實作細節
- `frontend/` - 前端開發對接
- `testing/` - 測試相關指南

### troubleshooting/ - 故障排除
**用途**: 累積常見問題的診斷與解決方案
**特點**:
- 問題導向
- 提供完整解決步驟
- 持續更新

**組織方式**:
按問題領域分類（database、api、deployment 等）

### archive/ - 歷史歸檔
**用途**: 保存已完成或過時的文件
**特點**:
- 保留歷史記錄
- 可追溯決策過程
- 定期清理

**子目錄**:
- `requirement-changes/` - 需求變更歷史
- `status-snapshots/` - 實作狀態快照
- `adr/` - 技術決策記錄（ADR 格式）
- `legacy/` - 已棄用文件

## 🎨 命名規範

### 文件命名
```
# 一般文件
[功能名稱]-[類型].md
範例: elasticsearch-setup.md

# 日期相關
YYYY-MM-DD-[描述].md
範例: 2025-10-08-permission-fix.md

# ADR 格式
序號-[決策內容].md
範例: 0001-use-timescaledb.md
```

### 目錄命名
```
# 使用小寫、連字符
[功能模組名稱]/
範例: elasticsearch/, user-management/

# 功能類型分類
api/, database/, deployment/
```

## 📝 文件模板

### 目錄必備文件

| 目錄 | 必須包含 | 說明 |
|------|---------|------|
| docs/ | README.md | 總導航索引 |
| spec/api/ | README.md, openapi.yml | API 規格與說明 |
| spec/database/ | architecture.md, schema.md | 資料庫設計 |
| guides/setup/ | development.md | 開發環境設置 |
| guides/frontend/ | api-integration.md | 前端對接指南 |

## 🚀 快速設置

### 1. 創建目錄結構
```bash
# 在專案根目錄執行
mkdir -p docs/{spec/{api,database,permissions,business},guides/{setup,implementation,frontend,testing},troubleshooting/{database,api,deployment},archive/{requirement-changes,status-snapshots,adr,legacy},template}
```

### 2. 創建必要文件
```bash
# 總導航
touch docs/README.md

# API 規格
touch docs/spec/api/{README.md,openapi.yml}

# 資料庫規格
touch docs/spec/database/{architecture.md,schema.md}

# 權限規格
touch docs/spec/permissions/rbac-guide.md

# 設置指南
touch docs/guides/setup/{development.md,deployment.md}

# 前端整合
touch docs/guides/frontend/api-integration.md
```

### 3. 複製模板
```bash
# 從本專案複製模板到新專案
cp -r docs/template /path/to/new-project/docs/
```

## 📐 擴展指南

### 新增功能模組
```bash
# 1. 在 spec/api/ 新增規格
touch docs/spec/api/[module-name]-api.md

# 2. 在 guides/implementation/ 新增實作指南
touch docs/guides/implementation/[module-name]-setup.md

# 3. 在 troubleshooting/ 新增故障排除目錄
mkdir -p docs/troubleshooting/[module-name]
touch docs/troubleshooting/[module-name]/common-issues.md

# 4. 更新 README.md 導航
```

### 新增子系統
```bash
# 1. 在 spec/ 新增子系統規格目錄
mkdir -p docs/spec/[subsystem-name]
touch docs/spec/[subsystem-name]/{architecture.md,specification.md}

# 2. 在 guides/ 新增對應指南
mkdir -p docs/guides/[subsystem-name]

# 3. 更新總 README.md
```

## ✅ 最佳實踐

1. **規格先行** - 實作前必須完成 spec/ 中的規格定義
2. **文件同步** - 程式碼更新時同步更新文件
3. **範例豐富** - 每個指南都應包含實際可執行的範例
4. **定期審查** - 每季度審查並歸檔過時文件
5. **版本標記** - 重要文件標註版本號與更新日期
6. **交叉引用** - 文件間使用相對路徑互相引用
7. **Markdown 規範** - 遵循一致的 Markdown 格式

## 🔄 維護流程

### 日常維護
- 新增功能 → 更新 spec/ 和 guides/
- 修復問題 → 記錄到 troubleshooting/
- 需求變更 → 更新規格並記錄到 archive/requirement-changes/

### 定期維護（建議每季度）
1. 審查所有文件的時效性
2. 歸檔已完成項目的文件到 archive/
3. 更新 README.md 導航
4. 檢查並修復失效的連結
5. 統一格式與風格

---

**模板版本**: 1.0.0
**最後更新**: 2025-10-08
**適用專案類型**: Web 應用、API 服務、微服務
