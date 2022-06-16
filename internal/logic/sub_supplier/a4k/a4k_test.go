package a4k

import (
	"testing"

	"github.com/allanpk716/ChineseSubFinder/internal/pkg/cache_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/global_value"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/random_auth_key"

	"github.com/allanpk716/ChineseSubFinder/internal/logic/file_downloader"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
)

func TestSupplier_searchKeyword(t *testing.T) {

	keyword := "绝命律师 第四季"
	defInstance()
	gotOutSubInfos, err := a4kInstance.searchKeyword(keyword)
	if err != nil {
		t.Fatal(err)
	}

	for i, searchResultItem := range gotOutSubInfos {
		println(i, searchResultItem.Title)
	}
}

var a4kInstance *Supplier

func defInstance() {

	my_util.ReadCustomAuthFile(log_helper.GetLogger4Tester())

	authKey := random_auth_key.AuthKey{
		BaseKey:  global_value.BaseKey(),
		AESKey16: global_value.AESKey16(),
		AESIv16:  global_value.AESIv16(),
	}

	nowSettings := settings.GetSettings()
	nowSettings.ExperimentalFunction.ShareSubSettings.ShareSubEnabled = true

	a4kInstance = NewSupplier(file_downloader.NewFileDownloader(
		cache_center.NewCacheCenter("test", nowSettings, log_helper.GetLogger4Tester()), authKey))
}
