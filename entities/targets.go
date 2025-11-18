package entities

import (
	"log-detect/models"
)

type Target struct {
	models.Common
	ID      int    `gorm:"primaryKey;index" json:"id" form:"id"`
	Subject string `gorm:"type:varchar(50)" json:"subject" form:"subject"`
	To      to     `gorm:"serializer:json"  json:"to" form:"to"`
	Enable  bool   `json:"enable"`
	// Indices []Index `gorm:"foreignKey:TargetID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT;"`
	Indices []Index `gorm:"many2many:indices_targets;foreignKey:ID;reference:ID;" json:"indices"`
}

type Receiver struct {
	models.Common
	ID   int  `gorm:"primaryKey;index" json:"id" form:"id"`
	Name Name `gorm:"serializer:json" json:"name"`
}

type IndicesTargets struct {
	// models.Common
	TargetID int `gorm:"primaryKey" form:"target_id"`
	IndexID  int `gorm:"primaryKey" form:"index_id"`
}

type Name []string
type to []string

type Index struct {
	models.Common
	ID int `gorm:"primaryKey;index" json:"id" form:"id"`
	// TargetID    int    `gorm:"index" json:"target_id" form:"target_id"`
	Targets     []Target `gorm:"many2many:indices_targets;"`
	Pattern     string   `gorm:"type:varchar(50)" json:"pattern" form:"pattern"`
	DeviceGroup string   `gorm:"type:varchar(50)" json:"device_group" form:"device_group"`
	Logname     string   `gorm:"type:varchar(50)" json:"logname" form:"logname"`
	Period      string   `gorm:"type:varchar(50)" json:"period" form:"period"`
	Unit        int      `type:"int" json:"unit" form:"unit"`
	Field       string   `gorm:"type:varchar(50)" json:"field" form:"field"`
}

type Device struct {
	models.Common
	ID          int    `gorm:"primaryKey;index" json:"id" form:"id"`
	DeviceGroup string `gorm:"type:varchar(50)" json:"device_group" form:"device_group"`
	Name        string `gorm:"type:varchar(50)" json:"name" form:"name"`
}

type CronList struct {
	models.Common
	EntryID  int `gorm:"index" json:"entry_id" form:"entry_id"`
	TargetID int `gorm:"index" json:"target_id" form:"target_id"`
	IndexID  int `gorm:"index" json:"index_id" form:"index_id"`
}

type Table_counts struct {
	DeviceGroup  string `json:"device_group" form:"device_group"`
	DevicesCount int64  `json:"devices_count" form:"devices_count"`
}

type GroupName struct {
	DeviceGroup string `json:"device_group" form:"device_group"`
}

type History struct {
	models.Common
	// 基本信息
	Logname     string `gorm:"type:varchar(50);index" json:"logname" form:"logname"`
	DeviceGroup string `gorm:"type:varchar(50);index" json:"device_group" form:"device_group"`
	Name        string `gorm:"type:varchar(100);index" json:"name" form:"name"`
	TargetID    int    `gorm:"index" json:"target_id" form:"target_id"`
	IndexID     int    `gorm:"index" json:"index_id" form:"index_id"`

	// 檢查結果
	Status  string `gorm:"type:varchar(20)" json:"status" form:"status"` // "online", "offline", "warning", "error"
	Lost    string `gorm:"type:varchar(10)" json:"lost" form:"lost"`     // "true", "false"
	LostNum int    `gorm:"default:0" json:"lost_num" form:"lost_num"`

	// 時間信息
	Date      string `gorm:"type:varchar(10);index" json:"date" form:"date"`           // YYYY-MM-DD
	Time      string `gorm:"type:varchar(8);index" json:"time" form:"time"`            // HH:MM:SS
	DateTime  string `gorm:"type:varchar(19);index" json:"date_time" form:"date_time"` // YYYY-MM-DD HH:MM:SS
	Timestamp int64  `gorm:"index" json:"timestamp" form:"timestamp"`                  // Unix timestamp

	// 檢查配置
	Period string `gorm:"type:varchar(20)" json:"period" form:"period"`
	Unit   int    `gorm:"default:1" json:"unit" form:"unit"`

	// 性能指標
	ResponseTime int64 `gorm:"default:0" json:"response_time" form:"response_time"` // 響應時間(毫秒)
	DataCount    int64 `gorm:"default:0" json:"data_count" form:"data_count"`       // 檢查到的數據量

	// 錯誤信息
	ErrorMsg  string `gorm:"type:text" json:"error_msg" form:"error_msg"`
	ErrorCode string `gorm:"type:varchar(50)" json:"error_code" form:"error_code"`

	// 額外元數據
	Metadata string `gorm:"type:json" json:"metadata" form:"metadata"` // JSON 格式的額外信息
}

