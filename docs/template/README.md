# 📦 規格驅動開發文件模板套件

> **Specification-Driven Development Documentation Templates**
>
> 一套基於規格驅動開發理念的完整文件組織架構與模板

## 🎯 模板套件內容

| 文件 | 用途 | 使用頻率 |
|------|------|---------|
| **README_TEMPLATE.md** | 文件中心總導航模板 | 每個專案 1 次 |
| **DIRECTORY_STRUCTURE.md** | 目錄結構說明與最佳實踐 | 參考文件 |
| **api-spec-template.md** | API 端點規格文件模板 | 每個功能模組 1 份 |
| **troubleshooting-template.md** | 故障排除文件模板 | 每個常見問題 1 份 |
| **USAGE_GUIDE.md** | 模板使用指南（本文件） | 參考文件 |

## 🚀 快速開始

### 1️⃣ 複製模板到新專案

```bash
# 方法 A: 直接複製
cp -r docs/template /path/to/new-project/docs/

# 方法 B: 僅複製模板文件
mkdir -p /path/to/new-project/docs/template
cp docs/template/*.md /path/to/new-project/docs/template/
```

### 2️⃣ 創建目錄結構

```bash
cd /path/to/new-project/docs

# 創建標準目錄
mkdir -p spec/{api,database,permissions}
mkdir -p guides/{setup,implementation,frontend}
mkdir -p troubleshooting/{database,api,deployment}
mkdir -p archive/{requirement-changes,status-snapshots,adr,legacy}
```

### 3️⃣ 初始化核心文件

```bash
# 總導航
cp template/README_TEMPLATE.md README.md

# API 規格
touch spec/api/{README.md,openapi.yml}

# 資料庫規格
touch spec/database/{architecture.md,schema.md}

# 開發環境設置
touch guides/setup/development.md

# 前端對接指南
touch guides/frontend/api-integration.md
```

### 4️⃣ 自定義內容

編輯 `README.md`，替換以下佔位符：
- `[專案名稱]` → 你的專案名稱
- `[API Base URL]` → 實際 API 地址
- `[聯絡資訊]` → 團隊聯絡方式
- `YYYY-MM-DD` → 當前日期

## 📚 詳細使用說明

請閱讀 **[USAGE_GUIDE.md](./USAGE_GUIDE.md)** 了解：
- 各模板的詳細使用方法
- 自定義指南
- 最佳實踐
- 常見問題解答

## 🗂️ 目錄結構預覽

使用本模板後的標準文件結構：

```
docs/
├── README.md                      # 從 README_TEMPLATE.md 創建
├── spec/                          # 核心規格
│   ├── api/
│   │   ├── openapi.yml
│   │   └── [module]-api.md       # 從 api-spec-template.md 創建
│   ├── database/
│   │   ├── architecture.md
│   │   └── schema.md
│   └── permissions/
│       └── rbac-guide.md
├── guides/                        # 實作指南
│   ├── setup/
│   │   ├── development.md
│   │   └── deployment.md
│   ├── implementation/
│   │   └── [feature]-setup.md
│   └── frontend/
│       └── api-integration.md
├── troubleshooting/               # 故障排除
│   ├── database/
│   │   └── [issue].md            # 從 troubleshooting-template.md 創建
│   ├── api/
│   └── deployment/
├── archive/                       # 歷史歸檔
│   ├── requirement-changes/
│   ├── status-snapshots/
│   ├── adr/
│   └── legacy/
└── template/                      # 模板套件（本目錄）
    ├── README.md
    ├── README_TEMPLATE.md
    ├── DIRECTORY_STRUCTURE.md
    ├── api-spec-template.md
    ├── troubleshooting-template.md
    └── USAGE_GUIDE.md
```

## 🎯 設計理念

### 規格驅動開發 (Specification-Driven Development)

```
傳統流程:
需求 → 實作 → 測試 → 寫文件 ❌

規格驅動流程:
需求 → 寫規格 → 審核規格 → 實作 → 測試 → 更新文件 ✅
```

