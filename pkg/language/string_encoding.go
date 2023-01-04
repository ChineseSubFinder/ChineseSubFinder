package language

import (
	"strings"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/charset"
	"github.com/axgle/mahonia"
	nzlov "github.com/nzlov/chardet"
	"github.com/sirupsen/logrus"
)

// ConvertToString 将字符串从原始编码转换到目标编码，需要配合字符串检测编码库使用 chardet.NewTextDetector()
func ConvertToString(log *logrus.Logger, src string, srcCode string, tagCode string) string {
	defer func() {
		if err := recover(); err != nil {
			log.Errorln("ConvertToString panic:", err)
		}
	}()
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}

// 感谢: https://blog.csdn.net/gaoluhua/article/details/109128154，解决了编码问题

// ChangeFileCoding2UTF8 自动检测文件的编码，然后转换到 UTF-8，但是导出 bytes 的时候会把头部的 BOM 信息去除
func ChangeFileCoding2UTF8(inBytes []byte) ([]byte, error) {
	best, err := detector.DetectBest(inBytes)
	utf8String := ""
	if err != nil {
		return nil, err
	}
	if best.Confidence < 90 {
		detectBest := nzlov.Mostlike(inBytes)
		utf8String, err = charset.ToUTF8(charset.Charset(detectBest), string(inBytes))
	} else {
		utf8String, err = charset.ToUTF8(charset.Charset(best.Charset), string(inBytes))
	}
	if err != nil {
		return nil, err
	}
	if utf8String == "" {
		return inBytes, nil
	}

	// 然后返回的时候需要去除头部的 BOM 信息
	dat := []byte(utf8String)
	if dat[0] == 0xef && dat[1] == 0xbb && dat[2] == 0xbf {
		dat = dat[3:]
	}
	// 在确认一次
	validUTF8String := strings.ToValidUTF8(string(dat[:]), "")

	return []byte(validUTF8String), nil
}

func ChangeFileCoding2GBK(inBytes []byte) ([]byte, error) {

	utf8Bytes, err := ChangeFileCoding2UTF8(inBytes)
	if err != nil {
		return nil, err
	}

	gbkString, err := charset.UTF8To("GBK", string(utf8Bytes))
	if err != nil {
		return nil, err
	}

	return []byte(gbkString), nil
}
