package chs_cht_changer

import (
	"github.com/longbridgeapp/opencc"
	"os"
)

// Process 使用前务必转换字幕文件为 UTF-8 来使用，否则会遇到乱码
func Process(srcSubFileFPath string, desChineseLanguageType int) error {

	// 默认 0 是 简体 ，1 是 繁体
	fBytes, err := os.ReadFile(srcSubFileFPath)
	if err != nil {
		return err
	}

	orgString := string(fBytes)
	outString := ""
	if desChineseLanguageType == 0 {
		t2s, err := opencc.New("t2s")
		if err != nil {
			return err
		}
		// 繁体转简体
		outString, err = t2s.Convert(orgString)
		if err != nil {
			return err
		}
	} else {
		// 繁体
		s2t, err := opencc.New("s2t")
		if err != nil {
			return err
		}
		// 简体转繁体
		outString, err = s2t.Convert(orgString)
		if err != nil {
			return err
		}
	}

	err = os.WriteFile(srcSubFileFPath, []byte(outString), os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