type Logname struct {
	Logname string `json:"logname" form:"logname"`
}

type HistoryData struct {
	Name string `json:"name" form:"name"`
	Time string `json:"time" form:"time"`
	Lost string `json:"lost" form:"lost"`
}

// 統計和視覺化相關實體

// HistoryStatistics 歷史統計數據
type HistoryStatistics struct {
	Date            string  `json:"date"`
	Logname         string  `json:"logname"`
	DeviceGroup     string  `json:"device_group"`
	TotalChecks     int64   `json:"total_checks"`
	OnlineCount     int64   `json:"online_count"`
	OfflineCount    int64   `json:"offline_count"`
	WarningCount    int64   `json:"warning_count"`
	ErrorCount      int64   `json:"error_count"`
	UptimeRate      float64 `json:"uptime_rate"` // 在線率百分比
	AvgResponseTime int64   `json:"avg_response_time"`
}

// DeviceTimeline 設備時間線數據
type DeviceTimeline struct {
	DeviceName string          `json:"device_name"`
	Logname    string          `json:"logname"`
	TimePoints []TimelinePoint `json:"time_points"`
}

// TimelinePoint 時間點數據
type TimelinePoint struct {
	Timestamp    int64  `json:"timestamp"`
	Status       string `json:"status"` // online, offline, warning, error
	ResponseTime int64  `json:"response_time"`
	DataCount    int64  `json:"data_count"`
	ErrorMsg     string `json:"error_msg,omitempty"`
}

// GroupStatistics 群組統計
type GroupStatistics struct {
	DeviceGroup    string  `json:"device_group"`
	Logname        string  `json:"logname"`
	TotalDevices   int64   `json:"total_devices"`
	OnlineDevices  int64   `json:"online_devices"`
	OfflineDevices int64   `json:"offline_devices"`
	UptimeRate     float64 `json:"uptime_rate"`
	LastCheckTime  string  `json:"last_check_time"`
}

// TrendData 趨勢數據
type TrendData struct {
	Date            string  `json:"date"`
	UptimeRate      float64 `json:"uptime_rate"`
	OfflineCount    int64   `json:"offline_count"`
	ErrorCount      int64   `json:"error_count"`
	AvgResponseTime int64   `json:"avg_response_time"`
}

// AlertHistory 告警歷史
type AlertHistory struct {
	models.Common
	Logname     string `gorm:"type:varchar(50);index" json:"logname"`
	DeviceGroup string `gorm:"type:varchar(50);index" json:"device_group"`
	DeviceName  string `gorm:"type:varchar(100);index" json:"device_name"`
	AlertType   string `gorm:"type:varchar(20)" json:"alert_type"` // offline, error, warning
	Severity    string `gorm:"type:varchar(10)" json:"severity"`   // low, medium, high, critical
	Message     string `gorm:"type:text" json:"message"`
	Status      string `gorm:"type:varchar(20)" json:"status"` // active, resolved, acknowledged
	ResolvedAt  *int64 `json:"resolved_at,omitempty"`
	ResolvedBy  string `gorm:"type:varchar(100)" json:"resolved_by,omitempty"`
}

// HistoryArchive 歷史記錄歸檔表
type HistoryArchive struct {
	models.Common
	// 與 History 實體完全相同，用於存儲已歸檔的歷史記錄
	Logname      string `gorm:"type:varchar(50);index" json:"logname" form:"logname"`
	DeviceGroup  string `gorm:"type:varchar(50);index" json:"device_group" form:"device_group"`
	Name         string `gorm:"type:varchar(100);index" json:"name" form:"name"`
	TargetID     int    `gorm:"index" json:"target_id" form:"target_id"`
	IndexID      int    `gorm:"index" json:"index_id" form:"index_id"`
	Status       string `gorm:"type:varchar(20)" json:"status" form:"status"`
	Lost         string `gorm:"type:varchar(10)" json:"lost" form:"lost"`
	LostNum      int    `gorm:"default:0" json:"lost_num" form:"lost_num"`
	Date         string `gorm:"type:varchar(10);index" json:"date" form:"date"`
	Time         string `gorm:"type:varchar(8);index" json:"time" form:"time"`
	DateTime     string `gorm:"type:varchar(19);index" json:"date_time" form:"date_time"`
	Timestamp    int64  `gorm:"index" json:"timestamp" form:"timestamp"`
	Period       string `gorm:"type:varchar(20)" json:"period" form:"period"`
	Unit         int    `gorm:"default:1" json:"unit" form:"unit"`
	ResponseTime int64  `gorm:"default:0" json:"response_time" form:"response_time"`
	DataCount    int64  `gorm:"default:0" json:"data_count" form:"data_count"`
	ErrorMsg     string `gorm:"type:text" json:"error_msg" form:"error_msg"`
	ErrorCode    string `gorm:"type:varchar(50)" json:"error_code" form:"error_code"`
	Metadata     string `gorm:"type:json" json:"metadata" form:"metadata"`
}

