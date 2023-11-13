package main

type ListParam struct {
	Favor     []string `form:"favor"`     // 多选
	Direction string   `form:"direction"` // 单选
	Clarity   []string `form:"clarity"`
	People    []string `form:"people"`
	Tag       []string `form:"tag"`
	// Sort
}

func MatchStringList(str_info string, list_param []string) bool {
	if list_param == nil || len(list_param) == 0 {
		return true
	}
	// go没有contain方法，需要手动遍历
	for _, val := range list_param {
		if str_info == val {
			return true
		}
	}
	return false
}

func MatchString(str_info string, str_param string) bool {
	if str_param != "" && str_info != str_param {
		return false
	}
	return true
}

func HaveIntersection(list_info []string, list_param []string) bool {
	if list_info == nil || len(list_info) == 0 { // info中list为空则必不匹配
		return false
	}
	if list_param == nil || len(list_param) == 0 { // param中list为空则跳过筛选
		return true
	}
	for _, v := range list_info {
		if MatchStringList(v, list_param) {
			return true
		}
	}
	return false
}

type UpdateFavorParam struct {
	VideoNameList []string `json:"video_name_list"`
	NewFavorName  string   `json:"new_favor_name"`
}

func FindFavorName(videoName string, favorMap FavorMap) string {
	for favorName, infoMap := range favorMap {
		for k, _ := range infoMap {
			if k == videoName {
				return favorName
			}
		}
	}
	return ""
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
