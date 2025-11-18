# Elasticsearch 告警 API 修復記錄

**日期**: 2025-10-22
**版本**: v1.3
**影響**: `/api/v1/elasticsearch/alerts` API 從完全無法使用恢復正常

---

## 🐛 問題概述

前端調用告警 API 時遇到兩個關鍵問題，導致 API 返回 500 錯誤：

1. **NULL 值掃描錯誤**：資料庫中的 NULL 值無法直接掃描到 Go string 類型
2. **PostgreSQL 數組參數錯誤**：Go 切片無法直接作為 PostgreSQL ANY() 參數

---

## 問題 1: NULL 值掃描錯誤

### 🔴 錯誤症狀

**請求**:
```bash
GET /api/v1/elasticsearch/alerts?page=1&page_size=20
```

**錯誤響應**:
```json
{
  "msg": "failed to scan alert: sql: Scan error on column index 10, name \"resolved_by\": converting NULL to string is unsupported",
  "success": false
}
```

### 🔍 根本原因

在 `services/es_alert_service.go` 中，`GetAlerts()` 和 `GetAlertByID()` 函數使用 `rows.Scan()` 時，直接將可空欄位掃描到 `string` 類型：

```go
// ❌ 錯誤代碼
err := rows.Scan(
    &alert.Time,
    &alert.MonitorID,
    &alert.AlertType,
    &alert.Severity,
    &alert.Status,
    &alert.Message,
    &alert.ClusterName,        // 可能為 NULL
    &alert.ThresholdValue,
    &alert.ActualValue,
    &alert.ResolvedAt,
    &alert.ResolvedBy,         // 可能為 NULL ❌
    &alert.ResolutionNote,     // 可能為 NULL ❌
    &alert.AcknowledgedAt,
    &alert.AcknowledgedBy,     // 可能為 NULL ❌
    &alert.Metadata,           // 可能為 NULL ❌
)
```

**資料庫欄位定義**:
```sql
CREATE TABLE es_alert_history (
    ...
    cluster_name TEXT,           -- 可為 NULL
    resolved_by TEXT,            -- 可為 NULL
    resolution_note TEXT,        -- 可為 NULL
    acknowledged_by TEXT,        -- 可為 NULL
    metadata JSONB               -- 可為 NULL
);
```

PostgreSQL 中未設置值的欄位為 NULL，Go 的 `database/sql` 包無法直接將 NULL 值掃描到非指針的 `string` 類型。

### ✅ 解決方案

使用 `sql.NullString` 來處理可空欄位：

```go
// ✅ 正確代碼
import (
    "database/sql"
    // ...
)

// 聲明 NullString 變量
var clusterName, resolvedBy, resolutionNote, acknowledgedBy, metadata sql.NullString

// 掃描到 NullString
err := rows.Scan(
    &alert.Time,
    &alert.MonitorID,
    &alert.AlertType,
    &alert.Severity,
    &alert.Status,
    &alert.Message,
    &clusterName,        // sql.NullString ✅
    &alert.ThresholdValue,
    &alert.ActualValue,
    &alert.ResolvedAt,
    &resolvedBy,         // sql.NullString ✅
    &resolutionNote,     // sql.NullString ✅
    &alert.AcknowledgedAt,
    &acknowledgedBy,     // sql.NullString ✅
    &metadata,           // sql.NullString ✅
)

// 檢查 Valid 屬性，只在非 NULL 時賦值
if clusterName.Valid {
    alert.ClusterName = clusterName.String
}
if resolvedBy.Valid {
    alert.ResolvedBy = resolvedBy.String
}
if resolutionNote.Valid {
    alert.ResolutionNote = resolutionNote.String
}
if acknowledgedBy.Valid {
    alert.AcknowledgedBy = acknowledgedBy.String
}
if metadata.Valid {
    alert.Metadata = metadata.String
}
```

### 📝 修改的函數

1. **GetAlerts()** (`services/es_alert_service.go:100-143`)
2. **GetAlertByID()** (`services/es_alert_service.go:148-202`)

---

## 問題 2: PostgreSQL 數組參數錯誤

### 🔴 錯誤症狀

**請求**:
```bash
GET /api/v1/elasticsearch/alerts?severity[]=critical&severity[]=medium
```

