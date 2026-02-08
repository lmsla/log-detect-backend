package services

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log-detect/entities"
	"log-detect/global"
	"log-detect/log"
	"net/http"
	"strings"
	"time"
)

// ESMonitorService ES 監控服務
type ESMonitorService struct {
	client *http.Client
}

// NewESMonitorService 創建 ES 監控服務
func NewESMonitorService() *ESMonitorService {
	return &ESMonitorService{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // 跳過證書驗證（生產環境建議設為 false）
				},
			},
		},
	}
}

// CheckESHealth 執行 ES 健康檢查
func (s *ESMonitorService) CheckESHealth(monitor entities.ElasticsearchMonitor) entities.ESHealthCheckResult {
	result := entities.ESHealthCheckResult{
		Success:   false,
		Status:    "offline",
		CheckTime: time.Now(),
	}

	startTime := time.Now()

	// 1. 檢查集群健康狀態
	clusterHealth, err := s.getClusterHealth(monitor)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to get cluster health: %v", err)
		return result
	}

	result.ResponseTime = time.Since(startTime).Milliseconds()
	result.Success = true
	result.Status = "online"
	result.ClusterName = clusterHealth["cluster_name"].(string)
	result.ClusterStatus = clusterHealth["status"].(string)

	// 保存 cluster health 資料供後續提取使用
	result.ClusterHealth = clusterHealth

	// 2. 檢查節點信息
	if monitor.CheckType == "" || contains(monitor.CheckType, "health") || contains(monitor.CheckType, "performance") {
		nodeInfo, err := s.getNodeStats(monitor)
		if err != nil {
			result.WarningMessage = fmt.Sprintf("Failed to get node stats: %v", err)
		} else {
			result.NodeInfo = nodeInfo
		}
	}

	// 3. 檢查集群統計
	if contains(monitor.CheckType, "performance") {
		clusterStats, err := s.getClusterStats(monitor)
		if err != nil {
			result.WarningMessage = fmt.Sprintf("Failed to get cluster stats: %v", err)
		} else {
			result.ClusterStats = clusterStats
		}
	}

	// 4. 檢查索引統計（performance 類型收集索引/搜索速率）
	if contains(monitor.CheckType, "performance") {
		indicesStats, err := s.getIndicesStats(monitor)
		if err != nil {
			result.WarningMessage = fmt.Sprintf("Failed to get indices stats: %v", err)
		} else {
			result.IndicesStats = indicesStats
		}
	}

	// 5. 評估狀態
	result.Status = s.evaluateStatus(result, monitor)

	return result
}

// getESConnection 從資料庫取得 ES 連線配置
func (s *ESMonitorService) getESConnection(monitor entities.ElasticsearchMonitor) (*entities.ESConnection, error) {
	// 優先使用已載入的 ESConnection
	if monitor.ESConnection != nil {
		log.Logrecord_no_rotate("DEBUG", fmt.Sprintf("Monitor %d (%s): using preloaded ESConnection %d (%s)",
			monitor.ID, monitor.Name, monitor.ESConnection.ID, monitor.ESConnection.Name))
		return monitor.ESConnection, nil
	}

	// ESConnection 未預載入，需要從資料庫讀取
	log.Logrecord_no_rotate("DEBUG", fmt.Sprintf("Monitor %d (%s): ESConnection not preloaded, loading from DB (ESConnectionID=%d)",
		monitor.ID, monitor.Name, monitor.ESConnectionID))

	if monitor.ESConnectionID == 0 {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Monitor %d (%s): ESConnectionID is 0, cannot load connection",
			monitor.ID, monitor.Name))
		return nil, fmt.Errorf("monitor %d has invalid ESConnectionID (0)", monitor.ID)
	}

	// 從資料庫讀取
	var conn entities.ESConnection
	if err := global.Mysql.Where("id = ? AND deleted_at IS NULL", monitor.ESConnectionID).First(&conn).Error; err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Monitor %d (%s): failed to load ESConnection %d from DB: %v",
			monitor.ID, monitor.Name, monitor.ESConnectionID, err))
		return nil, fmt.Errorf("ES connection %d not found: %w", monitor.ESConnectionID, err)
	}

	log.Logrecord_no_rotate("DEBUG", fmt.Sprintf("Monitor %d (%s): loaded ESConnection %d (%s) from DB",
		monitor.ID, monitor.Name, conn.ID, conn.Name))
	return &conn, nil
}

