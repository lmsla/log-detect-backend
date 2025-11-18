-- Migration: 003_alter_elasticsearch_monitors_add_es_connection
-- Description: 為 elasticsearch_monitors 表新增 ES 連線關聯欄位（可選）
-- Date: 2025-11-18

-- 新增 es_connection_id 欄位
ALTER TABLE `elasticsearch_monitors`
ADD COLUMN `es_connection_id` INT UNSIGNED DEFAULT NULL COMMENT 'ES 連線 ID（NULL=使用自己的 host/port，有值=複用指定的連線配置）'
AFTER `description`;

-- 建立外鍵約束（使用 SET NULL 策略，因為刪除連線時監控器應保留）
ALTER TABLE `elasticsearch_monitors`
ADD CONSTRAINT `fk_es_monitors_connection`
  FOREIGN KEY (`es_connection_id`)
  REFERENCES `es_connections`(`id`)
  ON UPDATE CASCADE
  ON DELETE SET NULL;

-- 建立索引
CREATE INDEX `idx_es_connection_id` ON `elasticsearch_monitors`(`es_connection_id`);
