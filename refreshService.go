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
	favorMap = Deserialize() // 旧json文件

	var newFavorMap FavorMap
	newFavorMap = make(FavorMap)

	favors, err := os.ReadDir(videoFolderPath)
	if err != nil {
		fmt.Println("can't read video folder path", videoFolderPath)
		return
	}
	for _, favor := range favors {
		// 注意这层只能有文件夹，若有其他文件，则会跳过refresh
		favorName := favor.Name() // 收藏文件夹
		favorPath := videoFolderPath + favorName
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
				updateTime := getInt64Value(entry, "time_update_stamp")
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
				intactOne(pagePath, mediaFolderName)

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

				// CustomInfo不能清零了，要从旧的对象中读
				var customInfo *CustomInfo
				customInfo = findCustomInfo(favorMap, key)
				if customInfo.CollectionTime == 0 { // 如果收藏时间为空，则设置为视频更新时间
					customInfo.CollectionTime = updateTime
				}
				if customInfo.VCodec == "" { // 如果视频编码为空，获取视频编码
					customInfo.VCodec = getVideoCodec(pagePath + "/" + mediaFolderName + "/video.mp4")
				}

				// 组装完整Info对象
				videoPage := &VideoInfo{
					Title:           getStringValue(entry, "title"),
					PageTitle:       pageTitle,
					PageOrder:       pageOrder,
					Transcoded:      PathExists(pagePath + "/intact.mp4"),
					Type:            videoType,
					OwnerId:         getInt64Value(entry, "owner_id"),
					OwnerName:       getStringValue(entry, "owner_name"),
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
	vCodec := getStringValue(stream, "codec_name")
	return vCodec
}

func intactOne(pagePath string, mediaFolderName string) {
	videoPath := pagePath + "/" + mediaFolderName + "/video.m4s"
	audioPath := pagePath + "/" + mediaFolderName + "/audio.m4s"
	intactPath := pagePath + "/intact.mp4"
	if PathExists(intactPath) {
		return
	}

	startTime := time.Now().Unix()
	// 执行ffmpeg
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-i", audioPath, "-vcodec", "copy", "-acodec", "copy", "-threads", "4", intactPath)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("ffmpeg intact error", pagePath)
		return
	}
	endTime := time.Now().Unix()
	costTime := endTime - startTime
	_ = costTime
	// fmt.Println("ffmpeg intact finished, endTime:", endTime, ", costTime: ", costTime, ", pagePath: ", pagePath)
}
