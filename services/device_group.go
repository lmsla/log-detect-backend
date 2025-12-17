package services

import (
	"fmt"
	"log-detect/entities"
	"log-detect/global"
	"log-detect/log"
	"log-detect/models"
)

// CreateDeviceGroup 創建設備群組
func CreateDeviceGroup(group entities.DeviceGroup) models.Response {
	res := models.Response{}
	res.Success = false

	// 檢查群組名稱是否已存在
	result := global.Mysql.Where("name = ?", group.Name).First(&entities.DeviceGroup{})
	if result.RowsAffected > 0 {
		res.Msg = "device group name already exists"
		return res
	}

	err := global.Mysql.Create(&group).Error
	if err != nil {
		res.Msg = "Create device group failed"
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Create DeviceGroup failed: %s", err.Error()))
		return res
	}

	res.Success = true
	res.Body = group
	res.Msg = "Create device group success"
	return res
}

// UpdateDeviceGroup 更新設備群組
func UpdateDeviceGroup(group entities.DeviceGroup) models.Response {
	res := models.Response{}
	res.Success = false

	// 檢查群組是否存在並獲取舊名稱
	var oldGroup entities.DeviceGroup
	result := global.Mysql.Where("id = ?", group.ID).First(&oldGroup)
	if result.RowsAffected == 0 {
		res.Msg = "device group does not exist"
		return res
	}

	// 使用事務確保數據一致性
	tx := global.Mysql.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新設備群組
	err := tx.Select("*").Where("id = ?", group.ID).Updates(&group).Error
	if err != nil {
		tx.Rollback()
		res.Msg = "Update device group failed"
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Update DeviceGroup failed: %s", err.Error()))
		return res
	}

	// 如果群組名稱有變更，需要同步更新 devices 和 indices 表
	if oldGroup.Name != group.Name {
		// 更新 devices 表中的 device_group
		err = tx.Model(&entities.Device{}).
			Where("device_group = ?", oldGroup.Name).
			Update("device_group", group.Name).Error
		if err != nil {
			tx.Rollback()
			res.Msg = "Failed to update devices table"
			log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Update devices device_group failed: %s", err.Error()))
			return res
		}

		// 更新 indices 表中的 device_group
		err = tx.Model(&entities.Index{}).
			Where("device_group = ?", oldGroup.Name).
			Update("device_group", group.Name).Error
		if err != nil {
			tx.Rollback()
			res.Msg = "Failed to update indices table"
			log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Update indices device_group failed: %s", err.Error()))
			return res
		}

		log.Logrecord_no_rotate("INFO", fmt.Sprintf("Device group name updated from '%s' to '%s', cascaded to devices and indices tables", oldGroup.Name, group.Name))
	}

	// 提交事務
	if err := tx.Commit().Error; err != nil {
		res.Msg = "Failed to commit transaction"
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Commit transaction failed: %s", err.Error()))
		return res
	}

	res.Success = true
	res.Body = group
	res.Msg = "Update device group success"
	return res
}

// DeleteDeviceGroup 刪除設備群組
func DeleteDeviceGroup(id int) models.Response {
	res := models.Response{}
	res.Success = false

	// 檢查群組是否存在
	var group entities.DeviceGroup
	result := global.Mysql.Where("id = ?", id).First(&group)
	if result.RowsAffected == 0 {
		res.Msg = "device group does not exist"
		return res
	}

	// 檢查群組下是否還有設備
	var deviceCount int64
	global.Mysql.Model(&entities.Device{}).Where("device_group = ?", group.Name).Count(&deviceCount)
	if deviceCount > 0 {
		res.Msg = fmt.Sprintf("cannot delete device group: %d devices still in this group", deviceCount)
		return res
	}

	err := global.Mysql.Where("id = ?", id).Delete(&entities.DeviceGroup{}).Error
	if err != nil {
		res.Msg = fmt.Sprintf("Delete device group failed: %s", err.Error())
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Delete DeviceGroup failed: %s", err.Error()))
		return res
	}

	res.Success = true
	res.Msg = "Delete device group success"
	return res
}

// GetAllDeviceGroups 獲取所有設備群組（含設備數量統計）
func GetAllDeviceGroups() models.Response {
	res := models.Response{}
	res.Success = false

	// 查詢所有群組並統計設備數量
	type DeviceGroupWithCount struct {
		entities.DeviceGroup
		DeviceCount int64 `json:"device_count"`
	}

	var groups []entities.DeviceGroup
	err := global.Mysql.Find(&groups).Error
	if err != nil {
		res.Msg = fmt.Sprintf("Get all device groups failed: %s", err.Error())
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Get all device groups failed: %s", err.Error()))
		return res
	}

	// 為每個群組統計設備數量
	var result []DeviceGroupWithCount
	for _, group := range groups {
		var count int64
		global.Mysql.Model(&entities.Device{}).Where("device_group = ?", group.Name).Count(&count)
		result = append(result, DeviceGroupWithCount{
			DeviceGroup: group,
			DeviceCount: count,
		})
	}

	res.Success = true
	res.Body = result
	res.Msg = "Get all device groups success"
	return res
}

