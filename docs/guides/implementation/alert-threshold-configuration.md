# ES 監控告警閾值配置指南

## 📋 概述

ES 監控系統提供**兩種方式**配置告警閾值：

1. **獨立欄位配置**（推薦）- 前端友好，使用表單控件
2. **JSON 配置**（高級選項）- 向後兼容，支援複雜配置

## ✨ 方式 1：獨立欄位配置（推薦）

### 優點
- ✅ 前端可使用數字輸入框、滑桿等控件
- ✅ 每個欄位可獨立驗證
- ✅ 不需要了解 JSON 格式
- ✅ 支援預設值和模板

### API 請求範例

```json
POST /api/v1/elasticsearch/monitors
{
  "name": "Production ES",
  "host": "localhost",
  "port": 9200,
  "check_type": "health,performance",
  "interval": 60,

  // 告警閾值（可選，不設置則使用預設值）
  "cpu_usage_high": 70.0,
  "cpu_usage_critical": 80.0,
  "memory_usage_high": 75.0,
  "memory_usage_critical": 85.0,
  "disk_usage_high": 80.0,
  "disk_usage_critical": 90.0,
  "response_time_high": 2000,
  "response_time_critical": 5000,
  "unassigned_shards_threshold": 2,

  "receivers": ["admin@example.com"],
  "enable_monitor": true
}
```

### 欄位說明

| 欄位名稱 | 類型 | 單位 | 預設值 | 說明 |
|---------|------|------|--------|------|
| `cpu_usage_high` | float64 | % | 75.0 | CPU 使用率-高閾值 |
| `cpu_usage_critical` | float64 | % | 85.0 | CPU 使用率-危險閾值 |
| `memory_usage_high` | float64 | % | 80.0 | 記憶體使用率-高閾值 |
| `memory_usage_critical` | float64 | % | 90.0 | 記憶體使用率-危險閾值 |
| `disk_usage_high` | float64 | % | 85.0 | 磁碟使用率-高閾值 |
| `disk_usage_critical` | float64 | % | 95.0 | 磁碟使用率-危險閾值 |
| `response_time_high` | int64 | ms | 3000 | 響應時間-高閾值 |
| `response_time_critical` | int64 | ms | 10000 | 響應時間-危險閾值 |
| `unassigned_shards_threshold` | int | 個 | 1 | 未分配分片閾值 |

### 預設閾值模板

#### 🟢 寬鬆模板（開發/測試環境）
```json
{
  "cpu_usage_high": 85.0,
  "cpu_usage_critical": 95.0,
  "memory_usage_high": 85.0,
  "memory_usage_critical": 95.0,
  "disk_usage_high": 90.0,
  "disk_usage_critical": 98.0,
  "response_time_high": 5000,
  "response_time_critical": 15000,
  "unassigned_shards_threshold": 5
}
```

#### 🟡 標準模板（一般生產環境）
```json
{
  "cpu_usage_high": 75.0,
  "cpu_usage_critical": 85.0,
  "memory_usage_high": 80.0,
  "memory_usage_critical": 90.0,
  "disk_usage_high": 85.0,
  "disk_usage_critical": 95.0,
  "response_time_high": 3000,
  "response_time_critical": 10000,
  "unassigned_shards_threshold": 1
}
```

#### 🔴 嚴格模板（核心業務系統）
```json
{
  "cpu_usage_high": 60.0,
  "cpu_usage_critical": 70.0,
  "memory_usage_high": 70.0,
  "memory_usage_critical": 80.0,
  "disk_usage_high": 75.0,
  "disk_usage_critical": 85.0,
  "response_time_high": 1000,
  "response_time_critical": 3000,
  "unassigned_shards_threshold": 0
}
```

## 📝 方式 2：JSON 配置（高級選項）

### 用途
- 向後兼容舊版本
- 批量配置多個監控器
- 腳本自動化部署

### API 請求範例

```json
POST /api/v1/elasticsearch/monitors
{
  "name": "Production ES",
  "host": "localhost",
  "port": 9200,
  "check_type": "health,performance",

  // 使用 JSON 配置（不推薦新用戶使用）
  "alert_threshold": "{\"cpu_usage_high\":75.0,\"cpu_usage_critical\":85.0,\"memory_usage_high\":80.0,\"memory_usage_critical\":90.0,\"disk_usage_high\":85.0,\"disk_usage_critical\":95.0,\"response_time_high\":3000,\"response_time_critical\":10000,\"unassigned_shards\":1}"
}
```

## 🔄 配置優先級

系統按以下順序決定使用哪個閾值：

1. **獨立欄位**（最高優先級）
   - 如果設置了 `cpu_usage_high` 等欄位，使用這些值

2. **JSON 配置**（向後兼容）
   - 如果獨立欄位未設置，嘗試解析 `alert_threshold` JSON

3. **預設值**（最低優先級）
   - 如果以上都沒有，使用系統預設值

### 範例說明

```json
{
  "cpu_usage_high": 70.0,           // ✅ 使用此值（優先級1）
  "alert_threshold": "{\"cpu_usage_high\":75.0}"  // ❌ 被忽略
}
```

```json
{
  "cpu_usage_high": null,           // ❌ 未設置
  "alert_threshold": "{\"cpu_usage_high\":75.0}"  // ✅ 使用此值（優先級2）
}
```

```json
{
  "cpu_usage_high": null,           // ❌ 未設置
  "alert_threshold": ""             // ❌ 未設置
}
// ✅ 使用預設值 75.0（優先級3）
```

