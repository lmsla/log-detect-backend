package services

import (
	"fmt"
	"log-detect/entities"
	"log-detect/global"
	"log-detect/log"
	"log-detect/models"
	"strings"
	"time"
)

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
	normalized := strings.ToLower(strings.TrimSpace(logname))
	err := global.Mysql.Where("LOWER(logname) = ?", normalized).Find(&indices).Error
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("find devices data error: %s", err.Error()))
		return indices, err
	}
	return indices, nil
}

// 以 logname , device name 查詢歷史紀錄
// hours = 0 表示查詢當天全部；hours > 0 表示查詢最近 N 小時
func GetHistoryDataByDeviceName(logname string, name string, hours int) []entities.History {
	return GetHistoryDataByDeviceName_TS(logname, name, hours)
}

// GenerateTimeArray 產生時間點陣列，用於前端格線補全
// startTime 為起始時間（查詢時間範圍起點），0 表示從當天 00:00 開始
func GenerateTimeArray(period string, unit int, startTime time.Time) []string {
	var timeArray []string

	now := time.Now()

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

	// 將 startTime 對齊到最近的 crontab 時間點（向下取整）
	alignedStart := alignToCrontab(startTime, duration)

	// 從 alignedStart 開始，根據時間間隔生成時間點陣列
	for t := alignedStart; t.Before(now) || t.Equal(now); t = t.Add(duration) {
		timeArray = append(timeArray, t.Format("15:04"))
	}

	return timeArray
}

// alignToCrontab 將時間向下對齊到最近的 crontab 時間點
func alignToCrontab(t time.Time, duration time.Duration) time.Time {
	startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	elapsed := t.Sub(startOfDay)
	aligned := elapsed.Truncate(duration)
	return startOfDay.Add(aligned)
}

// lostSeverity 回傳 lost 狀態的嚴重程度，數字越大越嚴重
// 用於同一時間點有多筆 row 時，保留最嚴重的狀態
func lostSeverity(lost string) int {
	switch lost {
	case "true":
		return 2
	case "false":
		return 1
	default: // "none"
		return 0
	}
}

// 處理 history data
// hours = 0 表示查詢當天全部；hours = 1/3/6 表示查詢最近 N 小時
func DataDealing(logname string, hours int) models.Response {
	logname = strings.TrimSpace(logname)

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

	// 決定時間範圍起始點
	var startTime time.Time
	if hours > 0 {
		startTime = time.Now().Add(-time.Duration(hours) * time.Hour)
	} else {
		// 全天：從當天 00:00 開始
		now := time.Now()
		startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	}

	timeArray := GenerateTimeArray(indicesData.Period, indicesData.Unit, startTime)

	for _, device := range device_list {
		history_data := GetHistoryDataByDeviceName(logname, device.Name, hours)

		// deduplicatedMap: key=hour_time, value=最嚴重的 lost 狀態
		// TimescaleDB 每次偵測都寫一行，同一分鐘可能有多筆 row，
		// 合併時優先保留 lost:"true"（lost > false > none）
		deduplicatedMap := make(map[string]string) // time -> lost
		for _, data := range history_data {
			existing, seen := deduplicatedMap[data.Time]
			if !seen || lostSeverity(data.Lost) > lostSeverity(existing) {
				deduplicatedMap[data.Time] = data.Lost
			}
		}

		var history_tmp_data []entities.HistoryData
		for timePoint, lost := range deduplicatedMap {
			history_tmp_data = append(history_tmp_data, entities.HistoryData{Name: device.Name, Time: timePoint, Lost: lost})
		}

		// Gap-fill：timeArray 中沒有資料的時間點補 "none"
		for _, timePoint := range timeArray {
			if _, ok := deduplicatedMap[timePoint]; !ok {
				history_tmp_data = append(history_tmp_data, entities.HistoryData{Name: device.Name, Time: timePoint, Lost: "none"})
			}
		}

		history_final_data = append(history_final_data, history_tmp_data...)
	}

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
