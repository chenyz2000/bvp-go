package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"
)

var randomSeed = time.Now().UnixNano()

type VideoService struct {
}

/*
筛选列表
*/
func (v *VideoService) List(param *ListParam) *ListResult {
	videoList := make([]*ListResultElement, 0)
	for favorName, infoMap := range favorMap {
		if param.ExcludeFavor != "" && ParamContainsString(favorName, param.ExcludeFavor) ||
			param.Favor != "" && !ParamContainsString(favorName, param.Favor) {
			continue
		}
		for name, videoInfo := range infoMap {
			if !matchKeywords(videoInfo, param.Keywords) {
				continue
			}
			if param.Direction != "" && videoInfo.Direction != param.Direction {
				continue
			}
			if param.Vcodec != "" && videoInfo.CustomInfo.Vcodec != param.Vcodec {
				continue
			}
			if param.ExcludePeople != "" && ParamIntersectsList(videoInfo.CustomInfo.People, param.ExcludePeople) ||
				param.People != "" && !ParamIntersectsList(videoInfo.CustomInfo.People, param.People) {
				continue
			}
			if param.ExcludeTag != "" && ParamIntersectsList(videoInfo.CustomInfo.Tag, param.ExcludeTag) ||
				param.Tag != "" && !ParamIntersectsList(videoInfo.CustomInfo.Tag, param.Tag) {
				continue
			}
			if param.MinDuration > 0 && videoInfo.Duration < param.MinDuration {
				continue
			}
			if param.MaxDuration > 0 && videoInfo.Duration > param.MaxDuration {
				continue
			}
			// 以下为不重要的选项
			if param.Clarity != "" && !ParamContainsString(videoInfo.Clarity, param.Clarity) {
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
	desc := false
	if sortType == 0 { // 随机排序
		// 需要先排序，因为从map拿出来的数据是乱序的
		sort.Slice(videoList, func(i, j int) bool {
			info1 := videoList[i].VideoInfo
			info2 := videoList[j].VideoInfo
			if info1.CustomInfo.PublishTime != info2.CustomInfo.PublishTime {
				return info1.CustomInfo.PublishTime > info2.CustomInfo.PublishTime
			}
			if info1.Title != info2.Title {
				return !gbkLess(info1.Title, info2.Title)
			}
			return !gbkLess(info1.PageTitle, info2.PageTitle)
		})
		if param.NewRandomSort {
			randomSeed = time.Now().UnixNano()
		}
		rand.Seed(randomSeed)
		//for i := len(videoList) - 1; i > 0; i-- { // Fisher–Yates shuffle
		//	j := rand.Intn(i + 1)
		//	videoList[i], videoList[j] = videoList[j], videoList[i]
		//}
		rand.Shuffle(len(videoList), func(i, j int) {
			videoList[i], videoList[j] = videoList[j], videoList[i]
		})
	} else {
		if sortType < 0 {
			desc = true
			sortType = -sortType
		}
		sort.Slice(videoList, func(i, j int) bool {
			info1 := videoList[i].VideoInfo
			info2 := videoList[j].VideoInfo
			switch sortType {
			// TODO 对pageOrder进行排序
			case 1: // 发布时间
				return info1.CustomInfo.PublishTime > info2.CustomInfo.PublishTime
			case 2: // 下载时间
				return info1.DownloadTime > info2.DownloadTime
			case 3: // 星级，只用倒序
				if info1.CustomInfo.StarLevel != info2.CustomInfo.StarLevel {
					return info1.CustomInfo.StarLevel > info2.CustomInfo.StarLevel
				}
				return info1.CustomInfo.PublishTime > info2.CustomInfo.PublishTime
			case 4: // 标题名称中文拼音排序，只用顺序
				if info1.Title != info2.Title {
					return !gbkLess(info1.Title, info2.Title)
				}
				if info1.PageTitle != info2.PageTitle {
					return !gbkLess(info1.PageTitle, info2.PageTitle)
				}
				return info1.CustomInfo.PublishTime < info2.CustomInfo.PublishTime
			case 5: // UP主名称中文拼音排序，只用顺序
				if info1.OwnerName != info2.OwnerName {
					return !gbkLess(info1.OwnerName, info2.OwnerName)
				}
				return info1.CustomInfo.PublishTime < info2.CustomInfo.PublishTime
			}
			return info1.CustomInfo.PublishTime > info2.CustomInfo.PublishTime
		})
		if !desc { // 顺序
			reverse(videoList)
		}
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

// 逗号表示与，竖划线表示或，减号表示非
func matchKeywords(info *VideoInfo, keywords string) bool {
	bytes, _ := json.MarshalIndent(info, "", "  ")
	s := string(bytes)
	for _, or := range strings.Split(keywords, "|") {
		match := true
		for _, and := range strings.Split(or, ",") {
			not := false
			if strings.HasPrefix(and, "-") {
				not = true
				and = strings.TrimPrefix(and, "-")
			}
			if strings.Contains(strings.ToLower(s), strings.ToLower(and)) == not {
				// 如果为非，not=ture，包含则不匹配；如果不为非，not=false，不包含则不匹配
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
	for _, videoName := range param.VideoNameList {
		oldFavorName := findFavorName(videoName, favorMap)
		if oldFavorName == param.NewFavorName {
			continue
		}

		tmp := strings.Split(videoName, videoNameConnector)
		itemName := tmp[0]
		pageName := tmp[1]
		// 移动文件夹
		err := MovePage(oldFavorName, itemName, pageName, param.NewFavorName, itemName, pageName)
		if err != nil {
			continue
		}
		// 修改favorMap对象
		favorMap[param.NewFavorName][videoName] = favorMap[oldFavorName][videoName]
		delete(favorMap[oldFavorName], videoName)
	}
	// 写入文件
	Serialize(favorMap)
	return nil
}

func MovePage(oldFavor string, oldItem string, oldPage string, newFavor string, newItem string, newPage string) error {
	oldItemPath := originDownloadFolderPath + oldFavor + "/" + oldItem
	oldPagePath := oldItemPath + "/" + oldPage
	newFavorPath := originDownloadFolderPath + newFavor
	newItemPath := newFavorPath + "/" + newItem
	newPagePath := newItemPath + "/" + newPage

	if !PathExists(newFavorPath) {
		err := os.MkdirAll(newFavorPath, 0777) // 创建favor文件夹
		favorMap[newFavor] = make(InfoMap)
		if err != nil {
			fmt.Println("文件夹创建错误", newFavorPath)
			return err
		}
	}

	if !PathExists(newItemPath) {
		err := os.MkdirAll(newItemPath, 0777) // 创建item文件夹
		if err != nil {
			fmt.Println("文件夹创建错误", newItemPath)
			return err
		}
	}

	err := os.Rename(oldPagePath, newPagePath)
	if err != nil {
		fmt.Println(oldPagePath + "移动至" + newPagePath + "错误")
		return err
	}
	// 若旧item目录为空，则删除
	dir, _ := os.ReadDir(oldItemPath)
	if len(dir) == 0 {
		err := os.Remove(oldItemPath)
		if err != nil {
			fmt.Println(oldItemPath + "删除错误")
			return err
		}
	}
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
				// 保留原有的不变数据
				param.CustomInfo.PublishTime = videoInfo.CustomInfo.PublishTime
				param.CustomInfo.CollectionTime = videoInfo.CustomInfo.CollectionTime
				param.CustomInfo.Vcodec = videoInfo.CustomInfo.Vcodec
				param.CustomInfo.OnlineDesc = videoInfo.CustomInfo.OnlineDesc
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
