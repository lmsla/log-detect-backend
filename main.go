package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"log-detect/clients"
	"log-detect/global"
	appLog "log-detect/log"
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

	global.SetMySQLRuntimeStatus(global.MySQLRuntimeStatus{
		Configured: true,
		Connected:  false,
		Host:       global.EnvConfig.Database.Host,
		Port:       global.EnvConfig.Database.Port,
		Database:   global.EnvConfig.Database.Db,
	})

	global.SetTimescaleRuntimeStatus(global.TimescaleRuntimeStatus{
		Configured:       global.EnvConfig.Features.TimescaleDB,
		Connected:        false,
		Host:             global.EnvConfig.Timescale.Host,
		Port:             global.EnvConfig.Timescale.Port,
		Error:            "",
		DegradedFeatures: services.TimescaleDependentFeatures(),
		LastCheckAt:      time.Now().Format(time.RFC3339),
	})

	global.SetESMonitoringRuntimeStatus(global.ESMonitoringRuntimeStatus{
		Enabled:          global.EnvConfig.Features.ESMonitoring,
		SchedulerStarted: false,
		Reason:           "not initialized",
	})

	clients.LoadDatabase()
	global.SetMySQLRuntimeStatus(global.MySQLRuntimeStatus{
		Configured: true,
		Connected:  true,
		Host:       global.EnvConfig.Database.Host,
		Port:       global.EnvConfig.Database.Port,
		Database:   global.EnvConfig.Database.Db,
	})

	mysql, _ := global.Mysql.DB()
	defer mysql.Close()

	// === Feature Toggle: TimescaleDB ===
	if global.EnvConfig.Features.TimescaleDB {
		if err := clients.LoadTimescaleDB(); err != nil {
			global.SetTimescaleRuntimeStatus(global.TimescaleRuntimeStatus{
				Configured:       true,
				Connected:        false,
				Host:             global.EnvConfig.Timescale.Host,
				Port:             global.EnvConfig.Timescale.Port,
				Error:            err.Error(),
				DegradedFeatures: services.TimescaleDependentFeatures(),
				LastCheckAt:      time.Now().Format(time.RFC3339),
			})

			msg := fmt.Sprintf(
				"TimescaleDB connection failed; host=%s port=%s configured=true connected=false reason=%q degraded_features=%v",
				global.EnvConfig.Timescale.Host,
				global.EnvConfig.Timescale.Port,
				err.Error(),
				services.TimescaleDependentFeatures(),
			)
			log.Printf("ERROR: %s", msg)
			appLog.Logrecord_no_rotate("ERROR", msg)
		} else {
			global.SetTimescaleRuntimeStatus(global.TimescaleRuntimeStatus{
				Configured:       true,
				Connected:        true,
				Host:             global.EnvConfig.Timescale.Host,
				Port:             global.EnvConfig.Timescale.Port,
				Error:            "",
				DegradedFeatures: nil,
				LastCheckAt:      time.Now().Format(time.RFC3339),
			})
			defer global.TimescaleDB.Close()

			// 批量寫入服務（依賴 TimescaleDB）
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
				log.Println("BatchWriter initialized successfully")
			}
		}
	} else {
		global.SetTimescaleRuntimeStatus(global.TimescaleRuntimeStatus{
			Configured:       false,
			Connected:        false,
			Host:             global.EnvConfig.Timescale.Host,
			Port:             global.EnvConfig.Timescale.Port,
			Error:            "feature disabled in setting.yml",
			DegradedFeatures: services.TimescaleDependentFeatures(),
			LastCheckAt:      time.Now().Format(time.RFC3339),
		})
		fmt.Println("TimescaleDB feature disabled, skipping initialization")
	}

	// 執行資料庫 migrations
	fmt.Println("Starting migrations...")
	if err := services.RunMigrations(); err != nil {
		fmt.Fprintf(os.Stderr, "Database migration failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Migrations completed")

	// === YML-to-DB 同步 ===
	if global.EnvConfig.ConfigSource == "yml" {
		fmt.Println("Config source is YML, syncing config.yml/devices.yml to database...")
		if err := services.SyncConfigToDB(); err != nil {
			log.Fatalf("Failed to sync config to DB: %v", err)
		}
		fmt.Println("Config sync completed")
		services.StartDevicesYMLReload()
	}

	// Initialize ES client after tables are created
	clients.SetElkClient()

	// === Feature Toggle: Auth ===
	if global.EnvConfig.Features.Auth {
		authService := services.NewAuthService()
		if err := authService.CreateDefaultRolesAndPermissions(); err != nil {
			log.Printf("Failed to create default roles and permissions: %v", err)
		}
		if err := authService.CreateDefaultAdmin(); err != nil {
			log.Printf("Failed to create default admin user: %v", err)
		}
	} else {
		fmt.Println("Auth feature disabled, skipping RBAC initialization")
	}

	services.LoadCrontab()

	// === Feature Toggle: ES Monitoring ===
	if global.EnvConfig.Features.ESMonitoring {
		if services.IsTimescaleReady() {
			services.InitESScheduler()
			if err := services.GlobalESScheduler.LoadAllMonitors(); err != nil {
				log.Printf("Failed to load ES monitors: %v", err)
				global.SetESMonitoringRuntimeStatus(global.ESMonitoringRuntimeStatus{
					Enabled:          true,
					SchedulerStarted: false,
					Reason:           fmt.Sprintf("failed to load monitors: %v", err),
				})
			} else {
				global.SetESMonitoringRuntimeStatus(global.ESMonitoringRuntimeStatus{
					Enabled:          true,
					SchedulerStarted: true,
					Reason:           "",
				})
			}
		} else {
			msg := fmt.Sprintf("ES Monitoring scheduler not started because TimescaleDB is unavailable: %s", services.TimescaleUnavailableMessage())
			log.Printf("ERROR: %s", msg)
			appLog.Logrecord_no_rotate("ERROR", msg)
			global.SetESMonitoringRuntimeStatus(global.ESMonitoringRuntimeStatus{
				Enabled:          true,
				SchedulerStarted: false,
				Reason:           services.TimescaleUnavailableMessage(),
			})
		}
	} else {
		global.SetESMonitoringRuntimeStatus(global.ESMonitoringRuntimeStatus{
			Enabled:          false,
			SchedulerStarted: false,
			Reason:           "feature disabled in setting.yml",
		})
		fmt.Println("ES Monitoring feature disabled, skipping scheduler initialization")
	}

	// 核心偵測（永遠執行）
	services.Control_center()

	r := router.LoadRouter()
	r.Run(global.EnvConfig.Server.Port)
}
