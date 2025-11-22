package services

import (
	"crypto/tls"
	"fmt"
	"log-detect/entities"
	"log-detect/global"
	"log-detect/log"
	"log-detect/models"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
)

// GetAllESConnections 取得所有 ES 連線配置
func GetAllESConnections() models.Response {
	res := models.Response{}
	res.Success = false
	res.Body = []entities.ESConnection{}

	var connections []entities.ESConnection
	err := global.Mysql.Where("deleted_at IS NULL").Order("is_default DESC, id ASC").Find(&connections).Error
	if err != nil {
		res.Msg = fmt.Sprintf("Failed to get ES connections: %s", err.Error())
		log.Logrecord_no_rotate("ERROR", res.Msg)
		return res
	}

	// 轉換為摘要格式（移除敏感資訊）
	summaries := make([]entities.ESConnectionSummary, len(connections))
	for i, conn := range connections {
		summaries[i] = conn.ToSummary()
	}

	res.Success = true
	res.Msg = "Get all ES connections success"
	res.Body = summaries
	return res
}

// GetESConnection 取得單一 ES 連線配置
func GetESConnection(id int) models.Response {
	res := models.Response{}
	res.Success = false

	var connection entities.ESConnection
	err := global.Mysql.Where("id = ? AND deleted_at IS NULL", id).First(&connection).Error
	if err != nil {
		res.Msg = fmt.Sprintf("ES connection not found: %s", err.Error())
		log.Logrecord_no_rotate("ERROR", res.Msg)
		return res
	}

	res.Success = true
	res.Msg = "Get ES connection success"
	res.Body = connection.ToSummary() // 返回摘要格式（不包含密碼）
	return res
}

// CreateESConnection 建立新的 ES 連線配置
func CreateESConnection(connection entities.ESConnection) models.Response {
	res := models.Response{}
	res.Success = false

	// 驗證連線配置
	if err := connection.Validate(); err != nil {
		res.Msg = fmt.Sprintf("Invalid ES connection config: %s", err.Error())
		log.Logrecord_no_rotate("ERROR", res.Msg)
		return res
	}

	// 檢查名稱是否已存在
	var existing entities.ESConnection
	result := global.Mysql.Where("name = ? AND deleted_at IS NULL", connection.Name).First(&existing)
	if result.RowsAffected > 0 {
		res.Msg = fmt.Sprintf("ES connection name '%s' already exists", connection.Name)
		log.Logrecord_no_rotate("WARNING", res.Msg)
		return res
	}

	// 如果設定為預設連線，先取消其他預設連線
	if connection.IsDefault {
		if err := global.Mysql.Model(&entities.ESConnection{}).
			Where("is_default = ? AND deleted_at IS NULL", true).
			Update("is_default", false).Error; err != nil {
			res.Msg = fmt.Sprintf("Failed to update existing default connection: %s", err.Error())
			log.Logrecord_no_rotate("ERROR", res.Msg)
			return res
		}
	}

	// 建立連線
	err := global.Mysql.Create(&connection).Error
	if err != nil {
		res.Msg = fmt.Sprintf("Failed to create ES connection: %s", err.Error())
		log.Logrecord_no_rotate("ERROR", res.Msg)
		return res
	}

	// 重新載入連線管理器
	manager := GetESConnectionManager()
	if err := manager.ReloadConnection(connection.ID); err != nil {
		log.Logrecord_no_rotate("WARNING", fmt.Sprintf("Failed to reload connection manager after creation: %s", err.Error()))
	}

	res.Success = true
	res.Msg = "Create ES connection success"
	res.Body = connection.ToSummary()
	log.Logrecord_no_rotate("INFO", fmt.Sprintf("Created ES connection: '%s' (ID: %d)", connection.Name, connection.ID))
	return res
}

