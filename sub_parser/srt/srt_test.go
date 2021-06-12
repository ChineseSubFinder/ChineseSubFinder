package srt

import (
	"testing"
)

func TestParser_DetermineFileType(t *testing.T) {
	//filePath := "C:\\Tmp\\saw9.srt"
	//filePath := "C:\\tmp\\[zimuku]_0_oslo.2021.1080p.web.h264-naisu.简体&英文.srt"
	filePath := "C:\\tmp\\oslo.2021.1080p.web.h264-naisu.简体&英文.srt"
	parser := NewParser()
	sfi, err := parser.DetermineFileTypeFromFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	println(sfi.Name, sfi.Lang.String(), sfi.Ext)
}
