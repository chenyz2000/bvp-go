package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strconv"
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

func getMapValueFromMap(mp map[string]interface{}, key string) map[string]interface{} {
	inter, ok := mp[key]
	if ok {
		v, ok := inter.(map[string]interface{})
		if ok {
			return v
		}
	}
	return nil
}

func getMapValueFromList(lst []interface{}, index int) map[string]interface{} {
	if len(lst) <= index {
		return nil
	}
	inter := lst[index]
	v, ok := inter.(map[string]interface{})
	if ok {
		return v
	}
	return nil
}

func getListValue(mp map[string]interface{}, key string) []interface{} {
	inter, ok := mp[key]
	if ok {
		v, ok := inter.([]interface{})
		if ok {
			return v
		}
	}
	return nil
}

/*
基础方法
*/
// 因为go没有contain方法，需要手动遍历
func ListContainsString(str string, lst []string) bool {
	// 如果list为空，返回false
	for _, val := range lst {
		if str == val {
			return true
		}
		// 当列表中的值以-开头时，视为匹配到
		if strings.HasPrefix(val, "-") && "-"+str == val {
			return true
		}
	}
	return false
}

func ListIntersection(infoList []string, paramList []string) bool {
	for _, v := range infoList {
		if ListContainsString(v, paramList) {
			return true
		}
	}
	return false
}

/*
供List接口筛选时匹配使用
*/
func ParamContainsString(infoStr string, paramStr string) bool {
	return ListContainsString(infoStr, strings.Split(paramStr, ","))
}

func ParamIntersectsList(infoList []string, paramStr string) bool {
	if infoList == nil || len(infoList) == 0 { // 若param中list不为空且info中list为空
		if strings.Contains(paramStr, "【未标注】") {
			return true
		}
		return false
	}
	return ListIntersection(infoList, strings.Split(paramStr, ","))
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

func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return true
	}
	return false
}

// 合并两个列表，且去重
func ConcatListUnique(first []string, second []string) []string {
	lst := first
	for _, s := range second {
		unique := true
		for _, s1 := range lst {
			if s == s1 {
				unique = false
				break
			}
		}
		if unique {
			lst = append(lst, s)
		}
	}
	res := make([]string, 0)
	for _, s := range lst {
		if strings.TrimSpace(s) != "" {
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

/*
举例：
data := GetOnlineVideoInfo("BV1PW41127BG")
getInt64Value(data, "pubdate") // 发布时间，时间戳（秒）
// 还可以获取视频简介，点赞观看投币数等
*/
func GetOnlineVideoInfo(avid int64) map[string]interface{} {
	client := http.Client{}
	response, err := client.Get("https://api.bilibili.com/x/web-interface/view?aid=" + strconv.FormatInt(avid, 10))
	if err != nil {
		fmt.Println("获取视频在线信息失败", avid)
		return nil
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		fmt.Println("响应码错误")
		return nil
	}
	data, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("读取响应数据失败")
		return nil
	}
	var mp map[string]interface{}
	err = json.Unmarshal(data, &mp)
	return getMapValueFromMap(mp, "data")
}
