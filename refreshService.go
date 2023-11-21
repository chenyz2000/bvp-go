package main

import (
	"encoding/json"
	"os"
	"strconv"
)

func RefreshService() {
	favorMap = Deserialize(jsonPath) // 旧json文件

	var newFavorMap FavorMap
	newFavorMap = make(FavorMap)

	favors, err := os.ReadDir(rootPath)
	if err != nil {
		return
	}
	for _, favor := range favors {
		if favor.Name() != jsonFileName { // go没有filter方法
			//fmt.Println(favor.Name())
			favorName := favor.Name() // 收藏文件夹
			favorPath := rootPath + favorName
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
				// fmt.Println(len(pages))		// 分片数量
				for _, page := range pages {
					pageName := page.Name() // 分片，一般以c_开头
					pagePath := itemPath + "/" + pageName
					entryPath := pagePath + "/entry.json"
					entry := ParseJSON(entryPath)

					// TODO 先不管番剧了，先支持普通视频的xml
					// TODO 待处理的字段
					cover := ""
					//

					quality1 := getStringValue(entry, "quality_pithy_description") // 4K、1080P或其他
					quality2 := getStringValue(entry, "quality_superscript")       // 高码率或空字符串
					clarity := quality1 + quality2
					updateTime := getInt64Value(entry, "time_update_stamp")

					// 在page_data中的数据
					var pageTitle, direction string
					var height, width int32
					pageData := getMapValue(entry, "page_data") // 子标签page_data转成的map
					if pageData != nil {
						pageTitle = getStringValue(pageData, "part")
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
					key := itemName + videoNameConnector + pageName
					customInfo = findCustomInfo(favorMap, key)

					// 组装完整Info对象
					videoPage := &VideoInfo{
						Title:      getStringValue(entry, "title"),
						PageTitle:  pageTitle,
						Type:       "",
						OwnerId:    getInt64Value(entry, "owner_id"),
						OwnerName:  getStringValue(entry, "owner_name"),
						Cover:      cover,
						UpdateTime: updateTime,
						Direction:  direction,
						Size:       getInt64Value(entry, "total_bytes"),
						Duration:   getInt64Value(entry, "total_time_milli"),
						Clarity:    clarity,
						Height:     height,
						Width:      width,
						Fps:        fps,
						Bvid:       getStringValue(entry, "bvid"),
						Avid:       getInt64Value(entry, "avid"),
						CustomInfo: customInfo,
					}

					infoMap[key] = videoPage
				}
			}
			newFavorMap[favorName] = infoMap
		}
	}

	favorMap = newFavorMap

	// 写入info.json文件
	Serialize(jsonPath, favorMap)
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
