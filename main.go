package main

import (
	"log-detect/clients"
	"log-detect/global"
	"log-detect/router"
	"log-detect/services"
	"log-detect/utils"
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