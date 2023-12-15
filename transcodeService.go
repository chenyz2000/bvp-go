package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

func Transcode() {
	fmt.Println("start transcode")
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
		if favor.Name() == "跳过转码" {
			continue
		}
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
					mediaFolderName := tmp.Name()                      // 媒体文件夹
					if !PathExists(pagePath + "/" + mediaFolderName) { // 防止在for循环中移动文件，导致旧路径下文件不存在
						continue
					}
					//如果文件大且未转码，则转移到“跳过转码”文件夹
					info, _ := os.Stat(pagePath + "/" + mediaFolderName + "/video.m4s")
					if info.Size() > 50*1024*1024 && !PathExists(pagePath+"/out.mp4") {
						// TODO 移动视频
						continue
					}

					transcodeOne(pagePath, mediaFolderName)
				}
			}
		}
	}

	//<-ch // 从channel消费，允许其他线程访问
	fmt.Println("finish transcode")
}

func transcodeOne(pagePath string, mediaFolderName string) {
	videoPath := pagePath + "/" + mediaFolderName + "/video.m4s"
	audioPath := pagePath + "/" + mediaFolderName + "/audio.m4s"
	outPath := pagePath + "/out_pre.mp4"
	if PathExists(outPath) {
		os.Remove(outPath) // 删除out.pre中间文件
	}

	startTime := time.Now().Unix()
	fmt.Println("ffmpeg transcode start, startTime:", startTime, ", pagePath: ", pagePath) // 时间单位：秒
	// 执行ffmpeg
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-i", audioPath, "-vcodec", "libx264", "-acodec", "copy", "-threads", "4", outPath)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("ffmpeg transcode error", pagePath)
		return
	}
	// 执行重命名，将out_pre.mp4转成out.mp4
	cmd = exec.Command("mv", pagePath+"/out_pre.mp4", pagePath+"/out.mp4")
	_, err = cmd.CombinedOutput()
	endTime := time.Now().Unix()
	fmt.Println("ffmpeg transcode finished, endTime:", endTime, ", costTime: ", endTime-startTime, ", pagePath: ", pagePath)

}
