-- 回滾：移除 devices 表的 ha_group 欄位
ALTER TABLE `devices` DROP INDEX `idx_devices_ha_group`;
ALTER TABLE `devices` DROP COLUMN `ha_group`;
