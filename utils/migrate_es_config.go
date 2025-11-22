package utils

import (
	"fmt"
	"log-detect/entities"
	"log-detect/global"
	"log-detect/log"
	"strings"
)

// MigrateESConfigToDB 將 setting.yml 的 ES 配置遷移到資料庫
func MigrateESConfigToDB() error {
	log.Logrecord_no_rotate("INFO", "Starting ES configuration migration from setting.yml to database...")

	// 檢查資料庫中是否已存在預設連線
	var existingDefault entities.ESConnection
	err := global.Mysql.Where("is_default = ?", true).First(&existingDefault).Error
	if err == nil {
		log.Logrecord_no_rotate("INFO", fmt.Sprintf("Default ES connection already exists in database: '%s' (ID: %d)", existingDefault.Name, existingDefault.ID))
		return fmt.Errorf("default connection already exists, migration skipped")
	}

	// 檢查 setting.yml 是否有 ES 配置
	if global.EnvConfig == nil || len(global.EnvConfig.ES.URL) == 0 {
		return fmt.Errorf("no ES configuration found in setting.yml")
	}

	// 解析 URL（可能包含 protocol）
	esURL := global.EnvConfig.ES.URL[0]
	host, port, useTLS := parseESURL(esURL)

	// 建立新的 ES 連線記錄
	newConnection := entities.ESConnection{
		Name:        "預設連線 (從 setting.yml 遷移)",
		Host:        host,
		Port:        port,
		Username:    global.EnvConfig.ES.SourceAccount,
		Password:    global.EnvConfig.ES.SourcePassword,
		EnableAuth:  global.EnvConfig.ES.SourceAccount != "",
		UseTLS:      useTLS,
		IsDefault:   true,
		Description: fmt.Sprintf("自動從 setting.yml 遷移的預設 ES 連線\n原始 URL: %s", esURL),
	}

	// 驗證配置
	if err := newConnection.Validate(); err != nil {
		return fmt.Errorf("invalid ES configuration in setting.yml: %w", err)
	}

	// 儲存到資料庫
	if err := global.Mysql.Create(&newConnection).Error; err != nil {
		return fmt.Errorf("failed to save ES connection to database: %w", err)
	}

	log.Logrecord_no_rotate("INFO", fmt.Sprintf("Successfully migrated ES configuration to database (ID: %d)", newConnection.ID))
	log.Logrecord_no_rotate("INFO", fmt.Sprintf("  Name: %s", newConnection.Name))
	log.Logrecord_no_rotate("INFO", fmt.Sprintf("  Host: %s", newConnection.Host))
	log.Logrecord_no_rotate("INFO", fmt.Sprintf("  Port: %d", newConnection.Port))
	log.Logrecord_no_rotate("INFO", fmt.Sprintf("  TLS: %v", newConnection.UseTLS))
	log.Logrecord_no_rotate("INFO", fmt.Sprintf("  Auth: %v", newConnection.EnableAuth))

	return nil
}

// AutoMigrateESConfigIfNeeded 自動檢測並遷移配置（如果需要）
func AutoMigrateESConfigIfNeeded() {
	log.Logrecord_no_rotate("INFO", "Checking if ES configuration migration is needed...")

	// 檢查資料庫中是否有任何 ES 連線
	var count int64
	if err := global.Mysql.Model(&entities.ESConnection{}).Count(&count).Error; err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to check ES connections: %v", err))
		return
	}

	if count > 0 {
		log.Logrecord_no_rotate("INFO", fmt.Sprintf("Found %d ES connection(s) in database, migration not needed", count))
		return
	}

	log.Logrecord_no_rotate("INFO", "No ES connections found in database, attempting auto-migration...")

	// 嘗試自動遷移
	if err := MigrateESConfigToDB(); err != nil {
		log.Logrecord_no_rotate("WARNING", fmt.Sprintf("Auto-migration failed: %v", err))
		log.Logrecord_no_rotate("INFO", "Please manually create ES connection in database or check setting.yml")
	} else {
		log.Logrecord_no_rotate("INFO", "✅ Auto-migration completed successfully")
	}
}

// parseESURL 解析 ES URL，提取 host, port 和 useTLS
func parseESURL(url string) (host string, port int, useTLS bool) {
	// 預設值
	port = 9200
	useTLS = true

	// 移除 protocol
	url = strings.TrimSpace(url)
	if strings.HasPrefix(url, "https://") {
		useTLS = true
		url = strings.TrimPrefix(url, "https://")
	} else if strings.HasPrefix(url, "http://") {
		useTLS = false
		url = strings.TrimPrefix(url, "http://")
	}

	// 移除尾部斜線
	url = strings.TrimSuffix(url, "/")

	// 解析 host:port
	if strings.Contains(url, ":") {
		parts := strings.Split(url, ":")
		host = parts[0]
		// 解析 port
		fmt.Sscanf(parts[1], "%d", &port)
	} else {
		host = url
	}

	return host, port, useTLS
}

// GetESConfigMigrationStatus 取得遷移狀態（用於診斷）
func GetESConfigMigrationStatus() map[string]interface{} {
	status := make(map[string]interface{})

	// 檢查資料庫中的連線數
	var dbCount int64
	global.Mysql.Model(&entities.ESConnection{}).Count(&dbCount)
	status["db_connections_count"] = dbCount

	// 檢查是否有預設連線
	var hasDefault bool
	err := global.Mysql.Where("is_default = ?", true).First(&entities.ESConnection{}).Error
	hasDefault = (err == nil)
	status["has_default_connection"] = hasDefault

	// 檢查 setting.yml 配置
	hasConfig := global.EnvConfig != nil && len(global.EnvConfig.ES.URL) > 0
	status["has_config_yml"] = hasConfig

	if hasConfig {
		status["config_yml_url"] = global.EnvConfig.ES.URL[0]
	}

	// 判斷是否需要遷移
	needsMigration := dbCount == 0 && hasConfig
	status["needs_migration"] = needsMigration

	return status
}
