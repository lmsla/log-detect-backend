package controller

import (
	"log-detect/global"
	"log-detect/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func requireTimescale(c *gin.Context, feature string) bool {
	if services.IsTimescaleReady() {
		return true
	}

	status := global.GetRuntimeStatus().TimescaleDB
	msg := "TimescaleDB connection failed"
	if !status.Configured {
		msg = "TimescaleDB feature is disabled"
	}

	c.JSON(http.StatusServiceUnavailable, gin.H{
		"success":           false,
		"service":           "timescaledb",
		"code":              "TIMESCALE_UNAVAILABLE",
		"msg":               msg,
		"reason":            services.TimescaleUnavailableMessage(),
		"affected_feature":  feature,
		"configured":        status.Configured,
		"connected":         status.Connected,
		"degraded_features": status.DegradedFeatures,
	})
	return false
}
