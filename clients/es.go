package clients

import (
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"log-detect/global"
	"log-detect/log"
	"log-detect/services"
)

var ES *elasticsearch.Client

// SetElkClient 初始化 ES 客戶端（使用新的連線管理器）
func SetElkClient() {
	log.Logrecord_no_rotate("INFO", "Initializing ES client via Connection Manager...")

	// 取得連線管理器單例
	manager := services.GetESConnectionManager()

	// 初始化連線管理器（從資料庫載入所有連線）
	if err := manager.Initialize(); err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to initialize ES Connection Manager: %v", err))
		log.Logrecord("Elasticsearch", "ES Connection Manager 初始化失敗")

		// 不再直接 panic，而是記錄錯誤並繼續
		// 這樣即使 ES 連線失敗，系統其他功能仍可運作
		fmt.Println("⚠️  ES Connection Manager 初始化失敗，部分功能可能無法使用")
		return
	}

	// 取得預設客戶端
	ES = manager.GetDefaultClient()

	// 同時設定 global.Elasticsearch（供其他 package 使用，避免 import cycle）
	global.Elasticsearch = ES

	if ES == nil {
		log.Logrecord_no_rotate("WARNING", "Default ES client is nil after initialization")
		log.Logrecord("Elasticsearch", "警告：無法取得預設 ES 客戶端")
		fmt.Println("⚠️  無法取得預設 ES 客戶端，請檢查 es_connections 表或 setting.yml 配置")
		return
	}

	// 測試連線
	res, err := ES.Info()
	if err != nil {
		log.Logrecord_no_rotate("WARNING", fmt.Sprintf("Default ES connection test failed: %v", err))
		log.Logrecord("Elasticsearch", fmt.Sprintf("ES 連線測試失敗: %s", err))
		fmt.Println("⚠️  ES 連線測試失敗，但連線管理器已初始化")
	} else {
		defer res.Body.Close()
		log.Logrecord_no_rotate("INFO", "Default ES connection test successful")
		log.Logrecord("Elasticsearch", "ES 連線成功")
		fmt.Println("✅ ES Connection Manager 初始化成功")
	}
}