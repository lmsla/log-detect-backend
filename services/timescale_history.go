package services

import (
	"fmt"
	"log-detect/entities"
	"log-detect/global"
	"log-detect/log"
	"log-detect/models"
	"time"
)

// GetHistoryDataByDeviceName_TS 從 TimescaleDB 查詢設備歷史 (替代 MySQL 版本)
func GetHistoryDataByDeviceName_TS(logname string, name string) []entities.History {
	histories := []entities.History{}
	date := time.Now().Format("2006-01-02")

	query := `
		SELECT device_id, device_group, logname, status,
		       CASE WHEN lost THEN 'true' ELSE 'false' END as lost,
		       lost_num, date, hour_time, date_time, timestamp_unix,
		       period, unit, COALESCE(target_id, 0), COALESCE(index_id, 0),
		       response_time, data_count,
		       COALESCE(error_msg, '') as error_msg,
		       COALESCE(error_code, '') as error_code
		FROM device_metrics
		WHERE logname = $1 AND device_id = $2 AND date = $3
		ORDER BY time DESC
	`

	rows, err := global.TimescaleDB.Query(query, logname, name, date)
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Get History Data By DeviceName error: %s", err.Error()))
		return histories
	}
	defer rows.Close()

	for rows.Next() {
		var h entities.History
		err := rows.Scan(
			&h.Name, &h.DeviceGroup, &h.Logname, &h.Status,
			&h.Lost, &h.LostNum, &h.Date, &h.Time, &h.DateTime, &h.Timestamp,
			&h.Period, &h.Unit, &h.TargetID, &h.IndexID, &h.ResponseTime, &h.DataCount,
			&h.ErrorMsg, &h.ErrorCode,
		)
		if err != nil {
			log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Scan history data error: %s", err.Error()))
			continue
		}
		histories = append(histories, h)
	}

	return histories
}

// GetLognameData_TS 從 TimescaleDB 獲取日誌名稱數據
func GetLognameData_TS() models.Response {
	res := models.Response{}
	res.Success = false

	// 查詢所有不同的 logname
	query := `SELECT DISTINCT logname FROM device_metrics ORDER BY logname`

	rows, err := global.TimescaleDB.Query(query)
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("failed to fetch lognames error: %s", err.Error()))
		res.Msg = "Query failed"
		return res
	}
	defer rows.Close()

	var lognames []string
	for rows.Next() {
		var logname string
		if err := rows.Scan(&logname); err != nil {
			continue
		}
		lognames = append(lognames, logname)
	}

	// 檢查每個 logname 的狀態
	checkResults := []entities.LognameCheck{}
	for _, name := range lognames {
		checkResult := CheckLogstatus_TS(name)
		checkResults = append(checkResults, checkResult)
	}

	res.Body = checkResults
	res.Success = true
	return res
}

// CheckLogstatus_TS 檢查日誌狀態 (TimescaleDB 版本)
func CheckLogstatus_TS(logname string) entities.LognameCheck {
	var indices entities.Index

	// 從 MySQL 查詢 indices 配置
	index_err := global.Mysql.Where("logname = ?", logname).Find(&indices).Error
	if index_err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("find indices data error: %s", index_err.Error()))
	}

	now := time.Now()
	lastCrontabTime := GetLastCrontabTime(now, indices.Period, indices.Unit)
	date := now.Format("2006-01-02")

	// 從 TimescaleDB 查詢歷史記錄
	query := `
		SELECT COUNT(*)
		FROM device_metrics
		WHERE logname = $1 AND date = $2 AND hour_time = $3
	`

	var count int
	err := global.TimescaleDB.QueryRow(query, logname, date, lastCrontabTime).Scan(&count)
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("check log status error: %s", err.Error()))
		return entities.LognameCheck{Name: logname, Lost: "true"}
	}

	if count == 0 {
		return entities.LognameCheck{Name: logname, Lost: "true"}
	}
	return entities.LognameCheck{Name: logname, Lost: "false"}
}

