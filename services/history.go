package services

import (
	"fmt"
	"log-detect/entities"
	"log-detect/global"
	"log-detect/log"
	"log-detect/models"
	"time"
)

// 新增 history
func CreateHistory(hisroty entities.History) models.Response {

	res := models.Response{}
	res.Success = false
	res.Body = entities.Index{}

	// Ensure metadata has valid JSON
	if hisroty.Metadata == "" {
		hisroty.Metadata = "{}"
	}

	err := global.Mysql.Create(&hisroty).Error
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Create history Fail: %s", err.Error()))
		res.Msg = "Create history Fail"
		return res
	}
	res.Success = true
	res.Body = hisroty
	res.Msg = "Create hisroty Success"

	return res
}

// 新增 mail history
func CreateMailHistory(mail_hisroty entities.MailHistory) models.Response {

	res := models.Response{}
	res.Success = false
	res.Body = entities.Index{}

	err := global.Mysql.Create(&mail_hisroty).Error
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Create mail history Fail: %s", err.Error()))
		res.Msg = "Create mail history Fail"
		return res
	}
	res.Success = true
	res.Body = mail_hisroty
	res.Msg = "Create mail hisroty Success"
	return res
}

func GetIndicesDataByLogname(logname string) (entities.Index, error) {

	indices := entities.Index{}
	err := global.Mysql.Where("logname = ?", logname).Find(&indices).Error
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("find devices data error: %s", err.Error()))
		return indices, err
	}
	return indices, nil
}

// 以 logname , device name 查詢歷史紀錄
func GetHistoryDataByDeviceName(logname string, name string) []entities.History {
	return GetHistoryDataByDeviceName_TS(logname, name)
}

func GenerateTimeArray(period string, unit int) []string {
	var timeArray []string

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 计算时间间隔
	var duration time.Duration
	switch period {
	case "minutes":
		duration = time.Minute * time.Duration(unit)
	case "hours":
		duration = time.Hour * time.Duration(unit)
	default:
		fmt.Println("Invalid period")
		return nil
	}

	// 從當天 00:00 開始，根據時間間隔生成時間數據數组
	for t := startOfDay; t.Before(now); t = t.Add(duration) {
		timeArray = append(timeArray, t.Format("15:04"))
	}

	return timeArray
}

// 處理 history data
func DataDealing(logname string) models.Response {

	res := models.Response{}
	res.Success = false

	indicesData, err := GetIndicesDataByLogname(logname)
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Get Indices Data By Logname error: %s", err.Error()))
	}

	device_list, err := GetDevicesDataByGroupName(indicesData.DeviceGroup)
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Get Devices Data By GroupName error: %s", err.Error()))
	}
	var history_final_data []entities.HistoryData

	timeArray := GenerateTimeArray(indicesData.Period, indicesData.Unit)
	// fmt.Println(len(timeArray))

	for _, device := range device_list {
		history_data := GetHistoryDataByDeviceName(logname, device.Name)
		var history_tmp_data []entities.HistoryData
		// fmt.Println("device.Name",device.Name)
		// fmt.Println(history_data)
		// 將歷史資料轉換為 map 方便查找
		historyMap := make(map[string]bool)

		for _, data := range history_data {
			history_tmp_data = append(history_tmp_data, entities.HistoryData{Name: data.Name, Time: data.Time, Lost: data.Lost})
			historyMap[data.Time] = true
		}

		// 匹配時間數組中的時間點與歷史資料中的時間
		for _, timePoint := range timeArray {
			// 如果時間點不在歷史資料中，則添加新的记录
			if _, ok := historyMap[timePoint]; !ok {
				// history_tmp_data = append(history_tmp_data, entities.HistoryData{Name: device.Name, Time: timePoint, Lost: "false"})
				history_tmp_data = append(history_tmp_data, entities.HistoryData{Name: device.Name, Time: timePoint, Lost: "none"})
			}
		}
		// 扁平化 Array 將 history_tmp_data 中的每個元件塞入 history_final_data 中
		history_final_data = append(history_final_data, history_tmp_data...)
		// fmt.Println(history_tmp_data)
		// fmt.Println(len(history_tmp_data))
	}
	// fmt.Println(history_final_data)
	// fmt.Println(len(history_final_data))
	res.Body = history_final_data
	res.Success = true
	return res

}

