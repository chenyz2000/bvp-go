package main

import "github.com/gin-gonic/gin"

type CommonApi struct {
}

func (api *CommonApi) Refresh(c *gin.Context) {
	RefreshService()
	c.JSON(200, favorMap)
}

func (api *CommonApi) Transcode(c *gin.Context) {
	//if len(ch) == CHANNEL_CAPACITY {
	//	c.JSON(200, "already in transcode")
	//	return
	//}
	//ch <- 1 // 向channel发送，如果能发送则可以调用
	//go Transcode()
	//c.JSON(200, "start transcode")
}

// 返回对属性的统计，属性包括favor、people、tag
func (api *CommonApi) ListProperty(c *gin.Context) {
	// 因为go不支持set，所以用map的key去重，value表示key出现的次数
	favorCount := make(CountMap)
	peopleCount := make(CountMap)
	tagCount := make(CountMap)
	clarityCount := make(CountMap)
	directionCount := make(CountMap)
	vcodecCount := make(CountMap)
	for favorName, infoMap := range favorMap {
		favorCount[favorName] = len(infoMap) // favor
		for _, videoInfo := range infoMap {
			if videoInfo.Clarity != "" { // clarity
				clarityCount[videoInfo.Clarity]++
			}
			if videoInfo.Direction != "" { // direction
				directionCount[videoInfo.Direction]++
			}
			customInfo := videoInfo.CustomInfo
			vcodecCount[customInfo.VCodec]++      // vCodecCount
			for _, v := range customInfo.People { //people
				peopleCount[v]++
			}
			for _, v := range customInfo.Tag { //tag
				tagCount[v]++
			}
		}
	}
	res := &Property{
		Favor:     favorCount,
		People:    peopleCount,
		Tag:       tagCount,
		Clarity:   clarityCount,
		Direction: directionCount,
		Vcodec:    vcodecCount,
	}
	c.JSON(200, *res)
}
