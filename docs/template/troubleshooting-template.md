# 🔧 [問題類型] 故障排除指南

> **最後更新**: YYYY-MM-DD
> **適用版本**: [版本範圍]
> **維護者**: [姓名/團隊]

## 📋 問題概述

**問題描述**: 簡要描述此問題的症狀與影響範圍

**常見表現**:
- 現象 1：具體描述
- 現象 2：具體描述
- 現象 3：具體描述

**影響範圍**:
- [ ] 開發環境
- [ ] 測試環境
- [ ] 生產環境

**優先級**: 🔴 高 / 🟡 中 / 🟢 低

---

## 🔍 問題診斷

### 快速檢查清單

在深入診斷前，先快速檢查以下項目：

- [ ] **檢查項目 1**: 具體檢查內容
  ```bash
  # 檢查指令
  command to check
  ```
  **預期結果**: 應該看到什麼

- [ ] **檢查項目 2**: 具體檢查內容
  ```bash
  # 檢查指令
  command to check
  ```
  **預期結果**: 應該看到什麼

- [ ] **檢查項目 3**: 具體檢查內容
  ```bash
  # 檢查指令
  command to check
  ```
  **預期結果**: 應該看到什麼

---

## 🎯 根本原因分析

### 原因 1: [具體原因描述]

**發生條件**:
- 條件 A
- 條件 B
- 條件 C

**診斷方法**:
```bash
# 查看相關日誌
grep "ERROR" /path/to/logfile.log

# 檢查相關配置
cat /path/to/config.yml
```

**判斷依據**:
如果看到以下錯誤訊息，表示是此原因：
```
ERROR: specific error message here
```

---

### 原因 2: [具體原因描述]

**發生條件**:
- 條件 A
- 條件 B

**診斷方法**:
```bash
# 診斷指令
diagnostic command
```

**判斷依據**:
描述如何確認是此原因

---

## ✅ 解決方案

### 方案 A: [方案名稱]（推薦）

**適用情況**: 描述何時使用此方案

**解決步驟**:

#### 步驟 1: 準備工作
```bash
# 備份相關數據（重要！）
cp /path/to/file /path/to/file.backup

# 停止相關服務
systemctl stop service-name
```

#### 步驟 2: 執行修復
```bash
# 執行修復指令
fix command here
```

**預期輸出**:
```
Success message or expected output
```

#### 步驟 3: 驗證結果
```bash
# 驗證指令
verification command
```

**驗證標準**:
- ✅ 檢查點 1
- ✅ 檢查點 2
- ✅ 檢查點 3

#### 步驟 4: 重啟服務
```bash
# 重啟服務
systemctl start service-name

# 檢查服務狀態
systemctl status service-name
```

---

### 方案 B: [替代方案名稱]

**適用情況**: 描述何時使用此方案（例如：方案 A 失敗時）

**解決步驟**:

#### 步驟 1: [步驟名稱]
```bash
# 指令
command here
```

詳細說明每個步驟的作用

#### 步驟 2: [步驟名稱]
```bash
# 指令
command here
```

---

### 方案 C: 完全重置（最後手段）

⚠️ **警告**: 此方案會清除所有數據，僅在其他方案都失敗時使用

**步驟**:
```bash
# 1. 完整備份
backup command

# 2. 清除數據
cleanup command

# 3. 重新初始化
initialization command

# 4. 恢復配置
restore command
```

---

## 📊 完整範例

### 情境描述
具體描述一個真實遇到的問題情境

### 完整執行過程
```bash
# 1. 診斷問題
$ diagnostic command
output showing the problem

# 2. 確認根本原因
$ check command
output confirming the cause

# 3. 執行修復
$ fix command
✅ Success message

# 4. 驗證結果
$ verify command
✅ All checks passed
```

---

## 🔐 權限相關

如果涉及權限問題，說明所需權限：

**必要權限**:
- 系統權限：sudo / root
- 數據庫權限：SUPERUSER / OWNER
- 應用權限：admin / specific_permission

**權限檢查**:
```bash
# 檢查當前用戶權限
whoami

# 檢查數據庫權限
psql -U postgres -c "\du"
```

---

## ⚠️ 注意事項

### 執行前必讀
1. **備份**: 必須先備份相關數據
2. **環境確認**: 確認在正確的環境執行
3. **權限確認**: 確保有足夠的操作權限
4. **影響範圍**: 了解操作可能影響的範圍

### 常見錯誤
❌ **錯誤 1**: 忘記備份數據
- **後果**: 數據可能永久丟失
- **預防**: 執行前必須備份

❌ **錯誤 2**: 在生產環境直接測試
- **後果**: 可能導致服務中斷
- **預防**: 先在測試環境驗證

---

## 🆘 如果仍無法解決

### 收集診斷資訊

執行以下指令收集診斷資訊：

```bash
# 1. 系統資訊
uname -a
cat /etc/os-release

# 2. 服務狀態
systemctl status service-name

# 3. 相關日誌（最近 100 行）
tail -100 /path/to/logfile.log

# 4. 配置檔案
cat /path/to/config.yml

# 5. 錯誤訊息（完整的 stack trace）
grep -A 20 "ERROR" /path/to/logfile.log
```

### 聯絡支援

提供以下資訊給技術支援：

1. **問題描述**: 詳細描述問題症狀
2. **復現步驟**: 如何重現此問題
3. **環境資訊**: 作業系統、版本、配置
4. **診斷結果**: 上述診斷資訊的輸出
5. **已嘗試方案**: 已經嘗試過哪些解決方案

**聯絡方式**:
- Email: support@example.com
- Slack: #tech-support
- Issue Tracker: [連結]

---

## 📝 預防措施

### 長期解決方案

為避免此問題再次發生：

1. **配置優化**: 具體的配置調整建議
   ```yaml
   # 建議配置
   config_key: recommended_value
   ```

2. **監控設置**: 設置監控以提早發現問題
   ```bash
   # 設置告警
   monitoring setup command
   ```

3. **自動化**: 將修復步驟自動化
   ```bash
   # 創建自動修復腳本
   script example
   ```

### 最佳實踐

- ✅ 定期檢查：每週執行健康檢查
- ✅ 日誌監控：設置日誌告警
- ✅ 版本管理：使用穩定版本
- ✅ 定期備份：每日自動備份

---

## 📚 相關文件

- [相關規格文件](../../spec/)
- [實作指南](../../guides/)
- [其他故障排除指南](../)

---

## 📊 問題統計

**發生頻率**: 罕見 / 偶爾 / 常見
**首次發現**: YYYY-MM-DD
**最後更新**: YYYY-MM-DD
**影響用戶數**: 約 X 人次

---

## 📋 變更歷史

| 日期 | 版本 | 變更內容 | 維護者 |
|------|------|---------|--------|
| YYYY-MM-DD | 1.0.0 | 初始版本 | [姓名] |
| YYYY-MM-DD | 1.1.0 | 新增方案 B | [姓名] |

---

**反饋**: 如果此文件幫助您解決問題，或有任何改進建議，請聯絡文件維護者。
