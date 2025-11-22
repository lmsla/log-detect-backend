package services

import (
	"encoding/json"
	"fmt"
	"log-detect/entities"
	"log-detect/global"
	"log-detect/models"
)

// CreateESMonitor 創建 ES 監控配置
func CreateESMonitor(monitor entities.ElasticsearchMonitor) models.Response {
	// 驗證必填欄位
	if monitor.Name == "" {
		return models.Response{
			Success: false,
			Msg:     "監控名稱不能為空",
		}
	}
	if monitor.ESConnectionID == 0 {
		return models.Response{
			Success: false,
			Msg:     "ES 連線 ID 不能為空",
		}
	}

	// 驗證 ESConnection 是否存在
	var esConn entities.ESConnection
	if err := global.Mysql.Where("id = ? AND deleted_at IS NULL", monitor.ESConnectionID).First(&esConn).Error; err != nil {
		return models.Response{
			Success: false,
			Msg:     "指定的 ES 連線不存在",
		}
	}

	// 設置默認值
	if monitor.Interval == 0 {
		monitor.Interval = 60
	}
	if monitor.CheckType == "" {
		monitor.CheckType = "health,performance"
	}

	// 設置默認告警閾值
	if monitor.AlertThreshold == "" {
		defaultThreshold := entities.DefaultESAlertThreshold()
		thresholdJSON, _ := json.Marshal(defaultThreshold)
		monitor.AlertThreshold = string(thresholdJSON)
	}

	// 創建監控配置
	if err := global.Mysql.Create(&monitor).Error; err != nil {
		return models.Response{
			Success: false,
			Msg:     fmt.Sprintf("創建監控配置失敗: %s", err.Error()),
		}
	}

	// 如果監控已啟用，啟動排程器
	if monitor.EnableMonitor && GlobalESScheduler != nil {
		if err := GlobalESScheduler.StartMonitor(monitor); err != nil {
			return models.Response{
				Success: false,
				Msg:     fmt.Sprintf("監控配置已創建但啟動失敗: %s", err.Error()),
			}
		}
	}

	return models.Response{
		Success: true,
		Msg:     "創建監控配置成功",
		Body:    monitor,
	}
}

// UpdateESMonitor 更新 ES 監控配置
func UpdateESMonitor(monitor entities.ElasticsearchMonitor) models.Response {
	if monitor.ID == 0 {
		return models.Response{
			Success: false,
			Msg:     "監控 ID 不能為空",
		}
	}

	// 檢查監控配置是否存在
	var existingMonitor entities.ElasticsearchMonitor
	if err := global.Mysql.First(&existingMonitor, monitor.ID).Error; err != nil {
		return models.Response{
			Success: false,
			Msg:     "監控配置不存在",
		}
	}

	// 更新監控配置
	if err := global.Mysql.Model(&existingMonitor).Updates(&monitor).Error; err != nil {
		return models.Response{
			Success: false,
			Msg:     fmt.Sprintf("更新監控配置失敗: %s", err.Error()),
		}
	}

	// 重新載入更新後的監控配置
	var updatedMonitor entities.ElasticsearchMonitor
	if err := global.Mysql.First(&updatedMonitor, monitor.ID).Error; err != nil {
		return models.Response{
			Success: false,
			Msg:     "無法載入更新後的監控配置",
		}
	}

	// 重啟排程器以應用新配置
	if GlobalESScheduler != nil {
		if err := GlobalESScheduler.RestartMonitor(updatedMonitor); err != nil {
			return models.Response{
				Success: false,
				Msg:     fmt.Sprintf("監控配置已更新但重啟失敗: %s", err.Error()),
			}
		}
	}

	return models.Response{
		Success: true,
		Msg:     "更新監控配置成功",
		Body:    updatedMonitor,
	}
}

// GetAllESMonitors 獲取所有 ES 監控配置
func GetAllESMonitors() models.Response {
	var monitors []entities.ElasticsearchMonitor

	if err := global.Mysql.Find(&monitors).Error; err != nil {
		return models.Response{
			Success: false,
			Msg:     fmt.Sprintf("查詢監控配置失敗: %s", err.Error()),
		}
	}

	return models.Response{
		Success: true,
		Msg:     "查詢監控配置成功",
		Body:    monitors,
	}
}

// GetESMonitorByID 根據 ID 獲取 ES 監控配置
func GetESMonitorByID(id int) models.Response {
	var monitor entities.ElasticsearchMonitor

	if err := global.Mysql.First(&monitor, id).Error; err != nil {
		return models.Response{
			Success: false,
			Msg:     "監控配置不存在",
		}
	}

	return models.Response{
		Success: true,
		Msg:     "查詢監控配置成功",
		Body:    monitor,
	}
}

// DeleteESMonitor 刪除 ES 監控配置
func DeleteESMonitor(id int) models.Response {
	var monitor entities.ElasticsearchMonitor

	// 檢查監控配置是否存在
	if err := global.Mysql.First(&monitor, id).Error; err != nil {
		return models.Response{
			Success: false,
			Msg:     "監控配置不存在",
		}
	}

	// 停止排程器
	if GlobalESScheduler != nil {
		if err := GlobalESScheduler.StopMonitor(id); err != nil {
			// 記錄錯誤但繼續刪除
			fmt.Printf("Warning: Failed to stop monitor scheduler: %s\n", err.Error())
		}
	}

	// 刪除監控配置
	if err := global.Mysql.Delete(&monitor).Error; err != nil {
		return models.Response{
			Success: false,
			Msg:     fmt.Sprintf("刪除監控配置失敗: %s", err.Error()),
		}
	}

	return models.Response{
		Success: true,
		Msg:     "刪除監控配置成功",
	}
}

// TestESMonitorConnection 已移除
// ES Monitor 現在使用 ESConnection，測試連線請使用 services.TestESConnection

// ToggleESMonitor 啟用/停用 ES 監控
func ToggleESMonitor(id int, enable bool) models.Response {
	var monitor entities.ElasticsearchMonitor

	// 檢查監控配置是否存在
	if err := global.Mysql.First(&monitor, id).Error; err != nil {
		return models.Response{
			Success: false,
			Msg:     "監控配置不存在",
		}
	}

	// 更新啟用狀態
	if err := global.Mysql.Model(&monitor).Update("enable_monitor", enable).Error; err != nil {
		return models.Response{
			Success: false,
			Msg:     fmt.Sprintf("更新監控狀態失敗: %s", err.Error()),
		}
	}

	// 重新載入更新後的監控配置
	var updatedMonitor entities.ElasticsearchMonitor
	if err := global.Mysql.First(&updatedMonitor, id).Error; err != nil {
		return models.Response{
			Success: false,
			Msg:     "無法載入更新後的監控配置",
		}
	}

	// 重啟排程器以應用新的啟用狀態
	if GlobalESScheduler != nil {
		if err := GlobalESScheduler.RestartMonitor(updatedMonitor); err != nil {
			return models.Response{
				Success: false,
				Msg:     fmt.Sprintf("監控狀態已更新但重啟失敗: %s", err.Error()),
			}
		}
	}

	statusMsg := "停用"
	if enable {
		statusMsg = "啟用"
	}

	return models.Response{
		Success: true,
		Msg:     fmt.Sprintf("監控已%s", statusMsg),
	}
}
