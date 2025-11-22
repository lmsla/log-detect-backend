-- ES 健康監控配置表
-- 連線配置統一使用 es_connections 表，不再維護獨立的 host/port/auth
CREATE TABLE IF NOT EXISTS `elasticsearch_monitors` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(100) NOT NULL COMMENT '監控名稱',
    `es_connection_id` INT UNSIGNED NOT NULL COMMENT 'ES 連線 ID（外鍵到 es_connections）',
    `check_type` VARCHAR(100) DEFAULT 'health,performance' COMMENT '檢查類型(逗號分隔)',
    `interval` INT NOT NULL DEFAULT 60 COMMENT '檢查間隔(秒,範圍:10-3600)',
    `enable_monitor` TINYINT(1) DEFAULT 1 COMMENT '是否啟用監控',
    `receivers` JSON COMMENT '告警收件人陣列',
    `subject` VARCHAR(255) COMMENT '告警主題',
    `description` TEXT COMMENT '監控描述',

    -- 告警閾值配置
    `cpu_usage_high` DECIMAL(5,2) COMMENT 'CPU使用率-高閾值(%)',
    `cpu_usage_critical` DECIMAL(5,2) COMMENT 'CPU使用率-危險閾值(%)',
    `memory_usage_high` DECIMAL(5,2) COMMENT '記憶體使用率-高閾值(%)',
    `memory_usage_critical` DECIMAL(5,2) COMMENT '記憶體使用率-危險閾值(%)',
    `disk_usage_high` DECIMAL(5,2) COMMENT '磁碟使用率-高閾值(%)',
    `disk_usage_critical` DECIMAL(5,2) COMMENT '磁碟使用率-危險閾值(%)',
    `response_time_high` BIGINT COMMENT '響應時間-高閾值(ms)',
    `response_time_critical` BIGINT COMMENT '響應時間-危險閾值(ms)',
    `unassigned_shards_threshold` INT COMMENT '未分配分片閾值',
    `alert_threshold` JSON COMMENT '告警閾值配置(JSON,高級選項)',
    `alert_dedupe_window` INT DEFAULT 300 COMMENT '告警去重時間窗口(秒)',

    -- 時間戳
    `created_at` INT UNSIGNED,
    `updated_at` INT UNSIGNED,
    `deleted_at` INT,

    -- 索引與外鍵
    INDEX `idx_es_monitors_connection_id` (`es_connection_id`),
    CONSTRAINT `fk_es_monitors_connection`
        FOREIGN KEY (`es_connection_id`) REFERENCES `es_connections`(`id`)
        ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Elasticsearch 健康監控配置表';