func CheckLogstatus(logname string) entities.LognameCheck {
	return CheckLogstatus_TS(logname)
}

func GetLognameData() models.Response {
	return GetLognameData_TS()
}

func GetLastCrontabTime(now time.Time, period string, unit int) string {
	var lastCrontabTime string

	switch period {
	case "minutes":
		// 將當前時間調整到最接近的上一個符合條件的時間點
		minutes := now.Minute()
		adjustedMinutes := minutes - (minutes % unit)
		lastCrontabTime = now.Add(-time.Duration(minutes-adjustedMinutes) * time.Minute).Format("15:04")
	case "hours":
		hour := now.Hour()
		// fmt.Println("hour:", hour)
		adjustedHour := hour - (hour % unit)
		// fmt.Println("adjustedHour:", adjustedHour)
		lastCrontabTime = time.Date(now.Year(), now.Month(), now.Day(), adjustedHour, 0, 0, 0, now.Location()).Format("15:04")
		// default:
		// 	lastCrontabTime = now
	}

	return lastCrontabTime
}

// GetDashboardData 獲取儀表板數據
func GetDashboardData() models.Response {
	return GetDashboardData_TS()
}

// GetHistoryStatistics 獲取歷史統計數據
func GetHistoryStatistics(logname, deviceGroup string, startDate, endDate string) models.Response {
	return GetHistoryStatistics_TS(logname, deviceGroup, startDate, endDate)
}

// GetDeviceTimeline 獲取設備時間線數據
func GetDeviceTimeline(deviceName, logname string, days int) models.Response {
	return GetDeviceTimeline_TS(deviceName, logname, days)
}

// GetTrendData 獲取趨勢數據
func GetTrendData(logname, deviceGroup string, days int) models.Response {
	return GetTrendData_TS(logname, deviceGroup, days)
}

// GetGroupStatistics 獲取群組統計
func GetGroupStatistics(logname string) models.Response {
	return GetGroupStatistics_TS(logname)
}

// CreateAlertHistory 創建告警歷史
func CreateAlertHistory(alert entities.AlertHistory) models.Response {
	res := models.Response{}
	res.Success = false

	if err := global.Mysql.Create(&alert).Error; err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Create alert history failed: %s", err.Error()))
		res.Msg = "Create alert history failed"
		return res
	}

	res.Success = true
	res.Body = alert
	res.Msg = "Create alert history success"
	return res
}

// 數據清理和歸檔策略

// CleanOldHistory 清理舊的歷史記錄
func CleanOldHistory(days int) models.Response {
	res := models.Response{}
	res.Success = false

	// 計算截止日期
	cutoffDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	// 刪除舊記錄
	result := global.Mysql.Where("date < ?", cutoffDate).Delete(&entities.History{})
	if result.Error != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to clean old history: %s", result.Error.Error()))
		res.Msg = "Clean old history failed"
		return res
	}

	res.Success = true
	res.Msg = fmt.Sprintf("Cleaned %d old history records", result.RowsAffected)
	return res
}

// ArchiveOldHistory 將舊記錄歸檔到歷史表
func ArchiveOldHistory(days int) models.Response {
	res := models.Response{}
	res.Success = false

	// 計算截止日期
	cutoffDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	// 開始事務
	tx := global.Mysql.Begin()

	// 將舊記錄複製到歷史歸檔表
	archiveSQL := `
		INSERT INTO history_archives
		SELECT * FROM histories
		WHERE date < ?
	`
	if err := tx.Exec(archiveSQL, cutoffDate).Error; err != nil {
		tx.Rollback()
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to archive old history: %s", err.Error()))
		res.Msg = "Archive old history failed"
		return res
	}

	// 刪除已歸檔的記錄
	if err := tx.Where("date < ?", cutoffDate).Delete(&entities.History{}).Error; err != nil {
		tx.Rollback()
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to delete archived records: %s", err.Error()))
		res.Msg = "Delete archived records failed"
		return res
	}

	// 提交事務
	if err := tx.Commit().Error; err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to commit archive transaction: %s", err.Error()))
		res.Msg = "Commit archive transaction failed"
		return res
	}

	res.Success = true
	res.Msg = "History archive completed successfully"
	return res
}

