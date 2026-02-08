-- 為 devices 表新增 ha_group 欄位，支援 HA 高可用裝置群組
-- ha_group 為空字串 = 獨立裝置；相同 ha_group 的裝置視為同一 HA 叢集
ALTER TABLE `devices` ADD COLUMN `ha_group` VARCHAR(50) NOT NULL DEFAULT '' AFTER `device_group`;
ALTER TABLE `devices` ADD INDEX `idx_devices_ha_group` (`ha_group`);
