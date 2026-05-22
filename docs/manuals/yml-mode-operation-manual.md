# Log Detect Backend YML 模式操作與設定手冊

## 1. 文件目的

本手冊提供測試人員在 **YML 模式** 下操作 `log-detect` 的完整流程。

適用情境：

- 不透過前端 API 維護監控設定
- 使用 `setting.yml`、`config.yml`、`devices.yml` 管理配置
- 使用啟動同步與 `devices.yml` reload 維護後端設定

不適用情境：

- `config_source: "api"` 的前端/API 管理模式
- 前端與 YML 混用的混合模式

---

## 2. YML 模式的核心觀念

YML 模式的運作方式如下：

1. `setting.yml`、`config.yml`、`devices.yml` 需與主程式放置在相同路徑下
2. 服務啟動時讀取 `setting.yml` & `config.yml`
3. 若存在 `devices.yml`，則用 `devices.yml` 的 `devices` 區段覆蓋 `config.yml` 的 `devices`
4. 若 `config_source: "yml"`，執行 `SyncConfigToDB()`
5. 將 YML 內容同步到 MySQL
6. 再從資料庫載入排程與監控設定開始執行

重點：

- **YML 是來源**
- **MySQL 是執行用資料**
- 修改 `config.yml` / `setting.yml` 後需重啟服務
- 修改 `devices.yml` 後，在 `yml_reload.devices_enabled: true` 時可自動 reload

---

## 3. 三個設定檔的角色

### 3.1 `setting.yml`

用途：

- 資料庫連線
- TimescaleDB 連線
- 功能開關 `features`
- 配置來源 `config_source`
- 郵件、CORS、SSO 等服務層設定

YML 模式最重要設定：

```yaml
config_source: "yml"
```

若要啟用 `devices.yml` 自動 reload，建議一併確認：

```yaml
yml_reload:
  devices_enabled: true
  devices_interval: "30s"
```

### 3.2 `config.yml`

用途：

- `es_connections`
- `targets`
- `indices`
- 可包含 `devices` 區段

這是主要的監控配置來源。

### 3.3 `devices.yml`

用途：

- 獨立維護 `devices` 清單
- 若檔案存在，會 **覆蓋 `config.yml` 的 `devices` 區段**

適合裝置清單經常變動，但不想頻繁修改 `config.yml` 的場景。

### 3.4 `disabled_devices`

若 `devices.yml` 內包含 `disabled_devices`，可將設備從監控體系中停用。

停用後的設備會：

- 從 `devices` table 移除
- 不再被 auto-discovery 自動補回
- 不再參與 HA 判斷
- 不再納入失聯告警

---

## 4. 啟用前建議

切換到 YML 模式前，建議先做以下動作：

- 備份 MySQL 資料庫
- 確認目前 DB 中是否已有前端/API 建立的 target、index、device、es_connection
- 確認是否接受「重啟後以 YML 覆蓋 DB」
- 確認是否要啟用 `devices.yml` hot reload

建議備份指令：

```bash
mysqldump -u <user> -p <database_name> > backup_before_yml_mode.sql
```

---

## 5. 設定步驟

### 5.1 修改 `setting.yml`

將配置來源改為 YML：

```yaml
config_source: "yml"
```

功能開關需依實際部署需求調整，並非 yml 模式就一定全部關閉，例如：

```yaml
features:
  timescaledb: false
  es_monitoring: false
  dashboard: false
  auth: false
```

### 5.2 編輯 `config.yml`

至少需要確認以下區段：

- `es_connections`
- `targets`
- `targets[].indices`

範例：

```yaml
es_connections:
  - name: "default-es"
    host: "10.99.1.213"
    port: 9200
    username: "elastic"
    password: "your_password"
    enable_auth: true
    use_tls: true
    is_default: true
    description: "主要 ES 叢集"

targets:
  - subject: "forti 防火牆離線告警"
    receiver:
      - "user@example.com"
    enable: true
    indices:
      - logname: "NCKU-forti"
        index: "logstash-forti-group"
        device_group: "forti"
        field: "host.keyword"
        period: "minutes"
        unit: 1
        es_connection: "default-es"
```

