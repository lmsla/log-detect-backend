package services

import (
	"fmt"
	"log-detect/entities"
	"log-detect/global"
	"log-detect/log"
	"log-detect/models"
)

// 新增device
func CreateDevice(device []entities.Device) models.Response {

	res := models.Response{}
	res.Success = false
	res.Body = []entities.Device{}

	// result := global.Mysql.Where("name = ?", receiver.Name).First(&entities.Receiver{})
	// if result.RowsAffected > 0 {
	// 	res.Msg = "receiver Name already existed"
	// 	return res
	// }

	err := global.Mysql.Create(&device).Error
	if err != nil {
		res.Msg = "Create Fail"
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Create Device Fail error: %s", err.Error()))
		return res
	}
	res.Success = true
	res.Body = device
	res.Msg = "Create Success"
	// global.Mysql.Where("name = ?", receiver.Name).First(&res.Body)

	return res
}

func UpdateDevice(device entities.Device) models.Response {

	res := models.Response{}
	res.Success = false
	res.Body = []entities.Device{}

	err := global.Mysql.Select("*").Where("id = ?", device.ID).Updates(&device).Error
	if err != nil {
		res.Msg = "Update Fail"
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Update Device Fail error: %s", err.Error()))
		return res
	}

	res.Success = true
	res.Body = device
	res.Msg = "Update Success"
	// global.Mysql.Where("id = ?", receiver.ID).First(&res.Body)

	return res
}

func DeleteDevice(id int) models.Response {

	res := models.Response{}
	res.Success = false
	res.Body = nil

	result := global.Mysql.Where("id = ?", id).First(&entities.Device{})
	if result.RowsAffected == 0 {
		res.Msg = "device ID does not exist"
		return res
	}

	err := global.Mysql.Where("id = ?", id).Delete(&entities.Device{}).Error
	if err != nil {
		res.Msg = fmt.Sprintf("Error when deleting device: %s", err.Error())
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Error when deleting device: %s", err.Error()))
		return res
	}

	res.Success = true
	res.Msg = "Delete Success"

	return res

}

func GetAllDevices() models.Response {

	res := models.Response{}
	res.Success = false
	res.Body = []entities.Device{}

	err := global.Mysql.Find(&res.Body).Error
	if err != nil {
		res.Msg = fmt.Sprintf("Error Get All Devices: %s", err.Error())
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Error Get All Devices: %s", err.Error()))
		return res
	}

	res.Success = true
	res.Msg = "Get All Devices Success"
	return res
}

func CountDevices() models.Response {
	var deviceGroupCounts []entities.Table_counts
	res := models.Response{}
	res.Success = false

	// 執行原生 SQL 查詢
	result := global.Mysql.Raw("SELECT device_group, COUNT(*) as devices_count FROM logdetect.devices GROUP BY device_group").Scan(&deviceGroupCounts)
	if result.Error != nil {
		// 處理錯誤
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("error querying device group counts: %s",result.Error.Error()))
	}
	res.Body = deviceGroupCounts
	res.Success = true
	// fmt.Println(deviceGroupCounts)
	return res
}

func GetDeviceGroup() models.Response {

	res := models.Response{}
	res.Success = false

	var groups []entities.GroupName
	err := global.Mysql.Model(&entities.Device{}).Distinct("device_group").Pluck("device_group", &groups).Error
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("error find device_group: %s",err.Error()))
		return res
	}

	res.Body = groups
	res.Success = true
	return res
}

//// 供程序處理 func

func GetDevicesDataByGroupName(device_group string) ([]entities.Device, error) {

	devices := []entities.Device{}
	err := global.Mysql.Debug().Where("device_group = ?", device_group).Find(&devices).Error
	if err != nil {
		fmt.Println("find devices data error", err.Error())
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("error find devices data: %s",err.Error()))
		return nil, err
	}

	return devices, nil
}
