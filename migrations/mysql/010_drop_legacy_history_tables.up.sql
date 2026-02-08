-- 移除已遷移至 TimescaleDB 的 legacy 歷史表
-- 時序資料已改由 TimescaleDB device_metrics 表處理
DROP TABLE IF EXISTS `histories`;
DROP TABLE IF EXISTS `history_archives`;
DROP TABLE IF EXISTS `history_daily_stats`;
