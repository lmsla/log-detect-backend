package entities

import (
	"encoding/json"
	"log-detect/models"
	"time"
)

// ElasticsearchMonitor ES 監控配置 (存儲在 MySQL)
type ElasticsearchMonitor struct {
	models.Common
	ID                int      `gorm:"primaryKey;index" json:"id" form:"id"`
	Name              string   `json:"name" gorm:"type:varchar(100);not null;comment:監控名稱"`
	Host              string   `json:"host" gorm:"type:varchar(255);not null;comment:ES 主機地址"`
	Port              int      `json:"port" gorm:"type:int;not null;default:9200;comment:ES 端口"`
	Username          string   `json:"username" gorm:"type:varchar(100);comment:認證用戶名"`
	Password          string   `json:"password" gorm:"type:varchar(255);comment:認證密碼"`
	EnableAuth        bool     `json:"enable_auth" gorm:"type:tinyint(1);default:0;comment:是否啟用認證"`
	CheckType         string   `json:"check_type" gorm:"type:varchar(100);default:'health,performance';comment:檢查類型(逗號分隔)"`
	Interval          int      `json:"interval" gorm:"type:int;not null;default:60;comment:檢查間隔(秒,範圍:10-3600)"`
	EnableMonitor     bool     `json:"enable_monitor" gorm:"type:tinyint(1);default:1;comment:是否啟用監控"`
	Receivers         []string `json:"receivers" gorm:"type:json;serializer:json;comment:告警收件人陣列"`
	Subject           string   `json:"subject" gorm:"type:varchar(255);comment:告警主題"`
	Description       string   `json:"description" gorm:"type:text;comment:監控描述"`

	// 告警閾值配置（獨立欄位，前端友好）
	CPUUsageHigh            *float64 `json:"cpu_usage_high" gorm:"type:decimal(5,2);comment:CPU使用率-高閾值(%)"`
	CPUUsageCritical        *float64 `json:"cpu_usage_critical" gorm:"type:decimal(5,2);comment:CPU使用率-危險閾值(%)"`
	MemoryUsageHigh         *float64 `json:"memory_usage_high" gorm:"type:decimal(5,2);comment:記憶體使用率-高閾值(%)"`
	MemoryUsageCritical     *float64 `json:"memory_usage_critical" gorm:"type:decimal(5,2);comment:記憶體使用率-危險閾值(%)"`
	DiskUsageHigh           *float64 `json:"disk_usage_high" gorm:"type:decimal(5,2);comment:磁碟使用率-高閾值(%)"`
	DiskUsageCritical       *float64 `json:"disk_usage_critical" gorm:"type:decimal(5,2);comment:磁碟使用率-危險閾值(%)"`
	ResponseTimeHigh        *int64   `json:"response_time_high" gorm:"type:bigint;comment:響應時間-高閾值(ms)"`
	ResponseTimeCritical    *int64   `json:"response_time_critical" gorm:"type:bigint;comment:響應時間-危險閾值(ms)"`
	UnassignedShardsThreshold *int   `json:"unassigned_shards_threshold" gorm:"type:int;comment:未分配分片閾值"`

	// 保留 JSON 欄位作為高級配置選項（向後兼容）
	AlertThreshold    string   `json:"alert_threshold" gorm:"type:json;comment:告警閾值配置(JSON,高級選項)"`
	AlertDedupeWindow int      `json:"alert_dedupe_window" gorm:"type:int;default:300;comment:告警去重時間窗口(秒,預設300秒=5分鐘)"`
}

// TableName 指定表名
func (ElasticsearchMonitor) TableName() string {
	return "elasticsearch_monitors"
}

// GetAlertThreshold 獲取告警閾值（優先使用獨立欄位，回退到 JSON 或預設值）
func (m *ElasticsearchMonitor) GetAlertThreshold() ESAlertThreshold {
	threshold := DefaultESAlertThreshold()

	// 優先使用獨立欄位
	if m.CPUUsageHigh != nil {
		threshold.CPUUsageHigh = *m.CPUUsageHigh
	}
	if m.CPUUsageCritical != nil {
		threshold.CPUUsageCritical = *m.CPUUsageCritical
	}
	if m.MemoryUsageHigh != nil {
		threshold.MemoryUsageHigh = *m.MemoryUsageHigh
	}
	if m.MemoryUsageCritical != nil {
		threshold.MemoryUsageCritical = *m.MemoryUsageCritical
	}
	if m.DiskUsageHigh != nil {
		threshold.DiskUsageHigh = *m.DiskUsageHigh
	}
	if m.DiskUsageCritical != nil {
		threshold.DiskUsageCritical = *m.DiskUsageCritical
	}
	if m.ResponseTimeHigh != nil {
		threshold.ResponseTimeHigh = *m.ResponseTimeHigh
	}
	if m.ResponseTimeCritical != nil {
		threshold.ResponseTimeCritical = *m.ResponseTimeCritical
	}
	if m.UnassignedShardsThreshold != nil {
		threshold.UnassignedShards = *m.UnassignedShardsThreshold
	}

	// 如果獨立欄位都沒設置，嘗試解析 AlertThreshold JSON（向後兼容）
	if m.AlertThreshold != "" && m.CPUUsageHigh == nil {
		// 只有當獨立欄位未設置時才使用 JSON 配置
		var jsonThreshold ESAlertThreshold
		if err := json.Unmarshal([]byte(m.AlertThreshold), &jsonThreshold); err == nil {
			return jsonThreshold
		}
	}

	return threshold
}

