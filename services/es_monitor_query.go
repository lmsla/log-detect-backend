package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log-detect/entities"
	"log-detect/global"
	"log-detect/log"
	"math"
	"time"
)

// ESMonitorQueryService ES 監控查詢服務
type ESMonitorQueryService struct {
	db *sql.DB
}

// NewESMonitorQueryService 創建 ES 監控查詢服務
func NewESMonitorQueryService() *ESMonitorQueryService {
	return &ESMonitorQueryService{
		db: global.TimescaleDB,
	}
}

// GetLatestMetrics 獲取最新監控指標
func (s *ESMonitorQueryService) GetLatestMetrics(monitorID int) (*entities.ESMetric, error) {
	query := `
		SELECT time, monitor_id, status, cluster_name, cluster_status, response_time,
		       cpu_usage, memory_usage, disk_usage, node_count, data_node_count,
		       query_latency, indexing_rate, search_rate, total_indices, total_documents,
		       total_size_bytes, active_shards, relocating_shards, unassigned_shards,
		       error_message, warning_message, metadata
		FROM es_metrics
		WHERE monitor_id = $1
		ORDER BY time DESC
		LIMIT 1
	`

	var metric entities.ESMetric
	var metadata sql.NullString

	err := s.db.QueryRow(query, monitorID).Scan(
		&metric.Time, &metric.MonitorID, &metric.Status, &metric.ClusterName,
		&metric.ClusterStatus, &metric.ResponseTime, &metric.CPUUsage, &metric.MemoryUsage,
		&metric.DiskUsage, &metric.NodeCount, &metric.DataNodeCount, &metric.QueryLatency,
		&metric.IndexingRate, &metric.SearchRate, &metric.TotalIndices, &metric.TotalDocuments,
		&metric.TotalSizeBytes, &metric.ActiveShards, &metric.RelocatingShards,
		&metric.UnassignedShards, &metric.ErrorMessage, &metric.WarningMessage, &metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no metrics found for monitor ID %d", monitorID)
		}
		return nil, err
	}

	if metadata.Valid {
		metric.Metadata = metadata.String
	} else {
		metric.Metadata = "{}"
	}

	return &metric, nil
}

// GetMetricsTimeSeries 獲取時間序列指標數據
func (s *ESMonitorQueryService) GetMetricsTimeSeries(monitorID int, startTime, endTime time.Time, interval string) ([]entities.ESMetricTimeSeries, error) {
	// 根據時間範圍自動調整聚合間隔
	if interval == "" {
		duration := endTime.Sub(startTime)
		if duration > 7*24*time.Hour {
			interval = "1 hour"
		} else if duration > 24*time.Hour {
			interval = "10 minutes"
		} else {
			interval = "1 minute"
		}
	}

	query := fmt.Sprintf(`
		SELECT
			time_bucket('%s', time) AS bucket_time,
			AVG(cpu_usage) AS avg_cpu,
			AVG(memory_usage) AS avg_memory,
			AVG(disk_usage) AS avg_disk,
			AVG(response_time) AS avg_response_time,
			AVG(indexing_rate) AS avg_indexing_rate,
			AVG(search_rate) AS avg_search_rate,
			AVG(active_shards) AS avg_active_shards,
			AVG(unassigned_shards) AS avg_unassigned_shards
		FROM es_metrics
		WHERE monitor_id = $1
		  AND time >= $2
		  AND time <= $3
		GROUP BY bucket_time
		ORDER BY bucket_time ASC
	`, interval)

	// 添加調試日誌
	log.Logrecord_no_rotate("DEBUG", fmt.Sprintf("GetMetricsTimeSeries - monitor_id: %d, startTime: %s, endTime: %s, interval: %s",
		monitorID, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339), interval))

	rows, err := s.db.Query(query, monitorID, startTime, endTime)
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to query ES metrics time series: %s", err.Error()))
		return nil, err
	}
	defer rows.Close()

	// 初始化為空切片，確保無數據時返回 [] 而不是 null
	results := make([]entities.ESMetricTimeSeries, 0)
	for rows.Next() {
		var ts entities.ESMetricTimeSeries
		var avgResponseTime, avgActiveShards, avgUnassignedShards sql.NullFloat64

		err := rows.Scan(
			&ts.Time, &ts.CPUUsage, &ts.MemoryUsage, &ts.DiskUsage,
			&avgResponseTime, &ts.IndexingRate, &ts.SearchRate,
			&avgActiveShards, &avgUnassignedShards,
		)
		if err != nil {
			log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to scan ES metric row: %s", err.Error()))
			continue
		}

		// 處理整數欄位的平均值（AVG 返回 float，需轉換）
		if avgResponseTime.Valid {
			ts.ResponseTime = int64(avgResponseTime.Float64)
		}
		if avgActiveShards.Valid {
			ts.ActiveShards = int(avgActiveShards.Float64)
		}
		if avgUnassignedShards.Valid {
			ts.UnassignedShards = int(avgUnassignedShards.Float64)
		}

		// 四捨五入到小數點後兩位
		ts.CPUUsage = math.Round(ts.CPUUsage*100) / 100
		ts.MemoryUsage = math.Round(ts.MemoryUsage*100) / 100
		ts.DiskUsage = math.Round(ts.DiskUsage*100) / 100
		ts.IndexingRate = math.Round(ts.IndexingRate*100) / 100
		ts.SearchRate = math.Round(ts.SearchRate*100) / 100

		results = append(results, ts)
	}

	// 添加查詢結果日誌
	log.Logrecord_no_rotate("DEBUG", fmt.Sprintf("GetMetricsTimeSeries - Returned %d data points for monitor_id: %d", len(results), monitorID))

	return results, nil
}

