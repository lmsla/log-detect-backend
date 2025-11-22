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