// getClusterHealth 獲取集群健康狀態
func (s *ESMonitorService) getClusterHealth(monitor entities.ElasticsearchMonitor) (map[string]interface{}, error) {
	conn, err := s.getESConnection(monitor)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/_cluster/health", conn.GetURL())
	return s.makeRequest(conn, "GET", url, nil)
}

// getNodeStats 獲取節點統計信息
func (s *ESMonitorService) getNodeStats(monitor entities.ElasticsearchMonitor) (map[string]interface{}, error) {
	conn, err := s.getESConnection(monitor)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/_nodes/stats/os,jvm,fs,indices", conn.GetURL())
	return s.makeRequest(conn, "GET", url, nil)
}

// getClusterStats 獲取集群統計信息
func (s *ESMonitorService) getClusterStats(monitor entities.ElasticsearchMonitor) (map[string]interface{}, error) {
	conn, err := s.getESConnection(monitor)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/_cluster/stats", conn.GetURL())
	return s.makeRequest(conn, "GET", url, nil)
}

// getIndicesStats 獲取索引統計信息
func (s *ESMonitorService) getIndicesStats(monitor entities.ElasticsearchMonitor) (map[string]interface{}, error) {
	conn, err := s.getESConnection(monitor)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/_stats", conn.GetURL())
	return s.makeRequest(conn, "GET", url, nil)
}

