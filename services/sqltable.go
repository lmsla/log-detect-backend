package services

import (
	"fmt"
	"log-detect/entities"
	"log-detect/global"
)

func CreateTable() {
	global.Mysql.Exec("USE logdetect")
	err := global.Mysql.AutoMigrate(
		// 用戶與權限
		&entities.User{},
		&entities.Role{},
		&entities.Permission{},

		// 裝置監控相關
		&entities.Device{},
		&entities.Receiver{},
		&entities.Index{},
		&entities.Target{},

		// 歷史記錄
		&entities.History{},
		&entities.HistoryArchive{},
		&entities.HistoryDailyStats{},
		&entities.MailHistory{},
		&entities.AlertHistory{},

		// 排程管理
		&entities.CronList{},

		// ES 連線與監控
		&entities.ESConnection{},           // ES 連線配置表（Issue #001）
		&entities.ElasticsearchMonitor{},   // ES 監控配置表

		// 系統模組
		&entities.Module{},                 // 系統模組管理
	)
	if err != nil {
		fmt.Println("Database migration error:", err)
	} else {
		fmt.Println("Database migration completed successfully")
	}
}
