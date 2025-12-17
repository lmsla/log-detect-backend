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

	// Public authentication routes (no auth required)
	auth := router.Group("/auth")
	{
		auth.POST("/login", controller.Login)
	}

	apiv1 := router.Group("api/v1")
	apiv1.Use(middleware.OptionalAuthMiddleware()) // Optional auth for backward compatibility

	// Protected authentication routes
	authProtected := apiv1.Group("/auth")
	authProtected.Use(middleware.AuthMiddleware())
	{
		authProtected.POST("/register", controller.Register)
		authProtected.GET("/profile", controller.GetProfile)
		authProtected.POST("/refresh", controller.RefreshToken)
		authProtected.GET("/users", controller.ListUsers)
		authProtected.GET("/users/:id", controller.GetUser)
		authProtected.PUT("/users/:id", controller.UpdateUser)
		authProtected.DELETE("/users/:id", controller.DeleteUser)
	}

	// Protected Receiver routes
	receiverGroup := apiv1.Group("/Receiver")
	receiverGroup.Use(middleware.AuthMiddleware())
	receiverGroup.Use(middleware.PermissionMiddleware("target", "read")) // Using target permission for receivers
	{
		receiverGroup.GET("/GetAll", controller.GetAllReceivers)
		receiverGroup.POST("/Create", controller.CreateReceiver).Use(middleware.PermissionMiddleware("target", "create"))
		receiverGroup.PUT("/Update", controller.UpdateReceiver).Use(middleware.PermissionMiddleware("target", "update"))
		receiverGroup.DELETE("/Delete/:id", controller.DeleteReceiver).Use(middleware.PermissionMiddleware("target", "delete"))
	}

	// Protected Target routes
	targetGroup := apiv1.Group("/Target")
	targetGroup.Use(middleware.AuthMiddleware())
	targetGroup.Use(middleware.PermissionMiddleware("target", "read"))
	{
		targetGroup.GET("/GetAll", controller.GetAllTargets)
		targetGroup.POST("/Create", controller.CreateTarget).Use(middleware.PermissionMiddleware("target", "create"))
		targetGroup.PUT("/Update", controller.UpdateTarget).Use(middleware.PermissionMiddleware("target", "update"))
		targetGroup.DELETE("/Delete/:id", controller.DeleteTarget).Use(middleware.PermissionMiddleware("target", "delete"))
	}

	// Protected Device routes
	deviceGroup := apiv1.Group("/Device")
	deviceGroup.Use(middleware.AuthMiddleware())
	deviceGroup.Use(middleware.PermissionMiddleware("device", "read"))
	{
		deviceGroup.POST("/Create", controller.CreateDevice).Use(middleware.PermissionMiddleware("device", "create"))
		deviceGroup.GET("/GetAll", controller.GetAllDevices)
		deviceGroup.PUT("/Update", controller.UpdateDevice).Use(middleware.PermissionMiddleware("device", "update"))
		deviceGroup.DELETE("/Delete/:id", controller.DeleteDevice).Use(middleware.PermissionMiddleware("device", "delete"))
		deviceGroup.GET("/count", controller.GetTableCounts)
		deviceGroup.GET("/GetGroup", controller.GetDeviceGroup)
	}

	// Protected DeviceGroup routes
	deviceGroupGroup := apiv1.Group("/DeviceGroup")
	deviceGroupGroup.Use(middleware.AuthMiddleware())
	deviceGroupGroup.Use(middleware.PermissionMiddleware("device", "read"))
	{
		deviceGroupGroup.POST("/Create", controller.CreateDeviceGroup).Use(middleware.PermissionMiddleware("device", "create"))
		deviceGroupGroup.GET("/GetAll", controller.GetAllDeviceGroups)
		deviceGroupGroup.GET("/Get/:id", controller.GetDeviceGroupByID)
		deviceGroupGroup.PUT("/Update", controller.UpdateDeviceGroup).Use(middleware.PermissionMiddleware("device", "update"))
		deviceGroupGroup.DELETE("/Delete/:id", controller.DeleteDeviceGroup).Use(middleware.PermissionMiddleware("device", "delete"))

		// Batch operations
		deviceGroupGroup.POST("/MoveDevices", controller.MoveDevicesToGroup).Use(middleware.PermissionMiddleware("device", "update"))
		deviceGroupGroup.POST("/MoveGroupDevices", controller.MoveGroupDevices).Use(middleware.PermissionMiddleware("device", "update"))
	}

	// Protected Indices routes
	indicesGroup := apiv1.Group("/Indices")
	indicesGroup.Use(middleware.AuthMiddleware())
	indicesGroup.Use(middleware.PermissionMiddleware("indices", "read"))
	{
		indicesGroup.POST("/Create", controller.CreateIndices).Use(middleware.PermissionMiddleware("indices", "create"))
		indicesGroup.GET("/GetAll", controller.GetAllIndices)
		indicesGroup.PUT("/Update", controller.UpdateIndices).Use(middleware.PermissionMiddleware("indices", "update"))
		indicesGroup.DELETE("/Delete/:id", controller.DeleteIndices).Use(middleware.PermissionMiddleware("indices", "delete"))
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

	// Dashboard routes (for visualization and monitoring)
	dashboardGroup := apiv1.Group("/dashboard")
	dashboardGroup.Use(middleware.AuthMiddleware())
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

		// Admin data management routes
		adminGroup := apiv1.Group("/admin")
		adminGroup.Use(middleware.AuthMiddleware())
		{
			dataGroup := adminGroup.Group("/data")
			dataGroup.DELETE("/clean-history", controller.CleanOldHistory)
			dataGroup.POST("/archive-history", controller.ArchiveOldHistory)
			dataGroup.POST("/create-aggregates", controller.CreateDailyAggregates)
			dataGroup.GET("/storage-stats", controller.GetStorageStats)
		}
	}

	// Protected ES Connection routes
	esConnectionGroup := apiv1.Group("/ESConnection")
	esConnectionGroup.Use(middleware.AuthMiddleware())
	esConnectionGroup.Use(middleware.PermissionMiddleware("indices", "read"))
	{
		esConnectionGroup.GET("/GetAll", controller.GetAllESConnections)
		esConnectionGroup.GET("/Get/:id", controller.GetESConnection)
		esConnectionGroup.POST("/Create", controller.CreateESConnection).Use(middleware.PermissionMiddleware("indices", "create"))
		esConnectionGroup.PUT("/Update", controller.UpdateESConnection).Use(middleware.PermissionMiddleware("indices", "update"))
		esConnectionGroup.DELETE("/Delete/:id", controller.DeleteESConnection).Use(middleware.PermissionMiddleware("indices", "delete"))
		esConnectionGroup.POST("/Test", controller.TestESConnection)
		esConnectionGroup.PUT("/SetDefault/:id", controller.SetDefaultESConnection).Use(middleware.PermissionMiddleware("indices", "update"))
		esConnectionGroup.PUT("/Reload/:id", controller.ReloadESConnection).Use(middleware.PermissionMiddleware("indices", "update"))
		esConnectionGroup.PUT("/ReloadAll", controller.ReloadAllESConnections).Use(middleware.PermissionMiddleware("indices", "update"))
	}

	// Protected Elasticsearch Monitor routes
	esGroup := apiv1.Group("/elasticsearch")
	esGroup.Use(middleware.AuthMiddleware())
	esGroup.Use(middleware.PermissionMiddleware("elasticsearch", "read"))
	{
		// Monitor configuration CRUD
		esGroup.GET("/monitors", controller.GetAllESMonitors)
		esGroup.GET("/monitors/:id", controller.GetESMonitorByID)
		esGroup.POST("/monitors", controller.CreateESMonitor).Use(middleware.PermissionMiddleware("elasticsearch", "create"))
		esGroup.PUT("/monitors", controller.UpdateESMonitor).Use(middleware.PermissionMiddleware("elasticsearch", "update"))
		esGroup.DELETE("/monitors/:id", controller.DeleteESMonitor).Use(middleware.PermissionMiddleware("elasticsearch", "delete"))

		// Monitor operations
		// Test API 已移除，請使用 /api/v1/ESConnection/Test 測試 ES 連線
		esGroup.POST("/monitors/:id/toggle", controller.ToggleESMonitor).Use(middleware.PermissionMiddleware("elasticsearch", "update"))

		// Monitor status and statistics
		esGroup.GET("/status", controller.GetAllESMonitorsStatus)
		esGroup.GET("/status/:id/history", controller.GetESMonitorHistory)
		esGroup.GET("/statistics", controller.GetESStatistics)

		// Alert management
		esGroup.GET("/alerts", controller.GetESAlerts)
		esGroup.GET("/alerts/:monitor_id", controller.GetESAlertByID)
		esGroup.POST("/alerts/:monitor_id/resolve", controller.ResolveESAlert).Use(middleware.PermissionMiddleware("elasticsearch", "update"))
		esGroup.PUT("/alerts/:monitor_id/acknowledge", controller.AcknowledgeESAlert).Use(middleware.PermissionMiddleware("elasticsearch", "update"))
	}

	// Environment routes (may need different permissions)
	apiv1.GET("/get-sso-url", controller.GetSSOURL)
	apiv1.GET("/user/get-server-menu", controller.GetServerMenu)
	apiv1.GET("/get-server-module", controller.GetServerModule)

	return router
}