// ESMetric ES 監控指標 (存儲在 TimescaleDB)
type ESMetric struct {
	Time               time.Time `json:"time"`
	MonitorID          int       `json:"monitor_id"`
	Status             string    `json:"status"` // online, offline, warning, error
	ClusterName        string    `json:"cluster_name"`
	ClusterStatus      string    `json:"cluster_status"` // green, yellow, red
	ResponseTime       int64     `json:"response_time"`  // 毫秒
	CPUUsage           float64   `json:"cpu_usage"`      // 百分比
	MemoryUsage        float64   `json:"memory_usage"`   // 百分比
	DiskUsage          float64   `json:"disk_usage"`     // 百分比
	NodeCount          int       `json:"node_count"`
	DataNodeCount      int       `json:"data_node_count"`
	QueryLatency       int64     `json:"query_latency"`     // 毫秒
	IndexingRate       float64   `json:"indexing_rate"`     // 索引並發數（index_current，非速率）
	SearchRate         float64   `json:"search_rate"`       // 搜尋並發數（query_current，非速率）
	TotalIndices       int       `json:"total_indices"`     // 索引總數
	TotalDocuments     int64     `json:"total_documents"`   // 文檔總數
	TotalSizeBytes     int64     `json:"total_size_bytes"`  // 總大小(字節)
	ActiveShards       int       `json:"active_shards"`     // 活躍分片數
	RelocatingShards   int       `json:"relocating_shards"` // 遷移中分片數
	UnassignedShards   int       `json:"unassigned_shards"` // 未分配分片數
	ErrorMessage       string    `json:"error_message"`
	WarningMessage     string    `json:"warning_message"`
	Metadata           string    `json:"metadata"` // JSON 格式的額外元數據
}

// ESAlert ES 告警記錄 (存儲在 TimescaleDB alert_history 表)
type ESAlert struct {
	Time            time.Time  `json:"time"`
	MonitorID       int        `json:"monitor_id"`
	AlertType       string     `json:"alert_type"` // health, performance, capacity, availability
	Severity        string     `json:"severity"`   // critical, high, medium, low
	Message         string     `json:"message"`
	Status          string     `json:"status"` // active, resolved, acknowledged
	ClusterName     string     `json:"cluster_name,omitempty"`
	ThresholdValue  *float64   `json:"threshold_value,omitempty"`  // 觸發閾值
	ActualValue     *float64   `json:"actual_value,omitempty"`     // 實際值
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy      string     `json:"resolved_by,omitempty"`
	ResolutionNote  string     `json:"resolution_note,omitempty"`
	AcknowledgedAt  *time.Time `json:"acknowledged_at,omitempty"`
	AcknowledgedBy  string     `json:"acknowledged_by,omitempty"`
	Metadata        string     `json:"metadata,omitempty"` // JSONB 存儲額外資訊
}

// ESAlertHistory ESAlert 的別名（用於服務層）
type ESAlertHistory = ESAlert

// ESMonitorStatus ES 監控狀態摘要 (用於 API 回應)
type ESMonitorStatus struct {
	MonitorID        int       `json:"monitor_id"`
	MonitorName      string    `json:"monitor_name"`
	Host             string    `json:"host"`
	Status           string    `json:"status"` // online, offline, warning, error
	ClusterStatus    string    `json:"cluster_status"` // green, yellow, red
	ClusterName      string    `json:"cluster_name"`
	ResponseTime     int64     `json:"response_time"` // 響應時間（單位：毫秒）
	CPUUsage         float64   `json:"cpu_usage"` // CPU 使用率（單位：百分比 0-100）
	MemoryUsage      float64   `json:"memory_usage"` // 記憶體使用率（單位：百分比 0-100）
	DiskUsage        float64   `json:"disk_usage"` // 磁碟使用率（單位：百分比 0-100）
	NodeCount        int       `json:"node_count"`
	ActiveShards     int       `json:"active_shards"`
	RelocatingShards int       `json:"relocating_shards"` // 遷移中的分片數
	UnassignedShards int       `json:"unassigned_shards"`
	LastCheckTime    time.Time `json:"last_check_time"` // ISO 8601 格式
	ErrorMessage     string    `json:"error_message,omitempty"`
	WarningMessage   string    `json:"warning_message,omitempty"`
}

