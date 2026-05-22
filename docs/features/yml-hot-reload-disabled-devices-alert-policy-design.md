# YML Hot Reload / Disabled Devices / Alert Policy / Recovery 設計草案

## 1. 文件目的

本文件用於規劃以下功能，**不包含 ES 作為 history backend**：

1. `devices.yml` 在 `yml` 模式下可定時自動 reload，不需重啟
2. `devices.yml` 支援 `disabled_devices`，可將設備從監控體系中停用
3. 失聯告警支援：
   - threshold：連續失聯達指定次數才寄出第一次告警
   - silence / notification cap：單次事故通知次數達上限後暫時靜音
4. 設備恢復後發送 recovery 通知

**Deferred（延後處理）**

- 使用 Elasticsearch 作為 history backend

---

## 2. 目前系統現況

### 2.1 `devices.yml`

- 目前只在服務啟動時載入一次
- 在 `config_source: "yml"` 時，`devices.yml` 會覆蓋 `config.yml` 的 `devices`
- 目前不支援 hot reload

### 2.2 設備 auto-discovery

- 若 ES 查詢結果中出現 DB 尚未存在的設備，會自動寫入 `devices` table
- 已改為只檢查同群組是否存在，因此同名設備可存在不同群組

### 2.3 失聯告警

- 目前 `trulyRemoved` 只要非空，就會立即寄信
- 沒有累積次數、靜音、事故狀態、恢復通知等機制

### 2.4 HA 邏輯

- `ha_group != ""` 才視為 HA 成員
- 若設備省略 `ha_group`，反序列化後視為空字串，等同獨立設備
- 若只剩單一成員的 HA group，現行邏輯仍可運作，但語意上屬於退化 HA 結構

---

## 3. 設計原則

### 3.1 `devices` 與 `disabled_devices` 語意分離

- `devices`：代表要持續監控的設備
- `disabled_devices`：代表明確停用、不再監控的設備

`disabled_devices` 的語意不是「從 HA group 移除但繼續監控」，而是：

- 從 `devices` table 移除
- 不再參與 auto-discovery
- 不再參與 HA 判斷
- 不再發送失聯告警

### 3.2 HA 設備退役標準流程

若 HA 群組中有一台設備退役，且另一台要保留監控，標準流程為：

1. 先在 `devices` 中將保留那台設備改為獨立設備
2. 再把退役那台放入 `disabled_devices`

範例：

```yaml
devices:
  - device_group: "firewall"
    names:
      - name: "fw-01-standby"

disabled_devices:
  - device_group: "firewall"
    devices:
      - name: "fw-01-primary"
        reason: "retired"
```

### 3.3 告警採用「事故狀態機」而非單次判斷

告警不應只依賴單次偵測結果，而應以設備狀態累積與轉移為基礎：

- `healthy`
- `offline_pending`
- `offline_alerting`
- `offline_silenced`
- `recovered`

---

## 4. YAML Schema 設計

## 4.1 `devices.yml`

### `devices`

```yaml
devices:
  - device_group: "firewall"
    names:
      - name: "fw-01-primary"
        ha_group: "fw-cluster-01"
      - name: "fw-01-standby"
        ha_group: "fw-cluster-01"
      - name: "fw-02"
      - name: "waf-01"
```

規則：

- `ha_group` 有值：HA 成員
- `ha_group` 空字串或省略：獨立設備

### `disabled_devices`

建議格式：

```yaml
disabled_devices:
  - device_group: "firewall"
    devices:
      - name: "fw-01-primary"
        reason: "retired"
      - name: "fw-03"
        reason: "decommissioned"
```

規則：

- 匹配鍵以 `(device_group, name)` 為主
- `reason` 為選填，供維運註記
- `disabled_devices` 優先級高於 `devices`

### 可選未來擴充：`disabled_ha_groups`

```yaml
disabled_ha_groups:
  - device_group: "firewall"
    ha_groups:
      - "fw-cluster-01"
```

此項目前可不做，保留為未來擴充。

---

## 4.2 `config.yml`

### `indices[].alert_policy`

建議放在 `indices` 底下，而不是 `targets`。

原因：

- 偵測頻率由 `logname/index` 決定
- threshold 的語意與該 index 的排程週期直接相關

範例：