// GetDeviceGroupByID 根據 ID 獲取設備群組
func GetDeviceGroupByID(id int) models.Response {
	res := models.Response{}
	res.Success = false

	var group entities.DeviceGroup
	err := global.Mysql.Where("id = ?", id).First(&group).Error
	if err != nil {
		res.Msg = fmt.Sprintf("Device group not found: %s", err.Error())
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Get DeviceGroup by ID failed: %s", err.Error()))
		return res
	}

	// 統計該群組的設備數量
	type DeviceGroupWithCount struct {
		entities.DeviceGroup
		DeviceCount int64 `json:"device_count"`
	}

	var count int64
	global.Mysql.Model(&entities.Device{}).Where("device_group = ?", group.Name).Count(&count)

	result := DeviceGroupWithCount{
		DeviceGroup: group,
		DeviceCount: count,
	}

	res.Success = true
	res.Body = result
	res.Msg = "Get device group success"
	return res
}

// MoveDevicesToGroup 批量遷移設備到指定群組（按設備 ID 列表）
func MoveDevicesToGroup(deviceIDs []int, targetGroupName string) models.Response {
	res := models.Response{}
	res.Success = false

	// 驗證目標群組是否存在
	var targetGroup entities.DeviceGroup
	result := global.Mysql.Where("name = ?", targetGroupName).First(&targetGroup)
	if result.RowsAffected == 0 {
		res.Msg = fmt.Sprintf("target device group '%s' does not exist", targetGroupName)
		return res
	}

	// 驗證設備 ID 列表不為空
	if len(deviceIDs) == 0 {
		res.Msg = "device ID list cannot be empty"
		return res
	}

	// 批量更新設備的群組
	updateResult := global.Mysql.Model(&entities.Device{}).
		Where("id IN ?", deviceIDs).
		Update("device_group", targetGroupName)

	if updateResult.Error != nil {
		res.Msg = fmt.Sprintf("Failed to move devices: %s", updateResult.Error.Error())
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("MoveDevicesToGroup failed: %s", updateResult.Error.Error()))
		return res
	}

	res.Success = true
	res.Body = map[string]interface{}{
		"moved_count":       updateResult.RowsAffected,
		"target_group_name": targetGroupName,
		"device_ids":        deviceIDs,
	}
	res.Msg = fmt.Sprintf("Successfully moved %d devices to group '%s'", updateResult.RowsAffected, targetGroupName)
	return res
}

// MoveGroupDevices 將整個群組的所有設備遷移到另一個群組
func MoveGroupDevices(sourceGroupName string, targetGroupName string) models.Response {
	res := models.Response{}
	res.Success = false

	// 驗證來源群組是否存在
	var sourceGroup entities.DeviceGroup
	result := global.Mysql.Where("name = ?", sourceGroupName).First(&sourceGroup)
	if result.RowsAffected == 0 {
		res.Msg = fmt.Sprintf("source device group '%s' does not exist", sourceGroupName)
		return res
	}

	// 驗證目標群組是否存在
	var targetGroup entities.DeviceGroup
	result = global.Mysql.Where("name = ?", targetGroupName).First(&targetGroup)
	if result.RowsAffected == 0 {
		res.Msg = fmt.Sprintf("target device group '%s' does not exist", targetGroupName)
		return res
	}

	// 防止來源和目標相同
	if sourceGroupName == targetGroupName {
		res.Msg = "source and target groups cannot be the same"
		return res
	}

	// 批量更新來源群組的所有設備
	updateResult := global.Mysql.Model(&entities.Device{}).
		Where("device_group = ?", sourceGroupName).
		Update("device_group", targetGroupName)

	if updateResult.Error != nil {
		res.Msg = fmt.Sprintf("Failed to move devices: %s", updateResult.Error.Error())
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("MoveGroupDevices failed: %s", updateResult.Error.Error()))
		return res
	}

	res.Success = true
	res.Body = map[string]interface{}{
		"moved_count":        updateResult.RowsAffected,
		"source_group_name":  sourceGroupName,
		"target_group_name":  targetGroupName,
	}
	res.Msg = fmt.Sprintf("Successfully moved %d devices from group '%s' to '%s'",
		updateResult.RowsAffected, sourceGroupName, targetGroupName)
	return res
}
