package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"os"
)

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func Deserialize(filepath string) FavorMap {
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil
	}
	var jsonfile FavorMap
	err = json.Unmarshal(bytes, &jsonfile)
	if err != nil {
		return nil
	}
	return jsonfile
}

func Serialize(filepath string, favorMap FavorMap) {
	bytes, _ := json.MarshalIndent(favorMap, "", "  ")
	err := os.WriteFile(filepath, bytes, 0666)
	if err != nil {
		return
	}
}

func ParseJSON(filepath string) map[string]interface{} {
	var mp map[string]interface{}
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil
	}
	err = json.Unmarshal(bytes, &mp)
	return mp
}

func getMapValue(mp map[string]interface{}, key string) map[string]interface{} {
	inter, ok := mp[key]
	if ok {
		v, ok := inter.(map[string]interface{})
		if ok {
			return v
		}
	}
	return nil
}

func getStringValue(mp map[string]interface{}, key string) string {
	inter, ok := mp[key]
	if ok {
		v, ok := inter.(string)
		if ok {
			return v
		}
	}
	return ""
}

func getFloat64Value(mp map[string]interface{}, key string) float64 {
	inter, ok := mp[key]
	if ok {
		v, ok := inter.(float64)
		if ok {
			return v
		}
	}
	return 0
}

func getInt64Value(mp map[string]interface{}, key string) int64 {
	num := getFloat64Value(mp, key)
	return int64(num)
}

func getInt32Value(mp map[string]interface{}, key string) int32 {
	num := getFloat64Value(mp, key)
	return int32(num)
}

func FindCustomInfo(favorMap FavorMap, key string) *CustomInfo {
	//var favor *InfoMap
	for _, infoMap := range favorMap {
		for k, videoInfo := range infoMap {
			if k == key {
				// 若找到对应的key则返回已有的CustomInfo
				customInfo := videoInfo.CustomInfo
				return customInfo
			}
		}
	}
	// 若没找到则初始化一个CustomInfo对象
	return &CustomInfo{
		People:      make([]string, 0),
		Tag:         make([]string, 0),
		Description: "",
		StarLevel:   0,
	}
}

func ReturnFalse(c *gin.Context, data string) {
	c.JSON(500, data)
}
