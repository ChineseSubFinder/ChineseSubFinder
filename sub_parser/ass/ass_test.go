package ass

import (
	"testing"
)

func TestParser_DetermineFileType(t *testing.T) {

	//filePath := "C:\\Tmp\\saw9.ass"
	//filePath := "C:\\tmp\\[zimuku]_0_oslo.2021.1080p.web.h264-naisu.简体&英文.ass"
	//filePath := "C:\\tmp\\oslo.2021.1080p.web.h264-naisu.简体&英文.ass"
	//filePath := "C:\\Tmp\\Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.chs[subhd].ass"
	//filePath := "C:\\Tmp\\oslo.2021.1080p.web.h264-naisu.繁体&英文.ass"
	filePath := "C:\\Tmp\\oslo.2021.1080p.web.h264-naisu.繁体.ass"
	parser := NewParser()
	sfi, err := parser.DetermineFileTypeFromFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	println(sfi.Name, sfi.Lang.String(), sfi.Ext)
}
