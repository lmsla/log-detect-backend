package main

import (
	"log"
	"time"

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

	// 初始化 TimescaleDB
	if err := clients.LoadTimescaleDB(); err != nil {
		log.Fatalf("Failed to initialize TimescaleDB: %v", err)
	}
	defer global.TimescaleDB.Close()

	// 初始化批量寫入服務
	if global.EnvConfig.BatchWriter.Enabled {
		flushInterval, err := time.ParseDuration(global.EnvConfig.BatchWriter.FlushInterval)
		if err != nil {
			flushInterval = 30 * time.Second
		}
		global.BatchWriter = services.NewBatchWriter(
			global.TimescaleDB,
			global.EnvConfig.BatchWriter.BatchSize,
			flushInterval,
		)
		defer global.BatchWriter.Stop()
		log.Println("✅ BatchWriter initialized successfully")
	}

	// Create tables first before initializing ES client (es_connections table must exist)
	services.CreateTable()

	// Initialize ES client after tables are created
	clients.SetElkClient()

	// Initialize authentication system (create default roles and admin user)
	authService := services.NewAuthService()
	if err := authService.CreateDefaultRolesAndPermissions(); err != nil {
		log.Printf("Failed to create default roles and permissions: %v", err)
	}

	if err := authService.CreateDefaultAdmin(); err != nil {
		log.Printf("Failed to create default admin user: %v", err)
	}

	services.LoadCrontab()

	// 初始化 ES 監控排程器
	services.InitESScheduler()
	if err := services.GlobalESScheduler.LoadAllMonitors(); err != nil {
		log.Printf("Failed to load ES monitors: %v", err)
	}

	services.Control_center()

	r := router.LoadRouter()
	r.Run(global.EnvConfig.Server.Port)

}
