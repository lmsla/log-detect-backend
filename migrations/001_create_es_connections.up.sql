-- Migration: 001_create_es_connections
-- Description: 建立 Elasticsearch 連線配置表
-- Date: 2025-11-18

-- 建立 es_connections 表
CREATE TABLE IF NOT EXISTS `es_connections` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(100) NOT NULL COMMENT '連線名稱，例如：生產環境-主集群',
  `host` VARCHAR(255) NOT NULL COMMENT 'ES 主機地址（例如：https://10.99.1.213）',
  `port` INT NOT NULL DEFAULT 9200 COMMENT 'ES 端口',
  `username` VARCHAR(100) DEFAULT NULL COMMENT '認證用戶名',
  `password` VARCHAR(255) DEFAULT NULL COMMENT '認證密碼',
  `enable_auth` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否啟用認證（0=否, 1=是）',
  `use_tls` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否使用 TLS（0=否, 1=是）',
  `is_default` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否為預設連線（0=否, 1=是）',
  `description` TEXT DEFAULT NULL COMMENT '連線描述',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '建立時間',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新時間',
  `deleted_at` TIMESTAMP NULL DEFAULT NULL COMMENT '刪除時間（軟刪除）',

  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_name` (`name`),
  INDEX `idx_is_default` (`is_default`),
  INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Elasticsearch 連線配置表';

-- 插入預設連線（從 setting.yml 遷移）
-- 注意：實際值需要根據 setting.yml 調整
-- INSERT INTO `es_connections` (`name`, `host`, `port`, `username`, `password`, `enable_auth`, `use_tls`, `is_default`, `description`)
-- VALUES ('預設連線', 'https://10.99.1.213', 9200, 'elastic', 'a12345678', 1, 1, 1, '從 setting.yml 自動遷移的預設 ES 連線');
