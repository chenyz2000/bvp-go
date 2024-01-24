package main

import (
	"encoding/json"
	"fmt"
	ffgo "github.com/u2takey/ffmpeg-go"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func RefreshService() {
	// 读取automap.json
	bytes, err := os.ReadFile(automapFilePath)
	if err != nil {
		fmt.Println("can't read automap.json")
		return
	}
	var autoMap map[string]AutoMapItem
	err = json.Unmarshal(bytes, &autoMap)
	owner2People := inverseMap(autoMap["peopleByOwner"])
	title2People := inverseMap(autoMap["peopleByTitle"])
	title2Tag := inverseMap(autoMap["tagByTitle"])

	/*
		更新FavorMap
	*/
	favorMap = Deserialize() // 旧json文件

	var newFavorMap FavorMap
	newFavorMap = make(FavorMap)

	favors, err := os.ReadDir(originDownloadFolderPath)
	if err != nil {
		fmt.Println("can't read video folder path", originDownloadFolderPath)
		return
	}
	for _, favor := range favors {
		// 注意这层只能有文件夹，若有其他文件，则会跳过refresh
		favorName := favor.Name() // 收藏文件夹
		favorPath := originDownloadFolderPath + favorName
		if !IsDir(favorPath) || strings.HasPrefix(favorName, ".") || strings.HasPrefix(favorName, "@") {
			fmt.Println("file not comply with refresh rule:", favorPath)
			continue
		}
		var infoMap InfoMap
		infoMap = make(InfoMap)

		items, err := os.ReadDir(favorPath)
		if err != nil {
			fmt.Println("can't read favor path", favorPath)
			continue
		}
		for _, item := range items {
			itemName := item.Name() // 视频条目
			itemPath := favorPath + "/" + itemName
			match, _ := regexp.MatchString("^[0-9]+$", itemName) // 匹配纯数字字符串
			if !IsDir(itemPath) || !match {
				fmt.Println("file not comply with refresh rule:", itemPath)
				continue
			}

			pages, err := os.ReadDir(itemPath)
			if err != nil {
				fmt.Println("can't read item path", itemPath)
				continue
			}
			for _, page := range pages {
				pageName := page.Name() // 分片，一般以c_开头
				pagePath := itemPath + "/" + pageName
				match, _ := regexp.MatchString("^(c_)*[0-9]+$", pageName) // 匹配数字和c_字符串
				if !IsDir(pagePath) || !match {
					fmt.Println("file not comply with refresh rule:", pagePath)
					continue
				}

				entryPath := pagePath + "/entry.json"
				if !PathExists(entryPath) {
					fmt.Println("entry file doesn't exist", entryPath)
					continue
				}
				entry := ParseJSON(entryPath)

				// 完全不管番剧了，因为番剧下载的实在太少了，只支持普通视频的xml
				key := itemName + videoNameConnector + pageName

				quality1 := getStringValue(entry, "quality_pithy_description") // 4K、1080P或其他
				quality2 := getStringValue(entry, "quality_superscript")       // 高码率或空字符串
				clarity := quality1 + quality2
				updateTime := getInt64Value(entry, "time_create_stamp")
				videoType := "single"
				if len(pages) > 1 {
					videoType = "multiple"
				}

				// cover
				coverOnlineUrl := getStringValue(entry, "cover")
				pictureName := key + filepath.Ext(coverOnlineUrl)
				picturePath := coverFolderPath + pictureName
				if !PathExists(picturePath) {
					success := DownloadPicture(coverOnlineUrl, picturePath)
					if !success {
						pictureName = ""
					}
					// 下载会有一些耗时，之后在这里输出一下日志
				}
				cover := pictureName

				// 合并为intact.mp4
				mediaFolderName := getStringValue(entry, "type_tag")
				if mediaFolderName == "" {
					fmt.Println("can't get media folder", pagePath)
					continue
				}
				mediaFolderPath := pagePath + "/" + mediaFolderName + "/"
				intactOne(mediaFolderPath, key)

				// 在page_data中的数据
				var pageTitle, direction string
				var pageOrder, height, width int32
				pageData := getMapValueFromMap(entry, "page_data") // 子标签page_data转成的map
				if pageData != nil {
					pageTitle = getStringValue(pageData, "part")
					pageOrder = getInt32Value(pageData, "page")
					height = getInt32Value(pageData, "height")
					width = getInt32Value(pageData, "width")
					if height > 0 && width > 0 {
						if width > height {
							direction = "horizontal"
						} else {
							direction = "vertical"
						}
					}
				}

				// 在index.json中
				var fps float64
				tmps, err := os.ReadDir(pagePath)
				if err != nil {
					fmt.Println("can't read page path", pagePath)
					continue
				}
				if len(tmps) == 0 {
					os.Remove(pagePath)
					continue
				}

				indexPath := pagePath + "/" + mediaFolderName + "/index.json"
				var indexJson Index_json
				bytes, err := os.ReadFile(indexPath)
				if err != nil {
					fmt.Println("can't read index path", indexPath)
					continue
				}
				err = json.Unmarshal(bytes, &indexJson)
				fps, _ = strconv.ParseFloat(indexJson.Video[0].FrameRate, 64)

				/*
					获取CustomInfo
				*/
				ownerName := getStringValue(entry, "owner_name")
				title := getStringValue(entry, "title")
				//从旧favormap中读，而不是每次赋新值
				customInfo := findCustomInfo(favorMap, key)
				// 如果收藏时间为空，则设置为视频更新时间
				if customInfo.CollectionTime == 0 {
					customInfo.CollectionTime = updateTime
				}
				// 如果视频编码为空，获取视频编码
				if customInfo.Vcodec == "" {
					customInfo.Vcodec = getVideoCodec(pagePath + "/" + mediaFolderName + "/video.m4s")
				}
				//automap—owner2People
				target := owner2People[ownerName]
				if target != "" && !ListContainsString(target, customInfo.People) {
					customInfo.People = append(customInfo.People, target)
					fmt.Println("owner2people, key:", key, ", owner:", ownerName, ", people:", target)
				}
				//automap—title2People
				for k := range title2People {
					if strings.Contains(strings.ToLower(title), strings.ToLower(k)) &&
						!ListContainsString(title2People[k], customInfo.People) {
						customInfo.People = append(customInfo.People, title2People[k])
						fmt.Println("title2People, key:", key, ", title:", title, ", people:", k)
					}
				}
				//automap—title2Tag
				for k := range title2Tag {
					if strings.Contains(strings.ToLower(title), strings.ToLower(k)) &&
						!ListContainsString(title2Tag[k], customInfo.Tag) {
						customInfo.Tag = append(customInfo.Tag, title2Tag[k])
						fmt.Println("title2Tag, key:", key, ", title:", title, ", tag:", k)
					}
				}

				// 组装完整Info对象
				videoPage := &VideoInfo{
					Title:           title,
					PageTitle:       pageTitle,
					PageOrder:       pageOrder,
					Type:            videoType,
					OwnerId:         getInt64Value(entry, "owner_id"),
					OwnerName:       ownerName,
					MediaFolderName: mediaFolderName,
					Cover:           cover,
					UpdateTime:      updateTime,
					Direction:       direction,
					Size:            getInt64Value(entry, "total_bytes"),
					Duration:        getInt64Value(entry, "total_time_milli"),
					Clarity:         clarity,
					Height:          height,
					Width:           width,
					Fps:             fps,
					Bvid:            getStringValue(entry, "bvid"),
					Avid:            getInt64Value(entry, "avid"),
					CustomInfo:      customInfo,
				}

				infoMap[key] = videoPage
			}
			pages, err = os.ReadDir(itemPath)
			if len(pages) == 0 {
				os.Remove(itemPath)
				continue
			}
		}
		items, err = os.ReadDir(favorPath)
		if len(items) == 0 && favorName != "【待分类】" {
			os.Remove(favorPath)
			continue
		}

		newFavorMap[favorName] = infoMap
	}

	favorMap = newFavorMap

	// 写入info.json文件
	Serialize(favorMap)
}

// 转置Map，取list中的字符串作为键
func inverseMap(item AutoMapItem) map[string]string {
	res := make(map[string]string)
	for key := range item {
		lst := item[key]
		for _, s := range lst {
			res[s] = key
		}
	}
	return res
}

// 根据videoName（item+page）获取其CustomInfo
func findCustomInfo(favorMap FavorMap, key string) *CustomInfo {
	//var favor *InfoMap
	for _, infoMap := range favorMap {
		for k, videoInfo := range infoMap {
			if k == key {
				// 若找到对应的key则返回已有的CustomInfo
				customInfo := videoInfo.CustomInfo
				return customInfo
			}
		}
	}
	// 若没找到则初始化一个CustomInfo对象
	return &CustomInfo{
		People:      make([]string, 0),
		Tag:         make([]string, 0),
		Description: "",
		StarLevel:   0,
		NeedH264:    false,
	}
}

func getVideoCodec(videoPath string) string {
	str, err := ffgo.Probe(videoPath)
	if err != nil {
		fmt.Println("get video codec error:", videoPath)
		return ""
	}
	var mp map[string]interface{}
	err = json.Unmarshal([]byte(str), &mp)
	if err != nil {
		fmt.Println("ffprobe unmarshal error:", videoPath)
		return ""
	}
	stream := getMapValueFromList(getListValue(mp, "streams"), 0)
	if stream == nil {
		fmt.Println("stream is nil", videoPath)
		return ""
	}
	vcodec := getStringValue(stream, "codec_name")
	return vcodec
}

func intactOne(mediaFolderPath string, key string) {
	videoPath := mediaFolderPath + "video.m4s"
	audioPath := mediaFolderPath + "audio.m4s"

	intactPath := intactVideoFolderPath + key + ".mp4"
	if PathExists(intactPath) {
		return
	}

	startTime := time.Now().Unix()
	// 执行ffmpeg
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-i", audioPath, "-vcodec", "copy", "-acodec", "copy", "-threads", "4", intactPath)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("ffmpeg intact error:", err, "path:", mediaFolderPath)
		return
	}
	endTime := time.Now().Unix()
	costTime := endTime - startTime
	_ = costTime
	//fmt.Println("ffmpeg intact finished, endTime:", endTime, ", costTime: ", costTime, ", key: ", key)
}
