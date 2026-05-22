# Log Detect Backend YML 模式測試 Checklist

本文件提供給測試同事快速執行 YML 模式驗證，不需先完整閱讀操作手冊。

完整手冊請參考：

- `yml-mode-operation-manual.md`

---

## 1. 測試前確認

- 已備份 MySQL 資料庫
- 已確認本次要測試的是 **YML 模式**
- 已準備好以下檔案：
  - `setting.yml`
  - `config.yml`
  - `devices.yml`（若有獨立維護裝置清單）

---

## 2. 設定檢查

### `setting.yml`

- `config_source: "yml"`
- `database` 連線資訊正確
- `timescale` 連線資訊正確
- `features.timescaledb` 已依需求開啟
- `features.es_monitoring` 已依需求開啟

### `config.yml`

- `es_connections` 已設定
- `targets` 已設定
- 每個 `targets[].indices` 均已設定：
  - `logname`
  - `index`
  - `device_group`
  - `field`
  - `period`
  - `unit`
  - `es_connection`

### `devices.yml`

- 若檔案存在，已確認內容正確
- 已知 `devices.yml` 會覆蓋 `config.yml` 的 `devices` 區段
- 若要停用設備，已使用 `disabled_devices`

### `yml_reload`

- 若希望 `devices.yml` 變更免重啟生效，已確認：
  - `devices_enabled: true`
  - `devices_interval` 設定合理

---

## 3. 啟動服務

在後端目錄執行：

```bash
cd 到主程式所在目錄

# 單次測試
./log-detect

# 背景執行
nohup ./log-detect &
```

---

## 4. 啟動日誌檢查

### 預期看到

- `SQL Database 連線成功`
- `✅ TimescaleDB connected successfully`
- `Starting migrations...`
- `Migrations completed`
- `Config source is YML, syncing config.yml to database...`
- `Config sync completed`

### 不應看到

- `偵測到 devices.yml：已載入 X 個裝置群組至記憶體（API 模式，不會同步至 DB）`

若看到上面這句，代表目前仍是 API 模式。

---

## 5. DB 驗證

登入 MySQL：

```sql
USE <database_name>;
```

### 檢查 ES 連線

```sql
SELECT id, name, host, port, is_default, deleted_at
FROM es_connections
ORDER BY id;
```

確認：

- YML 中的連線已建立或更新
- 不在 YML 的連線若未被引用，可能被軟刪除

### 檢查裝置群組

```sql
SELECT id, name, description
FROM device_groups
ORDER BY id;
```

確認：

- YML 中的群組已存在

### 檢查 Targets

```sql
SELECT id, subject, enable
FROM targets
ORDER BY id;
```

確認：

- 與 `config.yml` 一致

### 檢查 Indices

```sql
SELECT id, logname, pattern, device_group, period, unit, es_connection_id
FROM indices
ORDER BY id;
```

確認：

- 與 `config.yml` 一致

### 檢查 Devices

```sql
SELECT id, device_group, name, ha_group
FROM devices
ORDER BY device_group, id;
```

確認：

- `devices.yml` 中有的設備已同步
- 被列入 `disabled_devices` 的設備已移除

---

## 6. 功能測試

### 測試 1：新增設備

操作：

- 在 `devices.yml` 某群組新增一台設備
- 若 `yml_reload.devices_enabled: true`，等待 reload interval；否則重啟服務

確認：

- `devices` 表出現新設備

### 測試 2：移除設備

操作：

- 在 `devices.yml` 的 `disabled_devices` 中加入一台設備
- 若 `yml_reload.devices_enabled: true`，等待 reload interval；否則重啟服務

確認：

- `devices` 表中該設備已刪除，且後續不會被 auto-discovery 補回

### 測試 3：修改 ES 連線

操作：

- 修改 `config.yml` 的 `es_connections`
- 重啟服務

確認：

- `es_connections` 對應欄位已更新
- 相關 index / monitor 仍可正常工作

### 測試 4：修改 Target / Index

操作：

- 修改 `config.yml` 的 `targets` 或 `indices`
- 重啟服務

確認：

- `targets`、`indices`、`indices_targets` 與 YML 一致

---

## 7. 高風險注意事項

- YML 模式會在重啟時覆蓋 DB 中對應設定，不要與 API 模式混用
- `devices.yml` 若存在，請只維護它，不要只改 `config.yml` 的 `devices`
- `devices.yml` 未列出的設備不會自動刪除；要停用請用 `disabled_devices`
- 同名 `es_connection` 會直接更新既有記錄，不會新建第二筆
- `targets` / `indices` 在 YML 模式下是完整清單，YML 沒有的可能被移除

---

## 8. 測試完成回報建議

請測試回報以下內容：

- 使用的 `setting.yml` / `config.yml` / `devices.yml` 版本
- MySQL 驗證結果
- 是否成功完成新增 / 移除 / 修改測試
- 是否出現異常刪除、異常覆蓋、連線失效等問題
