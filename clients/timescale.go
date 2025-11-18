package clients

import (
	"database/sql"
	"fmt"
	"log-detect/global"
	"time"

	_ "github.com/lib/pq" // PostgreSQL 驅動
)

// LoadTimescaleDB 初始化 TimescaleDB 連接
func LoadTimescaleDB() error {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Taipei",
		global.EnvConfig.Timescale.Host,
		global.EnvConfig.Timescale.Port,
		global.EnvConfig.Timescale.User,
		global.EnvConfig.Timescale.Password,
		global.EnvConfig.Timescale.Db,
		global.EnvConfig.Timescale.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open TimescaleDB connection: %w", err)
	}

	// 測試連接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping TimescaleDB: %w", err)
	}

	// 連接池配置
	db.SetMaxOpenConns(int(global.EnvConfig.Timescale.MaxOpenConn))
	db.SetMaxIdleConns(int(global.EnvConfig.Timescale.MaxIdle))

	// 解析生命週期
	maxLifetime, err := time.ParseDuration(global.EnvConfig.Timescale.MaxLifeTime)
	if err != nil {
		maxLifetime = time.Hour
	}
	db.SetConnMaxLifetime(maxLifetime)

	// 儲存到全局變數
	global.TimescaleDB = db

	fmt.Println("✅ TimescaleDB connected successfully")
	return nil
}
