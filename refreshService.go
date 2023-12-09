package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
)

func RefreshService() {
	favorMap = Deserialize() // 旧json文件

	var newFavorMap FavorMap
	newFavorMap = make(FavorMap)

	favors, err := os.ReadDir(videoFolderPath)
	if err != nil {
		return
	}
	for _, favor := range favors {
		// 注意这层只能有文件夹，若有其他文件，则会跳过refresh
		favorName := favor.Name() // 收藏文件夹
		favorPath := videoFolderPath + favorName
		var infoMap InfoMap
		infoMap = make(InfoMap)

		items, err := os.ReadDir(favorPath)
		if err != nil {
			return
		}
		for _, item := range items {
			//fmt.Println(item.Name())
			itemName := item.Name() // 视频条目
			itemPath := favorPath + "/" + itemName

			pages, err := os.ReadDir(itemPath)
			if err != nil {
				return
			}
			for _, page := range pages {
				pageName := page.Name() // 分片，一般以c_开头
				pagePath := itemPath + "/" + pageName
				entryPath := pagePath + "/entry.json"
				entry := ParseJSON(entryPath)

				// 完全不管番剧了，因为番剧下载的实在太少了，只支持普通视频的xml
				key := itemName + videoNameConnector + pageName

				quality1 := getStringValue(entry, "quality_pithy_description") // 4K、1080P或其他
				quality2 := getStringValue(entry, "quality_superscript")       // 高码率或空字符串
				clarity := quality1 + quality2
				updateTime := getInt64Value(entry, "time_update_stamp")
				videoType := "single"
				if len(pages) > 1 {
					videoType = "multiple"
				}

				// cover
				// 之后可能需要根据web显示的要求调整这里的path，则需要根据旧cover值判断是否已下载
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

				// 在page_data中的数据
				var pageTitle, direction string
				var pageOrder, height, width int32
				pageData := getMapValue(entry, "page_data") // 子标签page_data转成的map
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
					return
				}
				if len(tmps) == 0 {
					os.Remove(pagePath)
					continue
				}
				for _, tmp := range tmps {
					if tmp.IsDir() {
						indexPath := pagePath + "/" + tmp.Name() + "/index.json"
						var indexJson Index_json
						bytes, err := os.ReadFile(indexPath)
						if err != nil {
							return
						}
						err = json.Unmarshal(bytes, &indexJson)
						fps, _ = strconv.ParseFloat(indexJson.Video[0].FrameRate, 64)
						break
					}
				}

				// CustomInfo不能清零了，要从旧的对象中读
				var customInfo *CustomInfo
				customInfo = findCustomInfo(favorMap, key)
				if customInfo.CollectionTime == 0 { // 如果收藏时间为空，则设置为视频更新时间
					customInfo.CollectionTime = updateTime
				}

				// 组装完整Info对象
				videoPage := &VideoInfo{
					Title:           getStringValue(entry, "title"),
					PageTitle:       pageTitle,
					PageOrder:       pageOrder,
					Transcoded:      PathExists(pagePath + "/out.mp4"),
					Type:            videoType,
					OwnerId:         getInt64Value(entry, "owner_id"),
					OwnerName:       getStringValue(entry, "owner_name"),
					MediaFolderName: getStringValue(entry, "type_tag"),
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
		if len(items) == 0 {
			os.Remove(favorPath)
			continue
		}

		newFavorMap[favorName] = infoMap
	}

	favorMap = newFavorMap

	// 写入info.json文件
	Serialize(favorMap)
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
	}
}
