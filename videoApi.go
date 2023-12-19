package main

import (
	"github.com/gin-gonic/gin"
	"strings"
)

type VideoApi struct {
	videoService *VideoService
}

func NewVideoApi() *VideoApi {
	return &VideoApi{
		videoService: &VideoService{},
	}
}

// 根据筛选条件返回favorMap
func (v *VideoApi) List(c *gin.Context) {
	param := &ListParam{}
	err := c.ShouldBindQuery(param)
	if err != nil {
		c.JSON(500, "参数绑定错误")
		return
	}
	listResult := v.videoService.List(param)
	c.JSON(200, listResult)
}

// TODO 批量修改收藏夹、人物、标签时，若遇到multiple类型的视频，给出同item下视频未被框选的提示
// 批量修改视频的收藏夹，视频原收藏夹不限，只能设置一个目的收藏夹
func (v *VideoApi) UpdateFavor(c *gin.Context) {
	param := &UpdateFavorParam{}
	err := c.ShouldBindJSON(param)
	if err != nil {
		c.JSON(500, "参数绑定错误")
		return
	}
	if strings.TrimSpace(param.NewFavorName) == "" {
		c.JSON(500, "收藏夹名称不能为空")
		return
	}
	err = v.videoService.UpdateFavor(param)
	if err != nil {
		c.JSON(500, err)
		return
	}
	c.JSON(200, nil)
}

// 修改视频的Custom信息，只能一次修改一个
func (v *VideoApi) UpdateCustomInfo(c *gin.Context) {
	param := &UpdateCustomInfoParam{}
	err := c.ShouldBindJSON(param)
	if err != nil {
		c.JSON(500, "参数绑定错误")
		return
	}
	v.videoService.UpdateCustomInfo(param)
	c.JSON(200, nil)
}

// 批量添加人物或标签信息
func (v *VideoApi) BatchAddPeopleOrTag(c *gin.Context) {
	param := &BatchAddPeopleOrTagParam{}
	err := c.ShouldBindJSON(param)
	if err != nil {
		c.JSON(500, "参数绑定错误")
		return
	}
	v.videoService.BatchAddPeopleOrTag(param)
	c.JSON(200, nil)
}
