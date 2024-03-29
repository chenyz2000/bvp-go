package main

const (
	assetsFolderPath         = "../bvp-assets/"
	originDownloadFolderPath = assetsFolderPath + "origin_download/"
	coverFolderPath          = assetsFolderPath + "cover/"
	intactVideoFolderPath    = assetsFolderPath + "intact_video/"

	jsonFolderPath       = assetsFolderPath + "data/"
	jsonBackupFolderPath = jsonFolderPath + "backup/"
	jsonFilePath         = jsonFolderPath + "info.json"

	automapFilePath = assetsFolderPath + "automap.json"

	videoNameConnector = ";" // itemName和pageName之间的分隔符

	CHANNEL_CAPACITY = 1
)

var favorMap FavorMap // 全局维护的对象
