package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	router := gin.Default()
	router.Use(cors.Default())

	// 将资源映射到url
	router.Static("/video", intactVideoFolderPath)
	router.Static("/cover", coverFolderPath)

	apiGroup := router.Group("/api")
	commonApi := &CommonApi{}
	apiGroup.GET("/refresh", commonApi.Refresh)
	apiGroup.GET("/get-property", commonApi.GetProperty)
	apiGroup.GET("/get-automap", commonApi.GetAutoMapJson)
	apiGroup.PUT("/update-automap", commonApi.UpdateAutoMap)
	//apiGroup.GET("/transcode", commonApi.Transcode)

	videoGroup := apiGroup.Group("/video")
	videoApi := NewVideoApi()
	videoGroup.GET("/list", videoApi.List)
	videoGroup.PUT("/update-favor", videoApi.UpdateFavor)
	videoGroup.PUT("/update-custom", videoApi.UpdateCustomInfo)
	videoGroup.PUT("/batch-add", videoApi.BatchAddPeopleOrTag)

	return router
}