### 5.3 編輯 `devices.yml`

若專案根目錄存在 `devices.yml`，啟動時會以它為準同步裝置清單。

範例：

```yaml
devices:
  - device_group: "forti"
    names:
      - name: "fw-01-primary"
        ha_group: "fw-cluster-01"
      - name: "fw-01-standby"
        ha_group: "fw-cluster-01"
      - name: "fw-02"
        ha_group: ""

  - device_group: "waf"
    names:
      - name: "waf-01"
        ha_group: ""
      - name: "waf-02"
        ha_group: ""

disabled_devices:
  - device_group: "forti"
    devices:
      - name: "fw-legacy-01"
        reason: "retired"
```

注意：

- 同名設備可以存在不同群組
- `ha_group` 為空字串表示獨立裝置
- `devices.yml` 中未列出的設備不會因同步自動刪除
- 若要明確停用設備，請使用 `disabled_devices`

---

## 6. 啟動服務

在後端專案目錄執行：

```bash
cd 到主程式所在目錄

./log-detect

# 背景執行
nohup ./log-detect &
```

YML 模式下預期會看到類似訊息：

```text
已從 devices.yml 載入 2 個裝置群組
SQL Database 連線成功
✅ TimescaleDB connected successfully
Starting migrations...
Migrations completed
Config source is YML, syncing config.yml to database...
Config sync completed
✅ ES Connection Manager 初始化成功
```

若看到：

```text
偵測到 devices.yml：已載入 X 個裝置群組至記憶體（API 模式，不會同步至 DB）
```

代表你目前仍是 **API 模式**，不是 YML 模式。

---

## 7. 啟動後會同步哪些資料

### 7.1 `es_connections`

同步規則：

- YML 有的 `name`：建立或更新
- DB 有但 YML 沒有的：若沒有被任何 `indices` / `elasticsearch_monitors` 參照，會被軟刪除
- 若同名連線已存在，會直接更新原記錄，不會建立第二筆

測試注意：

- 若 `default-es` 已被多個 index / monitor 使用，YML 中同名 `default-es` 會直接覆蓋該連線內容
- 這些引用者會一起切換到更新後的連線設定

### 7.2 `device_groups`

同步規則：

- 只做 upsert
- 不會刪除 DB 中既有但 YML 未出現的群組

### 7.3 `targets`

同步規則：

- YML 有的：建立或更新
- DB 有但 YML 沒有的：刪除

重點：

- `targets` 在 YML 模式下視為完整清單
- 若 `config.yml` 沒有 `targets`，現有 target 可能被清掉

### 7.4 `indices`

同步規則：

- 每個 target 底下的 index 以 `logname` 為識別鍵
- YML 有的：建立或更新
- 某 target 底下 DB 有但 YML 沒有的：解除關聯，必要時刪除孤立 index

### 7.5 `devices`

同步規則：

- 只處理 YML 中有列出的 `device_group`
- YML 有的：建立或更新 `ha_group`
- YML 沒列出的群組：不處理
- DB 中既有但未列於 `devices.yml` 的設備，不會因同步自動刪除

重點：

- 若 `devices.yml` 存在，會以 `devices.yml` 的 `devices` 為準
- 若你只在 `devices.yml` 放 `forti`，那就只會同步 `forti` 群組中有列出的設備

### 7.6 `disabled_devices`

同步規則：

- 命中的 `(device_group, name)` 會從 `devices` table 移除
- auto-discovery 會跳過這些設備，不會重新補回
- 若未來從 `disabled_devices` 移除，設備可再次被偵測建立

---

## 8. 建議測試流程

### 測試 1：首次同步

1. 將 `config_source` 設為 `yml`
2. 準備好 `config.yml`
3. 準備好 `devices.yml`
4. 啟動 `go run main.go`
5. 檢查啟動日誌是否出現 `Config sync completed`
6. 查詢 MySQL 確認資料已同步