**錯誤響應**:
```json
{
  "msg": "failed to count alerts: sql: converting argument $3 type: unsupported type []string, a slice of string",
  "success": false
}
```

### 🔍 根本原因

在構建 PostgreSQL 查詢時，直接將 Go 的 `[]string` 切片作為參數傳遞給 `ANY()` 函數：

```go
// ❌ 錯誤代碼
if len(params.Severity) > 0 {
    query += fmt.Sprintf(" AND severity = ANY($%d)", argIndex)
    args = append(args, params.Severity)  // ❌ []string 無法直接使用
    argIndex++
}
```

PostgreSQL 的 `database/sql` 驅動不支持直接綁定 Go 切片作為數組參數。

### ✅ 解決方案

使用 `pq.Array()` 包裝切片參數：

```go
// ✅ 正確代碼
import (
    "github.com/lib/pq"
    // ...
)

// 狀態過濾
if len(params.Status) > 0 {
    query += fmt.Sprintf(" AND status = ANY($%d)", argIndex)
    args = append(args, pq.Array(params.Status))  // ✅ 使用 pq.Array()
    argIndex++
}

// 嚴重性過濾
if len(params.Severity) > 0 {
    query += fmt.Sprintf(" AND severity = ANY($%d)", argIndex)
    args = append(args, pq.Array(params.Severity))  // ✅ 使用 pq.Array()
    argIndex++
}

// 告警類型過濾
if len(params.AlertType) > 0 {
    query += fmt.Sprintf(" AND alert_type = ANY($%d)", argIndex)
    args = append(args, pq.Array(params.AlertType))  // ✅ 使用 pq.Array()
    argIndex++
}
```

### 📝 修改位置

**文件**: `services/es_alert_service.go`

1. Import 添加：`"github.com/lib/pq"` (line 11)
2. Status 過濾：`pq.Array(params.Status)` (line 54)
3. Severity 過濾：`pq.Array(params.Severity)` (line 61)
4. AlertType 過濾：`pq.Array(params.AlertType)` (line 75)

---

## 📊 測試結果

### ✅ 測試 1: 基本查詢（無過濾器）

**請求**:
```bash
GET /api/v1/elasticsearch/alerts?page=1&page_size=5
Authorization: Bearer <token>
```

**響應**:
```json
{
  "body": {
    "items": [
      {
        "time": "2025-10-22T13:53:04.694643+08:00",
        "monitor_id": 3,
        "alert_type": "health",
        "severity": "high",
        "message": "Unassigned shards detected: 16",
        "status": "active",
        "cluster_name": "redhat9_elk",
        "threshold_value": 1,
        "actual_value": 16
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 5,
      "total": 268,
      "total_pages": 54
    }
  },
  "msg": "查詢成功",
  "success": true
}
```

**狀態**: ✅ 成功

---

### ✅ 測試 2: 帶時間範圍和 severity 過濾

**請求**:
```bash
GET /api/v1/elasticsearch/alerts?page=1&page_size=20&start_time=2025-10-21T05:53:10.940Z&end_time=2025-10-22T05:53:10.940Z&severity[]=critical&severity[]=medium
Authorization: Bearer <token>
```

**響應**:
```json
{
  "body": {
    "items": null,
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 0,
      "total_pages": 0
    }
  },
  "msg": "查詢成功",
  "success": true
}
```

**狀態**: ✅ 成功（該時間範圍內無 critical/medium 告警）

---

### ✅ 測試 3: 帶 high severity 過濾

**請求**:
```bash
GET /api/v1/elasticsearch/alerts?page=1&page_size=5&severity[]=high
Authorization: Bearer <token>
```

**響應**:
```json
{
  "body": {
    "items": [
      {
        "time": "2025-10-22T13:53:04.694643+08:00",
        "monitor_id": 3,
        "alert_type": "health",
        "severity": "high",
        "message": "Unassigned shards detected: 16",
        "status": "active",
        "cluster_name": "redhat9_elk",
        "threshold_value": 1,
        "actual_value": 16
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 5,
      "total": 19,
      "total_pages": 4
    }
  },
  "msg": "查詢成功",
  "success": true
}
```

**狀態**: ✅ 成功

---

### ✅ 測試 4: 多重過濾（severity + alert_type）