// GetAllMonitorsStatus 獲取所有監控器的當前狀態
func (s *ESMonitorQueryService) GetAllMonitorsStatus() ([]entities.ESMonitorStatus, error) {
	// 先從 MySQL 獲取所有監控配置（包含 ESConnection）
	var monitors []entities.ElasticsearchMonitor
	if err := global.Mysql.Preload("ESConnection").Find(&monitors).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch monitor configs: %w", err)
	}

	// 初始化為空切片，確保無數據時返回 [] 而不是 null
	statuses := make([]entities.ESMonitorStatus, 0)
	for _, monitor := range monitors {
		// 取得 ESConnection 資訊
		var host, connName string
		if monitor.ESConnection != nil {
			host = fmt.Sprintf("%s:%d", monitor.ESConnection.CleanHost(), monitor.ESConnection.Port)
			connName = monitor.ESConnection.Name
		}

		// 獲取每個監控器的最新指標
		metric, err := s.GetLatestMetrics(monitor.ID)
		if err != nil {
			// 如果沒有數據，標記為離線
			statuses = append(statuses, entities.ESMonitorStatus{
				MonitorID:        monitor.ID,
				MonitorName:      monitor.Name,
				ESConnectionID:   monitor.ESConnectionID,
				ESConnectionName: connName,
				Host:             host,
				Status:           "offline",
				ErrorMessage:     "No metrics data available",
			})
			continue
		}

		// 組裝狀態響應
		status := entities.ESMonitorStatus{
			MonitorID:        monitor.ID,
			MonitorName:      monitor.Name,
			ESConnectionID:   monitor.ESConnectionID,
			ESConnectionName: connName,
			Host:             host,
			Status:           metric.Status,
			ClusterStatus:    metric.ClusterStatus,
			ClusterName:      metric.ClusterName,
			ResponseTime:     metric.ResponseTime,
			CPUUsage:         metric.CPUUsage,
			MemoryUsage:      metric.MemoryUsage,
			DiskUsage:        metric.DiskUsage,
			NodeCount:        metric.NodeCount,
			ActiveShards:     metric.ActiveShards,
			RelocatingShards: metric.RelocatingShards,
			UnassignedShards: metric.UnassignedShards,
			LastCheckTime:    metric.Time,
			ErrorMessage:     metric.ErrorMessage,
			WarningMessage:   metric.WarningMessage,
		}
		statuses = append(statuses, status)
	}

	return statuses, nil
}

