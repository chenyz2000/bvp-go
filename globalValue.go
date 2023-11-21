package main

const (
	rootPath           = "../video/"
	jsonFileName       = "info.json"
	jsonPath           = rootPath + jsonFileName
	videoNameConnector = ";" // itemName和pageName之间的分隔符
)

var favorMap FavorMap // 全局维护的对象