```yaml
targets:
  - subject: "Firewall 離線告警"
    receiver:
      - "ops@example.com"
    enable: true
    indices:
      - logname: "FW-01"
        index: "logstash-firewall-*"
        device_group: "firewall"
        field: "host.keyword"
        period: "minutes"
        unit: 1
        es_connection: "default-es"
        alert_policy:
          offline_threshold: 3
          max_notifications_per_incident: 2
          silence_minutes: 60
          send_recovery: true
```

欄位定義：

- `offline_threshold`
  連續失聯幾次後才發第一次告警
- `max_notifications_per_incident`
  同一事故最多發幾次通知
- `silence_minutes`
  達通知上限後靜音多久
- `send_recovery`
  該設備恢復時是否發 recovery 通知

---

## 4.3 `setting.yml`

新增 YML runtime reload 設定：

```yaml
yml_reload:
  devices_enabled: true
  devices_interval: "30s"
```

規則：

- 僅在 `config_source: "yml"` 時生效
- `devices_enabled: false` 時，仍維持啟動時載入一次的現行行為

---

## 5. Hot Reload 設計

## 5.1 目標

在 `yml` 模式下，修改 `devices.yml` 後不需重啟服務，即可：

- 更新 `devices` table
- 更新 disabled 清單
- 更新記憶體中的 `global.YMLConfig.Devices`

## 5.2 建議做法

採用 **polling + checksum**，不使用 file watcher。

理由：

- 跨平台穩定
- 在 server / container 環境可預期
- 實作簡單，出錯面較小

## 5.3 Reload 流程

1. 週期性讀取 `devices.yml`
2. 計算 checksum
3. 若 checksum 未變，不處理
4. 若 checksum 改變：
   - parse 新 YAML
   - schema 驗證
   - transaction 同步 DB
   - 成功後原子更新記憶體配置

## 5.4 同步順序

1. parse `devices`
2. parse `disabled_devices`
3. `syncDeviceGroups`
4. `syncDevices`
5. `syncDisabledDevices`
6. 更新記憶體中的 active device config / disabled set

---

## 6. `disabled_devices` 的預期行為

### 6.1 同步時

若某設備出現在 `disabled_devices`：

- 從 `devices` table 刪除
- 即使該設備也出現在 `devices` 區塊中，仍以 disabled 優先

### 6.2 偵測時

在 `Detect()` auto-discovery 之前，若設備命中 disabled set：

- 不可自動寫入 `devices` table
- 不可被視為新設備
- 不應觸發失聯告警

### 6.3 若未來從 `disabled_devices` 中移除

則允許：

- 下次 reload 後重新回到可監控狀態
- 若在 `devices` 有列出，會被同步回 DB
- 若不在 `devices`，則可由 auto-discovery 再次加入

---

## 7. 告警狀態機設計

## 7.1 現有問題

目前 `Detect()` 對 `trulyRemoved` 只要有值就立即寄信，缺少：

- 累積失聯次數
- 單次事故通知次數上限
- 靜音窗口
- 恢復通知

## 7.2 新增資料表：`device_alert_states`

建議新增 MySQL table：

- `logname`
- `device_group`
- `device_name`
- `current_state`
- `consecutive_offline_count`
- `notification_count`
- `last_alert_at`
- `silenced_until`
- `incident_open`
- `last_seen_online_at`
- `last_seen_offline_at`
- `last_recovery_notified_at`
- `created_at`
- `updated_at`

唯一鍵建議：

- `(logname, device_group, device_name)`

## 7.3 狀態定義

- `healthy`
  正常，沒有打開中的事故
- `offline_pending`
  已開始累積失聯次數，但未達 threshold
- `offline_alerting`
  已達 threshold，事故成立，且允許發信
- `offline_silenced`
  已達通知上限，目前靜音中
- `recovered`
  從已告警事故恢復

## 7.4 狀態轉移

### 正常 → 失聯

- `healthy` → `offline_pending`
  - `consecutive_offline_count = 1`

### 持續失聯但未達門檻

- `offline_pending`
  - count 遞增
  - 未達 `offline_threshold` 不寄信

### 達門檻

- `offline_pending` → `offline_alerting`
  - 寄第一封告警信
  - `notification_count += 1`

### 事故持續

- 若仍失聯，且未超過 `max_notifications_per_incident`
  - 可依策略持續寄送

### 超過通知上限

- `offline_alerting` → `offline_silenced`
  - 不再發信
  - 等待 `silence_minutes` 到期或恢復

### 恢復

