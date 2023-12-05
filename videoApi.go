package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"os"
	"os/exec"
	"sort"
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
	videoList := make([]*ListResultElement, 0)
	for favorName, infoMap := range favorMap {
		if !MatchStringList(favorName, param.Favor) {
			continue
		}

		for name, videoInfo := range infoMap {
			if !MatchStringList(videoInfo.Direction, param.Direction) {
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
			// 筛完后，加入resList
			ele := &ListResultElement{
				FavorName: favorName,
				ItemName:  strings.Split(name, videoNameConnector)[0],
				PageName:  strings.Split(name, videoNameConnector)[1],
				VideoInfo: videoInfo,
			}
			videoList = append(videoList, ele)
		}
	}

	// 根据sort排序
	sortType := param.Sort
	if !MatchIntList(sortType, []int{-1, -2, -3, -4, 1, 2, 3, 4}) {
		sortType = -1
	}
	desc := false
	if sortType < 0 {
		desc = true
		sortType = -sortType
	}
	sort.Slice(videoList, func(i, j int) bool {
		info1 := videoList[i].VideoInfo
		info2 := videoList[j].VideoInfo
		switch sortType {
		// TODO 对pageOrder进行排序
		case 1: // 更新时间
			return info1.UpdateTime > info2.UpdateTime
		case 2: // 收藏时间
			if info1.CustomInfo.CollectionTime != info1.CustomInfo.CollectionTime {
				return info1.CustomInfo.CollectionTime > info1.CustomInfo.CollectionTime
			}
			return info1.UpdateTime > info2.UpdateTime
		//TODO 如果想要中文排序，好像需要将utf-8转换为GBK
		//case 3: // 名称
		//	if info1.Title != info2.Title {
		//		return info1.Title > info2.Title
		//	}
		//	if info1.PageTitle != info2.PageTitle {
		//		return info1.PageTitle > info2.PageTitle
		//	}
		//	return info1.UpdateTime > info2.UpdateTime
		case 4: // 星级
			if info1.CustomInfo.StarLevel != info1.CustomInfo.StarLevel {
				return info1.CustomInfo.StarLevel > info1.CustomInfo.StarLevel
			}
			return info1.UpdateTime > info2.UpdateTime
		}
		return info1.UpdateTime > info2.UpdateTime
	})
	if !desc { // 顺序
		sort.Slice(videoList, func(i, j int) bool {
			return true
		})
	}

	// 分页
	count := len(videoList)
	resList := videoList
	if param.Page > 0 && param.PageSize > 0 {
		left := (param.Page - 1) * param.PageSize
		right := param.Page * param.PageSize
		if left >= count {
			resList = make([]*ListResultElement, 0)
		} else {
			if right >= count {
				right = count
			}
			resList = videoList[left:right]
		}
	}

	c.JSON(200, &ListResult{
		Count: count,
		List:  resList,
	})
}

// TODO 批量修改收藏夹、人物、标签时，若遇到multiple类型的视频，给出同item下视频未被框选的提示
// 批量修改视频的收藏夹，视频原收藏夹不限，只能设置一个目的收藏夹
func (v *VideoApi) UpdateFavor(c *gin.Context) {
	param := &UpdateFavorParam{}
	err := c.ShouldBindJSON(param)
	if err != nil {
		ReturnFalse(c, "参数绑定错误")
		return
	}
	newFavorPath := videoFolderPath + param.NewFavorName
	for _, videoName := range param.VideoNameList {
		oldFavorName := findFavorName(videoName, favorMap)
		if oldFavorName == param.NewFavorName {
			continue
		}
		mediaFolderName := favorMap[oldFavorName][videoName].MediaFolderName

		tmp := strings.Split(videoName, videoNameConnector)
		itemName := tmp[0]
		pageName := tmp[1]
		if !PathExists(newFavorPath) {
			err := os.MkdirAll(newFavorPath, 0777) // 创建favor文件夹
			favorMap[param.NewFavorName] = make(InfoMap)
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
		// 移动文件夹，直接移动不行，采用复制后删除旧的
		oldPagePath := videoFolderPath + oldFavorName + "/" + itemName + "/" + pageName
		newPagePath := videoFolderPath + param.NewFavorName + "/" + itemName + "/" + pageName
		cmd := exec.Command("cp", "-r", oldPagePath, newPagePath)
		out, err := cmd.CombinedOutput()
		fmt.Println(out)
		if err != nil {
			ReturnFalse(c, oldPagePath+"复制至"+newPagePath+"错误")
		}
		if !compareMD5(oldPagePath, newPagePath, mediaFolderName) {
			ReturnFalse(c, oldPagePath+"和"+newPagePath+"MD5值不同")
		}
		err = os.RemoveAll(oldPagePath)
		if err != nil {
			ReturnFalse(c, "删除"+oldPagePath+"错误")
		}

		oldItemPath := videoFolderPath + oldFavorName + "/" + itemName
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
	Serialize(favorMap)
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
				// 保留原有的收藏时间
				collectionTime := videoInfo.CustomInfo.CollectionTime
				videoInfo.CustomInfo = param.CustomInfo
				videoInfo.CustomInfo.CollectionTime = collectionTime
				break
			}
		}
	}
	Serialize(favorMap)
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
	Serialize(favorMap)
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

func compareMD5(oldPath string, newPath string, mediaFolderName string) bool {
	if getFileMd5(oldPath+"/entry.json") != getFileMd5(newPath+"/entry.json") {
		return false
	}
	if getFileMd5(oldPath+"/"+mediaFolderName+"/audio.mp3") !=
		getFileMd5(newPath+"/"+mediaFolderName+"/audio.mp3") {
		return false
	}
	if getFileMd5(oldPath+"/"+mediaFolderName+"/video.mp4") !=
		getFileMd5(newPath+"/"+mediaFolderName+"/video.mp4") {
		return false
	}
	return true
}

func getFileMd5(filePath string) string {
	pFile, err := os.Open(filePath)
	if err != nil {
		fmt.Errorf("打开文件失败，filename=%v, err=%v", filePath, err)
		return ""
	}
	defer pFile.Close()
	md5h := md5.New()
	io.Copy(md5h, pFile)
	return hex.EncodeToString(md5h.Sum(nil))
}