## 🎨 前端實作建議

### React 範例（使用 Ant Design）

```tsx
import { Form, InputNumber, Slider, Select } from 'antd';

const ThresholdTemplate = {
  relaxed: {
    label: '寬鬆（開發環境）',
    values: { cpu_usage_high: 85.0, cpu_usage_critical: 95.0, ... }
  },
  standard: {
    label: '標準（一般生產）',
    values: { cpu_usage_high: 75.0, cpu_usage_critical: 85.0, ... }
  },
  strict: {
    label: '嚴格（核心業務）',
    values: { cpu_usage_high: 60.0, cpu_usage_critical: 70.0, ... }
  }
};

function AlertThresholdForm() {
  const [form] = Form.useForm();

  const applyTemplate = (templateName) => {
    form.setFieldsValue(ThresholdTemplate[templateName].values);
  };

  return (
    <Form form={form}>
      {/* 快速模板選擇 */}
      <Form.Item label="快速模板">
        <Select onChange={applyTemplate} placeholder="選擇預設模板">
          <Select.Option value="relaxed">🟢 寬鬆</Select.Option>
          <Select.Option value="standard">🟡 標準（推薦）</Select.Option>
          <Select.Option value="strict">🔴 嚴格</Select.Option>
        </Select>
      </Form.Item>

      {/* CPU 閾值 */}
      <Form.Item label="CPU 使用率 - 高" name="cpu_usage_high">
        <InputNumber min={0} max={100} step={1} suffix="%" />
      </Form.Item>

      <Form.Item label="CPU 使用率 - 危險" name="cpu_usage_critical">
        <Slider min={0} max={100} marks={{ 0: '0%', 100: '100%' }} />
      </Form.Item>

      {/* 記憶體閾值 */}
      <Form.Item label="記憶體使用率 - 高" name="memory_usage_high">
        <InputNumber min={0} max={100} step={1} suffix="%" />
      </Form.Item>

      {/* ... 其他閾值欄位 */}
    </Form>
  );
}
```

### Vue 範例（使用 Element Plus）

```vue
<template>
  <el-form :model="form">
    <!-- 快速模板 -->
    <el-form-item label="快速模板">
      <el-select @change="applyTemplate" placeholder="選擇模板">
        <el-option label="🟢 寬鬆" value="relaxed" />
        <el-option label="🟡 標準（推薦）" value="standard" />
        <el-option label="🔴 嚴格" value="strict" />
      </el-select>
    </el-form-item>

    <!-- CPU 閾值 -->
    <el-form-item label="CPU 使用率 - 高">
      <el-input-number
        v-model="form.cpu_usage_high"
        :min="0"
        :max="100"
        :step="1"
      />
      <span style="margin-left: 8px">%</span>
    </el-form-item>

    <el-form-item label="CPU 使用率 - 危險">
      <el-slider
        v-model="form.cpu_usage_critical"
        :min="0"
        :max="100"
        show-stops
      />
    </el-form-item>
  </el-form>
</template>

<script setup>
import { reactive } from 'vue';

const form = reactive({
  cpu_usage_high: 75.0,
  cpu_usage_critical: 85.0,
  // ... 其他欄位
});

const templates = {
  relaxed: { cpu_usage_high: 85.0, cpu_usage_critical: 95.0, ... },
  standard: { cpu_usage_high: 75.0, cpu_usage_critical: 85.0, ... },
  strict: { cpu_usage_high: 60.0, cpu_usage_critical: 70.0, ... }
};

const applyTemplate = (templateName) => {
  Object.assign(form, templates[templateName]);
};
</script>
```

## 🚀 部署步驟

### 1. 更新資料庫

```bash
mysql -u monitor -p config < docs/troubleshooting/add_threshold_fields.sql
```

### 2. 重啟應用

```bash
# GORM AutoMigrate 會自動添加新欄位（如果尚未手動添加）
# 重啟應用即可
```

### 3. 驗證功能

```bash
# 測試 API
curl -X POST http://localhost:8080/api/v1/elasticsearch/monitors \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Monitor",
    "host": "localhost",
    "port": 9200,
    "cpu_usage_high": 70.0,
    "cpu_usage_critical": 80.0
  }'
```

## ⚠️ 注意事項

1. **閾值範圍**
   - 百分比類（CPU、記憶體、磁碟）：0-100
   - 響應時間：建議 100-30000ms
   - 未分配分片：建議 0-10

2. **High vs Critical**
   - `high` 閾值應 < `critical` 閾值
   - 建議差距：5-10%

3. **向後兼容**
   - 舊版本使用 JSON 配置的監控器仍可正常運作
   - 建議逐步遷移到獨立欄位配置

4. **NULL 值處理**
   - 欄位為 NULL 時使用預設值
   - 不會報錯，保證系統穩定性

## 📊 監控建議

| 環境類型 | 建議模板 | 說明 |
|---------|---------|------|
| 開發環境 | 🟢 寬鬆 | 減少告警干擾 |
| 測試環境 | 🟡 標準 | 平衡監控和容錯 |
| 預發布環境 | 🟡 標準 | 接近生產配置 |
| 生產環境（一般） | 🟡 標準 | 推薦配置 |
| 生產環境（核心） | 🔴 嚴格 | 及早發現問題 |

## 🔗 相關文檔

- [ES 監控 API 規格](../../spec/api/elasticsearch-api-spec.md)
- [告警去重配置](./CHANGELOG-ES-ALERT-DEDUPE.md)
- [故障排除](../../troubleshooting/)
