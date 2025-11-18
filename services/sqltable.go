package services

import (
	"fmt"
	"log-detect/entities"
	"log-detect/global"
)

func CreateTable() {
	global.Mysql.Exec("USE logdetect")
	err := global.Mysql.AutoMigrate(
		&entities.User{},
		&entities.Role{},
		&entities.Permission{},
		&entities.Device{},
		&entities.Receiver{},
		&entities.Index{},
		&entities.Target{},
		&entities.History{},
		&entities.HistoryArchive{},
		&entities.HistoryDailyStats{},
		&entities.MailHistory{},
		&entities.AlertHistory{},
		&entities.CronList{},
		&entities.ElasticsearchMonitor{}, // ES 監控配置表
	)
	if err != nil {
		fmt.Println("Database migration error:", err)
	} else {
		fmt.Println("Database migration completed successfully")
	}
}
