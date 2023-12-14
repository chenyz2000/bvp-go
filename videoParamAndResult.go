package main

type ListParam struct {
	Keywords  string `form:"keywords"` // 筛选条件都是多选，以逗号分割
	Favor     string `form:"favor"`
	Direction string `form:"direction"`
	Clarity   string `form:"clarity"`
	People    string `form:"people"`
	Tag       string `form:"tag"`
	Sort      int    `form:"sort"` // -1更新时间倒序、-2收藏时间倒序、-3名称倒序、-4星级倒序，1~4为对应的顺序，默认为-1
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
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
