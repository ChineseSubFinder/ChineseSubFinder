package cache_center

import (
	"github.com/allanpk716/ChineseSubFinder/pkg/types/language"
	"github.com/allanpk716/ChineseSubFinder/pkg/types/supplier"
	"testing"

	"github.com/allanpk716/ChineseSubFinder/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/pkg/settings"
)

func TestCacheCenter_DownloadFileAdd(t *testing.T) {
	cc := NewCacheCenter("testFile", settings.GetSettings(), log_helper.GetLogger4Tester())

	subInfo := supplier.NewSubInfo(
		"test",
		1,
		"name",
		language.ChineseSimple,
		"url123123",
		0,
		0,
		"ext",
		[]byte{1, 2, 3, 4, 5},
	)
	err := cc.DownloadFileAdd(subInfo)
	if err != nil {
		t.Fatal(err)
	}

	bok, getSubInfo, err := cc.DownloadFileGet(subInfo.GetUID())
	if err != nil {
		t.Fatal(err)
	}
	if bok == false {
		t.Fatal("bok == false")
	}

	if subInfo.FileUrl != getSubInfo.FileUrl {
		t.Fatal("subInfo.FileUrl != getSubInfo.FileUrl")
	}
}