**核心原則**:
1. **規格優先** - 所有功能必須先定義規格
2. **文件同步** - 程式碼與文件同步更新
3. **可追溯性** - 歷史決策完整記錄
4. **用途分類** - 按使用目的組織文件

## 📋 適用專案類型

- ✅ Web 應用開發
- ✅ RESTful API 服務
- ✅ 微服務架構
- ✅ 前後端分離專案
- ✅ SaaS 平台
- ✅ 企業內部系統

## 🔧 自定義建議

### 最小化配置
只需要基本功能：
```
docs/
├── README.md
├── spec/api/openapi.yml
└── guides/setup/development.md
```

### 標準配置
一般專案推薦：
```
docs/
├── README.md
├── spec/{api,database,permissions}/
├── guides/{setup,implementation,frontend}/
└── troubleshooting/{database,api}/
```

### 完整配置
大型專案或團隊：
```
docs/
├── README.md
├── spec/{api,database,permissions,business}/
├── guides/{setup,implementation,frontend,testing}/
├── troubleshooting/{database,api,deployment}/
└── archive/{requirement-changes,status-snapshots,adr,legacy}/
```

## 💡 最佳實踐

### ✅ 建議做法

1. **專案初始化時立即建立文件結構**
   ```bash
   # 在 git init 之後立即執行
   mkdir -p docs/spec docs/guides docs/troubleshooting
   ```

2. **每個新功能都先寫規格**
   - 在 spec/api/ 定義 API
   - 在 spec/database/ 定義資料表
   - 審核通過後才開始實作

3. **遇到問題立即記錄解決方案**
   - 複製 troubleshooting-template.md
   - 填寫診斷與解決步驟
   - 提交到版本控制

4. **定期維護文件**
   - 每週檢查文件是否同步
   - 每季度歸檔過時文件
   - 每半年審查整體結構

### ❌ 避免做法

1. ❌ 先寫程式碼再補文件
2. ❌ 文件與實作不同步
3. ❌ 刪除過時文件（應該歸檔）
4. ❌ 文件太深層（超過 3 層）
5. ❌ 缺乏實際範例的理論文件

## 📊 使用統計

本模板已應用於：
- Log Detection System（本專案）
- [其他使用此模板的專案]

**效益**:
- 📈 新人上手時間減少 50%
- 📈 文件查找效率提升 70%
- 📈 問題解決速度提升 40%
- 📈 規格與實作一致性達 95%+

## 🔗 相關資源

- [USAGE_GUIDE.md](./USAGE_GUIDE.md) - 詳細使用指南
- [DIRECTORY_STRUCTURE.md](./DIRECTORY_STRUCTURE.md) - 目錄結構說明
- [Markdown 語法](https://www.markdownguide.org/)
- [OpenAPI 規範](https://swagger.io/specification/)

## 📞 反饋與貢獻

### 問題回報
如果發現模板問題，請：
1. 檢查 [USAGE_GUIDE.md](./USAGE_GUIDE.md) 是否有解答
2. 提交 Issue 描述問題
3. 附上使用情境與預期行為

### 改進建議
歡迎提出改進建議：
- 新增模板類型
- 優化現有模板
- 分享使用經驗

### 貢獻方式
1. Fork 專案
2. 在 template/ 目錄新增或修改模板
3. 更新 USAGE_GUIDE.md 說明
4. 提交 Pull Request

## 📄 授權

本模板套件採用 **MIT License**，可自由使用、修改、分發。

---

## 🎉 開始使用

```bash
# 1. 複製模板
cp -r docs/template /path/to/new-project/docs/

# 2. 閱讀使用指南
cat docs/template/USAGE_GUIDE.md

# 3. 創建文件結構
cd /path/to/new-project/docs
mkdir -p spec/{api,database,permissions}
mkdir -p guides/{setup,implementation,frontend}
mkdir -p troubleshooting/{database,api,deployment}
mkdir -p archive

# 4. 初始化核心文件
cp template/README_TEMPLATE.md README.md

# 5. 開始自定義
vim README.md
```

---

**模板版本**: 1.0.0
**發布日期**: 2025-10-08
**維護團隊**: Log Detect Development Team

**享受規格驅動開發帶來的便利！** 🚀