// makeRequest 發送 HTTP 請求到 ES
func (s *ESMonitorService) makeRequest(conn *entities.ESConnection, method, url string, body interface{}) (map[string]interface{}, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	// 添加認證（從 ESConnection 取得）
	if conn.EnableAuth {
		req.SetBasicAuth(conn.Username, conn.Password)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// ParseMetricsFromCheckResult 從檢查結果中解析指標
func (s *ESMonitorService) ParseMetricsFromCheckResult(monitor entities.ElasticsearchMonitor, result entities.ESHealthCheckResult) entities.ESMetric {
	metric := entities.ESMetric{
		Time:           result.CheckTime,
		MonitorID:      monitor.ID,
		Status:         result.Status,
		ClusterName:    result.ClusterName,
		ClusterStatus:  result.ClusterStatus,
		ResponseTime:   result.ResponseTime,
		ErrorMessage:   result.ErrorMessage,
		WarningMessage: result.WarningMessage,
	}

	// 解析節點信息
	if result.NodeInfo != nil {
		metric.NodeCount = s.extractNodeCount(result.NodeInfo)
		metric.DataNodeCount = s.extractDataNodeCount(result.NodeInfo)
		metric.CPUUsage = s.extractCPUUsage(result.NodeInfo)
		metric.MemoryUsage = s.extractMemoryUsage(result.NodeInfo)
		metric.DiskUsage = s.extractDiskUsage(result.NodeInfo)
		metric.QueryLatency = s.extractQueryLatency(result.NodeInfo)
	}

	// 解析集群統計
	if result.ClusterStats != nil {
		metric.TotalIndices = s.extractTotalIndices(result.ClusterStats)
		metric.TotalDocuments = s.extractTotalDocuments(result.ClusterStats)
		metric.TotalSizeBytes = s.extractTotalSizeBytes(result.ClusterStats)
	}

	// 解析索引統計
	if result.IndicesStats != nil {
		metric.IndexingRate = s.extractIndexingRate(result.IndicesStats)
		metric.SearchRate = s.extractSearchRate(result.IndicesStats)
	}

	// 從集群健康中提取分片信息
	metric.ActiveShards = s.extractActiveShards(result)
	metric.RelocatingShards = s.extractRelocatingShards(result)
	metric.UnassignedShards = s.extractUnassignedShards(result)

	// 生成元數據
	metadata := map[string]interface{}{
		"check_type": monitor.CheckType,
		"interval":   monitor.Interval,
	}
	if jsonData, err := json.Marshal(metadata); err == nil {
		metric.Metadata = string(jsonData)
	}

	return metric
}

// 輔助函數：從結果中提取各種指標

func (s *ESMonitorService) extractNodeCount(nodeInfo map[string]interface{}) int {
	if nodes, ok := nodeInfo["nodes"].(map[string]interface{}); ok {
		return len(nodes)
	}
	return 0
}

func (s *ESMonitorService) extractDataNodeCount(nodeInfo map[string]interface{}) int {
	count := 0
	if nodes, ok := nodeInfo["nodes"].(map[string]interface{}); ok {
		for _, node := range nodes {
			if nodeMap, ok := node.(map[string]interface{}); ok {
				if roles, ok := nodeMap["roles"].([]interface{}); ok {
					for _, role := range roles {
						if role == "data" {
							count++
							break
						}
					}
				}
			}
		}
	}
	return count
}

func (s *ESMonitorService) extractCPUUsage(nodeInfo map[string]interface{}) float64 {
	totalCPU := 0.0
	nodeCount := 0

	if nodes, ok := nodeInfo["nodes"].(map[string]interface{}); ok {
		for _, node := range nodes {
			if nodeMap, ok := node.(map[string]interface{}); ok {
				if os, ok := nodeMap["os"].(map[string]interface{}); ok {
					if cpu, ok := os["cpu"].(map[string]interface{}); ok {
						if percent, ok := cpu["percent"].(float64); ok {
							totalCPU += percent
							nodeCount++
						}
					}
				}
			}
		}
	}

	if nodeCount > 0 {
		return totalCPU / float64(nodeCount)
	}
	return 0.0
}

func (s *ESMonitorService) extractMemoryUsage(nodeInfo map[string]interface{}) float64 {
	totalUsed := int64(0)
	totalMax := int64(0)

	if nodes, ok := nodeInfo["nodes"].(map[string]interface{}); ok {
		for _, node := range nodes {
			if nodeMap, ok := node.(map[string]interface{}); ok {
				if jvm, ok := nodeMap["jvm"].(map[string]interface{}); ok {
					if mem, ok := jvm["mem"].(map[string]interface{}); ok {
						if heapUsedBytes, ok := mem["heap_used_in_bytes"].(float64); ok {
							totalUsed += int64(heapUsedBytes)
						}
						if heapMaxBytes, ok := mem["heap_max_in_bytes"].(float64); ok {
							totalMax += int64(heapMaxBytes)
						}
					}
				}
			}
		}
	}

	if totalMax > 0 {
		return float64(totalUsed) / float64(totalMax) * 100
	}
	return 0.0
}

func (s *ESMonitorService) extractDiskUsage(nodeInfo map[string]interface{}) float64 {
	totalCapacity := int64(0)
	totalAvailable := int64(0)

	if nodes, ok := nodeInfo["nodes"].(map[string]interface{}); ok {
		for _, node := range nodes {
			if nodeMap, ok := node.(map[string]interface{}); ok {
				if fs, ok := nodeMap["fs"].(map[string]interface{}); ok {
					if total, ok := fs["total"].(map[string]interface{}); ok {
						// total_in_bytes 是磁碟總容量
						if capacityBytes, ok := total["total_in_bytes"].(float64); ok {
							totalCapacity += int64(capacityBytes)
						}
						// available_in_bytes 是可用空間
						if availableBytes, ok := total["available_in_bytes"].(float64); ok {
							totalAvailable += int64(availableBytes)
						}
					}
				}
			}
		}
	}

	// 正確計算：已用 = 總容量 - 可用空間
	// disk_usage = (total - available) / total * 100
	if totalCapacity > 0 {
		usedBytes := totalCapacity - totalAvailable
		return float64(usedBytes) / float64(totalCapacity) * 100
	}
	return 0.0
}

func (s *ESMonitorService) extractQueryLatency(nodeInfo map[string]interface{}) int64 {
	totalLatency := int64(0)
	queryCount := 0

	if nodes, ok := nodeInfo["nodes"].(map[string]interface{}); ok {
		for _, node := range nodes {
			if nodeMap, ok := node.(map[string]interface{}); ok {
				if indices, ok := nodeMap["indices"].(map[string]interface{}); ok {
					if search, ok := indices["search"].(map[string]interface{}); ok {
						if queryTimeMs, ok := search["query_time_in_millis"].(float64); ok {
							if queryTotal, ok := search["query_total"].(float64); ok && queryTotal > 0 {
								totalLatency += int64(queryTimeMs / queryTotal)
								queryCount++
							}
						}
					}
				}
			}
		}
	}

	if queryCount > 0 {
		return totalLatency / int64(queryCount)
	}
	return 0
}

func (s *ESMonitorService) extractTotalIndices(clusterStats map[string]interface{}) int {
	if indices, ok := clusterStats["indices"].(map[string]interface{}); ok {
		if count, ok := indices["count"].(float64); ok {
			return int(count)
		}
	}
	return 0
}

func (s *ESMonitorService) extractTotalDocuments(clusterStats map[string]interface{}) int64 {
	if indices, ok := clusterStats["indices"].(map[string]interface{}); ok {
		if docs, ok := indices["docs"].(map[string]interface{}); ok {
			if count, ok := docs["count"].(float64); ok {
				return int64(count)
			}
		}
	}
	return 0
}

func (s *ESMonitorService) extractTotalSizeBytes(clusterStats map[string]interface{}) int64 {
	if indices, ok := clusterStats["indices"].(map[string]interface{}); ok {
		if store, ok := indices["store"].(map[string]interface{}); ok {
			if sizeBytes, ok := store["size_in_bytes"].(float64); ok {
				return int64(sizeBytes)
			}
		}
	}
	return 0
}

func (s *ESMonitorService) extractIndexingRate(indicesStats map[string]interface{}) float64 {
	if all, ok := indicesStats["_all"].(map[string]interface{}); ok {
		if primaries, ok := all["primaries"].(map[string]interface{}); ok {
			if indexing, ok := primaries["indexing"].(map[string]interface{}); ok {
				// 使用 index_current（當前正在進行的索引操作數）
				// 注意：這是"並發數"而非"速率"
				// 值為 0-N，表示當前有幾個索引操作正在執行
				if current, ok := indexing["index_current"].(float64); ok {
					return current
				}
			}
		}
	}
	return 0.0
}

func (s *ESMonitorService) extractSearchRate(indicesStats map[string]interface{}) float64 {
	if all, ok := indicesStats["_all"].(map[string]interface{}); ok {
		if primaries, ok := all["primaries"].(map[string]interface{}); ok {
			if search, ok := primaries["search"].(map[string]interface{}); ok {
				// 使用 query_current（當前正在執行的查詢數）
				// 注意：這是"並發數"而非"速率"
				// 值為 0-N，表示當前有幾個查詢正在執行
				if current, ok := search["query_current"].(float64); ok {
					return current
				}
			}
		}
	}
	return 0.0
}

func (s *ESMonitorService) extractActiveShards(result entities.ESHealthCheckResult) int {
	// 從 cluster health API 回應中提取 active_shards
	if result.ClusterHealth != nil {
		if activeShards, ok := result.ClusterHealth["active_shards"].(float64); ok {
			return int(activeShards)
		}
	}
	return 0
}

func (s *ESMonitorService) extractRelocatingShards(result entities.ESHealthCheckResult) int {
	// 從 cluster health API 回應中提取 relocating_shards
	if result.ClusterHealth != nil {
		if relocatingShards, ok := result.ClusterHealth["relocating_shards"].(float64); ok {
			return int(relocatingShards)
		}
	}
	return 0
}

func (s *ESMonitorService) extractUnassignedShards(result entities.ESHealthCheckResult) int {
	// 從 cluster health API 回應中提取 unassigned_shards
	if result.ClusterHealth != nil {
		if unassignedShards, ok := result.ClusterHealth["unassigned_shards"].(float64); ok {
			return int(unassignedShards)
		}
	}
	return 0
}

// evaluateStatus 評估 ES 監控狀態
func (s *ESMonitorService) evaluateStatus(result entities.ESHealthCheckResult, _ entities.ElasticsearchMonitor) string {
	// 如果檢查失敗，返回 offline
	if !result.Success {
		return "offline"
	}

	// 如果集群狀態是 red，返回 error
	if result.ClusterStatus == "red" {
		return "error"
	}

	// 如果集群狀態是 yellow 或有警告訊息，返回 warning
	if result.ClusterStatus == "yellow" || result.WarningMessage != "" {
		return "warning"
	}

	// 其他情況返回 online
	// TODO: 未來可根據 monitor.AlertThreshold 做更精細的狀態判斷
	return "online"
}

// MonitorESCluster 監控 ES 集群（主函數）
func (s *ESMonitorService) MonitorESCluster(monitor entities.ElasticsearchMonitor) {
	log.Logrecord_no_rotate("INFO", fmt.Sprintf("Starting ES monitor check for: %s", monitor.Name))

	// 1. 執行健康檢查
	result := s.CheckESHealth(monitor)

	// 2. 解析指標
	metric := s.ParseMetricsFromCheckResult(monitor, result)

	// 3. 寫入 TimescaleDB (使用 BatchWriter)
	if global.BatchWriter != nil {
		if err := global.BatchWriter.AddHistory(metric); err != nil {
			log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to add ES metric to batch: %s", err.Error()))
		}
	}

	// 4. 檢查告警條件
	alerts := s.CheckAlertConditions(monitor, metric)
	if len(alerts) > 0 {
		for _, alert := range alerts {
			// 寫入告警記錄（帶去重邏輯）
			created := s.CreateAlert(monitor, alert)
			// 只有成功創建新告警時才發送通知（避免重複通知）
			if created && len(monitor.Receivers) > 0 {
				s.SendAlertNotification(monitor, alert)
			}
		}
	}

	log.Logrecord_no_rotate("INFO", fmt.Sprintf("ES monitor check completed for: %s, status: %s", monitor.Name, metric.Status))
}

// CheckAlertConditions 檢查告警條件
func (s *ESMonitorService) CheckAlertConditions(monitor entities.ElasticsearchMonitor, metric entities.ESMetric) []entities.ESAlert {
	var alerts []entities.ESAlert

	// 解析 check_type 配置
	checkTypes := strings.Split(monitor.CheckType, ",")
	enabledTypes := make(map[string]bool)
	for _, ct := range checkTypes {
		enabledTypes[strings.TrimSpace(ct)] = true
	}

	// 獲取告警閾值（優先使用獨立欄位，向後兼容 JSON 配置）
	threshold := monitor.GetAlertThreshold()

	now := time.Now()

	// CPU 使用率告警 (performance)
	if enabledTypes["performance"] && metric.CPUUsage >= threshold.CPUUsageCritical {
		thresholdVal := threshold.CPUUsageCritical
		actualVal := metric.CPUUsage
		alerts = append(alerts, entities.ESAlert{
			Time:           now,
			MonitorID:      metric.MonitorID,
			AlertType:      "performance",
			Severity:       "critical",
			Message:        fmt.Sprintf("CPU usage critical: %.2f%% (threshold: %.2f%%)", metric.CPUUsage, threshold.CPUUsageCritical),
			Status:         "active",
			ClusterName:    metric.ClusterName,
			ThresholdValue: &thresholdVal,
			ActualValue:    &actualVal,
		})
	} else if enabledTypes["performance"] && metric.CPUUsage >= threshold.CPUUsageHigh {
		thresholdVal := threshold.CPUUsageHigh
		actualVal := metric.CPUUsage
		alerts = append(alerts, entities.ESAlert{
			Time:           now,
			MonitorID:      metric.MonitorID,
			AlertType:      "performance",
			Severity:       "high",
			Message:        fmt.Sprintf("CPU usage high: %.2f%% (threshold: %.2f%%)", metric.CPUUsage, threshold.CPUUsageHigh),
			Status:         "active",
			ClusterName:    metric.ClusterName,
			ThresholdValue: &thresholdVal,
			ActualValue:    &actualVal,
		})
	}

	// 記憶體使用率告警 (performance)
	if enabledTypes["performance"] && metric.MemoryUsage >= threshold.MemoryUsageCritical {
		thresholdVal := threshold.MemoryUsageCritical
		actualVal := metric.MemoryUsage
		alerts = append(alerts, entities.ESAlert{
			Time:           now,
			MonitorID:      metric.MonitorID,
			AlertType:      "performance",
			Severity:       "critical",
			Message:        fmt.Sprintf("Memory usage critical: %.2f%% (threshold: %.2f%%)", metric.MemoryUsage, threshold.MemoryUsageCritical),
			Status:         "active",
			ClusterName:    metric.ClusterName,
			ThresholdValue: &thresholdVal,
			ActualValue:    &actualVal,
		})
	} else if enabledTypes["performance"] && metric.MemoryUsage >= threshold.MemoryUsageHigh {
		thresholdVal := threshold.MemoryUsageHigh
		actualVal := metric.MemoryUsage
		alerts = append(alerts, entities.ESAlert{
			Time:           now,
			MonitorID:      metric.MonitorID,
			AlertType:      "performance",
			Severity:       "high",
			Message:        fmt.Sprintf("Memory usage high: %.2f%% (threshold: %.2f%%)", metric.MemoryUsage, threshold.MemoryUsageHigh),
			Status:         "active",
			ClusterName:    metric.ClusterName,
			ThresholdValue: &thresholdVal,
			ActualValue:    &actualVal,
		})
	}

	// 磁碟使用率告警 (capacity)
	if enabledTypes["capacity"] && metric.DiskUsage >= threshold.DiskUsageCritical {
		thresholdVal := threshold.DiskUsageCritical
		actualVal := metric.DiskUsage
		alerts = append(alerts, entities.ESAlert{
			Time:           now,
			MonitorID:      metric.MonitorID,
			AlertType:      "capacity",
			Severity:       "critical",
			Message:        fmt.Sprintf("Disk usage critical: %.2f%% (threshold: %.2f%%)", metric.DiskUsage, threshold.DiskUsageCritical),
			Status:         "active",
			ClusterName:    metric.ClusterName,
			ThresholdValue: &thresholdVal,
			ActualValue:    &actualVal,
		})
	} else if enabledTypes["capacity"] && metric.DiskUsage >= threshold.DiskUsageHigh {
		thresholdVal := threshold.DiskUsageHigh
		actualVal := metric.DiskUsage
		alerts = append(alerts, entities.ESAlert{
			Time:           now,
			MonitorID:      metric.MonitorID,
			AlertType:      "capacity",
			Severity:       "high",
			Message:        fmt.Sprintf("Disk usage high: %.2f%% (threshold: %.2f%%)", metric.DiskUsage, threshold.DiskUsageHigh),
			Status:         "active",
			ClusterName:    metric.ClusterName,
			ThresholdValue: &thresholdVal,
			ActualValue:    &actualVal,
		})
	}

	// 響應時間告警 (performance)
	if enabledTypes["performance"] && metric.ResponseTime >= threshold.ResponseTimeCritical {
		thresholdVal := float64(threshold.ResponseTimeCritical)
		actualVal := float64(metric.ResponseTime)
		alerts = append(alerts, entities.ESAlert{
			Time:           now,
			MonitorID:      metric.MonitorID,
			AlertType:      "performance",
			Severity:       "critical",
			Message:        fmt.Sprintf("Response time critical: %dms (threshold: %dms)", metric.ResponseTime, threshold.ResponseTimeCritical),
			Status:         "active",
			ClusterName:    metric.ClusterName,
			ThresholdValue: &thresholdVal,
			ActualValue:    &actualVal,
		})
	} else if enabledTypes["performance"] && metric.ResponseTime >= threshold.ResponseTimeHigh {
		thresholdVal := float64(threshold.ResponseTimeHigh)
		actualVal := float64(metric.ResponseTime)
		alerts = append(alerts, entities.ESAlert{
			Time:           now,
			MonitorID:      metric.MonitorID,
			AlertType:      "performance",
			Severity:       "high",
			Message:        fmt.Sprintf("Response time high: %dms (threshold: %dms)", metric.ResponseTime, threshold.ResponseTimeHigh),
			Status:         "active",
			ClusterName:    metric.ClusterName,
			ThresholdValue: &thresholdVal,
			ActualValue:    &actualVal,
		})
	}

	// 未分配分片告警 (health)
	if enabledTypes["health"] && metric.UnassignedShards >= threshold.UnassignedShards {
		thresholdVal := float64(threshold.UnassignedShards)
		actualVal := float64(metric.UnassignedShards)
		alerts = append(alerts, entities.ESAlert{
			Time:           now,
			MonitorID:      metric.MonitorID,
			AlertType:      "health",
			Severity:       "high",
			Message:        fmt.Sprintf("Unassigned shards detected: %d", metric.UnassignedShards),
			Status:         "active",
			ClusterName:    metric.ClusterName,
			ThresholdValue: &thresholdVal,
			ActualValue:    &actualVal,
		})
	}

	// 集群狀態 red 告警 (health)
	if enabledTypes["health"] && metric.ClusterStatus == "red" {
		alerts = append(alerts, entities.ESAlert{
			Time:        now,
			MonitorID:   metric.MonitorID,
			AlertType:   "health",
			Severity:    "critical",
			Message:     "Cluster status is RED",
			Status:      "active",
			ClusterName: metric.ClusterName,
		})
	}

	return alerts
}

// CreateAlert 創建告警記錄（帶去重邏輯）
// 返回 true 表示成功創建新告警，false 表示跳過（重複告警或創建失敗）
func (s *ESMonitorService) CreateAlert(monitor entities.ElasticsearchMonitor, alert entities.ESAlert) bool {
	// 告警去重：檢查最近是否有相同告警
	if s.isDuplicateAlert(monitor, alert) {
		log.Logrecord_no_rotate("DEBUG", fmt.Sprintf("Skipping duplicate alert for monitor %d: %s", alert.MonitorID, alert.Message))
		return false
	}

	// 處理 metadata：如果為空字符串，設為 null；否則確保是有效 JSON
	var metadata interface{}
	if alert.Metadata == "" {
		metadata = nil
	} else {
		// 驗證是否為有效 JSON
		var jsonCheck interface{}
		if err := json.Unmarshal([]byte(alert.Metadata), &jsonCheck); err != nil {
			log.Logrecord_no_rotate("WARN", fmt.Sprintf("Invalid JSON in metadata, setting to null: %s", err.Error()))
			metadata = nil
		} else {
			metadata = alert.Metadata
		}
	}

	// 寫入到 TimescaleDB es_alert_history 表
	query := `
		INSERT INTO es_alert_history (
			time, monitor_id, alert_type, severity, status, message,
			cluster_name, threshold_value, actual_value,
			resolved_at, resolved_by, resolution_note,
			acknowledged_at, acknowledged_by,
			metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := global.TimescaleDB.Exec(
		query,
		alert.Time,
		alert.MonitorID,
		alert.AlertType,
		alert.Severity,
		alert.Status,
		alert.Message,
		alert.ClusterName,
		alert.ThresholdValue,
		alert.ActualValue,
		nil, // resolved_at - 創建時為 NULL
		nil, // resolved_by - 創建時為 NULL
		nil, // resolution_note - 創建時為 NULL
		nil, // acknowledged_at - 創建時為 NULL
		nil, // acknowledged_by - 創建時為 NULL
		metadata,
	)

	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to create ES alert: %s", err.Error()))
		return false
	}

	log.Logrecord_no_rotate("WARN", fmt.Sprintf("ES Alert Created [%s][%s]: %s", alert.Severity, alert.AlertType, alert.Message))
	return true
}

// isDuplicateAlert 檢查是否為重複告警（去重邏輯）
// 防止在短時間內對同一監控器、相同類型和嚴重性的告警重複發送
func (s *ESMonitorService) isDuplicateAlert(monitor entities.ElasticsearchMonitor, alert entities.ESAlert) bool {
	// 去重時間窗口：使用監控器配置，預設 300 秒（5 分鐘）
	dedupeSeconds := monitor.AlertDedupeWindow
	if dedupeSeconds <= 0 {
		dedupeSeconds = 300 // 預設 5 分鐘
	}
	dedupeWindow := time.Duration(dedupeSeconds) * time.Second
	startTime := alert.Time.Add(-dedupeWindow)

	query := `
		SELECT COUNT(*) FROM es_alert_history
		WHERE monitor_id = $1
		  AND alert_type = $2
		  AND severity = $3
		  AND status = 'active'
		  AND time BETWEEN $4 AND $5
	`

	var count int
	err := global.TimescaleDB.QueryRow(
		query,
		alert.MonitorID,
		alert.AlertType,
		alert.Severity,
		startTime,
		alert.Time,
	).Scan(&count)

	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to check duplicate alert: %s", err.Error()))
		return false // 出錯時不阻止告警
	}

	return count > 0
}

// SendAlertNotification 發送告警通知
func (s *ESMonitorService) SendAlertNotification(monitor entities.ElasticsearchMonitor, alert entities.ESAlert) {
	if len(monitor.Receivers) == 0 {
		log.Logrecord_no_rotate("WARN", fmt.Sprintf("No receivers configured for monitor: %s", monitor.Name))
		return
	}

	// 構建告警郵件主題
	subject := fmt.Sprintf("[ES 告警][%s] %s - %s",
		strings.ToUpper(alert.Severity),
		monitor.Name,
		alert.AlertType)

	// 如果配置了自定義主題，使用自定義主題
	if monitor.Subject != "" {
		subject = monitor.Subject
	}

	// 構建告警詳情列表
	details := []string{
		fmt.Sprintf("監控名稱: %s", monitor.Name),
		fmt.Sprintf("集群名稱: %s", alert.ClusterName),
		fmt.Sprintf("告警類型: %s", alert.AlertType),
		fmt.Sprintf("嚴重程度: %s", alert.Severity),
		fmt.Sprintf("告警時間: %s", alert.Time.Format("2006-01-02 15:04:05")),
		fmt.Sprintf("告警訊息: %s", alert.Message),
	}

	// 添加閾值資訊（如果有）
	if alert.ThresholdValue != nil && alert.ActualValue != nil {
		details = append(details,
			fmt.Sprintf("閾值: %.2f", *alert.ThresholdValue),
			fmt.Sprintf("實際值: %.2f", *alert.ActualValue),
		)
	}

	// 添加監控描述（如果有）
	if monitor.Description != "" {
		details = append(details, fmt.Sprintf("說明: %s", monitor.Description))
	}

	// 發送郵件給所有收件人
	Mail4(monitor.Receivers, nil, nil, subject, monitor.Name, details)

	log.Logrecord_no_rotate("INFO", fmt.Sprintf("Alert notification sent to %d receivers for monitor: %s",
		len(monitor.Receivers), monitor.Name))
}

// 輔助函數
func contains(str, substr string) bool {
	return len(str) > 0 && len(substr) > 0 && (str == substr ||
		(len(str) > len(substr) && (str[:len(substr)] == substr || str[len(str)-len(substr):] == substr)))
}
