package router

import (
	"zbx-monitor/api"
	"zbx-monitor/logger"

	"github.com/gin-gonic/gin"
)

func Setup() *gin.Engine {

	r := gin.New()
	r.Use(logger.GinLogger(), logger.GinRecovery(true))

	// azureCloud API router
	v1 := r.Group("/v1")
	{
		v1.GET("/getZabbixStat/:id/", api.GetZabbixStat)
		// v1.GET("/azureCloud", azureCloud.Login)
		// v1.GET("/test", azureCloud.Logout)
	}

	return r
}