**請求**:
```bash
GET /api/v1/elasticsearch/alerts?page=1&page_size=5&severity[]=high&severity[]=critical&alert_type[]=health
Authorization: Bearer <token>
```

**響���**:
```json
{
  "body": {
    "items": [
      {
        "time": "2025-10-22T14:22:06.637943+08:00",
        "monitor_id": 3,
        "alert_type": "health",
        "severity": "high",
        "message": "Unassigned shards detected: 16",
        "status": "active",
        "cluster_name": "redhat9_elk",
        "threshold_value": 1,
        "actual_value": 16
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 5,
      "total": 24,
      "total_pages": 5
    }
  },
  "msg": "查詢成功",
  "success": true
}
```

**狀態**: ✅ 成功

---

## 📁 修改的文件清單

| 文件 | 修改內容 | 行數 |
|------|---------|------|
| `services/es_alert_service.go` | 添加 `database/sql` import | 4 |
| `services/es_alert_service.go` | 添加 `github.com/lib/pq` import | 11 |
| `services/es_alert_service.go` | 修復 GetAlerts() NULL 值處理 | 103-143 |
| `services/es_alert_service.go` | 修復 GetAlertByID() NULL 值處理 | 151-202 |
| `services/es_alert_service.go` | 修復 Status 過濾參數 | 54 |
| `services/es_alert_service.go` | 修復 Severity 過濾參數 | 61 |
| `services/es_alert_service.go` | 修復 AlertType 過濾參數 | 75 |

---

## 🎯 影響範圍

### ✅ 修復的 API 端點

| 端點 | 方法 | 狀態 |
|------|------|------|
| `/api/v1/elasticsearch/alerts` | GET | ✅ 修復成功 |
| `/api/v1/elasticsearch/alerts/{monitor_id}` | GET | ✅ 修復成功 |

### ✅ 支援的查詢參數

- ✅ `page` - 頁碼
- ✅ `page_size` - 每頁筆數
- ✅ `monitor_id` - 監控器 ID
- ✅ `start_time` - 開始時間（ISO 8601）
- ✅ `end_time` - 結束時間（ISO 8601）
- ✅ `status[]` - 狀態過濾（可多選）
- ✅ `severity[]` - 嚴重性過濾（可多選）
- ✅ `alert_type[]` - 告警類型過濾（可多選）

---

## 📚 相關知識

### sql.NullString 用法

`sql.NullString` 是 Go 標準庫提供的結構，用於處理資料庫中的 NULL 值：

```go
type NullString struct {
    String string  // 實際的字串值
    Valid  bool    // true 表示非 NULL，false 表示 NULL
}
```

**使用場景**:
- 資料庫欄位允許 NULL
- 需要區分空字串 (`""`) 和 NULL

### pq.Array() 用法

`pq.Array()` 是 PostgreSQL 驅動提供的函數，用於將 Go 切片轉換為 PostgreSQL 數組：

```go
// 支援的類型
pq.Array([]string{"a", "b", "c"})
pq.Array([]int{1, 2, 3})
pq.Array([]float64{1.1, 2.2, 3.3})
```

**使用場景**:
- 使用 PostgreSQL 的 `ANY()` 函數
- 使用 PostgreSQL 的數組操作符

---

## 🔗 相關文件

- [Elasticsearch API 規格](../../spec/api/elasticsearch-api-spec.md)
- [實作狀態報告](../../spec/api/elasticsearch-implementation-status.md)
- [TimescaleDB 遷移指南](../../guides/implementation/timescaledb-migration-guide.md)

---

## 👥 維護者

**開發者**: Log Detect 開發團隊
**修復日期**: 2025-10-22
**測試日期**: 2025-10-22
**部署狀態**: ✅ 已部署至開發環境

---

## 📋 檢查清單

修復前端問題前的檢查清單：

- [x] 識別錯誤信息
- [x] 分析根本原因
- [x] 實作修復方案
- [x] 重新編譯後端
- [x] 重啟後端服務
- [x] 測試基本查詢
- [x] 測試帶過濾條件的查詢
- [x] 測試多重過濾條件
- [x] 更新文檔
- [x] 提交代碼變更

---

**狀態**: ✅ 完成
**下次檢查**: 無需檢查（問題已完全修復）