// HistoryDailyStats 每日統計匯總表
type HistoryDailyStats struct {
	models.Common
	Date            string  `gorm:"type:varchar(10);primaryKey" json:"date"`
	Logname         string  `gorm:"type:varchar(50);primaryKey" json:"logname"`
	DeviceGroup     string  `gorm:"type:varchar(50);primaryKey" json:"device_group"`
	TotalChecks     int64   `gorm:"default:0" json:"total_checks"`
	OnlineCount     int64   `gorm:"default:0" json:"online_count"`
	OfflineCount    int64   `gorm:"default:0" json:"offline_count"`
	WarningCount    int64   `gorm:"default:0" json:"warning_count"`
	ErrorCount      int64   `gorm:"default:0" json:"error_count"`
	UptimeRate      float64 `gorm:"type:decimal(5,2);default:0.00" json:"uptime_rate"`
	AvgResponseTime float64 `gorm:"type:decimal(10,2);default:0.00" json:"avg_response_time"`
}

// DashboardData 儀表板數據
type DashboardData struct {
	TotalTargets   int64   `json:"total_targets"`
	ActiveTargets  int64   `json:"active_targets"`
	TotalDevices   int64   `json:"total_devices"`
	OnlineDevices  int64   `json:"online_devices"`
	OfflineDevices int64   `json:"offline_devices"`
	UptimeRate     float64 `json:"uptime_rate"`
	ActiveAlerts   int64   `json:"active_alerts"`
	LastUpdateTime string  `json:"last_update_time"`
}

// 兼容性實體
type LognameCheck struct {
	Name string `json:"name" form:"name"`
	Lost string `json:"lost" form:"lost"`
}

type MailHistory struct {
	models.Common
	Date    string `gorm:"type:varchar(50)" json:"date" form:"date"`
	Time    string `gorm:"type:varchar(50)" json:"time" form:"time"`
	Logname string `gorm:"type:varchar(50)" json:"logname" form:"logname"`
	Sended  bool   `gorm:"type:boolean" json:"sended" form:"sended"`
}

// User represents a system user
type User struct {
	models.Common
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Username string `gorm:"uniqueIndex;type:varchar(50);not null" json:"username" form:"username"`
	Email    string `gorm:"uniqueIndex;type:varchar(100);not null" json:"email" form:"email"`
	Password string `gorm:"type:varchar(255);not null" json:"-" form:"password"` // Never return password in JSON
	Role     Role   `gorm:"foreignKey:RoleID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"role"`
	RoleID   uint   `gorm:"not null" json:"role_id" form:"role_id"`
	IsActive bool   `gorm:"default:true" json:"is_active" form:"is_active"`
}

// Role represents a user role with permissions
type Role struct {
	models.Common
	ID          uint         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string       `gorm:"uniqueIndex;type:varchar(50);not null" json:"name" form:"name"`
	Description string       `gorm:"type:varchar(200)" json:"description" form:"description"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
}

// Permission represents a specific permission
type Permission struct {
	models.Common
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"uniqueIndex;type:varchar(100);not null" json:"name" form:"name"`
	Resource    string `gorm:"type:varchar(100);not null" json:"resource" form:"resource"` // e.g., "device", "target", "user"
	Action      string `gorm:"type:varchar(50);not null" json:"action" form:"action"`      // e.g., "create", "read", "update", "delete"
	Description string `gorm:"type:varchar(200)" json:"description" form:"description"`
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// AuthClaims represents JWT claims
type AuthClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	RoleID   uint   `json:"role_id"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

// TableName specifies the table name for Role
func (Role) TableName() string {
	return "roles"
}

// TableName specifies the table name for Permission
func (Permission) TableName() string {
	return "permissions"
}