// ESMetricTimeSeries ES 指標時序數據 (用於圖表)
type ESMetricTimeSeries struct {
	Time           time.Time `json:"time"`
	CPUUsage       float64   `json:"cpu_usage"`
	MemoryUsage    float64   `json:"memory_usage"`
	DiskUsage      float64   `json:"disk_usage"`
	ResponseTime   int64     `json:"response_time"`
	IndexingRate   float64   `json:"indexing_rate"`
	SearchRate     float64   `json:"search_rate"`
	ActiveShards   int       `json:"active_shards"`
	UnassignedShards int     `json:"unassigned_shards"`
}

// ESStatistics ES 統計數據 (用於儀表板)
type ESStatistics struct {
	TotalMonitors    int     `json:"total_monitors"`
	OnlineMonitors   int     `json:"online_monitors"`
	OfflineMonitors  int     `json:"offline_monitors"`
	WarningMonitors  int     `json:"warning_monitors"`
	TotalNodes       int     `json:"total_nodes"`
	TotalIndices     int     `json:"total_indices"`
	TotalDocuments   int64   `json:"total_documents"`
	TotalSizeGB      float64 `json:"total_size_gb"`
	AvgResponseTime  float64 `json:"avg_response_time"` // 平均響應時間（單位：毫秒）
	AvgCPUUsage      float64 `json:"avg_cpu_usage"` // 平均 CPU 使用率（單位：百分比 0-100）
	AvgMemoryUsage   float64 `json:"avg_memory_usage"` // 平均記憶體使用率（單位：百分比 0-100）
	ActiveAlerts     int     `json:"active_alerts"`
	LastUpdateTime   string  `json:"last_update_time"`
}

// ESHealthCheckResult ES 健康檢查結果
type ESHealthCheckResult struct {
	Success        bool                   `json:"success"`
	Status         string                 `json:"status"` // online, offline, warning, error
	ClusterName    string                 `json:"cluster_name"`
	ClusterStatus  string                 `json:"cluster_status"`
	ResponseTime   int64                  `json:"response_time"`
	ClusterHealth  map[string]interface{} `json:"cluster_health,omitempty"` // 集群健康資料（含 shards 資訊）
	NodeInfo       map[string]interface{} `json:"node_info,omitempty"`
	ClusterStats   map[string]interface{} `json:"cluster_stats,omitempty"`
	IndicesStats   map[string]interface{} `json:"indices_stats,omitempty"`
	ErrorMessage   string                 `json:"error_message,omitempty"`
	WarningMessage string                 `json:"warning_message,omitempty"`
	CheckTime      time.Time              `json:"check_time"`
}

// ESAlertThreshold ES 告警閾值配置
type ESAlertThreshold struct {
	CPUUsageHigh        float64 `json:"cpu_usage_high"`         // CPU 使用率高閾值
	CPUUsageCritical    float64 `json:"cpu_usage_critical"`     // CPU 使用率危險閾值
	MemoryUsageHigh     float64 `json:"memory_usage_high"`      // 記憶體使用率高閾值
	MemoryUsageCritical float64 `json:"memory_usage_critical"`  // 記憶體使用率危險閾值
	DiskUsageHigh       float64 `json:"disk_usage_high"`        // 磁碟使用率高閾值
	DiskUsageCritical   float64 `json:"disk_usage_critical"`    // 磁碟使用率危險閾值
	ResponseTimeHigh    int64   `json:"response_time_high"`     // 響應時間高閾值(ms)
	ResponseTimeCritical int64  `json:"response_time_critical"` // 響應時間危險閾值(ms)
	UnassignedShards    int     `json:"unassigned_shards"`      // 未分配分片閾值
}

// DefaultESAlertThreshold 默認告警閾值
func DefaultESAlertThreshold() ESAlertThreshold {
	return ESAlertThreshold{
		CPUUsageHigh:         75.0,
		CPUUsageCritical:     85.0,
		MemoryUsageHigh:      80.0,
		MemoryUsageCritical:  90.0,
		DiskUsageHigh:        85.0,
		DiskUsageCritical:    95.0,
		ResponseTimeHigh:     3000,
		ResponseTimeCritical: 10000,
		UnassignedShards:     1,
	}
}
