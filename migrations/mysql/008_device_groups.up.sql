-- 設備群組表
CREATE TABLE IF NOT EXISTS `device_groups` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(50) NOT NULL UNIQUE COMMENT '群組名稱',
    `description` VARCHAR(255) COMMENT '群組描述',
    `created_at` INT UNSIGNED,
    `updated_at` INT UNSIGNED,
    `deleted_at` INT,

    INDEX `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='設備群組表';

-- 將現有的 device_group 遷移到 device_groups 表
INSERT INTO device_groups (name, created_at, updated_at)
SELECT DISTINCT device_group, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
FROM devices
WHERE device_group IS NOT NULL AND device_group != ''
ON DUPLICATE KEY UPDATE
    updated_at = VALUES(updated_at);
