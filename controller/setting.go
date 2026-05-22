package controller

import (
	"log-detect/global"
	"log-detect/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary Get SSO URL
// @Tags    Env
// @Accept  json
// @Produce json
// @Success 200 {object} string
// @Router /get-sso-url [get]
func GetSSOURL(c *gin.Context) {
	c.JSON(http.StatusOK, global.EnvConfig.SSO.URL)
}

// @Summary  Get Log-detect Menu
// @Tags     Env
// @Accept   json
// @Produce  json
// @Success  200 {object} []entities.MainMenu
// @Security ApiKeyAuth
// @Router   /user/get-server-menu [get]
func GetServerMenu(c *gin.Context) {
	c.JSON(http.StatusOK, services.GetServerMenu())
}

// @Summary Get Server Module
// @Tags    Env
// @Accept  json
// @Produce json
// @Success 200 {object} []entities.Module
// @Router /get-server-module [get]
func GetServerModule(c *gin.Context) {
	res := services.GetServerModule()
	if len(res) == 0 {
		c.JSON(http.StatusBadRequest, res)
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary Get System Runtime Status
// @Tags    Env
// @Accept  json
// @Produce json
// @Success 200 {object} gin.H
// @Router /system/status [get]
func GetSystemStatus(c *gin.Context) {
	status := global.GetRuntimeStatus()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"body":    status,
	})
}
