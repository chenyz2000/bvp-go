package main

type ListParam struct {
	Keywords  string `form:"keywords"`  // 逗号、竖划线分割
	Favor     string `form:"favor"`     // 多选，逗号分割
	Direction string `form:"direction"` // 单选。取值：horizontal，vertical
	People    string `form:"people"`    // 多选，逗号分割
	Tag       string `form:"tag"`       // 多选，逗号分割
	// 第二行
	Clarity       string `form:"clarity"`        // 多选，逗号分割
	Vcodec        string `form:"vcodec"`         // 单选
	ExcludeFavor  string `form:"exclude_favor"`  // 多选，逗号分割
	ExcludePeople string `form:"exclude_people"` // 多选，逗号分割
	ExcludeTag    string `form:"exclude_tag"`    // 多选，逗号分割
	MinDuration   int64  `form:"min_duration"`
	MaxDuration   int64  `form:"max_duration"`
	// 其他
	Sort          int  `form:"sort"`            // 负数为倒序，正数为顺序，0为随机。1发布时间、2下载时间、3名称、4星级，5UP主名。默认为-1
	NewRandomSort bool `form:"new_random_sort"` // 随机排序时，是否保留旧的排序
	Page          int  `form:"page"`
	PageSize      int  `form:"page_size"`
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
