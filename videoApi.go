package main

import (
	"github.com/gin-gonic/gin"
	"os"
	"strings"
)

type VideoApi struct {
}

// 根据筛选条件返回favorMap
func (v *VideoApi) List(c *gin.Context) {
	param := &ListParam{}
	err := c.ShouldBindQuery(param)
	if err != nil {
		ReturnFalse(c, "参数绑定错误")
	}
	newFavorMap := make(FavorMap)
	for favorName, infoMap := range favorMap {
		if !MatchStringList(favorName, param.Favor) {
			continue
		}
		newInfoMap := make(InfoMap)

		for name, videoInfo := range infoMap {
			if !MatchString(videoInfo.Direction, param.Direction) {
				continue
			}
			if !MatchStringList(videoInfo.Clarity, param.Clarity) {
				continue
			}
			if !HaveIntersection(videoInfo.CustomInfo.People, param.People) {
				continue
			}
			if !HaveIntersection(videoInfo.CustomInfo.Tag, param.Tag) {
				continue
			}
			// 其他条件
			// 筛完后，加入map
			newInfoMap[name] = videoInfo
		}

		// TODO param添加sort，返回sort后的列表，而不是返回map
		// 若某个收藏夹下筛选后不为空，则加入newFavorMap
		if len(newInfoMap) > 0 {
			newFavorMap[favorName] = newInfoMap
		}
	}
	c.JSON(200, newFavorMap)
}

// 批量修改视频的收藏夹，视频原收藏夹不限，只能设置一个目的收藏夹
func (v *VideoApi) UpdateFavor(c *gin.Context) {
	param := &UpdateFavorParam{}
	err := c.ShouldBindJSON(param)
	if err != nil {
		ReturnFalse(c, "参数绑定错误")
		return
	}
	newFavorPath := rootPath + param.NewFavorName
	for _, videoName := range param.VideoNameList {
		oldFavorName := findFavorName(videoName, favorMap)
		if oldFavorName == param.NewFavorName {
			continue
		}

		tmp := strings.Split(videoName, videoNameConnector)
		itemName := tmp[0]
		pageName := tmp[1]
		if !PathExists(newFavorPath) {
			err := os.MkdirAll(newFavorPath, 0777) // 创建favor文件夹
			if err != nil {
				ReturnFalse(c, newFavorPath+"文件夹创建错误")
				return
			}
		}
		newItemPath := newFavorPath + "/" + itemName
		if !PathExists(newItemPath) {
			err := os.MkdirAll(newItemPath, 0777) // 创建item文件夹
			if err != nil {
				ReturnFalse(c, newItemPath+"文件夹创建错误")
				return
			}
		}
		// 移动文件夹
		oldPagePath := rootPath + oldFavorName + "/" + itemName + "/" + pageName
		newPagePath := rootPath + param.NewFavorName + "/" + itemName + "/" + pageName
		err = os.Rename(oldPagePath, newPagePath)
		if err != nil {
			ReturnFalse(c, oldPagePath+"移动至"+newPagePath+"错误")
			return
		}
		oldItemPath := rootPath + oldFavorName + "/" + itemName
		dir, _ := os.ReadDir(oldItemPath)
		if len(dir) == 0 { // 若旧item目录为空，则删除
			err := os.Remove(oldItemPath)
			if err != nil {
				ReturnFalse(c, oldItemPath+"删除错误")
				return
			}
		}
		// 修改favorMap对象
		// TODO 修改CustomInfo的FavorName字段
		favorMap[param.NewFavorName][videoName] = favorMap[oldFavorName][videoName]
		delete(favorMap[oldFavorName], videoName)
	}
	// 写入文件
	Serialize(jsonPath, favorMap)
	c.JSON(200, nil)
}

// 根据videoName返回其favorName
func findFavorName(videoName string, favorMap FavorMap) string {
	for favorName, infoMap := range favorMap {
		for k, _ := range infoMap {
			if k == videoName {
				return favorName
			}
		}
	}
	return ""
}

// 修改视频的Custom信息，只能一次修改一个
func (v *VideoApi) UpdateCustomInfo(c *gin.Context) {
	param := &UpdateCustomInfoParam{}
	err := c.ShouldBindJSON(param)
	if err != nil {
		ReturnFalse(c, "参数绑定错误")
		return
	}

	for _, infoMap := range favorMap {
		for k, videoInfo := range infoMap {
			if k == param.VideoName {
				videoInfo.CustomInfo = param.CustomInfo
				break
			}
		}
	}
	Serialize(jsonPath, favorMap)
	c.JSON(200, nil)
}

// 批量添加人物或标签信息
func (v *VideoApi) BatchAddPeopleOrTag(c *gin.Context) {
	param := &BatchAddPeopleOrTagParam{}
	err := c.ShouldBindJSON(param)
	if err != nil {
		ReturnFalse(c, "参数绑定错误")
		return
	}

	for _, videoName := range param.VideoNameList {
		addPeopleOrTag(videoName, param)
	}
	Serialize(jsonPath, favorMap)
	c.JSON(200, nil)
}

func addPeopleOrTag(videoName string, param *BatchAddPeopleOrTagParam) {
	for _, infoMap := range favorMap {
		for k, videoInfo := range infoMap {
			if k == videoName {
				customInfo := videoInfo.CustomInfo
				customInfo.People = ConcatListUnique(customInfo.People, param.PeopleList)
				customInfo.Tag = ConcatListUnique(customInfo.Tag, param.TagList)
				return
			}
		}
	}
}
