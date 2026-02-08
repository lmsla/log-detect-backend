package controller

import (
	"log-detect/entities"
	"log-detect/global"
	"log-detect/middleware"
	"log-detect/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// @Summary Get Dashboard Data
// @Tags Dashboard
// @Accept  json
// @Produce  json
// @Success 200 {object} models.Response
// @Failure 401 {object} models.Response
// @Security ApiKeyAuth
// @Router /dashboard/overview [get]
func GetDashboardOverview(c *gin.Context) {
	res := services.GetDashboardData()

	if !res.Success {
		c.JSON(http.StatusInternalServerError, res)
		return
	}

	c.JSON(http.StatusOK, res.Body)
}

// @Summary Get History Statistics
// @Tags Dashboard
// @Accept  json
// @Produce  json
// @Param logname query string false "Log name filter"
// @Param device_group query string false "Device group filter"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} models.Response
// @Failure 401 {object} models.Response
// @Security ApiKeyAuth
// @Router /dashboard/statistics [get]
func GetHistoryStatistics(c *gin.Context) {
	logname := c.Query("logname")
	deviceGroup := c.Query("device_group")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	res := services.GetHistoryStatistics(logname, deviceGroup, startDate, endDate)

	if !res.Success {
		c.JSON(http.StatusInternalServerError, res)
		return
	}

	c.JSON(http.StatusOK, res.Body)
}

// @Summary Get Device Timeline
// @Tags Dashboard
// @Accept  json
// @Produce  json
// @Param device_name path string true "Device name"
// @Param logname query string true "Log name"
// @Param days query int false "Number of days (default: 7)"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Security ApiKeyAuth
// @Router /dashboard/devices/{device_name}/timeline [get]
func GetDeviceTimeline(c *gin.Context) {
	deviceName := c.Param("device_name")
	logname := c.Query("logname")

	if deviceName == "" || logname == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_name and logname are required"})
		return
	}

	daysStr := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 || days > 90 {
		days = 7 // default to 7 days
	}

	res := services.GetDeviceTimeline(deviceName, logname, days)

	if !res.Success {
		c.JSON(http.StatusInternalServerError, res)
		return
	}

	c.JSON(http.StatusOK, res.Body)
}

// @Summary Get Trend Data
// @Tags Dashboard
// @Accept  json
// @Produce  json
// @Param logname query string false "Log name filter"
// @Param device_group query string false "Device group filter"
// @Param days query int false "Number of days (default: 30)"
// @Success 200 {object} models.Response
// @Failure 401 {object} models.Response
// @Security ApiKeyAuth
// @Router /dashboard/trends [get]
func GetTrendData(c *gin.Context) {
	logname := c.Query("logname")
	deviceGroup := c.Query("device_group")

	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 || days > 365 {
		days = 30 // default to 30 days
	}

	res := services.GetTrendData(logname, deviceGroup, days)

	if !res.Success {
		c.JSON(http.StatusInternalServerError, res)
		return
	}

	c.JSON(http.StatusOK, res.Body)
}

// @Summary Get Group Statistics
// @Tags Dashboard
// @Accept  json
// @Produce  json
// @Param logname query string true "Log name"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Security ApiKeyAuth
// @Router /dashboard/groups/statistics [get]
func GetGroupStatistics(c *gin.Context) {
	logname := c.Query("logname")

	if logname == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "logname is required"})
		return
	}

	res := services.GetGroupStatistics(logname)

	if !res.Success {
		c.JSON(http.StatusInternalServerError, res)
		return
	}

	c.JSON(http.StatusOK, res.Body)
}

// @Summary Get Device Status Overview
// @Tags Dashboard
// @Accept  json
// @Produce  json
// @Param logname query string false "Log name filter"
// @Param device_group query string false "Device group filter"
// @Success 200 {object} models.Response
// @Failure 401 {object} models.Response
// @Security ApiKeyAuth
// @Router /dashboard/devices/status [get]
func GetDeviceStatusOverview(c *gin.Context) {
	logname := c.Query("logname")

	// 使用現有的 DataDealing 函數來獲取設備狀態
	res := services.DataDealing(logname)

	if !res.Success {
		c.JSON(http.StatusInternalServerError, res)
		return
	}

	c.JSON(http.StatusOK, res.Body)
}

// @Summary Get Recent Alerts
// @Tags Dashboard
// @Accept  json
// @Produce  json
// @Param limit query int false "Number of alerts to return (default: 10)"
// @Param status query string false "Alert status filter (active, resolved, acknowledged)"
// @Success 200 {object} models.Response
// @Failure 401 {object} models.Response
// @Security ApiKeyAuth
// @Router /dashboard/alerts/recent [get]
func GetRecentAlerts(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10
	}

	status := c.Query("status")

	query := global.Mysql.Model(&entities.AlertHistory{}).Order("created_at DESC").Limit(limit)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var alerts []entities.AlertHistory
	if err := query.Find(&alerts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch alerts"})
		return
	}

	c.JSON(http.StatusOK, alerts)
}

// @Summary Create Alert
// @Tags Dashboard
// @Accept  json
// @Produce  json
// @Param alert body entities.AlertHistory true "Alert data"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Security ApiKeyAuth
// @Router /dashboard/alerts [post]
func CreateAlert(c *gin.Context) {
	var alert entities.AlertHistory

	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	// 設置默認值
	if alert.Status == "" {
		alert.Status = "active"
	}
	if alert.Severity == "" {
		alert.Severity = "medium"
	}

	// 獲取當前用戶信息
	if user, exists := middleware.GetCurrentUser(c); exists {
		alert.ResolvedBy = user.Username
	}

	res := services.CreateAlertHistory(alert)

	if !res.Success {
		c.JSON(http.StatusInternalServerError, res)
		return
	}

	c.JSON(http.StatusOK, res.Body)
}

// @Summary Update Alert Status
// @Tags Dashboard
// @Accept  json
// @Produce  json
// @Param id path int true "Alert ID"
// @Param status body object true "Status update"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Failure 404 {object} models.Response
// @Security ApiKeyAuth
// @Router /dashboard/alerts/{id}/status [put]
func UpdateAlertStatus(c *gin.Context) {
	alertIDStr := c.Param("id")
	alertID, err := strconv.Atoi(alertIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	var statusUpdate struct {
		Status  string `json:"status" binding:"required"`
		Comment string `json:"comment,omitempty"`
	}

	if err := c.ShouldBindJSON(&statusUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	// 獲取當前用戶
	var resolvedBy string
	if user, exists := middleware.GetCurrentUser(c); exists {
		resolvedBy = user.Username
	}

	// 更新告警狀態
	updates := map[string]interface{}{
		"status": statusUpdate.Status,
	}

	if statusUpdate.Status == "resolved" || statusUpdate.Status == "acknowledged" {
		updates["resolved_at"] = time.Now().Unix()
		updates["resolved_by"] = resolvedBy
	}

	if err := global.Mysql.Model(&entities.AlertHistory{}).Where("id = ?", alertID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update alert status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert status updated successfully"})
}