- 若狀態為 `offline_alerting` 或 `offline_silenced`
  - 且本輪變為 `online` 或 `standby`
  - 依 `send_recovery` 決定是否寄 recovery
  - 狀態回 `healthy`
  - 清空事故相關計數

### 注意

- 若只曾進入 `offline_pending`，但未達告警門檻，恢復時**不寄 recovery**

---

## 8. HA 與 Recovery 的交互規則

告警與 recovery 都應該以 **HA 過濾後的結果** 為準。

### 失聯判斷

- 使用 `filterHAGroups()` 得到：
  - `trulyRemoved`
  - `standbyDevices`

### Recovery 判斷

若設備曾處於 `offline_alerting` / `offline_silenced`，而本輪變為：

- `intersection`
- 或 `standbyDevices`

都可視為已恢復可服務狀態

建議：

- `standby` 視為 recovery 的合法終態
- 郵件可註記恢復型態：
  - `online`
  - `standby`

---

## 9. Recovery 郵件規則

Recovery mail 建議條件：

- 事故曾成立（曾寄出至少一封 offline alert）
- 本輪狀態恢復
- `send_recovery = true`

主旨建議：

```text
[恢復通知] <subject> - <device_name>
```

內容建議包含：

- logname
- device_group
- device_name
- 原事故開始時間
- 本次恢復時間
- 最終恢復狀態（online / standby）
- 若屬 HA，附帶 ha_group

---

## 10. 與現有程式碼的整合點

### 10.1 `utils/utils.go`

- 新增 `devices.yml` reload service
- 支援 `disabled_devices` 解析

### 10.2 `services/config_sync.go`

- 既有 `syncDevices()` 保留
- 新增 `syncDisabledDevices()`

### 10.3 `services/detect.go`

- auto-discovery 前檢查 disabled set
- 導入 alert state machine
- offline / recovery mail 改由狀態機驅動

### 10.4 Migration

新增：

- `device_alert_states` table

可選：

- 若未來要追蹤 disabled 設備來源，可再加 `disabled_device_audits`

---

## 11. 實作順序

### Phase 1

- `devices.yml` schema 擴充
- `disabled_devices`
- hot reload
- disabled set 與 auto-discovery 整合

### Phase 2

- `alert_policy` schema
- `device_alert_states` migration
- threshold / silence state machine

### Phase 3

- recovery notification
- HA + recovery 整合驗證

### Deferred

- Elasticsearch 作為 history backend

---

## 12. 測試情境

### 12.1 devices.yml reload

- 修改 `devices.yml`
- 不重啟服務
- 確認 `devices` table 更新成功

### 12.2 disabled_devices

- 將既有設備加入 `disabled_devices`
- 確認設備從 `devices` table 消失
- 確認下一輪偵測不會自動補回

### 12.3 HA 退役標準流程

原始：

```yaml
devices:
  - device_group: "firewall"
    names:
      - name: "fw-01-primary"
        ha_group: "fw-cluster-01"
      - name: "fw-01-standby"
        ha_group: "fw-cluster-01"
```

調整後：

```yaml
devices:
  - device_group: "firewall"
    names:
      - name: "fw-01-standby"

disabled_devices:
  - device_group: "firewall"
    devices:
      - name: "fw-01-primary"
        reason: "retired"
```

確認：

- `fw-01-primary` 停用
- `fw-01-standby` 轉為獨立設備

### 12.4 告警 threshold

- 設 `offline_threshold: 3`
- 前 2 次失聯不寄信
- 第 3 次才寄第一封

### 12.5 notification cap / silence

- 設 `max_notifications_per_incident: 2`
- 達上限後不再重複通知

### 12.6 recovery

- 設備曾進入已告警事故
- 恢復後寄 recovery mail

---

## 13. 開放問題

1. `offline_threshold` 是否允許設為 `1`（代表維持現行立即告警）
2. 靜音是「直到恢復」還是「到期後仍離線則再寄」？
3. recovery mail 是否需要獨立主旨模板？
4. `standby` 是否一律視為 recovery，或需額外配置？

---

## 14. 建議結論

建議先做：

1. `devices.yml` hot reload
2. `disabled_devices`
3. `alert_policy` + threshold / silence
4. recovery notification

`ES history backend` 延後。

這樣可以先解決實際維運最痛的問題，且改動範圍仍可控制在：

- YML parsing / sync
- detect service
- MySQL migration
- mail notification flow

