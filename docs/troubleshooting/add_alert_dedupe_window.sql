-- 添加告警去重時間窗口配置欄位到 elasticsearch_monitors 表
-- 執行方式：連接到 MySQL config 數據庫
-- mysql -u monitor -p config < add_alert_dedupe_window.sql

-- 添加欄位
ALTER TABLE elasticsearch_monitors
  ADD COLUMN IF NOT EXISTS alert_dedupe_window INT DEFAULT 300 COMMENT '告警去重時間窗口(秒,預設300秒=5分鐘)';

-- 驗證欄位
DESCRIBE elasticsearch_monitors;

-- 說明：
-- alert_dedupe_window 控制告警去重的時間窗口
-- - 值為 300（預設）：同一監控器、相同類型和嚴重性的告警，在 5 分鐘內只發送一次
-- - 值為 60：1 分鐘內去重
-- - 值為 600：10 分鐘內去重
-- - 值為 0 或負數：使用預設值 300 秒

-- 使用範例：
-- 1. 創建監控時指定去重窗口為 10 分鐘（600 秒）
-- INSERT INTO elasticsearch_monitors (..., alert_dedupe_window) VALUES (..., 600);

-- 2. 更新現有監控的去重窗口為 3 分鐘（180 秒）
-- UPDATE elasticsearch_monitors SET alert_dedupe_window = 180 WHERE id = 1;

-- 3. 高頻告警場景（檢查間隔 30 秒）建議設置較短的去重窗口（60-120 秒）
-- UPDATE elasticsearch_monitors SET alert_dedupe_window = 60 WHERE interval < 60;

-- 4. 低頻告警場景（檢查間隔 5 分鐘以上）建議設置較長的去重窗口（600-1800 秒）
-- UPDATE elasticsearch_monitors SET alert_dedupe_window = 600 WHERE interval >= 300;