// GetESStatistics 獲取 ES 監控統計數據
func (s *ESMonitorQueryService) GetESStatistics() (*entities.ESStatistics, error) {
	var stats entities.ESStatistics

	// 從 MySQL 獲取監控器總數
	var totalMonitors int64
	if err := global.Mysql.Model(&entities.ElasticsearchMonitor{}).Count(&totalMonitors).Error; err != nil {
		return nil, err
	}
	stats.TotalMonitors = int(totalMonitors)

	// 從 TimescaleDB 獲取最新狀態統計
	query := `
		WITH latest_metrics AS (
			SELECT DISTINCT ON (monitor_id)
				monitor_id,
				status,
				cluster_status,
				response_time,
				cpu_usage,
				memory_usage,
				disk_usage,
				node_count,
				total_indices,
				total_documents,
				total_size_bytes,
				time
			FROM es_metrics
			WHERE time > NOW() - INTERVAL '1 hour'
			ORDER BY monitor_id, time DESC
		)
		SELECT
			COUNT(*) FILTER (WHERE status = 'online') AS online_count,
			COUNT(*) FILTER (WHERE status = 'offline') AS offline_count,
			COUNT(*) FILTER (WHERE status IN ('warning', 'error')) AS warning_count,
			COALESCE(SUM(node_count), 0) AS total_nodes,
			COALESCE(SUM(total_indices), 0) AS total_indices,
			COALESCE(SUM(total_documents), 0) AS total_documents,
			COALESCE(SUM(total_size_bytes), 0) AS total_size_bytes,
			COALESCE(AVG(response_time), 0) AS avg_response_time,
			COALESCE(AVG(cpu_usage), 0) AS avg_cpu_usage,
			COALESCE(AVG(memory_usage), 0) AS avg_memory_usage,
			MAX(time) AS last_update
		FROM latest_metrics
	`

	var lastUpdate sql.NullTime
	err := s.db.QueryRow(query).Scan(
		&stats.OnlineMonitors,
		&stats.OfflineMonitors,
		&stats.WarningMonitors,
		&stats.TotalNodes,
		&stats.TotalIndices,
		&stats.TotalDocuments,
		&stats.TotalSizeGB,
		&stats.AvgResponseTime,
		&stats.AvgCPUUsage,
		&stats.AvgMemoryUsage,
		&lastUpdate,
	)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// 轉換單位：bytes -> GB 並四捨五入到小數點後兩位
	stats.TotalSizeGB = math.Round(float64(stats.TotalSizeGB)/(1024*1024*1024)*100) / 100

	// 四捨五入平均值到小數點後兩位
	stats.AvgResponseTime = math.Round(stats.AvgResponseTime*100) / 100
	stats.AvgCPUUsage = math.Round(stats.AvgCPUUsage*100) / 100
	stats.AvgMemoryUsage = math.Round(stats.AvgMemoryUsage*100) / 100

	// 格式化最後更新時間
	if lastUpdate.Valid {
		stats.LastUpdateTime = lastUpdate.Time.Format("2006-01-02 15:04:05")
	} else {
		stats.LastUpdateTime = "N/A"
	}

	// 查詢活躍告警數量
	alertQuery := `
		SELECT COUNT(*)
		FROM es_alert_history
		WHERE status = 'active'
		  AND time > NOW() - INTERVAL '24 hours'
	`
	_ = s.db.QueryRow(alertQuery).Scan(&stats.ActiveAlerts)

	return &stats, nil
}

