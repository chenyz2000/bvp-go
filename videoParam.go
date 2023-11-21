package main

type ListParam struct {
	Favor     []string `form:"favor"`     // 多选
	Direction string   `form:"direction"` // 单选
	Clarity   []string `form:"clarity"`
	People    []string `form:"people"`
	Tag       []string `form:"tag"`
	// Sort
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
