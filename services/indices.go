package services

import (
	"fmt"
	"log-detect/entities"
	"log-detect/global"
	"log-detect/log"
	"log-detect/models"
)

// 新增 indices
func CreateIndices(indices entities.Index) models.Response {

	res := models.Response{}
	res.Success = false
	res.Body = entities.Index{}

	// result := global.Mysql.Where("name = ?", receiver.Name).First(&entities.Receiver{})
	// if result.RowsAffected > 0 {
	// 	res.Msg = "receiver Name already existed"
	// 	return res
	// }

	err := global.Mysql.Create(&indices).Error
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Create indices Fail: %s", err.Error()))
		res.Msg = "Create indices Fail"
		return res
		// } else {
		// Control_center_by_IndiceID(indices.ID)
	}

	res.Success = true
	res.Body = indices
	res.Msg = "Create indice Success"
	// global.Mysql.Where("name = ?", receiver.Name).First(&res.Body)

	return res
}

func UpdateIndices(indices entities.Index) models.Response {

	res := models.Response{}
	res.Success = false
	res.Body = []entities.Index{}

	entriesTable, err := GetEntryByIndiceID(indices.ID)
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Error when get data in CronList: %s", err.Error()))
		res.Msg = fmt.Sprintf("Error when get data in CronList, err: %s", err.Error())
		return res
	}

	// 找出目前 Entries 中 與CronList 對應的 entryID 並刪除 Entries 中的 entry
	entries := global.Crontab.Entries()
	for _, entry := range entries {
		for _, data := range entriesTable {
			if data.EntryID == int(entry.ID) {
				global.Crontab.Remove(entry.ID)
			}
		}
	}

	// 先取得 CronList 中 index_id 對應的 target_id
	targetdata, err := GetTargetIDByIndiceID(indices.ID)
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Get TargetID By IndiceID Fail: %s", err.Error()))
	}

	// 刪除 CronList 中對應的 entry ID
	err = global.Mysql.Where("index_id = ?", indices.ID).Delete(&entities.CronList{}).Error
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Error when deleting data in CronList table: %s", err.Error()))
		res.Msg = fmt.Sprintf("Error when deleting data in CronList table, err: %s", err.Error())
		return res
	}

	err = global.Mysql.Select("*").Where("id = ?", indices.ID).Updates(&indices).Error
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Update indices Fail: %s", err.Error()))
		res.Msg = "Update Fail"
		return res
	} else {
		Control_center_by_IndiceID(indices.ID, targetdata)
	}

	res.Success = true
	res.Body = indices
	res.Msg = "Update indice Success"
	// global.Mysql.Where("id = ?", receiver.ID).First(&res.Body)

	return res

}

func DeleteIndice(id int) models.Response {

	res := models.Response{}
	res.Success = false
	res.Body = nil

	result := global.Mysql.Where("id = ?", id).First(&entities.Index{})
	if result.RowsAffected == 0 {
		res.Msg = "indice ID does not exist"
		return res
	}

	entriesTable, lerr := GetEntryByIndiceID(id)
	if lerr != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Error when get data in CronList, err: %s", lerr.Error()))
		res.Msg = fmt.Sprintf("Error when get data in CronList, err: %s", lerr)
		return res
	}

	// 找出目前 Entries 中 與 CronList 對應的 entryID 並刪除 Entries 中的 entry
	entries := global.Crontab.Entries()
	for _, entry := range entries {
		for _, data := range entriesTable {
			if data.EntryID == int(entry.ID) {
				global.Crontab.Remove(entry.ID)
			}
		}
	}

	// 刪除 CronList 中對應的 entry ID
	entry_err := global.Mysql.Where("index_id = ?", id).Delete(&entities.CronList{}).Error
	if entry_err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Error when deleting data in CronList table, err: %s", entry_err.Error()))
		res.Msg = fmt.Sprintf("Error when deleting data in CronList table, err: %s", entry_err.Error())
		return res
	}

	err := global.Mysql.Where("id = ?", id).Delete(&entities.Index{}).Error
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Error when deleting device, err: %s", err.Error()))
		res.Msg = fmt.Sprintf("Error when deleting device, err: %s", err.Error())
		return res
	}

	res.Success = true
	res.Msg = "Delete indice Success"

	return res

}

func GetAllIndices() models.Response {

	res := models.Response{}
	res.Success = false
	res.Body = []models.Index{}

	err := global.Mysql.Find(&res.Body).Error
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Get All Indices err: %s", err.Error()))
		res.Msg = fmt.Sprintf("Get All Indices err: %s", err.Error())
		return res
	}

	res.Success = true
	res.Msg = "Get All Indices Success"
	return res
}

// 查單一 Index by TargetID
func GetIndicesByTargetID(TargetID int) ([]entities.Index, error) {

	target := entities.Target{}

	err := global.Mysql.Debug().Where("id = ?", TargetID).Preload("Indices").Find(&target).Error
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Get Indices By TargetID err: %s", err.Error()))
		return nil, err
	}
	return target.Indices, nil

}

//// 供程序處理 func

func GetEntryByIndiceID(IndicesID int) ([]entities.CronList, error) {
	cronlist := []entities.CronList{}
	// err := global.Mysql.Debug().Where("id = ?",scheduleID).Preload("Reports").Find(&schedule).Error

	err := global.Mysql.Debug().Where("index_id = ?", IndicesID).Find(&cronlist).Error
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Get Entry By IndiceID err: %s", err.Error()))
		return cronlist, err
	}

	return cronlist, nil

}

func GetLogname() models.Response {

	res := models.Response{}
	res.Success = false

	var lognames []entities.Logname
	err := global.Mysql.Model(&entities.Index{}).Distinct("logname").Pluck("logname", &lognames).Error
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("find device_group error: %s", err.Error()))
		return res
	}

	res.Body = lognames
	res.Success = true
	// fmt.Println(res)
	return res
}

func GetIndicesName() models.Response {

	res := models.Response{}
	res.Success = false

	var groups []entities.GroupName
	err := global.Mysql.Model(&entities.Device{}).Distinct("device_group").Pluck("device_group", &groups).Error
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("find device_group error: %s", err.Error()))
		return res
	}

	res.Body = groups
	res.Success = true
	return res

}

func GetIndicesData(logname string) models.Response {

	res := models.Response{}
	res.Success = false

	indices := entities.Index{}
	err := global.Mysql.Debug().Where("logname = ?", logname).Find(&indices).Error
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("find indices data error: %s", err.Error()))
		fmt.Println("find indices data error", err.Error())
		return res
	}
	res.Body = indices
	res.Success = true
	return res

}
