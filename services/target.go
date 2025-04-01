package services

import (
	"fmt"
	"log-detect/entities"
	"log-detect/global"
	"log-detect/models"
	"log-detect/log"
)

func GetAllTargets() models.Response {

	res := models.Response{}
	res.Success = false
	res.Body = []entities.Target{}

	err := global.Mysql.Preload("Indices").Find(&res.Body).Error
	if err != nil {
		res.Msg = err.Error()
		return res
	}

	res.Success = true
	res.Msg = "Get All Targets Success"
	return res
}

// 新增 target
func CreateTarget(target entities.Target) models.Response {

	res := models.Response{}
	res.Success = false
	res.Body = []entities.Target{}

	result := global.Mysql.Where("subject = ?", target.Subject).First(&entities.Target{})
	if result.RowsAffected > 0 {
		res.Msg = "subject already existed"
		return res
	}

	err := global.Mysql.Create(&target).Error
	if err != nil {

		res.Msg = fmt.Sprintf("Create Target Fail: %s", err.Error())
		log.Logrecord_no_rotate("ERROR", res.Msg)

		return res
	} else {
		fmt.Println("target.ID", target.ID)
		Control_center_by_TargetID(target.ID)
	}

	res.Success = true
	res.Body = target
	res.Msg = "Create Target Success"
	// global.Mysql.Where("name = ?", receiver.Name).First(&res.Body)
	return res
}

func UpdateTarget(target entities.Target) models.Response {

	res := models.Response{}
	res.Success = false
	res.Body = []entities.Target{}

	result := global.Mysql.Debug().Where("id != ? AND subject = ?", target.ID, target.Subject).First(&entities.Target{})
	if result.RowsAffected > 0 {
		res.Msg = "Subject already existed"
		return res
	}
	// //先刪除 indices 中相對應的資料
	// err := global.Mysql.Where("target_id = ?", target.ID).Delete(&entities.Index{}).Error
	// if err != nil {
	// 	res.Msg = fmt.Sprintf("Error when deleting related elements, err: %s", err)
	// 	return res
	// }

	//先刪除 IndicesTargets 中相應的資料
	target_err := global.Mysql.Where("target_id = ?", target.ID).Delete(&entities.IndicesTargets{}).Error
	if target_err != nil {
		res.Msg = fmt.Sprintf("Error when deleting data in IndicesTargets, err: %s", target_err)
		log.Logrecord_no_rotate("ERROR", res.Msg)
		return res
	}

	// cronlist table
	entriesTable, err := GetEntryByTargetID(target.ID)
	if err != nil {
		res.Msg = fmt.Sprintf("Error when get data in CronList, err: %s", err)
		log.Logrecord_no_rotate("ERROR", res.Msg)
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

	// 刪除 CronList 中對應的 entry ID
	err = global.Mysql.Where("target_id = ?", target.ID).Delete(&entities.CronList{}).Error
	if err != nil {
		res.Msg = fmt.Sprintf("Error when deleting data in CronList table, err: %s", err)
		log.Logrecord_no_rotate("ERROR", res.Msg)
		return res
	}


	err = global.Mysql.Debug().Select("*").Where("id = ?", target.ID).Updates(&target).Error
	if err != nil {
		res.Msg = "Update Fail"
		return res
	} else {
		Control_center_by_TargetID(target.ID)
	}

	res.Success = true
	res.Msg = "Update Success"
	global.Mysql.Where("id = ?", target.ID).First(&res.Body)

	return res

}

func DeleteTarget(id int) models.Response {

	res := models.Response{}
	res.Success = false
	res.Body = nil

	result := global.Mysql.Where("id = ?", id).First(&entities.Target{})
	if result.RowsAffected == 0 {
		res.Msg = "Target ID does not exist"
		return res
	}

	//先刪除 indices 中相對應的資料
	err := global.Mysql.Where("target_id = ?", id).Delete(&entities.Index{}).Error
	if err != nil {
		res.Msg = fmt.Sprintf("Error when deleting related indices, err: %s", err)
		log.Logrecord_no_rotate("ERROR", res.Msg)
		return res
	}

	entriesTable, err := GetEntryByTargetID(id)
	if err != nil {
		res.Msg = fmt.Sprintf("Error when get data in CronList, err: %s", err)
		log.Logrecord_no_rotate("ERROR", res.Msg)
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

	// 刪除 CronList 中對應的 entry ID
	err = global.Mysql.Where("target_id = ?", id).Delete(&entities.CronList{}).Error
	if err != nil {
		res.Msg = fmt.Sprintf("Error when deleting data in CronList table, err: %s", err)
		log.Logrecord_no_rotate("ERROR", res.Msg)
		return res
	}

	err = global.Mysql.Where("id = ?", id).Delete(&entities.Target{}).Error
	if err != nil {
		res.Msg = fmt.Sprintf("Error when deleting receiver, err: %s", err)
		log.Logrecord_no_rotate("ERROR", res.Msg)
		return res
	}

	res.Success = true
	res.Msg = "Delete Success"

	return res

}

//// 供程序處理 func

// 查單一Target
func GetTargetIDByIndiceID(indiceID int) (entities.IndicesTargets, error) {

	indicestarget := entities.IndicesTargets{}

	err := global.Mysql.Debug().Where("index_id = ?", indiceID).First(&indicestarget).Error
	if err != nil {
		fmt.Println("find targets id error (GetTargetIDByIndiceID)", err.Error())
	}
	// TargetID = indicestarget.TargetID
	return indicestarget, nil

}

// 查單一Target
func GetTargetByID(targetID int) (entities.Target, error) {

	var target entities.Target
	target.ID = targetID
	err := global.Mysql.Preload("Indices").First(&target).Error
	if err != nil {
		return target, err
	}

	return target, nil

}

func GetAllTargetsData() ([]entities.Target, error) {

	targets := []entities.Target{}
	err := global.Mysql.Preload("Indices").Find(&targets).Error
	if err != nil {
		fmt.Println("find targets data error", err.Error())
		return nil, err
	}
	return targets, nil
}

func GetEntryByTargetID(TargetID int) ([]entities.CronList, error) {
	cronlist := []entities.CronList{}
	// err := global.Mysql.Debug().Where("id = ?",scheduleID).Preload("Reports").Find(&schedule).Error

	err := global.Mysql.Debug().Where("target_id = ?", TargetID).Find(&cronlist).Error
	if err != nil {
		return cronlist, err
	}
	return cronlist, nil

}