// GetHistoryStatistics_TS TimescaleDB 高性能統計查詢
func GetHistoryStatistics_TS(logname, deviceGroup string, startDate, endDate string) models.Response {
	res := models.Response{}
	res.Success = false

	var statistics []entities.HistoryStatistics

	// 使用 TimescaleDB 的高性能聚合
	query := `
		SELECT
			date,
			logname,
			device_group,
			COUNT(*) as total_checks,
			COUNT(*) FILTER (WHERE status = 'online') as online_count,
			COUNT(*) FILTER (WHERE status = 'offline') as offline_count,
			COUNT(*) FILTER (WHERE status = 'warning') as warning_count,
			COUNT(*) FILTER (WHERE status = 'error') as error_count,
			ROUND(AVG(response_time), 2) as avg_response_time,
			ROUND(
				(COUNT(*) FILTER (WHERE status = 'online')::DECIMAL / NULLIF(COUNT(*), 0)) * 100,
				2
			) as uptime_rate
		FROM device_metrics
		WHERE 1=1
	`

	args := []any{}
	argIndex := 1

	if logname != "" {
		query += fmt.Sprintf(" AND logname = $%d", argIndex)
		args = append(args, logname)
		argIndex++
	}
	if deviceGroup != "" {
		query += fmt.Sprintf(" AND device_group = $%d", argIndex)
		args = append(args, deviceGroup)
		argIndex++
	}
	if startDate != "" {
		query += fmt.Sprintf(" AND date >= $%d", argIndex)
		args = append(args, startDate)
		argIndex++
	}
	if endDate != "" {
		query += fmt.Sprintf(" AND date <= $%d", argIndex)
		args = append(args, endDate)
	}

	query += " GROUP BY date, logname, device_group ORDER BY date DESC"

	rows, err := global.TimescaleDB.Query(query, args...)
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to get history statistics: %s", err.Error()))
		res.Msg = "Query failed"
		return res
	}
	defer rows.Close()

	for rows.Next() {
		var stat entities.HistoryStatistics
		err := rows.Scan(
			&stat.Date, &stat.Logname, &stat.DeviceGroup,
			&stat.TotalChecks, &stat.OnlineCount, &stat.OfflineCount,
			&stat.WarningCount, &stat.ErrorCount, &stat.AvgResponseTime, &stat.UptimeRate,
		)
		if err != nil {
			log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Scan statistics error: %s", err.Error()))
			continue
		}
		statistics = append(statistics, stat)
	}

	res.Body = statistics
	res.Success = true
	return res
}

// GetDeviceTimeline_TS 從 TimescaleDB 獲取設備時間線
func GetDeviceTimeline_TS(deviceName, logname string, days int) models.Response {
	res := models.Response{}
	res.Success = false

	var timeline entities.DeviceTimeline
	timeline.DeviceName = deviceName
	timeline.Logname = logname

	// 計算開始日期
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	query := `
		SELECT timestamp_unix, status, response_time, data_count, COALESCE(error_msg, '')
		FROM device_metrics
		WHERE device_id = $1 AND logname = $2 AND date >= $3
		ORDER BY time ASC
	`

	rows, err := global.TimescaleDB.Query(query, deviceName, logname, startDate)
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to get device timeline: %s", err.Error()))
		res.Msg = "Query failed"
		return res
	}
	defer rows.Close()

	for rows.Next() {
		var point entities.TimelinePoint
		err := rows.Scan(&point.Timestamp, &point.Status, &point.ResponseTime, &point.DataCount, &point.ErrorMsg)
		if err != nil {
			log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Scan timeline error: %s", err.Error()))
			continue
		}
		timeline.TimePoints = append(timeline.TimePoints, point)
	}

	res.Body = timeline
	res.Success = true
	return res
}

// GetTrendData_TS 從 TimescaleDB 獲取趨勢數據
func GetTrendData_TS(logname, deviceGroup string, days int) models.Response {
	res := models.Response{}
	res.Success = false

	var trends []entities.TrendData
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	query := `
		SELECT
			date,
			ROUND(AVG(CASE WHEN NOT lost THEN 100 ELSE 0 END), 2) as uptime_rate,
			SUM(CASE WHEN lost THEN 1 ELSE 0 END) as offline_count,
			SUM(CASE WHEN status = 'error' THEN 1 ELSE 0 END) as error_count,
			ROUND(AVG(response_time), 0) as avg_response_time
		FROM device_metrics
		WHERE date >= $1
	`

	args := []any{startDate}
	argIndex := 2

	if logname != "" {
		query += fmt.Sprintf(" AND logname = $%d", argIndex)
		args = append(args, logname)
		argIndex++
	}
	if deviceGroup != "" {
		query += fmt.Sprintf(" AND device_group = $%d", argIndex)
		args = append(args, deviceGroup)
	}

	query += " GROUP BY date ORDER BY date ASC"

	rows, err := global.TimescaleDB.Query(query, args...)
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to get trend data: %s", err.Error()))
		res.Msg = "Query failed"
		return res
	}
	defer rows.Close()

	for rows.Next() {
		var trend entities.TrendData
		err := rows.Scan(&trend.Date, &trend.UptimeRate, &trend.OfflineCount, &trend.ErrorCount, &trend.AvgResponseTime)
		if err != nil {
			log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Scan trend error: %s", err.Error()))
			continue
		}
		trends = append(trends, trend)
	}

	res.Body = trends
	res.Success = true
	return res
}

