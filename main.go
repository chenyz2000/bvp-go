package main

import "os"

//var ch = make(chan int, CHANNEL_CAPACITY)

func main() {
	// 自动创建必要的文件
	if !PathExists(jsonFilePath) {
		os.Create(jsonFilePath)
	}
	if !PathExists(jsonBackupFolderPath) {
		os.Mkdir(jsonBackupFolderPath, 0777)
	}
	if !PathExists(coverFolderPath) {
		os.Mkdir(coverFolderPath, 0777)
	}
	if !PathExists(intactVideoFolderPath) {
		os.Mkdir(intactVideoFolderPath, 0777)
	}

	// 启动时应自动调用一次Transcode和refreshService方法
	//ch <- 1 // 向channel发送，如果能发送则可以调用
	//go Transcode()
	RefreshService()

	router := NewRouter()
	err := router.Run(":1024")
	if err != nil {
		return
	}
}
