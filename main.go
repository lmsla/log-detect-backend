package main

import (
	"fmt"
	"log-detect/clients"
	"log-detect/global"
	"log-detect/router"
	"log-detect/services"
	"log-detect/utils"
	"time"
	"log-detect/log"
)

// @title Log Detect Golang API
// @version 1.0
// @description Golang API 專案描述
// @termsOfService http://swagger.io/terms/
// @contact.name Russell
// @contact.email support@swagger.io
//// @host 10.99.1.133:8006
// @host localhost:8006
// @BasePath  /api/v1
// @query.collection.format multi
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @schemes http
func main() {

	utils.LoadEnvironment()

	clients.LoadDatabase()
	mysql, _ := global.Mysql.DB()
	defer mysql.Close()

	clients.SetElkClient()

	services.LoadCrontab()

	services.Control_center()

	r := router.LoadRouter()
	r.Run(global.EnvConfig.Server.Port)
	// services.CreateTable()
	

}

func main1() {

	utils.LoadEnvironment()

	clients.LoadDatabase()
	mysql, _ := global.Mysql.DB()
	defer mysql.Close()

	clients.SetElkClient()

	log.Logrecord_no_rotate("dfds","test")

	// services.GetDeviceGroup()
	// services.GetDevicesDataByGroupName("vip")
	// services.GetServerMenu()
	// services.Control_center()

	date := time.Now().Format("2006-01-02")
	time1 := time.Now().Format("15:04")
	fmt.Println("date", date)
	fmt.Println("time", time1)
	// services.GetLogname()
	// services.GenerateTimeArray("minutes",3)
	// services.DataDealing("some-fire-log")

	now1 := time.Now()
	// 根據 period 和 unit 計算上一次 crontab 的執行時間
	lastCrontabTime := services.GetLastCrontabTime(now1,"hours", 5)

	fmt.Println("Last crontab time:", lastCrontabTime)

	// services.GetLognameData()

	
}
