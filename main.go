package main

import (
	"fmt"
	"log"
	"os"
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

	// 連接 MySQL
	fmt.Println("Connecting to MySQL...")
	clients.LoadDatabase()
	mysql, err := global.Mysql.DB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to get MySQL connection: %v\n", err)
		os.Exit(1)
	}
	defer mysql.Close()
	fmt.Println("✅ MySQL connected")

	// 初始化 TimescaleDB
	fmt.Println("Connecting to TimescaleDB...")
	if err := clients.LoadTimescaleDB(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to initialize TimescaleDB: %v\n", err)
		os.Exit(1)
	}
	defer global.TimescaleDB.Close()
	fmt.Println("✅ TimescaleDB connected")

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

	// 執行資料庫 migrations
	fmt.Println("Starting migrations...")
	if err := services.RunMigrations(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Database migration failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Migrations completed")

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
