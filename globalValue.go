package main

const (
	assetsFolderPath     = "../assets/"
	videoFolderPath      = assetsFolderPath + "video/"
	jsonFolderPath       = assetsFolderPath + "data/"
	jsonBackupFolderPath = jsonFolderPath + "backup/"
	coverFolderPath      = assetsFolderPath + "cover/"

	jsonFileName = "info.json"
	jsonFilePath = jsonFolderPath + jsonFileName

	videoNameConnector = ";" // itemName和pageName之间的分隔符
)

var favorMap FavorMap // 全局维护的对象
