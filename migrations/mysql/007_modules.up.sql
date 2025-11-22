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
