-- Rollback initial schema
-- Version: 001

DROP TABLE IF EXISTS `modules`;
DROP TABLE IF EXISTS `elasticsearch_monitors`;
DROP TABLE IF EXISTS `cron_lists`;
DROP TABLE IF EXISTS `alert_histories`;
DROP TABLE IF EXISTS `mail_histories`;
DROP TABLE IF EXISTS `history_daily_stats`;
DROP TABLE IF EXISTS `history_archives`;
DROP TABLE IF EXISTS `histories`;
DROP TABLE IF EXISTS `indices_targets`;
DROP TABLE IF EXISTS `indices`;
DROP TABLE IF EXISTS `targets`;
DROP TABLE IF EXISTS `receivers`;
DROP TABLE IF EXISTS `devices`;
DROP TABLE IF EXISTS `es_connections`;
DROP TABLE IF EXISTS `users`;
DROP TABLE IF EXISTS `role_permissions`;
DROP TABLE IF EXISTS `permissions`;
DROP TABLE IF EXISTS `roles`;
