package services

import (
	// "context"
	"fmt"
	"log-detect/entities"
	"log-detect/global"
	"log-detect/log"
	"time"

	"github.com/robfig/cron/v3"
	// "sync"
)

func LoadCrontab() {

	global.Crontab = cron.New()
	global.Crontab.Start()
}

func ExecuteCrontab(target_id int, index_id int, cronjob string, index string, field string, period string, unit int, receiver []string, subject string, logname string, device_group string) {

	EntryID, err := global.Crontab.AddFunc(cronjob, func() {
		execute_time := time.Now()
		Detect(execute_time, index, field, period, unit, receiver, subject, logname, device_group)
	})
	if err != nil {
		log.Logrecord_no_rotate("ERROR",fmt.Sprintf("Crontab AddFunc error: %s",err.Error()))
	}

	// fmt.Println("EntryID", EntryID)
	// 清除 cronlist table 中的紀錄
	// if err := global.Mysql.Delete(&entities.CronList{}).Error; err != nil {
	// 	log.Logrecord_no_rotate("ERROR","error when delete cronlist table")
	// }

	// 紀錄 cronjob 的 entry ID
	cronlist := entities.CronList{TargetID: target_id, IndexID: index_id, EntryID: int(EntryID)}
	result := global.Mysql.Create(&cronlist).Error
	if result != nil {
		log.Logrecord_no_rotate("ERROR",fmt.Sprintf("cronlist Create error: %s",result.Error()))
	}

	if err != nil {
		msg := fmt.Sprintf("標的名稱: %s ,日誌名稱: %s , 初始化失敗", subject, logname)
		log.Logrecord_no_rotate("排程 ", msg)
		log.Logrecord_no_rotate("ERROR", err.Error())
	} else {
		msg := fmt.Sprintf("標的名稱: %s ,日誌名稱: %s , 初始化成功", subject, logname)
		log.Logrecord_no_rotate("排程 ", msg)
		global.Crontab.Start()
	}
}

// 重啟服務時初始化所有 targets
func Control_center() {

	targets, err := GetAllTargetsData()
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("get targets data error: %s",err.Error()))
	}

	if err := global.Mysql.Exec("TRUNCATE TABLE cron_lists").Error; err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("error when truncate cronlist table: %s",err.Error()))
	}

	var cronjob string
	for _, target := range targets {
		for _, index := range target.Indices {
				// var cronjob string
				if index.Period == "minutes" {
					cronjob = fmt.Sprintf("*/%v * * * *", index.Unit)

				} else if index.Period == "hours" {
					cronjob = fmt.Sprintf("0 */%v * * *", index.Unit)
				}
				// fmt.Println("crontab", cronjob)
				// Detect(index.Pattern, index.Field, index.Period, index.Unit, target.To, target.Subject, index.Logname, index.DeviceGroup)
				ExecuteCrontab(target.ID, index.ID, cronjob, index.Pattern, index.Field, index.Period, index.Unit, target.To, target.Subject, index.Logname, index.DeviceGroup)

		}
	}
}

func Control_center_by_TargetID(targetID int) {

	target, err := GetTargetByID(targetID)
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("get targets data erro: %s",err.Error()))
	}
	// if err := global.Mysql.Exec("TRUNCATE TABLE cron_lists").Error; err != nil {
	// 	log.Logrecord_no_rotate("ERROR", "error when truncate cronlist table")
	// }
	var cronjob string

	for _, index := range target.Indices {
		fmt.Println("index", index)

		if index.Period == "minutes" {
			cronjob = fmt.Sprintf("*/%v * * * *", index.Unit)

		} else if index.Period == "hours" {
			cronjob = fmt.Sprintf("* */%v * * *", index.Unit)
		}
		// fmt.Println("crontab", cronjob)
		// Detect(index.Pattern, index.Field, index.Period, index.Unit, target.To, target.Subject, index.Logname, index.DeviceGroup)
		ExecuteCrontab(targetID, index.ID, cronjob, index.Pattern, index.Field, index.Period, index.Unit, target.To, target.Subject, index.Logname, index.DeviceGroup)

	}

}

func Control_center_by_IndiceID(indicesID int, targetdata entities.IndicesTargets) {

	var cronjob string
	
	target,err := GetTargetByID(targetdata.TargetID)
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("get targets data error (Control_center_by_IndiceID - GetTargetByID): %s",err.Error()))
	}
	for _, index := range target.Indices {
		if index.ID == indicesID {
			// var cronjob string
			if index.Period == "minutes" {
				cronjob = fmt.Sprintf("*/%v * * * *", index.Unit)
			} else if index.Period == "hours" {
				cronjob = fmt.Sprintf("* */%v * * *", index.Unit)
			}
			// fmt.Println("crontab", cronjob)
			// Detect(index.Pattern, index.Field, index.Period, index.Unit, target.To, target.Subject, index.Logname, index.DeviceGroup)
			ExecuteCrontab(target.ID, index.ID, cronjob, index.Pattern, index.Field, index.Period, index.Unit, target.To, target.Subject, index.Logname, index.DeviceGroup)
		}
	}

}
