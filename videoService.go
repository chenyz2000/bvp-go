package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"os"
	"sort"
	"strings"
)

type VideoService struct {
}

/*
筛选列表
*/
func (v *VideoService) List(param *ListParam) *ListResult {
	videoList := make([]*ListResultElement, 0)
	for favorName, infoMap := range favorMap {
		if !MatchStringList(favorName, param.Favor) {
			continue
		}

		for name, videoInfo := range infoMap {
			if !matchKeywords(videoInfo, param.Keywords) {
				continue
			}
			if param.Direction != "" && videoInfo.Direction != param.Direction {
				continue
			}
			if !HaveIntersection(videoInfo.CustomInfo.People, param.People) {
				continue
			}
			// 以下为不重要的选项
			if !HaveIntersection(videoInfo.CustomInfo.Tag, param.Tag) {
				continue
			}
			if !MatchStringList(videoInfo.Clarity, param.Clarity) {
				continue
			}
			if param.PeopleMarked != "" {
				peopleMarked := "未标注"
				if len(videoInfo.CustomInfo.People) > 0 {
					peopleMarked = "已标注"
				}
				if peopleMarked != param.PeopleMarked {
					continue
				}
			}
			if param.Transcode != "" {
				transcode := "未转码"
				if videoInfo.Transcoded {
					transcode = "已转码"
				}
				if transcode != param.Transcode {
					continue
				}
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
			if info1.CustomInfo.CollectionTime != info2.CustomInfo.CollectionTime {
				return info1.CustomInfo.CollectionTime > info2.CustomInfo.CollectionTime
			}
			return info1.UpdateTime > info2.UpdateTime
		case 3: // 星级，只用倒序
			if info1.CustomInfo.StarLevel != info2.CustomInfo.StarLevel {
				return info1.CustomInfo.StarLevel > info2.CustomInfo.StarLevel
			}
			return info1.UpdateTime > info2.UpdateTime
		case 4: // 标题名称中文拼音排序，只用顺序
			if info1.Title != info2.Title {
				return !gbkLess(info1.Title, info2.Title)
			}
			if info1.PageTitle != info2.PageTitle {
				return !gbkLess(info1.PageTitle, info2.PageTitle)
			}
			return info1.UpdateTime < info2.UpdateTime
		case 5: // UP主名称中文拼音排序，只用顺序
			if info1.OwnerName != info2.OwnerName {
				return !gbkLess(info1.OwnerName, info2.OwnerName)
			}
			return info1.UpdateTime < info2.UpdateTime
		}
		return info1.UpdateTime > info2.UpdateTime
	})
	if !desc { // 顺序
		reverse(videoList)
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

	return &ListResult{
		Count: count,
		List:  resList,
	}
}

func utf2gbk(src string) ([]byte, error) {
	GB18030 := simplifiedchinese.All[0]
	return io.ReadAll(transform.NewReader(bytes.NewReader([]byte(src)), GB18030.NewEncoder()))
}

func gbkLess(str1, str2 string) bool {
	a, _ := utf2gbk(str1)
	b, _ := utf2gbk(str2)
	bLen := len(b)
	for idx, chr := range a {
		if idx > bLen-1 {
			return false
		}
		if chr != b[idx] {
			return chr < b[idx]
		}
	}
	return true
}

func matchKeywords(info *VideoInfo, keywords string) bool {
	bytes, _ := json.MarshalIndent(info, "", "  ")
	s := string(bytes)
	for _, or := range strings.Split(keywords, "|") {
		match := true
		for _, and := range strings.Split(or, ",") {
			if !strings.Contains(s, and) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func reverse(s []*ListResultElement) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

/*
更新视频Favor
*/
func (v *VideoService) UpdateFavor(param *UpdateFavorParam) error {
	newFavorPath := originDownloadFolderPath + param.NewFavorName
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
			favorMap[param.NewFavorName] = make(InfoMap)
			if err != nil {
				fmt.Println("文件夹创建错误", newFavorPath)
				return err
			}
		}
		newItemPath := newFavorPath + "/" + itemName
		if !PathExists(newItemPath) {
			err := os.MkdirAll(newItemPath, 0777) // 创建item文件夹
			if err != nil {
				fmt.Println("文件夹创建错误", newItemPath)
				return err
			}
		}
		// 移动文件夹
		oldPagePath := originDownloadFolderPath + oldFavorName + "/" + itemName + "/" + pageName
		newPagePath := originDownloadFolderPath + param.NewFavorName + "/" + itemName + "/" + pageName
		err := os.Rename(oldPagePath, newPagePath)
		if err != nil {
			fmt.Println(oldPagePath + "移动至" + newPagePath + "错误")
			return err
		}
		oldItemPath := originDownloadFolderPath + oldFavorName + "/" + itemName
		dir, _ := os.ReadDir(oldItemPath)
		if len(dir) == 0 { // 若旧item目录为空，则删除
			err := os.Remove(oldItemPath)
			if err != nil {
				fmt.Println(oldItemPath + "删除错误")
				return err
			}
		}
		// 修改favorMap对象
		favorMap[param.NewFavorName][videoName] = favorMap[oldFavorName][videoName]
		delete(favorMap[oldFavorName], videoName)
	}
	// 写入文件
	Serialize(favorMap)
	return nil
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

/*
修改CustomInfo
*/
func (v *VideoService) UpdateCustomInfo(param *UpdateCustomInfoParam) {
	for _, infoMap := range favorMap {
		for k, videoInfo := range infoMap {
			if k == param.VideoName {
				// 保留原有的收藏时间
				param.CustomInfo.CollectionTime = videoInfo.CustomInfo.CollectionTime
				param.CustomInfo.VCodec = videoInfo.CustomInfo.VCodec
				videoInfo.CustomInfo = param.CustomInfo
				break
			}
		}
	}
	Serialize(favorMap)
}

/*
批量修改People或Tag
*/
func (v *VideoService) BatchAddPeopleOrTag(param *BatchAddPeopleOrTagParam) {
	for _, videoName := range param.VideoNameList {
		addPeopleOrTag(videoName, param)
	}
	Serialize(favorMap)
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
