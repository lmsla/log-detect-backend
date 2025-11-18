-- 修復 es_alert_history 表缺少的欄位和權限問題
-- 執行方式：以 postgres 超級用戶身份執行
-- psql -U postgres -d monitoring -f fix_es_alert_history_columns.sql

-- 添加缺少的欄位
ALTER TABLE es_alert_history
  ADD COLUMN IF NOT EXISTS cluster_name TEXT,
  ADD COLUMN IF NOT EXISTS threshold_value DOUBLE PRECISION,
  ADD COLUMN IF NOT EXISTS actual_value DOUBLE PRECISION,
  ADD COLUMN IF NOT EXISTS resolved_by TEXT,
  ADD COLUMN IF NOT EXISTS acknowledged_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS acknowledged_by TEXT,
  ADD COLUMN IF NOT EXISTS metadata JSONB;

-- 授予 logdetect 用戶完整權限
GRANT ALL PRIVILEGES ON TABLE es_alert_history TO logdetect;

-- 驗證欄位
SELECT column_name, data_type
FROM information_schema.columns
WHERE table_name = 'es_alert_history'
ORDER BY ordinal_position;

-- 驗證權限
SELECT grantee, privilege_type
FROM information_schema.table_privileges
WHERE table_name = 'es_alert_history' AND grantee = 'logdetect';

-- 預期應該有以下欄位：
-- time, monitor_id, alert_type, severity, message, status,
-- cluster_name, threshold_value, actual_value,
-- resolved_at, resolved_by, resolution_note,
-- acknowledged_at, acknowledged_by, metadata

-- 預期權限：
-- logdetect | INSERT
-- logdetect | SELECT
-- logdetect | UPDATE
-- logdetect | DELETE
-- logdetect | TRUNCATE
-- logdetect | REFERENCES
-- logdetect | TRIGGER