// GetMonitorMetricsByTimeRange 獲取指定時間範圍內的原始指標
func (s *ESMonitorQueryService) GetMonitorMetricsByTimeRange(monitorID int, startTime, endTime time.Time, limit int) ([]entities.ESMetric, error) {
	if limit <= 0 || limit > 10000 {
		limit = 1000 // 默認返回 1000 條
	}

	query := `
		SELECT time, monitor_id, status, cluster_name, cluster_status, response_time,
		       cpu_usage, memory_usage, disk_usage, node_count, data_node_count,
		       query_latency, indexing_rate, search_rate, total_indices, total_documents,
		       total_size_bytes, active_shards, relocating_shards, unassigned_shards,
		       error_message, warning_message, metadata
		FROM es_metrics
		WHERE monitor_id = $1
		  AND time >= $2
		  AND time <= $3
		ORDER BY time DESC
		LIMIT $4
	`

	rows, err := s.db.Query(query, monitorID, startTime, endTime, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 初始化為空切片，確保無數據時返回 [] 而不是 null
	metrics := make([]entities.ESMetric, 0)
	for rows.Next() {
		var metric entities.ESMetric
		var metadata sql.NullString

		err := rows.Scan(
			&metric.Time, &metric.MonitorID, &metric.Status, &metric.ClusterName,
			&metric.ClusterStatus, &metric.ResponseTime, &metric.CPUUsage, &metric.MemoryUsage,
			&metric.DiskUsage, &metric.NodeCount, &metric.DataNodeCount, &metric.QueryLatency,
			&metric.IndexingRate, &metric.SearchRate, &metric.TotalIndices, &metric.TotalDocuments,
			&metric.TotalSizeBytes, &metric.ActiveShards, &metric.RelocatingShards,
			&metric.UnassignedShards, &metric.ErrorMessage, &metric.WarningMessage, &metadata,
		)

		if err != nil {
			log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to scan ES metric: %s", err.Error()))
			continue
		}

		if metadata.Valid {
			metric.Metadata = metadata.String
		} else {
			metric.Metadata = "{}"
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// GetClusterHealthHistory 獲取集群健康狀態歷史
func (s *ESMonitorQueryService) GetClusterHealthHistory(monitorID int, hours int) (map[string]int, error) {
	if hours <= 0 || hours > 720 {
		hours = 24 // 默認 24 小時
	}

	query := `
		SELECT cluster_status, COUNT(*) as count
		FROM es_metrics
		WHERE monitor_id = $1
		  AND time > NOW() - INTERVAL '1 hour' * $2
		GROUP BY cluster_status
	`

	rows, err := s.db.Query(query, monitorID, hours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			continue
		}
		result[status] = count
	}

	return result, nil
}

// GetPerformanceTrend 獲取性能趨勢分析
func (s *ESMonitorQueryService) GetPerformanceTrend(monitorID int, metric string, hours int) ([]map[string]any, error) {
	if hours <= 0 || hours > 720 {
		hours = 24
	}

	// 驗證指標名稱防止 SQL 注入
	validMetrics := map[string]bool{
		"cpu_usage":         true,
		"memory_usage":      true,
		"disk_usage":        true,
		"response_time":     true,
		"indexing_rate":     true,
		"search_rate":       true,
		"query_latency":     true,
		"active_shards":     true,
		"unassigned_shards": true,
	}

	if !validMetrics[metric] {
		return nil, fmt.Errorf("invalid metric name: %s", metric)
	}

	query := fmt.Sprintf(`
		SELECT
			time_bucket('5 minutes', time) AS bucket_time,
			AVG(%s) AS avg_value,
			MIN(%s) AS min_value,
			MAX(%s) AS max_value
		FROM es_metrics
		WHERE monitor_id = $1
		  AND time > NOW() - INTERVAL '1 hour' * $2
		GROUP BY bucket_time
		ORDER BY bucket_time ASC
	`, metric, metric, metric)

	rows, err := s.db.Query(query, monitorID, hours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 初始化為空切片，確保無數據時返回 [] 而不是 null
	results := make([]map[string]any, 0)
	for rows.Next() {
		var bucketTime time.Time
		var avgValue, minValue, maxValue float64

		if err := rows.Scan(&bucketTime, &avgValue, &minValue, &maxValue); err != nil {
			continue
		}

		results = append(results, map[string]any{
			"time": bucketTime.Format("2006-01-02 15:04:05"),
			"avg":  avgValue,
			"min":  minValue,
			"max":  maxValue,
		})
	}

	return results, nil
}

// ExportMetricsToJSON 導出指標為 JSON 格式
func (s *ESMonitorQueryService) ExportMetricsToJSON(monitorID int, startTime, endTime time.Time) (string, error) {
	metrics, err := s.GetMonitorMetricsByTimeRange(monitorID, startTime, endTime, 10000)
	if err != nil {
		return "", err
	}

	jsonData, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}
