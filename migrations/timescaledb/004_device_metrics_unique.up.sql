-- 為 device_metrics 新增唯一約束，防止同一設備在同一時間點重複寫入
-- TimescaleDB hypertable 的 unique constraint 必須包含分區鍵（time 欄位）
-- 唯一鍵：(time, device_id, logname) — 相同裝置、相同 logname、相同 timestamp 只允許一筆
CREATE UNIQUE INDEX IF NOT EXISTS uniq_device_metrics_time_device_logname
    ON device_metrics (time, device_id, logname);