### 測試 2：新增一台裝置

1. 在 `devices.yml` 某個群組下新增一台設備
2. 若 `yml_reload.devices_enabled: true`，等待 reload interval；否則重啟服務
3. 查詢 `devices` 表，確認新設備已建立

### 測試 3：移除一台裝置

1. 在 `devices.yml` 的 `disabled_devices` 中加入目標設備
2. 若 `yml_reload.devices_enabled: true`，等待 reload interval；否則重啟服務
3. 查詢 `devices` 表，確認該設備已被移除

### 測試 4：修改 ES 連線

1. 修改 `config.yml` 的 `es_connections`
2. 重啟服務
3. 查詢 `es_connections` 表，確認欄位已更新
4. 驗證對應 index / monitor 仍可正常使用

### 測試 5：修改 target 或 index

1. 修改 `config.yml` 中的 `targets` / `indices`
2. 重啟服務
3. 查詢 `targets`、`indices`、`indices_targets`
4. 確認排程與偵測行為已更新

---

## 9. MySQL 驗證 SQL

先登入資料庫：

```sql
USE <database_name>;
```

### 9.1 檢查 ES 連線

```sql
SELECT id, name, host, port, is_default, deleted_at
FROM es_connections
ORDER BY id;
```

### 9.2 檢查裝置群組

```sql
SELECT id, name, description
FROM device_groups
ORDER BY id;
```

### 9.3 檢查 Targets

```sql
SELECT id, subject, enable
FROM targets
ORDER BY id;
```

### 9.4 檢查 Indices

```sql
SELECT id, logname, pattern, device_group, period, unit, es_connection_id
FROM indices
ORDER BY id;
```

### 9.5 檢查 Target 與 Index 關聯

```sql
SELECT target_id, index_id
FROM indices_targets
ORDER BY target_id, index_id;
```

### 9.6 檢查 Devices

```sql
SELECT id, device_group, name, ha_group
FROM devices
ORDER BY device_group, id;
```

### 9.7 檢查某個群組的設備

```sql
SELECT id, device_group, name, ha_group
FROM devices
WHERE device_group = 'forti'
ORDER BY id;
```

---

## 10. 常見注意事項

### 10.1 YML 模式與 API 模式不要混用

若先用前端/API 改 DB，再切回 `yml` 模式重啟，YML 會把對應資料蓋回去。

### 10.2 `devices.yml` 會覆蓋 `config.yml` 的 `devices`

若 `devices.yml` 存在，請以它為準維護裝置清單，不要只改 `config.yml` 的 `devices`。

### 10.3 同名 ES 連線會直接覆蓋

若 `config.yml` 中使用與 DB 相同的 `es_connections.name`，啟動後會直接更新既有那筆資料。

### 10.4 target / index 是全量覆蓋

YML 中沒有的 target/index，可能在同步時被刪除或解除關聯。

### 10.5 device 同步只作用於 YML 有列出的群組

若某群組未出現在 `devices.yml`，YML 模式不會主動整理該群組的設備。

### 10.6 `devices.yml` 可 hot reload，但 `config.yml` / `setting.yml` 仍需重啟

- `devices.yml` 在 `yml_reload.devices_enabled: true` 時可自動 reload
- `config.yml` 與 `setting.yml` 變更後，仍需重新啟動服務才會重新同步/載入

---

## 11. 建議的測試最小流程

1. 備份 DB
2. 確認 `setting.yml` 為 `config_source: "yml"`
3. 檢查 `config.yml`
4. 檢查 `devices.yml`
5. 執行 `go run main.go`
6. 確認出現 `Config sync completed`
7. 執行本文件第 9 節 SQL 驗證
8. 再開始前端或 API 層功能測試

---

## 12. 相關檔案

- `setting.yml`
- `config.yml`
- `devices.yml`
- `services/config_sync.go`
- `services/detect.go`
- `main.go`
