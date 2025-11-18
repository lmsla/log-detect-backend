-- Rollback: 002_alter_indices_add_es_connection
-- Description: 移除 indices 表的 ES 連線關聯欄位
-- Date: 2025-11-18

-- 移除外鍵約束
ALTER TABLE `indices`
DROP FOREIGN KEY `fk_indices_es_connection`;

-- 移除索引
DROP INDEX `idx_es_connection_id` ON `indices`;

-- 移除欄位
ALTER TABLE `indices`
DROP COLUMN `es_connection_id`;