// CreateDailyAggregates 生成每日統計匯總
func CreateDailyAggregates(targetDate string) models.Response {
	res := models.Response{}
	res.Success = false

	if targetDate == "" {
		targetDate = time.Now().AddDate(0, 0, -1).Format("2006-01-02") // 昨天
	}

	// 插入或更新每日統計
	aggregateSQL := `
		INSERT INTO history_daily_stats (
			date, logname, device_group,
			total_checks, online_count, offline_count, warning_count, error_count,
			uptime_rate, avg_response_time
		)
		SELECT
			date,
			logname,
			device_group,
			COUNT(*) as total_checks,
			SUM(CASE WHEN status = 'online' THEN 1 ELSE 0 END) as online_count,
			SUM(CASE WHEN status = 'offline' THEN 1 ELSE 0 END) as offline_count,
			SUM(CASE WHEN status = 'warning' THEN 1 ELSE 0 END) as warning_count,
			SUM(CASE WHEN status = 'error' THEN 1 ELSE 0 END) as error_count,
			ROUND(
				(SUM(CASE WHEN status = 'online' THEN 1 ELSE 0 END) * 100.0) / COUNT(*),
				2
			) as uptime_rate,
			ROUND(AVG(response_time), 2) as avg_response_time
		FROM histories
		WHERE date = ?
		GROUP BY date, logname, device_group
		ON DUPLICATE KEY UPDATE
			total_checks = VALUES(total_checks),
			online_count = VALUES(online_count),
			offline_count = VALUES(offline_count),
			warning_count = VALUES(warning_count),
			error_count = VALUES(error_count),
			uptime_rate = VALUES(uptime_rate),
			avg_response_time = VALUES(avg_response_time)
	`

	if err := global.Mysql.Exec(aggregateSQL, targetDate).Error; err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to create daily aggregates: %s", err.Error()))
		res.Msg = "Create daily aggregates failed"
		return res
	}

	res.Success = true
	res.Msg = fmt.Sprintf("Daily aggregates created for date: %s", targetDate)
	return res
}

// GetStorageStats 獲取存儲統計信息
func GetStorageStats() models.Response {
	res := models.Response{}
	res.Success = false

	type StorageStats struct {
		TableName    string  `json:"table_name"`
		RecordCount  int64   `json:"record_count"`
		DataSizeMB   float64 `json:"data_size_mb"`
		IndexSizeMB  float64 `json:"index_size_mb"`
		TotalSizeMB  float64 `json:"total_size_mb"`
		OldestRecord string  `json:"oldest_record"`
		NewestRecord string  `json:"newest_record"`
	}

	var stats []StorageStats

	// 查詢各表存儲統計
	statsSQL := `
		SELECT
			table_name,
			table_rows as record_count,
			data_length/1024/1024 as data_size_mb,
			index_length/1024/1024 as index_size_mb,
			(data_length + index_length)/1024/1024 as total_size_mb
		FROM information_schema.tables
		WHERE table_schema = 'logdetect'
			AND table_name IN ('histories', 'history_archives', 'history_daily_stats', 'alert_histories')
		ORDER BY total_size_mb DESC
	`

	if err := global.Mysql.Raw(statsSQL).Scan(&stats).Error; err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to get storage stats: %s", err.Error()))
		res.Msg = "Get storage stats failed"
		return res
	}

	// 獲取記錄時間範圍
	for i := range stats {
		var dateRange struct {
			Oldest string
			Newest string
		}

		tableName := stats[i].TableName
		if tableName == "histories" || tableName == "history_archives" {
			global.Mysql.Table(tableName).Select("MIN(date) as oldest, MAX(date) as newest").Scan(&dateRange)
		}

		stats[i].OldestRecord = dateRange.Oldest
		stats[i].NewestRecord = dateRange.Newest
	}

	res.Body = stats
	res.Success = true
	return res
}
