-- Initial schema for log-detect-backend
-- Version: 001
-- Created: 2025-11-22

-- =============================================
-- 用戶與權限
-- =============================================

CREATE TABLE IF NOT EXISTS `roles` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(50) NOT NULL UNIQUE,
    `description` VARCHAR(200),
    `created_at` INT UNSIGNED,
    `updated_at` INT UNSIGNED,
    `deleted_at` INT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `permissions` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(100) NOT NULL UNIQUE,
    `resource` VARCHAR(100) NOT NULL,
    `action` VARCHAR(50) NOT NULL,
    `description` VARCHAR(200),
    `created_at` INT UNSIGNED,
    `updated_at` INT UNSIGNED,
    `deleted_at` INT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `role_permissions` (
    `role_id` INT UNSIGNED NOT NULL,
    `permission_id` INT UNSIGNED NOT NULL,
    PRIMARY KEY (`role_id`, `permission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `users` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `username` VARCHAR(50) NOT NULL UNIQUE,
    `email` VARCHAR(100) NOT NULL UNIQUE,
    `password` VARCHAR(255) NOT NULL,
    `role_id` INT UNSIGNED NOT NULL,
    `is_active` TINYINT(1) DEFAULT 1,
    `created_at` INT UNSIGNED,
    `updated_at` INT UNSIGNED,
    `deleted_at` INT,
    INDEX `idx_users_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- =============================================
-- ES 連線配置
-- =============================================

CREATE TABLE IF NOT EXISTS `es_connections` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(100) NOT NULL UNIQUE,
    `host` VARCHAR(255) NOT NULL,
    `port` INT NOT NULL DEFAULT 9200,
    `username` VARCHAR(100),
    `password` VARCHAR(255),
    `enable_auth` TINYINT(1) DEFAULT 0,
    `use_tls` TINYINT(1) DEFAULT 1,
    `is_default` TINYINT(1) DEFAULT 0,
    `description` TEXT,
    `created_at` INT UNSIGNED,
    `updated_at` INT UNSIGNED,
    `deleted_at` INT,
    INDEX `idx_es_connections_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- =============================================
-- 裝置監控相關
-- =============================================

CREATE TABLE IF NOT EXISTS `devices` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `device_group` VARCHAR(50),
    `name` VARCHAR(50),
    `created_at` INT UNSIGNED,
    `updated_at` INT UNSIGNED,
    `deleted_at` INT,
    INDEX `idx_devices_device_group` (`device_group`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `receivers` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `name` JSON,
    `created_at` INT UNSIGNED,
    `updated_at` INT UNSIGNED,
    `deleted_at` INT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `targets` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `subject` VARCHAR(50),
    `to` JSON,
    `enable` TINYINT(1) DEFAULT 0,
    `created_at` INT UNSIGNED,
    `updated_at` INT UNSIGNED,
    `deleted_at` INT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `indices` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `pattern` VARCHAR(50),
    `device_group` VARCHAR(50),
    `logname` VARCHAR(50),
    `period` VARCHAR(50),
    `unit` INT,
    `field` VARCHAR(50),
    `es_connection_id` INT,
    `created_at` INT UNSIGNED,
    `updated_at` INT UNSIGNED,
    `deleted_at` INT,
    INDEX `idx_indices_es_connection_id` (`es_connection_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `indices_targets` (
    `target_id` INT NOT NULL,
    `index_id` INT NOT NULL,
    PRIMARY KEY (`target_id`, `index_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- =============================================
-- 歷史記錄
-- =============================================

CREATE TABLE IF NOT EXISTS `histories` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `logname` VARCHAR(50),
    `device_group` VARCHAR(50),
    `name` VARCHAR(100),
    `target_id` INT,
    `index_id` INT,
    `status` VARCHAR(20),
    `lost` VARCHAR(10),
    `lost_num` INT DEFAULT 0,
    `date` VARCHAR(10),
    `time` VARCHAR(8),
    `date_time` VARCHAR(19),
    `timestamp` BIGINT,
    `period` VARCHAR(20),
    `unit` INT DEFAULT 1,
    `response_time` BIGINT DEFAULT 0,
    `data_count` BIGINT DEFAULT 0,
    `error_msg` TEXT,
    `error_code` VARCHAR(50),
    `metadata` JSON,
    `created_at` INT UNSIGNED,
    `updated_at` INT UNSIGNED,
    `deleted_at` INT,
    INDEX `idx_histories_logname` (`logname`),
    INDEX `idx_histories_device_group` (`device_group`),
    INDEX `idx_histories_name` (`name`),
    INDEX `idx_histories_date` (`date`),
    INDEX `idx_histories_time` (`time`),
    INDEX `idx_histories_date_time` (`date_time`),
    INDEX `idx_histories_timestamp` (`timestamp`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `history_archives` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `logname` VARCHAR(50),
    `device_group` VARCHAR(50),
    `name` VARCHAR(100),
    `target_id` INT,
    `index_id` INT,
    `status` VARCHAR(20),
    `lost` VARCHAR(10),
    `lost_num` INT DEFAULT 0,
    `date` VARCHAR(10),
    `time` VARCHAR(8),
    `date_time` VARCHAR(19),
    `timestamp` BIGINT,
    `period` VARCHAR(20),
    `unit` INT DEFAULT 1,
    `response_time` BIGINT DEFAULT 0,
    `data_count` BIGINT DEFAULT 0,
    `error_msg` TEXT,
    `error_code` VARCHAR(50),
    `metadata` JSON,
    `created_at` INT UNSIGNED,
    `updated_at` INT UNSIGNED,
    `deleted_at` INT,
    INDEX `idx_history_archives_logname` (`logname`),
    INDEX `idx_history_archives_device_group` (`device_group`),
    INDEX `idx_history_archives_name` (`name`),
    INDEX `idx_history_archives_date` (`date`),
    INDEX `idx_history_archives_timestamp` (`timestamp`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `history_daily_stats` (
    `date` VARCHAR(10) NOT NULL,
    `logname` VARCHAR(50) NOT NULL,
    `device_group` VARCHAR(50) NOT NULL,
    `total_checks` BIGINT DEFAULT 0,
    `online_count` BIGINT DEFAULT 0,
    `offline_count` BIGINT DEFAULT 0,
    `warning_count` BIGINT DEFAULT 0,
    `error_count` BIGINT DEFAULT 0,
    `uptime_rate` DECIMAL(5,2) DEFAULT 0.00,
    `avg_response_time` DECIMAL(10,2) DEFAULT 0.00,
    `created_at` INT UNSIGNED,
    `updated_at` INT UNSIGNED,
    `deleted_at` INT,
    PRIMARY KEY (`date`, `logname`, `device_group`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `mail_histories` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `date` VARCHAR(50),
    `time` VARCHAR(50),
    `logname` VARCHAR(50),
    `sended` TINYINT(1),
    `created_at` INT UNSIGNED,
    `updated_at` INT UNSIGNED,
    `deleted_at` INT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `alert_histories` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `logname` VARCHAR(50),
    `device_group` VARCHAR(50),
    `device_name` VARCHAR(100),
    `alert_type` VARCHAR(20),
    `severity` VARCHAR(10),
    `message` TEXT,
    `status` VARCHAR(20),
    `resolved_at` BIGINT,
    `resolved_by` VARCHAR(100),
    `created_at` INT UNSIGNED,
    `updated_at` INT UNSIGNED,
    `deleted_at` INT,
    INDEX `idx_alert_histories_logname` (`logname`),
    INDEX `idx_alert_histories_device_group` (`device_group`),
    INDEX `idx_alert_histories_device_name` (`device_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- =============================================
-- 排程管理
-- =============================================

CREATE TABLE IF NOT EXISTS `cron_lists` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `entry_id` INT,
    `target_id` INT,
    `index_id` INT,
    `created_at` INT UNSIGNED,
    `updated_at` INT UNSIGNED,
    `deleted_at` INT,
    INDEX `idx_cron_lists_entry_id` (`entry_id`),
    INDEX `idx_cron_lists_target_id` (`target_id`),
    INDEX `idx_cron_lists_index_id` (`index_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- =============================================
-- ES 監控配置
-- =============================================

CREATE TABLE IF NOT EXISTS `elasticsearch_monitors` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(100) NOT NULL,
    `host` VARCHAR(255) NOT NULL,
    `port` INT NOT NULL DEFAULT 9200,
    `username` VARCHAR(100),
    `password` VARCHAR(255),
    `enable_auth` TINYINT(1) DEFAULT 0,
    `check_type` VARCHAR(100) DEFAULT 'health,performance',
    `interval` INT NOT NULL DEFAULT 60,
    `enable_monitor` TINYINT(1) DEFAULT 1,
    `receivers` JSON,
    `subject` VARCHAR(255),
    `description` TEXT,
    `es_connection_id` INT,
    `cpu_usage_high` DECIMAL(5,2),
    `cpu_usage_critical` DECIMAL(5,2),
    `memory_usage_high` DECIMAL(5,2),
    `memory_usage_critical` DECIMAL(5,2),
    `disk_usage_high` DECIMAL(5,2),
    `disk_usage_critical` DECIMAL(5,2),
    `response_time_high` BIGINT,
    `response_time_critical` BIGINT,
    `unassigned_shards_threshold` INT,
    `alert_threshold` JSON,
    `alert_dedupe_window` INT DEFAULT 300,
    `created_at` INT UNSIGNED,
    `updated_at` INT UNSIGNED,
    `deleted_at` INT,
    INDEX `idx_elasticsearch_monitors_es_connection_id` (`es_connection_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- =============================================
-- 系統模組
-- =============================================

CREATE TABLE IF NOT EXISTS `modules` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(100),
    `url` VARCHAR(255),
    `license_key` VARCHAR(255),
    `disabled` TINYINT(1) DEFAULT 0,
    `created_at` INT UNSIGNED,
    `updated_at` INT UNSIGNED,
    `deleted_at` INT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
