package main

import (
	"github.com/gin-gonic/gin"
	"os"
	"strings"
)

type VideoApi struct {
}

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

		// 若某个收藏夹下筛选后不为空，则加入newFavorMap
		if len(newInfoMap) > 0 {
			newFavorMap[favorName] = newInfoMap
		}
	}
	c.JSON(200, newFavorMap)
}

// 可以批量修改，视频原收藏夹不限，只能设置一个目的收藏夹
func (v *VideoApi) UpdateFavor(c *gin.Context) {
	param := &UpdateFavorParam{}
	err := c.ShouldBindJSON(param)
	if err != nil {
		ReturnFalse(c, "参数绑定错误")
		return
	}
	newFavorPath := rootPath + param.NewFavorName
	for _, videoName := range param.VideoNameList {
		oldFavorName := FindFavorName(videoName, favorMap)
		if oldFavorName == param.NewFavorName {
			continue
		}

		tmp := strings.Split(videoName, ";")
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
		favorMap[param.NewFavorName][videoName] = favorMap[oldFavorName][videoName]
		delete(favorMap[oldFavorName], videoName)
	}
	// 写入文件
	Serialize(jsonPath, favorMap)
	c.JSON(200, nil)
}

// 只能一次修改一个
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

// 将append列表中的元素加到origin列表中，且去重
func ConcatListUnique(first []string, second []string) []string {
	res := first
	for _, s := range second {
		unique := true
		for _, s1 := range res {
			if s == s1 {
				unique = false
				break
			}
		}
		if unique {
			res = append(res, s)
		}
	}
	return res
}