// GetGroupStatistics_TS 從 TimescaleDB 獲取群組統計
func GetGroupStatistics_TS(logname string) models.Response {
	res := models.Response{}
	res.Success = false

	var statistics []entities.GroupStatistics
	today := time.Now().Format("2006-01-02")

	query := `
		SELECT
			d.device_group,
			$1 as logname,
			COUNT(DISTINCT d.id) as total_devices,
			COUNT(DISTINCT CASE WHEN h.lost = false THEN h.device_id END) as online_devices,
			COUNT(DISTINCT CASE WHEN h.lost = true THEN h.device_id END) as offline_devices,
			MAX(h.date_time) as last_check_time,
			ROUND(
				(COUNT(DISTINCT CASE WHEN h.lost = false THEN h.device_id END)::DECIMAL /
				 NULLIF(COUNT(DISTINCT d.id), 0)) * 100,
				2
			) as uptime_rate
		FROM devices d
		LEFT JOIN device_metrics h ON d.name = h.device_id AND h.logname = $1 AND h.date = $2
		GROUP BY d.device_group
	`

	rows, err := global.TimescaleDB.Query(query, logname, today)
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to get group statistics: %s", err.Error()))
		res.Msg = "Query failed"
		return res
	}
	defer rows.Close()

	for rows.Next() {
		var stat entities.GroupStatistics
		var lastCheckTime *string
		err := rows.Scan(
			&stat.DeviceGroup, &stat.Logname, &stat.TotalDevices,
			&stat.OnlineDevices, &stat.OfflineDevices, &lastCheckTime, &stat.UptimeRate,
		)
		if err != nil {
			log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Scan group statistics error: %s", err.Error()))
			continue
		}
		if lastCheckTime != nil {
			stat.LastCheckTime = *lastCheckTime
		}
		statistics = append(statistics, stat)
	}

	res.Body = statistics
	res.Success = true
	return res
}

// GetDashboardData_TS 從 TimescaleDB 獲取儀表板數據
func GetDashboardData_TS() models.Response {
	res := models.Response{}
	res.Success = false

	var dashboard entities.DashboardData
	today := time.Now().Format("2006-01-02")

	// 從 MySQL 獲取目標統計
	if err := global.Mysql.Model(&entities.Target{}).Count(&dashboard.TotalTargets).Error; err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to count targets: %s", err.Error()))
		return res
	}

	if err := global.Mysql.Model(&entities.Target{}).Where("enable = ?", true).Count(&dashboard.ActiveTargets).Error; err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to count active targets: %s", err.Error()))
		return res
	}

	// 從 MySQL 獲取總設備數
	if err := global.Mysql.Model(&entities.Device{}).Count(&dashboard.TotalDevices).Error; err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to count devices: %s", err.Error()))
		return res
	}

	// 從 TimescaleDB 獲取在線設備數
	query := `
		SELECT COUNT(DISTINCT device_id)
		FROM device_metrics
		WHERE date = $1 AND lost = false
	`

	err := global.TimescaleDB.QueryRow(query, today).Scan(&dashboard.OnlineDevices)
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to count online devices: %s", err.Error()))
		return res
	}

	dashboard.OfflineDevices = dashboard.TotalDevices - dashboard.OnlineDevices

	// 計算在線率
	if dashboard.TotalDevices > 0 {
		dashboard.UptimeRate = float64(dashboard.OnlineDevices) / float64(dashboard.TotalDevices) * 100
	}

	// 從 MySQL 獲取活躍告警數
	if err := global.Mysql.Model(&entities.AlertHistory{}).Where("status = ?", "active").Count(&dashboard.ActiveAlerts).Error; err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to count active alerts: %s", err.Error()))
		return res
	}

	dashboard.LastUpdateTime = time.Now().Format("2006-01-02 15:04:05")

	res.Body = dashboard
	res.Success = true
	return res
}
