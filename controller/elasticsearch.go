package controller

import (
	"log-detect/entities"
	"log-detect/models"
	"log-detect/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// @Summary Create ES Monitor
// @Tags Elasticsearch
// @Accept  json
// @Produce  json
// @Param monitor body entities.ElasticsearchMonitor true "ES Monitor Config"
// @Success 200 {object} models.Response
// @Router /api/v1/elasticsearch/monitors [post]
func CreateESMonitor(c *gin.Context) {
	var monitor entities.ElasticsearchMonitor

	if err := c.ShouldBindJSON(&monitor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res := services.CreateESMonitor(monitor)

	if !res.Success {
		c.JSON(http.StatusBadRequest, res)
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Summary Update ES Monitor
// @Tags Elasticsearch
// @Accept  json
// @Produce  json
// @Param monitor body entities.ElasticsearchMonitor true "ES Monitor Config"
// @Success 200 {object} models.Response
// @Router /api/v1/elasticsearch/monitors [put]
func UpdateESMonitor(c *gin.Context) {
	var monitor entities.ElasticsearchMonitor

	if err := c.ShouldBindJSON(&monitor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res := services.UpdateESMonitor(monitor)

	if !res.Success {
		c.JSON(http.StatusBadRequest, res)
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Summary Get All ES Monitors
// @Tags Elasticsearch
// @Accept  json
// @Produce  json
// @Success 200 {object} models.Response
// @Router /api/v1/elasticsearch/monitors [get]
func GetAllESMonitors(c *gin.Context) {
	res := services.GetAllESMonitors()

	if !res.Success {
		c.JSON(http.StatusBadRequest, res)
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Summary Get ES Monitor by ID
// @Tags Elasticsearch
// @Accept  json
// @Produce  json
// @Param id path int true "Monitor ID"
// @Success 200 {object} models.Response
// @Router /api/v1/elasticsearch/monitors/{id} [get]
func GetESMonitorByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	res := services.GetESMonitorByID(id)

	if !res.Success {
		c.JSON(http.StatusNotFound, res)
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Summary Delete ES Monitor
// @Tags Elasticsearch
// @Accept  json
// @Produce  json
// @Param id path int true "Monitor ID"
// @Success 200 {object} models.Response
// @Router /api/v1/elasticsearch/monitors/{id} [delete]
func DeleteESMonitor(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	res := services.DeleteESMonitor(id)

	if !res.Success {
		c.JSON(http.StatusBadRequest, res)
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Summary Test ES Connection
// @Tags Elasticsearch
// @Accept  json
// @Produce  json
// @Param id path int true "Monitor ID"
// @Success 200 {object} models.Response
// @Router /api/v1/elasticsearch/monitors/{id}/test [post]
func TestESMonitorConnection(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// 從資料庫取得監控配置
	monitorRes := services.GetESMonitorByID(id)
	if !monitorRes.Success {
		c.JSON(http.StatusNotFound, monitorRes)
		return
	}

	monitor, ok := monitorRes.Body.(entities.ElasticsearchMonitor)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid monitor data"})
		return
	}

	res := services.TestESMonitorConnection(monitor)

	if !res.Success {
		c.JSON(http.StatusBadRequest, res)
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Summary Toggle ES Monitor
// @Tags Elasticsearch
// @Accept  json
// @Produce  json
// @Param id path int true "Monitor ID"
// @Param body body object{enable=bool} true "Enable/Disable"
// @Success 200 {object} models.Response
// @Router /api/v1/elasticsearch/monitors/{id}/toggle [post]
func ToggleESMonitor(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var body struct {
		Enable bool `json:"enable"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res := services.ToggleESMonitor(id, body.Enable)

	if !res.Success {
		c.JSON(http.StatusBadRequest, res)
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Summary Get All ES Monitors Status
// @Tags Elasticsearch
// @Accept  json
// @Produce  json
// @Success 200 {object} models.Response
// @Router /api/v1/elasticsearch/status [get]
func GetAllESMonitorsStatus(c *gin.Context) {
	queryService := services.NewESMonitorQueryService()
	statuses, err := queryService.GetAllMonitorsStatus()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"msg":     err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "查詢成功",
		"body":    statuses,
	})
}

// @Summary Get ES Statistics
// @Tags Elasticsearch
// @Accept  json
// @Produce  json
// @Success 200 {object} models.Response
// @Router /api/v1/elasticsearch/statistics [get]
func GetESStatistics(c *gin.Context) {
	queryService := services.NewESMonitorQueryService()
	stats, err := queryService.GetESStatistics()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"msg":     err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "查詢成功",
		"body":    stats,
	})
}

// @Summary Get ES Monitor History
// @Tags Elasticsearch
// @Accept  json
// @Produce  json
// @Param id path int true "Monitor ID"
// @Param hours query int false "Hours to query (default: 24, max: 720)"
// @Param interval query string false "Time bucket interval (e.g., '1 minute', '5 minutes', '1 hour')"
// @Success 200 {object} models.Response
// @Router /api/v1/elasticsearch/status/{id}/history [get]
func GetESMonitorHistory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// 解析查詢參數
	hours := 24
	if h := c.Query("hours"); h != "" {
		if parsedHours, err := strconv.Atoi(h); err == nil && parsedHours > 0 && parsedHours <= 720 {
			hours = parsedHours
		}
	}

	interval := c.Query("interval")
	if interval == "" {
		if hours <= 1 {
			interval = "1 minute"
		} else if hours <= 24 {
			interval = "5 minutes"
		} else {
			interval = "1 hour"
		}
	}

	// 計算時間範圍
	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(hours) * time.Hour)

	queryService := services.NewESMonitorQueryService()
	history, err := queryService.GetMetricsTimeSeries(id, startTime, endTime, interval)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"msg":     err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "查詢成功",
		"body": gin.H{
			"monitor_id": id,
			"start_time": startTime.Format(time.RFC3339),
			"end_time":   endTime.Format(time.RFC3339),
			"interval":   interval,
			"data":       history,
		},
	})
}

// ==================== 告警管理 API ====================

// @Summary Get ES Alerts
// @Tags Elasticsearch
// @Accept  json
// @Produce  json
// @Param status query []string false "Status filter (active, resolved, acknowledged)" collectionFormat(multi)
// @Param severity query []string false "Severity filter (critical, high, medium, low)" collectionFormat(multi)
// @Param alert_type query []string false "Alert type filter" collectionFormat(multi)
// @Param monitor_id query int false "Monitor ID"
// @Param start_time query string false "Start time (ISO 8601)"
// @Param end_time query string false "End time (ISO 8601)"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20, max: 100)"
// @Success 200 {object} models.Response
// @Router /api/v1/elasticsearch/alerts [get]
func GetESAlerts(c *gin.Context) {
	var params models.ESAlertQueryParams

	// 解析查詢參數
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"msg":     "Invalid query parameters",
		})
		return
	}

	// 設置預設值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	// 如果沒有指定時間範圍，預設查詢最近 7 天
	if params.StartTime.IsZero() && params.EndTime.IsZero() {
		params.EndTime = time.Now()
		params.StartTime = params.EndTime.Add(-7 * 24 * time.Hour)
	}

	alertService := services.NewESAlertService()
	alerts, total, err := alertService.GetAlerts(params)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"msg":     err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "查詢成功",
		"body": gin.H{
			"items": alerts,
			"pagination": gin.H{
				"page":        params.Page,
				"page_size":   params.PageSize,
				"total":       total,
				"total_pages": (total + int64(params.PageSize) - 1) / int64(params.PageSize),
			},
		},
	})
}

// @Summary Get ES Alert By ID
// @Tags Elasticsearch
// @Accept  json
// @Produce  json
// @Param monitor_id path int true "Monitor ID"
// @Param alert_time query string true "Alert time (ISO 8601)"
// @Success 200 {object} models.Response
// @Router /api/v1/elasticsearch/alerts/{monitor_id} [get]
func GetESAlertByID(c *gin.Context) {
	monitorID, err := strconv.Atoi(c.Param("monitor_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid monitor ID"})
		return
	}

	alertTimeStr := c.Query("alert_time")
	if alertTimeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alert_time is required"})
		return
	}

	alertTime, err := time.Parse(time.RFC3339, alertTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert_time format (use ISO 8601)"})
		return
	}

	alertService := services.NewESAlertService()
	alert, err := alertService.GetAlertByID(monitorID, alertTime)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"msg":     "Alert not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "查詢成功",
		"body":    alert,
	})
}

// @Summary Resolve ES Alert
// @Tags Elasticsearch
// @Accept  json
// @Produce  json
// @Param monitor_id path int true "Monitor ID"
// @Param body body object true "Resolve params"
// @Success 200 {object} models.Response
// @Router /api/v1/elasticsearch/alerts/{monitor_id}/resolve [post]
func ResolveESAlert(c *gin.Context) {
	monitorID, err := strconv.Atoi(c.Param("monitor_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid monitor ID"})
		return
	}

	var req struct {
		AlertTime      string `json:"alert_time" binding:"required"`
		ResolvedBy     string `json:"resolved_by"`
		ResolutionNote string `json:"resolution_note"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	alertTime, err := time.Parse(time.RFC3339, req.AlertTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert_time format (use ISO 8601)"})
		return
	}

	// 如果沒有提供 resolved_by，使用當前用戶（從 context 獲取）
	resolvedBy := req.ResolvedBy
	if resolvedBy == "" {
		if username, exists := c.Get("username"); exists {
			resolvedBy = username.(string)
		} else {
			resolvedBy = "system"
		}
	}

	alertService := services.NewESAlertService()
	err = alertService.ResolveAlert(monitorID, alertTime, resolvedBy, req.ResolutionNote)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"msg":     err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "告警已標記為已解決",
	})
}

// @Summary Acknowledge ES Alert
// @Tags Elasticsearch
// @Accept  json
// @Produce  json
// @Param monitor_id path int true "Monitor ID"
// @Param body body object true "Acknowledge params"
// @Success 200 {object} models.Response
// @Router /api/v1/elasticsearch/alerts/{monitor_id}/acknowledge [put]
func AcknowledgeESAlert(c *gin.Context) {
	monitorID, err := strconv.Atoi(c.Param("monitor_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid monitor ID"})
		return
	}

	var req struct {
		AlertTime      string `json:"alert_time" binding:"required"`
		AcknowledgedBy string `json:"acknowledged_by"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	alertTime, err := time.Parse(time.RFC3339, req.AlertTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert_time format (use ISO 8601)"})
		return
	}

	// 如果沒有提供 acknowledged_by，使用當前用戶
	acknowledgedBy := req.AcknowledgedBy
	if acknowledgedBy == "" {
		if username, exists := c.Get("username"); exists {
			acknowledgedBy = username.(string)
		} else {
			acknowledgedBy = "system"
		}
	}

	alertService := services.NewESAlertService()
	err = alertService.AcknowledgeAlert(monitorID, alertTime, acknowledgedBy)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"msg":     err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "告警已確認",
	})
}