// UpdateESConnection 更新 ES 連線配置
func UpdateESConnection(connection entities.ESConnection) models.Response {
	res := models.Response{}
	res.Success = false

	// 驗證連線配置
	if err := connection.Validate(); err != nil {
		res.Msg = fmt.Sprintf("Invalid ES connection config: %s", err.Error())
		log.Logrecord_no_rotate("ERROR", res.Msg)
		return res
	}

	// 檢查連線是否存在
	var existing entities.ESConnection
	if err := global.Mysql.Where("id = ? AND deleted_at IS NULL", connection.ID).First(&existing).Error; err != nil {
		res.Msg = fmt.Sprintf("ES connection not found: %s", err.Error())
		log.Logrecord_no_rotate("ERROR", res.Msg)
		return res
	}

	// 檢查名稱是否與其他連線衝突
	result := global.Mysql.Where("id != ? AND name = ? AND deleted_at IS NULL", connection.ID, connection.Name).First(&entities.ESConnection{})
	if result.RowsAffected > 0 {
		res.Msg = fmt.Sprintf("ES connection name '%s' already exists", connection.Name)
		log.Logrecord_no_rotate("WARNING", res.Msg)
		return res
	}

	// 如果設定為預設連線，先取消其他預設連線
	if connection.IsDefault {
		if err := global.Mysql.Model(&entities.ESConnection{}).
			Where("id != ? AND is_default = ? AND deleted_at IS NULL", connection.ID, true).
			Update("is_default", false).Error; err != nil {
			res.Msg = fmt.Sprintf("Failed to update existing default connection: %s", err.Error())
			log.Logrecord_no_rotate("ERROR", res.Msg)
			return res
		}
	}

	// 更新連線
	err := global.Mysql.Model(&entities.ESConnection{}).Where("id = ?", connection.ID).Updates(&connection).Error
	if err != nil {
		res.Msg = fmt.Sprintf("Failed to update ES connection: %s", err.Error())
		log.Logrecord_no_rotate("ERROR", res.Msg)
		return res
	}

	// 重新載入連線管理器
	manager := GetESConnectionManager()
	if err := manager.ReloadConnection(connection.ID); err != nil {
		log.Logrecord_no_rotate("WARNING", fmt.Sprintf("Failed to reload connection manager after update: %s", err.Error()))
	}

	res.Success = true
	res.Msg = "Update ES connection success"
	res.Body = connection.ToSummary()
	log.Logrecord_no_rotate("INFO", fmt.Sprintf("Updated ES connection: '%s' (ID: %d)", connection.Name, connection.ID))
	return res
}

// DeleteESConnection 刪除 ES 連線配置（軟刪除）
func DeleteESConnection(id int) models.Response {
	res := models.Response{}
	res.Success = false

	// 檢查連線是否存在
	var connection entities.ESConnection
	if err := global.Mysql.Where("id = ? AND deleted_at IS NULL", id).First(&connection).Error; err != nil {
		res.Msg = fmt.Sprintf("ES connection not found: %s", err.Error())
		log.Logrecord_no_rotate("ERROR", res.Msg)
		return res
	}

	// 檢查是否有 Index 正在使用此連線
	var indexCount int64
	global.Mysql.Model(&entities.Index{}).Where("es_connection_id = ?", id).Count(&indexCount)
	if indexCount > 0 {
		res.Msg = fmt.Sprintf("Cannot delete ES connection: %d index(es) are still using it", indexCount)
		log.Logrecord_no_rotate("WARNING", res.Msg)
		return res
	}

	// 檢查是否有 Monitor 正在使用此連線
	var monitorCount int64
	global.Mysql.Model(&entities.ElasticsearchMonitor{}).Where("es_connection_id = ?", id).Count(&monitorCount)
	if monitorCount > 0 {
		res.Msg = fmt.Sprintf("Cannot delete ES connection: %d monitor(s) are still using it", monitorCount)
		log.Logrecord_no_rotate("WARNING", res.Msg)
		return res
	}

	// 執行軟刪除
	err := global.Mysql.Delete(&connection).Error
	if err != nil {
		res.Msg = fmt.Sprintf("Failed to delete ES connection: %s", err.Error())
		log.Logrecord_no_rotate("ERROR", res.Msg)
		return res
	}

	// 從連線管理器移除
	manager := GetESConnectionManager()
	manager.RemoveConnection(id)

	res.Success = true
	res.Msg = "Delete ES connection success"
	log.Logrecord_no_rotate("INFO", fmt.Sprintf("Deleted ES connection: '%s' (ID: %d)", connection.Name, connection.ID))
	return res
}

