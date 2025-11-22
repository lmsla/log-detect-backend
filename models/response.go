package models

import "time"

type Response struct {
	Success bool   `json:"success" form:"success"`
	Msg     string `json:"msg" form:"msg"`
	Body    interface{}
}
// ESAlertQueryParams ES 告警查詢參數
type ESAlertQueryParams struct {
	Status    []string  `form:"status[]" json:"status"`       // active, resolved, acknowledged
	Severity  []string  `form:"severity[]" json:"severity"`   // critical, high, medium, low
	AlertType []string  `form:"alert_type[]" json:"alert_type"` // health, performance, capacity
	MonitorID int       `form:"monitor_id" json:"monitor_id"` // 監控器 ID
	StartTime time.Time `form:"start_time" json:"start_time"` // 開始時間
	EndTime   time.Time `form:"end_time" json:"end_time"`     // 結束時間
	Page      int       `form:"page" json:"page"`             // 頁碼（預設 1）
	PageSize  int       `form:"page_size" json:"page_size"`   // 每頁筆數（預設 20）
}

// ESAlertStatistics ES 告警統計
type ESAlertStatistics struct {
	Total    int `json:"total"`
	Active   int `json:"active"`
	Resolved int `json:"resolved"`
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
}
