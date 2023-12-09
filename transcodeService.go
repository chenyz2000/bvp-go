package main

import (
	"fmt"
	"os"
	"os/exec"
)

func Transcode() {
	cmd := exec.Command("ffmpeg", "-help")
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("don't have ffmpeg")
		return
		//ReturnFalse(c, "don't have ffmpeg")
	}

	favors, err := os.ReadDir(videoFolderPath)
	if err != nil {
		return
	}
	for _, favor := range favors {
		favorPath := videoFolderPath + favor.Name() // 收藏文件夹

		items, err := os.ReadDir(favorPath)
		if err != nil {
			return
		}
		for _, item := range items {
			itemPath := favorPath + "/" + item.Name() // 视频条目

			pages, err := os.ReadDir(itemPath)
			if err != nil {
				return
			}
			for _, page := range pages {
				pagePath := itemPath + "/" + page.Name() // 分片，一般以c_开头
				if PathExists(pagePath + "/out.mp4") {   // 已经转码过
					continue
				}

				tmps, err := os.ReadDir(pagePath)
				if err != nil {
					return
				}
				for _, tmp := range tmps {
					if !tmp.IsDir() {
						continue
					}
					mediaFolderName := tmp.Name() // 媒体文件夹
					transcodeOne(pagePath, mediaFolderName)
				}
			}
		}
	}
}

func transcodeOne(pagePath string, mediaFolderName string) {
	videoPath := pagePath + "/" + mediaFolderName + "/video.m4s"
	audioPath := pagePath + "/" + mediaFolderName + "/audio.m4s"
	outPath := pagePath + "/out_pre.mp4"
	if PathExists(outPath) {
		os.Remove(outPath)
	}
	fmt.Println("ffmpeg transcode start", pagePath)
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-i", audioPath, "-vcodec", "libx264", "-acodec", "copy", "-threads", "4", outPath)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("ffmpeg transcode error", pagePath)
		return
	}
	cmd = exec.Command("mv", pagePath+"/out_pre.mp4", pagePath+"/out.mp4")
	_, err = cmd.CombinedOutput()
	fmt.Println("ffmpeg transcode finished", pagePath)
}
