package main

import "github.com/gin-gonic/gin"

func NewRouter() *gin.Engine {
	// 启动时应自动调用一次refreshService方法
	RefreshService()

	router := gin.Default()
	apiGroup := router.Group("/api")
	commonApi := &CommonApi{}
	apiGroup.GET("/refresh", commonApi.Refresh)
	apiGroup.GET("/get-property", commonApi.ListProperty)

	videoGroup := apiGroup.Group("/video")
	videoApi := &VideoApi{}
	videoGroup.GET("/list", videoApi.List)
	videoGroup.PUT("/update-favor", videoApi.UpdateFavor)
	videoGroup.PUT("/update-custom", videoApi.UpdateCustomInfo)
	videoGroup.PUT("/batch-add", videoApi.BatchAddPeopleOrTag)

	return router
}
