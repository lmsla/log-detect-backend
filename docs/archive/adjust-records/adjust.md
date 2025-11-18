## Path/Endpoint Mismatches

Test endpoint: docs use POST /api/v1/elasticsearch/monitors/{id}/test; OpenAPI has POST /api/v1/elasticsearch/monitors/test（無 id）.
Per‑monitor status: docs含 GET /status/{id}、/status/{id}/history、/status/{id}/trends；OpenAPI 只有整體 GET /status（缺 per‑monitor 端點）。
Alerts: docs含 GET /alerts、GET /alerts/{id}、POST /alerts/{id}/resolve、PUT /alerts/{id}/acknowledge；OpenAPI 未定義。
Schema/Field Inconsistencies

receivers 型別: OpenAPI 定義為 string（JSON 字串），實務/文件應為 array[string]。
interval 命名: MySQL DDL 用 interval_seconds；OpenAPI 用 interval。需統一名稱與單位（秒）。
Response Contract

ES 端點多回 { success, msg, body } 封裝；既有模組多回直接資料。建議統一（維持封裝就定義 Envelope<T> 並一致使用，或改為直接資料）。
Missing Query Parameters

時間範圍: GET /status/{id}、/history 應支援 start, end（或 hours）。
分頁與過濾: GET /monitors、/status、/alerts 應支援 page, page_size, q（模糊查詢）, status[], severity[] 等。
Units/Formats Clarification

response_time: 明確標註「毫秒」。
*_usage（cpu/memory/disk）: 明確標註「百分比」。
last_update_time/時間欄位: 建議 format: date-time（ISO 8601）。
Permissions/Tags（可選強化）

在路徑加 x-permissions（例如 ['elasticsearch:read']、['elasticsearch:update']），便於前端權限控制。
用 vendor x-module: es 標註模組，前端可自動歸類到 ES 模組。

## 建議最小修正清單
統一路徑：
改為 POST /elasticsearch/monitors/{id}/test（或另訂 POST /monitors/test-connection，二擇一）。
補 GET /elasticsearch/status/{id}、GET /elasticsearch/status/{id}/history（含 bucket）、（可選）GET /elasticsearch/status/{id}/trends。
補 GET /elasticsearch/alerts、GET /elasticsearch/alerts/{id}、POST /elasticsearch/alerts/{id}/resolve、PUT /elasticsearch/alerts/{id}/acknowledge。
調整 Schema：
ElasticsearchMonitor.receivers → array[string]
interval 與 DDL 一致（建議沿用 interval 秒）。
參數與格式：
為上述查詢端點加 start/end 或 hours、分頁參數與必要過濾。
為時間/比率/單位加上清楚描述與 format。