-- Migration: 002_alter_indices_add_es_connection
-- Description: 為 indices 表新增 ES 連線關聯欄位
-- Date: 2025-11-18

-- 新增 es_connection_id 欄位
ALTER TABLE `indices`
ADD COLUMN `es_connection_id` INT UNSIGNED DEFAULT NULL COMMENT 'ES 連線 ID（外鍵到 es_connections，NULL 表示使用預設連線）'
AFTER `field`;

-- 建立外鍵約束
ALTER TABLE `indices`
ADD CONSTRAINT `fk_indices_es_connection`
  FOREIGN KEY (`es_connection_id`)
  REFERENCES `es_connections`(`id`)
  ON UPDATE CASCADE
  ON DELETE RESTRICT;

-- 建立索引以提升查詢效能
CREATE INDEX `idx_es_connection_id` ON `indices`(`es_connection_id`);
