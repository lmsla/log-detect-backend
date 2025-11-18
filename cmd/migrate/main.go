package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"log-detect/clients"
	"log-detect/global"
	"log-detect/utils"
)

func main() {
	// 定義命令列參數
	action := flag.String("action", "up", "Migration action: up, down, version, force, goto")
	version := flag.Uint("version", 0, "Target version for 'goto' or 'force' action")
	migrationPath := flag.String("path", "migrations", "Path to migrations directory")
	flag.Parse()

	// 顯示歡迎訊息
	printBanner()

	// 載入環境配置
	log.Println("📋 Loading environment configuration...")
	utils.LoadEnvironment()

	// 初始化資料庫連線
	log.Println("🔌 Connecting to databases...")
	clients.LoadDatabase()
	mysqlDB, err := global.Mysql.DB()
	if err != nil {
		log.Fatalf("❌ Failed to get MySQL DB instance: %v", err)
	}
	defer mysqlDB.Close()

	// 初始化 TimescaleDB
	var timescaleDB *sql.DB
	if err := clients.LoadTimescaleDB(); err != nil {
		log.Printf("⚠️  TimescaleDB not available: %v", err)
		log.Println("ℹ️  Will only run MySQL migrations")
	} else {
		timescaleDB = global.TimescaleDB
		defer timescaleDB.Close()
	}

	// 取得專案根目錄的絕對路徑
	rootDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("❌ Failed to get current directory: %v", err)
	}
	fullMigrationPath := filepath.Join(rootDir, *migrationPath)

	// 建立 Migration 管理器
	log.Println("🛠️  Creating migration manager...")
	manager, err := utils.NewMigrationManager(mysqlDB, timescaleDB, fullMigrationPath)
	if err != nil {
		log.Fatalf("❌ Failed to create migration manager: %v", err)
	}
	defer manager.Close()

	// 執行 Migration 動作
	log.Printf("🚀 Executing action: %s\n", *action)
	switch *action {
	case "up":
		if err := manager.MigrateUp(); err != nil {
			log.Fatalf("❌ Migration up failed: %v", err)
		}
		log.Println("✅ All migrations completed successfully!")

	case "down":
		if err := manager.MigrateDown(); err != nil {
			log.Fatalf("❌ Migration down failed: %v", err)
		}
		log.Println("✅ Migration rolled back successfully!")

	case "goto":
		if *version == 0 {
			log.Fatal("❌ Please specify target version with -version flag")
		}
		if err := manager.MigrateTo(*version); err != nil {
			log.Fatalf("❌ Migration to version %d failed: %v", *version, err)
		}
		log.Printf("✅ Migrated to version %d successfully!\n", *version)

	case "version":
		mysqlVer, mysqlDirty, tsVer, tsDirty, err := manager.GetVersion()
		if err != nil {
			log.Fatalf("❌ Failed to get version: %v", err)
		}

		log.Println("\n📊 Current Migration Versions:")
		log.Println("================================")

		// MySQL
		if mysqlVer == 0 && !mysqlDirty {
			log.Println("MySQL:       No migrations applied yet")
		} else {
			dirtyStatus := ""
			if mysqlDirty {
				dirtyStatus = " (⚠️  DIRTY - needs manual intervention)"
			}
			log.Printf("MySQL:       Version %d%s\n", mysqlVer, dirtyStatus)
		}

		// TimescaleDB
		if timescaleDB != nil {
			if tsVer == 0 && !tsDirty {
				log.Println("TimescaleDB: No migrations applied yet")
			} else {
				dirtyStatus := ""
				if tsDirty {
					dirtyStatus = " (⚠️  DIRTY - needs manual intervention)"
				}
				log.Printf("TimescaleDB: Version %d%s\n", tsVer, dirtyStatus)
			}
		} else {
			log.Println("TimescaleDB: Not connected")
		}
		log.Println("================================")

	case "force":
		if *version == 0 {
			log.Fatal("❌ Please specify version with -version flag")
		}
		log.Printf("⚠️  WARNING: Forcing database to version %d\n", *version)
		log.Println("⚠️  This should only be used to fix dirty migration state!")
		if err := manager.Force(int(*version)); err != nil {
			log.Fatalf("❌ Force failed: %v", err)
		}
		log.Printf("✅ Forced to version %d successfully!\n", *version)

	default:
		log.Fatalf("❌ Unknown action: %s. Use: up, down, version, goto, or force", *action)
	}

	log.Println("\n✨ Migration tool finished!")
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════╗
║   Log Detect - Database Migration Tool        ║
║   Version: 1.0.0                              ║
╚═══════════════════════════════════════════════╝
`
	fmt.Println(banner)
}
