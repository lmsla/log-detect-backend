package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"log-detect/controller"
	"log-detect/middleware"
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
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Log Detect API is running"})
	})

	authEnabled := global.EnvConfig.Features.Auth

	// === Helper: 根據 auth feature 決定使用真實 middleware 或 pass-through ===
	authMW := noopMiddleware()
	permMW := func(resource, action string) gin.HandlerFunc { return noopMiddleware() }
	optionalAuthMW := noopMiddleware()
	if authEnabled {
		authMW = middleware.AuthMiddleware()
		permMW = middleware.PermissionMiddleware
		optionalAuthMW = middleware.OptionalAuthMiddleware()
	}

	// === Public authentication routes（僅在 auth 啟用時註冊）===
	if authEnabled {
		auth := router.Group("/auth")
		{
			auth.POST("/login", controller.Login)
		}
	}

	apiv1 := router.Group("api/v1")
	apiv1.Use(optionalAuthMW)

	// === Protected authentication routes（僅在 auth 啟用時註冊）===
	if authEnabled {
		authProtected := apiv1.Group("/auth")
		authProtected.Use(authMW)
		{
			authProtected.POST("/register", controller.Register)
			authProtected.GET("/profile", controller.GetProfile)
			authProtected.POST("/refresh", controller.RefreshToken)
			authProtected.GET("/users", controller.ListUsers)
			authProtected.GET("/users/:id", controller.GetUser)
			authProtected.PUT("/users/:id", controller.UpdateUser)
			authProtected.DELETE("/users/:id", controller.DeleteUser)
		}
	}

	// Protected Receiver routes
	receiverGroup := apiv1.Group("/Receiver")
	receiverGroup.Use(authMW)
	receiverGroup.Use(permMW("target", "read"))
	{
		receiverGroup.GET("/GetAll", controller.GetAllReceivers)
		receiverGroup.POST("/Create", controller.CreateReceiver).Use(permMW("target", "create"))
		receiverGroup.PUT("/Update", controller.UpdateReceiver).Use(permMW("target", "update"))
		receiverGroup.DELETE("/Delete/:id", controller.DeleteReceiver).Use(permMW("target", "delete"))
	}

	// Protected Target routes
	targetGroup := apiv1.Group("/Target")
	targetGroup.Use(authMW)
	targetGroup.Use(permMW("target", "read"))
	{
		targetGroup.GET("/GetAll", controller.GetAllTargets)
		targetGroup.POST("/Create", controller.CreateTarget).Use(permMW("target", "create"))
		targetGroup.PUT("/Update", controller.UpdateTarget).Use(permMW("target", "update"))
		targetGroup.DELETE("/Delete/:id", controller.DeleteTarget).Use(permMW("target", "delete"))
	}

	// Protected Device routes
	deviceGroup := apiv1.Group("/Device")
	deviceGroup.Use(authMW)
	deviceGroup.Use(permMW("device", "read"))
	{
		deviceGroup.POST("/Create", controller.CreateDevice).Use(permMW("device", "create"))
		deviceGroup.GET("/GetAll", controller.GetAllDevices)
		deviceGroup.PUT("/Update", controller.UpdateDevice).Use(permMW("device", "update"))
		deviceGroup.DELETE("/Delete/:id", controller.DeleteDevice).Use(permMW("device", "delete"))
		deviceGroup.GET("/count", controller.GetTableCounts)
		deviceGroup.GET("/GetGroup", controller.GetDeviceGroup)
	}

	// Protected DeviceGroup routes
	deviceGroupGroup := apiv1.Group("/DeviceGroup")
	deviceGroupGroup.Use(authMW)
	deviceGroupGroup.Use(permMW("device", "read"))
	{
		deviceGroupGroup.POST("/Create", controller.CreateDeviceGroup).Use(permMW("device", "create"))
		deviceGroupGroup.GET("/GetAll", controller.GetAllDeviceGroups)
		deviceGroupGroup.GET("/Get/:id", controller.GetDeviceGroupByID)
		deviceGroupGroup.PUT("/Update", controller.UpdateDeviceGroup).Use(permMW("device", "update"))
		deviceGroupGroup.DELETE("/Delete/:id", controller.DeleteDeviceGroup).Use(permMW("device", "delete"))

		// Batch operations
		deviceGroupGroup.POST("/MoveDevices", controller.MoveDevicesToGroup).Use(permMW("device", "update"))
		deviceGroupGroup.POST("/MoveGroupDevices", controller.MoveGroupDevices).Use(permMW("device", "update"))
	}

	// Protected Indices routes
	indicesGroup := apiv1.Group("/Indices")
	indicesGroup.Use(authMW)
	indicesGroup.Use(permMW("indices", "read"))
	{
		indicesGroup.POST("/Create", controller.CreateIndices).Use(permMW("indices", "create"))
		indicesGroup.GET("/GetAll", controller.GetAllIndices)
		indicesGroup.PUT("/Update", controller.UpdateIndices).Use(permMW("indices", "update"))
		indicesGroup.DELETE("/Delete/:id", controller.DeleteIndices).Use(permMW("indices", "delete"))
		indicesGroup.GET("/GetIndicesByLogname/:logname", controller.GetIndiceData)
		indicesGroup.GET("/GetIndicesByTargetID/:id", controller.GetIndicesByTargetID)
		indicesGroup.GET("/GetLogname", controller.GetLogname)
	}

	// Protected History routes
	historyGroup := apiv1.Group("/History")
	// historyGroup.Use(middleware.AuthMiddleware())
	{
		historyGroup.GET("/GetData/:logname", controller.GetHistoryData)
		historyGroup.GET("/GetLognameData", controller.GetLognameData)
	}

	// === Feature Toggle: Dashboard ===
	if global.EnvConfig.Features.Dashboard {
		dashboardGroup := apiv1.Group("/dashboard")
		dashboardGroup.Use(authMW)
		{
			dashboardGroup.GET("/overview", controller.GetDashboardOverview)
			dashboardGroup.GET("/statistics", controller.GetHistoryStatistics)
			dashboardGroup.GET("/trends", controller.GetTrendData)
			dashboardGroup.GET("/groups/statistics", controller.GetGroupStatistics)
			dashboardGroup.GET("/devices/status", controller.GetDeviceStatusOverview)
			dashboardGroup.GET("/devices/:device_name/timeline", controller.GetDeviceTimeline)
			dashboardGroup.GET("/alerts/recent", controller.GetRecentAlerts)
			dashboardGroup.POST("/alerts", controller.CreateAlert)
			dashboardGroup.PUT("/alerts/:id/status", controller.UpdateAlertStatus)

		}
	}

	// Protected ES Connection routes
	esConnectionGroup := apiv1.Group("/ESConnection")
	esConnectionGroup.Use(authMW)
	esConnectionGroup.Use(permMW("indices", "read"))
	{
		esConnectionGroup.GET("/GetAll", controller.GetAllESConnections)
		esConnectionGroup.GET("/Get/:id", controller.GetESConnection)
		esConnectionGroup.POST("/Create", controller.CreateESConnection).Use(permMW("indices", "create"))
		esConnectionGroup.PUT("/Update", controller.UpdateESConnection).Use(permMW("indices", "update"))
		esConnectionGroup.DELETE("/Delete/:id", controller.DeleteESConnection).Use(permMW("indices", "delete"))
		esConnectionGroup.POST("/Test", controller.TestESConnection)
		esConnectionGroup.PUT("/SetDefault/:id", controller.SetDefaultESConnection).Use(permMW("indices", "update"))
		esConnectionGroup.PUT("/Reload/:id", controller.ReloadESConnection).Use(permMW("indices", "update"))
		esConnectionGroup.PUT("/ReloadAll", controller.ReloadAllESConnections).Use(permMW("indices", "update"))
	}

	// === Feature Toggle: ES Monitoring ===
	if global.EnvConfig.Features.ESMonitoring {
		esGroup := apiv1.Group("/elasticsearch")
		esGroup.Use(authMW)
		esGroup.Use(permMW("elasticsearch", "read"))
		{
			// Monitor configuration CRUD
			esGroup.GET("/monitors", controller.GetAllESMonitors)
			esGroup.GET("/monitors/:id", controller.GetESMonitorByID)
			esGroup.POST("/monitors", controller.CreateESMonitor).Use(permMW("elasticsearch", "create"))
			esGroup.PUT("/monitors", controller.UpdateESMonitor).Use(permMW("elasticsearch", "update"))
			esGroup.DELETE("/monitors/:id", controller.DeleteESMonitor).Use(permMW("elasticsearch", "delete"))

			// Monitor operations
			esGroup.POST("/monitors/:id/toggle", controller.ToggleESMonitor).Use(permMW("elasticsearch", "update"))

			// Monitor status and statistics
			esGroup.GET("/status", controller.GetAllESMonitorsStatus)
			esGroup.GET("/status/:id/history", controller.GetESMonitorHistory)
			esGroup.GET("/statistics", controller.GetESStatistics)

			// Alert management
			esGroup.GET("/alerts", controller.GetESAlerts)
			esGroup.GET("/alerts/:monitor_id", controller.GetESAlertByID)
			esGroup.POST("/alerts/:monitor_id/resolve", controller.ResolveESAlert).Use(permMW("elasticsearch", "update"))
			esGroup.PUT("/alerts/:monitor_id/acknowledge", controller.AcknowledgeESAlert).Use(permMW("elasticsearch", "update"))
		}
	}

	// Environment routes (may need different permissions)
	apiv1.GET("/get-sso-url", controller.GetSSOURL)
	apiv1.GET("/user/get-server-menu", controller.GetServerMenu)
	apiv1.GET("/get-server-module", controller.GetServerModule)

	return router
}

// noopMiddleware 不做任何驗證，直接放行
func noopMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
