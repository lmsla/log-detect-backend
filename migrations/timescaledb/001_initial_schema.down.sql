-- TimescaleDB Rollback Schema for log-detect-backend
-- Version: 001
-- Created: 2025-11-22
--
-- 此檔案用於回滾 001_initial_schema.up.sql 的變更
-- 注意：執行此檔案會刪除所有時序數據！

-- 刪除 hypertable（會連帶刪除所有數據和索引）
DROP TABLE IF EXISTS device_metrics CASCADE;
DROP TABLE IF EXISTS es_metrics CASCADE;
DROP TABLE IF EXISTS es_alert_history CASCADE;
