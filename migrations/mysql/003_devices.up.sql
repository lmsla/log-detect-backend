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
