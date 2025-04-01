package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"log-detect/controller"
	"log-detect/utils"

	_ "log-detect/docs"
	"log-detect/global"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func LoadRouter() *gin.Engine {

	gin.SetMode(global.EnvConfig.Server.Mode)
	router := gin.Default()

	router.Use(utils.CorsConfig())

	// swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("healthcheck", func(c *gin.Context) {
		c.JSON(http.StatusOK, "")
	})
	apiv1 := router.Group("api/v1")

	apiv1.GET("/Receiver/GetAll", controller.GetAllReceivers)
	apiv1.POST("/Receiver/Create", controller.CreateReceiver)
	apiv1.PUT("/Receiver/Update", controller.UpdateReceiver)
	apiv1.DELETE("/Receiver/Delete/:id", controller.DeleteReceiver)

	apiv1.GET("/Target/GetAll", controller.GetAllTargets)
	apiv1.POST("/Target/Create", controller.CreateTarget)
	apiv1.PUT("/Target/Update", controller.UpdateTarget)
	apiv1.DELETE("/Target/Delete/:id", controller.DeleteTarget)

	apiv1.POST("/Device/Create", controller.CreateDevice)
	apiv1.GET("/Device/GetAll", controller.GetAllDevices)
	apiv1.PUT("/Device/Update", controller.UpdateDevice)
	apiv1.DELETE("/Device/Delete/:id", controller.DeleteDevice)
	apiv1.GET("/Device/count", controller.GetTableCounts)
	apiv1.GET("/Device/GetGroup", controller.GetDeviceGroup)

	apiv1.POST("/Indices/Create", controller.CreateIndices)
	apiv1.GET("/Indices/GetAll", controller.GetAllIndices)
	apiv1.PUT("/Indices/Update", controller.UpdateIndices)
	apiv1.DELETE("/Indices/Delete/:id", controller.DeleteIndices)
	apiv1.GET("/Indices/GetIndicesByLogname/:logname",controller.GetIndiceData)
	apiv1.GET("/Indices/GetIndicesByTargetID/:id", controller.GetIndicesByTargetID)
	apiv1.GET("/Indices/GetLogname", controller.GetLogname)

	// history
	apiv1.GET("/History/GetData/:logname", controller.GetHistoryData)
	apiv1.GET("/History/GetLognameData", controller.GetLognameData)

	//*** enviroment ***//
	apiv1.GET("/get-sso-url", controller.GetSSOURL)
	apiv1.GET("/user/get-server-menu", controller.GetServerMenu)
	apiv1.GET("/get-server-module", controller.GetServerModule)

	return router
}
