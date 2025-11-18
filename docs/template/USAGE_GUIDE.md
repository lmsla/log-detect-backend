# 📖 文件模板使用指南

> 如何使用這套規格驅動開發文件模板

## 🎯 模板概述

本模板提供一套完整的文件組織架構，適用於：
- Web 應用專案
- RESTful API 服務
- 微服務架構
- 前後端分離專案

**核心理念**: 規格驅動開發 (Specification-Driven Development)

---

## 🚀 快速開始

### 1. 複製模板到新專案

```bash
# 方法 A: 複製整個 template 目錄
cp -r /path/to/log-detect-backend/docs/template /path/to/new-project/docs/

# 方法 B: 使用 Git（推薦）
cd /path/to/new-project
mkdir -p docs
cd docs
git clone <template-repo-url> template
```

### 2. 創建目錄結構

```bash
# 進入專案 docs 目錄
cd /path/to/new-project/docs

# 創建標準目錄結構
mkdir -p spec/{api,database,permissions,business}
mkdir -p guides/{setup,implementation,frontend,testing}
mkdir -p troubleshooting/{database,api,deployment}
mkdir -p archive/{requirement-changes,status-snapshots,adr,legacy}
```

### 3. 複製並自定義 README

```bash
# 複製 README 模板
cp template/README_TEMPLATE.md README.md

# 使用編輯器替換佔位符
# [專案名稱] → 你的專案名稱
# [姓名/團隊] → 實際維護者
# YYYY-MM-DD → 當前日期
```

### 4. 創建初始文件

```bash
# API 規格
touch spec/api/{README.md,openapi.yml}

# 資料庫規格
touch spec/database/{architecture.md,schema.md}

# 開發環境設置
touch guides/setup/development.md

# 前端對接指南
touch guides/frontend/api-integration.md
```

---

## 📝 各模板使用說明

### README_TEMPLATE.md
**用途**: 文件中心總導航
**使用時機**: 專案初始化時創建

**自定義步驟**:
1. 替換 `[專案名稱]` 為實際專案名稱
2. 更新 API Base URL 為實際地址
3. 調整功能模組列表（根據專案實際功能）
4. 更新聯絡資訊
5. 刪除不適用的章節

**範例**:
```markdown
# 從
[專案名稱] 文件中心

# 改為
Log Detection System 文件中心
```

---

### DIRECTORY_STRUCTURE.md
**用途**: 目錄結構說明與最佳實踐
**使用時機**: 作為團隊參考文件

**自定義步驟**:
1. 根據專案特性調整目錄結構
2. 移除不需要的目錄（如 business/）
3. 新增專案特有的目錄
4. 更新命名規範以符合團隊習慣

---

### api-spec-template.md
**用途**: API 端點規格文件範本
**使用時機**: 每個新增的功能模組都應創建一份

**使用流程**:
```bash
# 1. 複製模板
cp template/api-spec-template.md spec/api/user-management-api.md

# 2. 自定義內容
# - 替換 [功能模組名稱] 為實際模組名稱
# - 替換 [module]/[resource] 為實際路徑
# - 填寫每個端點的詳細資訊
# - 更新資料模型定義

# 3. 移除不適用的端點
# 例如：如果只有 GET 和 POST，刪除 PUT 和 DELETE 章節
```

**關鍵區塊**:
- **認證與授權**: 定義所需權限
- **API 端點清單**: 每個端點的完整規格
- **資料模型**: 完整的資料結構定義
- **錯誤處理**: 統一的錯誤格式

---

### troubleshooting-template.md
**用途**: 故障排除文件範本
**使用時機**: 遇到問題並解決後立即記錄

**使用流程**:
```bash
# 1. 確定問題分類
# database/ | api/ | deployment/ | [module-name]/

# 2. 複製模板
cp template/troubleshooting-template.md \
   troubleshooting/database/connection-timeout.md

# 3. 填寫內容
# - 問題描述：詳細描述症狀
# - 診斷過程：記錄診斷步驟
# - 解決方案：提供具體可執行的解決步驟
# - 範例：提供完整的執行範例
```

**編寫原則**:
- ✅ 提供可直接複製執行的指令
- ✅ 包含預期輸出與實際輸出對比
- ✅ 說明每個步驟的作用
- ✅ 提供多個解決方案（從簡單到複雜）
- ✅ 加入預防措施

---

## 🎨 自定義指南

### 調整目錄結構

根據專案需求調整：

