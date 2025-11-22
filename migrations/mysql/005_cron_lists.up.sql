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
