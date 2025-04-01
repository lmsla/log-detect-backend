package controller

import (
	"net/http"
	// "strconv"
	// "log-detect/handler"
	// "log-detect/models"
	// "log-detect/entities"
	"log-detect/services"
	"log-detect/global"
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

	// user := c.Keys["user"].(models.SSOUser)

	// var roleName string
	// // 根據 realm 跟 role 給 conf
	// if user.IsAdmin {
	// 	roleName = global.EnvConfig.SSO.AdminRole
	// } else {
	// 	roleName = global.EnvConfig.SSO.UserRole
	// }

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