```bash
# 範例：新增 monitoring/ 規格目錄（監控系統專案）
mkdir -p spec/monitoring
touch spec/monitoring/{metrics.md,alerts.md}

# 範例：新增 mobile/ 前端指南（有移動端的專案）
mkdir -p guides/mobile
touch guides/mobile/api-integration.md
```

### 新增自定義模板

```bash
# 1. 在 template/ 創建新模板
touch template/deployment-guide-template.md

# 2. 參考現有模板編寫內容

# 3. 更新 USAGE_GUIDE.md 說明新模板用途
```

### 調整模板風格

**表情符號使用**:
- 如果團隊不喜歡 emoji，可全部移除
- 或統一使用特定風格的 emoji

**語言調整**:
- 模板使用繁體中文，可改為簡體中文或英文
- 使用 sed 批量替換：
  ```bash
  sed -i 's/資料庫/数据库/g' *.md
  ```

---

## 📋 使用檢查清單

### 專案初始化檢查清單

- [ ] 複製 template 目錄到專案
- [ ] 創建標準目錄結構
- [ ] 自定義 README.md
- [ ] 創建 spec/api/openapi.yml
- [ ] 創建 spec/database/schema.md
- [ ] 創建 guides/setup/development.md
- [ ] 創建 guides/frontend/api-integration.md
- [ ] 設置 Git 忽略規則（如果有敏感配置）

### 新增功能模組檢查清單

- [ ] 在 spec/api/ 創建 API 規格文件
- [ ] 在 guides/implementation/ 創建實作指南
- [ ] 更新 README.md 導航
- [ ] 更新 openapi.yml 新增端點定義
- [ ] 創建對應的 troubleshooting 目錄

### 文件維護檢查清單（每季度）

- [ ] 審查所有文件的時效性
- [ ] 歸檔已完成項目的文件
- [ ] 更新 README.md 導航
- [ ] 檢查並修復失效連結
- [ ] 統一文件格式與風格
- [ ] 更新變更歷史與版本號

---

## 💡 最佳實踐

### 1. 規格先行
```
錯誤流程: 實作 → 測試 → 寫文件
正確流程: 寫規格 → 審核 → 實作 → 更新文件
```

### 2. 保持同步
- 程式碼變更 → 立即更新文件
- API 調整 → 更新 openapi.yml
- 資料表變更 → 更新 schema.md

### 3. 範例豐富
每個指南都應包含：
- ✅ 完整可執行的指令
- ✅ 實際的輸出範例
- ✅ 常見錯誤與解決方法

### 4. 版本管理
重要文件應標註版本：
```markdown
> **規格版本**: 2.1.0
> **最後更新**: 2025-10-08
```

### 5. 交叉引用
文件間使用相對路徑引用：
```markdown
詳見 [API 規格](../../spec/api/openapi.yml)
```

---

## 🔧 常見問題

### Q1: 是否每個小功能都要創建獨立的 API 規格文件？

**A**: 視複雜度而定
- 簡單功能（1-3 個端點）→ 可以合併在主 API 規格中
- 複雜模組（5+ 個端點）→ 建議獨立文件

### Q2: troubleshooting 文件何時創建？

**A**: 遇到問題並解決後立即創建
- 不要等到「有空」才寫
- 趁記憶猶新時記錄完整步驟
- 即使是小問題也值得記錄

### Q3: 如何處理過時的文件？

**A**: 移到 archive/ 而非刪除
```bash
# 錯誤
rm old-document.md

# 正確
mv old-document.md archive/legacy/old-document-YYYY-MM-DD.md
```

### Q4: 前端和後端使用同一套文件嗎？

**A**: 共用規格，分離指南
- `spec/` - 前後端共用（API 規格、資料模型）
- `guides/implementation/` - 後端專用
- `guides/frontend/` - 前端專用

### Q5: 是否需要中英文雙語文件？

**A**: 視團隊需求
- 國際團隊 → 英文優先
- 本地團隊 → 母語優先
- 大型專案 → 考慮雙語

---

## 🔗 相關資源

- [Markdown 語法指南](https://www.markdownguide.org/)
- [OpenAPI 3.0 規範](https://swagger.io/specification/)
- [Architecture Decision Records](https://adr.github.io/)
- [Specification by Example](https://en.wikipedia.org/wiki/Specification_by_example)

---

## 📞 支援

**問題回報**: 如果發現模板問題或有改進建議，請透過以下方式聯絡：
- 提交 Issue
- 發送 Pull Request
- 聯絡維護團隊

---

**模板版本**: 1.0.0
**最後更新**: 2025-10-08
**維護者**: Log Detect Development Team