// TestESConnection 測試 ES 連線（不儲存到資料庫）
func TestESConnection(connection entities.ESConnection) models.Response {
	res := models.Response{}
	res.Success = false

	// 如果有 ID 且啟用認證但缺少密碼，從資料庫讀取
	if connection.ID > 0 && connection.EnableAuth && connection.Password == "" {
		var existing entities.ESConnection
		if err := global.Mysql.Where("id = ? AND deleted_at IS NULL", connection.ID).First(&existing).Error; err == nil {
			// 使用資料庫中的 username 和 password（如果前端沒提供）
			if connection.Username == "" {
				connection.Username = existing.Username
			}
			connection.Password = existing.Password
		}
	}

	// 驗證連線配置
	if err := connection.Validate(); err != nil {
		res.Msg = fmt.Sprintf("Invalid ES connection config: %s", err.Error())
		return res
	}

	// 建立臨時客戶端進行測試
	testClient, err := createTestClient(&connection)
	if err != nil {
		res.Msg = fmt.Sprintf("Failed to create test client: %s", err.Error())
		return res
	}

	// 測試連線
	esRes, err := testClient.Info()
	if err != nil {
		res.Msg = fmt.Sprintf("Connection test failed: %s", err.Error())
		return res
	}
	defer esRes.Body.Close()

	if esRes.IsError() {
		res.Msg = fmt.Sprintf("ES returned error: %s", esRes.String())
		return res
	}

	res.Success = true
	res.Msg = "Connection test successful"
	res.Body = map[string]interface{}{
		"status":      "online",
		"url":         connection.GetURL(),
		"response":    esRes.Status(),
		"enable_auth": connection.EnableAuth,
	}
	return res
}

// createTestClient 建立測試用的 ES 客戶端（私有函數）
func createTestClient(conn *entities.ESConnection) (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: conn.GetURLs(),
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	if conn.EnableAuth {
		cfg.Username = conn.Username
		cfg.Password = conn.Password
	}

	return elasticsearch.NewClient(cfg)
}

// SetDefaultESConnection 設定預設連線
func SetDefaultESConnection(id int) models.Response {
	res := models.Response{}
	res.Success = false

	// 檢查連線是否存在
	var connection entities.ESConnection
	if err := global.Mysql.Where("id = ? AND deleted_at IS NULL", id).First(&connection).Error; err != nil {
		res.Msg = fmt.Sprintf("ES connection not found: %s", err.Error())
		log.Logrecord_no_rotate("ERROR", res.Msg)
		return res
	}

	// 取消所有預設連線
	if err := global.Mysql.Model(&entities.ESConnection{}).
		Where("deleted_at IS NULL").
		Update("is_default", false).Error; err != nil {
		res.Msg = fmt.Sprintf("Failed to clear default connections: %s", err.Error())
		log.Logrecord_no_rotate("ERROR", res.Msg)
		return res
	}

	// 設定新的預設連線
	if err := global.Mysql.Model(&entities.ESConnection{}).
		Where("id = ?", id).
		Update("is_default", true).Error; err != nil {
		res.Msg = fmt.Sprintf("Failed to set default connection: %s", err.Error())
		log.Logrecord_no_rotate("ERROR", res.Msg)
		return res
	}

	// 重新載入所有連線
	manager := GetESConnectionManager()
	if err := manager.ReloadAllConnections(); err != nil {
		log.Logrecord_no_rotate("WARNING", fmt.Sprintf("Failed to reload connection manager: %s", err.Error()))
	}

	res.Success = true
	res.Msg = "Set default ES connection success"
	log.Logrecord_no_rotate("INFO", fmt.Sprintf("Set default ES connection: '%s' (ID: %d)", connection.Name, connection.ID))
	return res
}

// ReloadESConnection 重新載入指定的 ES 連線
func ReloadESConnection(id int) models.Response {
	res := models.Response{}
	res.Success = false

	// 檢查連線是否存在
	var connection entities.ESConnection
	if err := global.Mysql.Where("id = ? AND deleted_at IS NULL", id).First(&connection).Error; err != nil {
		res.Msg = fmt.Sprintf("ES connection not found: %s", err.Error())
		log.Logrecord_no_rotate("ERROR", res.Msg)
		return res
	}

	// 重新載入連線管理器中的此連線
	manager := GetESConnectionManager()
	if err := manager.ReloadConnection(id); err != nil {
		res.Msg = fmt.Sprintf("Failed to reload connection: %s", err.Error())
		log.Logrecord_no_rotate("ERROR", res.Msg)
		return res
	}

	res.Success = true
	res.Msg = "Reload ES connection success"
	log.Logrecord_no_rotate("INFO", fmt.Sprintf("Reloaded ES connection: '%s' (ID: %d)", connection.Name, connection.ID))
	return res
}

// ReloadAllESConnections 重新載入所有 ES 連線
func ReloadAllESConnections() models.Response {
	res := models.Response{}
	res.Success = false

	manager := GetESConnectionManager()
	if err := manager.ReloadAllConnections(); err != nil {
		res.Msg = fmt.Sprintf("Failed to reload all connections: %s", err.Error())
		log.Logrecord_no_rotate("ERROR", res.Msg)
		return res
	}

	res.Success = true
	res.Msg = "Reload all ES connections success"
	log.Logrecord_no_rotate("INFO", "Reloaded all ES connections")
	return res
}
