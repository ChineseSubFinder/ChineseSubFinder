package model

import (
	"testing"
)

func TestUnArchiveFile(t *testing.T) {

	desRoot := "C:\\Tmp"
	//file := "C:\\Tmp\\[subhd]_0_162236051219240.zip"
	file := "C:\\Tmp\\123.zip"
	//file := "C:\\Tmp\\Tmp.7z"
	//file := "C:\\Tmp\\[zimuku]_0_[zmk.pw]奥斯陆.Oslo.[WEB.1080P]中英文字幕.zip"

	//err := archiver.Unarchive(file, desRoot)
	//if err != nil {
	//	t.Fatal(err)
	//}
	err := UnArchiveFile(file, desRoot)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnArr(t *testing.T) {
	desRoot := "C:\\Tmp"
	//file := "C:\\Tmp\\[subhd]_0_162236051219240.zip"
	file := "C:\\Tmp\\123.zip"
	//file := "C:\\Tmp\\Tmp.7z"
	//file := "C:\\Tmp\\[zimuku]_0_[zmk.pw]奥斯陆.Oslo.[WEB.1080P]中英文字幕.zip"
	err := UnArr(file, desRoot)
	if err != nil {
		t.Fatal(err)
	}
}