package v1

import (
	b64 "encoding/base64"
	"fmt"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/decode"
	backend2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/backend"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
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
	// base64 解码
	videoFPathUrlEncodeStr, err := b64.StdEncoding.DecodeString(videoFPathBase64)
	if err != nil {
		return
	}
	// url 解码
	videoFPath, err := url.QueryUnescape(string(videoFPathUrlEncodeStr))
	if err != nil {
		return
	}

	// 暂时不支持蓝光的预览
	if pkg.IsFile(videoFPath) == false {
		bok, _, _ := decode.IsFakeBDMVWorked(videoFPath)
		if bok == true {
			c.JSON(http.StatusOK, backend2.ReplyCommon{Message: "not support blu-ray preview"})
			return
		} else {
			c.JSON(http.StatusOK, backend2.ReplyCommon{Message: "video file not found"})
			return
		}
	}

	// segments/720/0/videofpathbase64
	template := fmt.Sprintf("/%s/preview/segments/{{.Resolution}}/{{.Segment}}/%v", cb.GetVersion(), videoFPathBase64)
	err = cb.hslCenter.WritePlaylist(template, videoFPath, c.Writer)
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
	// base64 解码
	videoFPathUrlEncodeStr, err := b64.StdEncoding.DecodeString(videoFPathBase64)
	if err != nil {
		return
	}
	// url 解码
	videoFPath, err := url.QueryUnescape(string(videoFPathUrlEncodeStr))
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
	err = cb.hslCenter.WriteSegment(videoFPath, segmentInt64, resolutionInt64, c.Writer)
	if err != nil {
		return
	}
}
