package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
)

/*
	Deserialize和Serialize方法，用于json字符串和favorMap之间的转换
*/
// 将字符串解析成favorMap
func Deserialize() FavorMap {
	bytes, err := os.ReadFile(jsonFilePath)
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

// 将favorMap序列化为string
func Serialize(favorMap FavorMap) {
	oldFavorMap := Deserialize()
	equal := reflect.DeepEqual(oldFavorMap, favorMap)
	if !equal {
		// 将旧的info.json转移到backup文件夹中
		backupPath := jsonBackupFolderPath + "info_" + time.Now().Format("20060102_150405") + ".json"
		err := os.Rename(jsonFilePath, backupPath)
		if err != nil {
			return
		}
		bytes, _ := json.MarshalIndent(favorMap, "", "  ")
		err = os.WriteFile(jsonFilePath, bytes, 0666)
		if err != nil {
			return
		}
	}
}

/*
ParseJSON和下面的getxxxValue方法，
是用于将json字符串，解析为map[string]interface{}，
以及从map[string]interface{}中读取值，转换为对应的数据类型
*/
func ParseJSON(filepath string) map[string]interface{} {
	var mp map[string]interface{}
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil
	}
	err = json.Unmarshal(bytes, &mp)
	return mp
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

/*
供List接口筛选时匹配使用
*/
func MatchIntList(num int, num_list []int) bool {
	if num_list == nil || len(num_list) == 0 {
		return false
	}
	// go没有contain方法，需要手动遍历
	for _, val := range num_list {
		if num == val {
			return true
		}
	}
	return false
}

func MatchStringList(str_info string, str_param string) bool { // str_param是以逗号分割的字符串
	if str_param == "" {
		return true
	}
	for _, val := range strings.Split(str_param, ",") {
		if str_info == val {
			return true
		}
	}
	return false
}

func HaveIntersection(list_info []string, str_param string) bool {
	if str_param == "" { // param中list为空则跳过筛选
		return true
	}
	if list_info == nil || len(list_info) == 0 { // 若param中list不为空且info中list为空则必不匹配
		return false
	}
	for _, v := range list_info {
		if MatchStringList(v, str_param) {
			return true
		}
	}
	return false
}

/*
不属于上面两组的其他方法
*/
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

// 合并两个列表，且去重
func ConcatListUnique(first []string, second []string) []string {
	res := first
	for _, s := range second {
		unique := true
		for _, s1 := range res {
			if s == s1 {
				unique = false
				break
			}
		}
		if unique {
			res = append(res, s)
		}
	}
	return res
}

func DownloadPicture(url string, filepath string) bool {
	client := http.Client{}
	response, err := client.Get(url)
	// TODO 访问不到咋办
	if err != nil {
		fmt.Println("获取图片失败")
		return false
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		fmt.Println("响应码错误")
		return false
	}
	data, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("读取响应数据失败")
		return false
	}
	err = os.WriteFile(filepath, data, 666)
	if err != nil {
		fmt.Println("写入文件失败")
		return false
	}
	return true
}

func ReturnFalse(c *gin.Context, data string) {
	c.JSON(500, data)
}
