-- Rollback: 003_alter_elasticsearch_monitors_add_es_connection
-- Description: 移除 elasticsearch_monitors 表的 ES 連線關聯欄位
-- Date: 2025-11-18

-- 移除外鍵約束
ALTER TABLE `elasticsearch_monitors`
DROP FOREIGN KEY `fk_es_monitors_connection`;

-- 移除索引
DROP INDEX `idx_es_connection_id` ON `elasticsearch_monitors`;

-- 移除欄位
ALTER TABLE `elasticsearch_monitors`
DROP COLUMN `es_connection_id`;
