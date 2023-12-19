package main

type ListParam struct {
	Keywords  string `form:"keywords"`  // 逗号、竖划线分割
	Favor     string `form:"favor"`     // 多选，逗号分割
	Direction string `form:"direction"` // 单选。取值：horizontal，vertical
	People    string `form:"people"`    // 多选，逗号分割
	// 第二行
	Tag          string `form:"tag"`           // 多选，逗号分割
	Clarity      string `form:"clarity"`       // 多选，逗号分割
	PeopleMarked string `form:"people_marked"` // 含义：已标注people。单选。取值：已标注，未标注
	Transcode    string `form:"transcode"`     // 单选。取值：已转码，未转码
	// 其他
	Sort     int `form:"sort"` // -1更新时间倒序、-2收藏时间倒序、-3名称倒序、-4星级倒序，1~4为对应的顺序，默认为-1
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type UpdateFavorParam struct {
	VideoNameList []string `json:"video_name_list"`
	NewFavorName  string   `json:"new_favor_name"`
}

type UpdateCustomInfoParam struct {
	VideoName  string      `json:"video_name"`
	CustomInfo *CustomInfo `json:"custom_info"`
}

type BatchAddPeopleOrTagParam struct {
	VideoNameList []string `json:"video_name_list"`
	PeopleList    []string `json:"people_list"`
	TagList       []string `json:"tag_list"`
}

type ListResult struct {
	Count int                  `json:"count"`
	List  []*ListResultElement `json:"list"`
}

type ListResultElement struct {
	FavorName string     `json:"favor_name"`
	ItemName  string     `json:"item_name"`
	PageName  string     `json:"page_name"`
	VideoInfo *VideoInfo `json:"video_info"`
}
