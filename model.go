package main

/*
对应info.json
*/
type CustomInfo struct {
	People      []string `json:"people"`
	Tag         []string `json:"tag"`
	Description string   `json:"description"`
	StarLevel   int32    `json:"star_level"` // 1仅收藏，2值得下载，3优质

	/* 以下字段只支持自动修改 */
	CollectionTime int64 `json:"colletion_time"` // 收藏时间
}

type VideoInfo struct { // 视频的每一个分片对应一个VideoInfo
	Title           string `json:"title"`
	PageTitle       string `json:"page_title"`
	PageOrder       int32  `json:"page_order"` // 当前分P的序号
	Transcoded      bool   `json:"transcoded"` // 已转码
	Type            string `json:"type"`       // single、multiple、TVseries
	OwnerId         int64  `json:"owner_id"`
	OwnerName       string `json:"owner_name"`
	MediaFolderName string `json:"media_folder_name"` // 媒体文件夹的名称，如16、112
	Cover           string `json:"cover"`
	//DownloadTime int64       `json:"download_time"`
	UpdateTime int64       `json:"update_time"`
	Direction  string      `json:"direction"`
	Size       int64       `json:"size"`     // 大小，单位字节
	Duration   int64       `json:"duration"` // 时长，单位毫秒
	Clarity    string      `json:"clarity"`  // 视频清晰度，如4K
	Height     int32       `json:"height"`
	Width      int32       `json:"width"`
	Fps        float64     `json:"fps"`
	Bvid       string      `json:"bvid"`
	Avid       int64       `json:"avid"`
	CustomInfo *CustomInfo `json:"custom_info"`
}

type InfoMap map[string]*VideoInfo

type FavorMap map[string]InfoMap

/*
对应index.json
*/
type Index_video struct {
	FrameRate string `json:"frame_rate"`
}

type Index_json struct {
	Video []*Index_video `json:"video"`
}

/*
propertyList接口使用
*/
type Property struct {
	Favor     CountMap `json:"favor"`
	People    CountMap `json:"people"`
	Tag       CountMap `json:"tag"`
	Clarity   CountMap `json:"clarity"` // 视频清晰度，如4K
	Direction CountMap `json:"direction"`
	Transcode CountMap `json:"transcode"`
}

type CountMap map[string]int
