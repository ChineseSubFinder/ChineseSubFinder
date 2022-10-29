package v1

import (
	b64 "encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
)

// HlsPlaylist 获取 m3u8 列表
func (cb *ControllerBase) HlsPlaylist(c *gin.Context) {

	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "HlsPlaylist", err)
	}()

	videoFPathBase64 := c.Param("videofpathbase64")
	videoFPath, err := b64.StdEncoding.DecodeString(videoFPathBase64)
	if err != nil {
		return
	}

	// segments/720/0/videofpathbase64
	template := fmt.Sprintf("/%s/preview/segments/{{.Resolution}}/{{.Segment}}/%v", cb.GetVersion(), videoFPathBase64)
	err = cb.hslCenter.WritePlaylist(template, string(videoFPath), c.Writer)
	if err != nil {
		return
	}
}

// HlsSegment 获取具体一个 ts 文件
func (cb *ControllerBase) HlsSegment(c *gin.Context) {

	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "HlsSegment", err)
	}()

	resolution := c.Param("resolution")
	segment := c.Param("segment")
	videoFPathBase64 := c.Param("videofpathbase64")
	videoFPath, err := b64.StdEncoding.DecodeString(videoFPathBase64)
	if err != nil {
		return
	}
	segmentInt64, err := strconv.ParseInt(segment, 0, 64)
	if err != nil {
		return
	}
	resolutionInt64, err := strconv.ParseInt(resolution, 0, 64)
	if err != nil {
		return
	}
	err = cb.hslCenter.WriteSegment(string(videoFPath), segmentInt64, resolutionInt64, c.Writer)
	if err != nil {
		return
	}
}
