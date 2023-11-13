package main

import "github.com/gin-gonic/gin"

type CommonApi struct {
}

func (api *CommonApi) Refresh(c *gin.Context) {
	RefreshService()
	c.JSON(200, favorMap)
}

func (api *CommonApi) ListProperty(c *gin.Context) {
	// favor、people、tag
	// 因为go不支持set，所以用map的key去重，value表示key出现的次数
	favorCount := make(CountMap)
	peopleCount := make(CountMap)
	tagCount := make(CountMap)

	for favorName, infoMap := range favorMap {
		favorCount[favorName] = len(infoMap) // favor
		for _, videoInfo := range infoMap {
			customInfo := videoInfo.CustomInfo
			for _, v := range customInfo.People { //people
				peopleCount[v]++
			}
			for _, v := range customInfo.Tag { //tag
				tagCount[v]++
			}
		}
	}
	res := &Property{
		Favor:  favorCount,
		People: peopleCount,
		Tag:    tagCount,
	}
	c.JSON(200, *res)
}
