package main

type CustomInfo struct {
	People      []string `json:"people"`
	Tag         []string `json:"tag"`
	Description string   `json:"description"`
	StarLevel   int32    `json:"star_level"`
}

type VideoInfo struct { // 视频的每一个分片对应一个VideoInfo
	Title        string      `json:"title"`
	PageTitle    string      `json:"page_title"`
	Type         string      `json:"type"` // single、multiple、TVseries
	OwnerId      int64       `json:"owner_id"`
	OwnerName    string      `json:"owner_name"`
	Cover        string      `json:"cover"`
	DownloadTime int64       `json:"download_time"`
	UpdateTime   int64       `json:"update_time"`
	Direction    string      `json:"direction"`
	Size         int64       `json:"size"`    // 大小，单位字节
	Duration     int64       `json:"length"`  // 时长，单位毫秒
	Clarity      string      `json:"quality"` // 视频清晰度，如4K
	Height       int32       `json:"height"`
	Width        int32       `json:"width"`
	Fps          float64     `json:"fps"`
	Bvid         string      `json:"bvid"`
	Avid         int64       `json:"avid"`
	CustomInfo   *CustomInfo `json:"custom_info"`
}

type InfoMap map[string]*VideoInfo

type FavorMap map[string]InfoMap

// 对应index.json中需要的结构
type Index_video struct {
	FrameRate string `json:"frame_rate"`
}

type Index_json struct {
	Video []*Index_video `json:"video"`
}

// propertyList方法使用
type CountMap map[string]int

type Property struct {
	Favor  CountMap `json:"favor"`
	People CountMap `json:"people"`
	Tag    CountMap `json:"tag"`
}
