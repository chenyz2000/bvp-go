package main

const (
	assetsFolderPath     = "../assets/"
	videoFolderPath      = assetsFolderPath + "video/"
	jsonFolderPath       = assetsFolderPath + "data/"
	jsonBackupFolderPath = jsonFolderPath + "backup/"

	jsonFileName = "info.json"
	jsonFilePath = jsonFolderPath + jsonFileName

	webFolderPath   = "../bvp-web/"
	coverFolderPath = webFolderPath + "public/cover/"

	videoNameConnector = ";" // itemName和pageName之间的分隔符

	CHANNEL_CAPACITY = 1
)

var favorMap FavorMap // 全局维护的对象
