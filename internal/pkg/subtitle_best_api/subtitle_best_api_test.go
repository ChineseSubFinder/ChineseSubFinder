package subtitle_best_api

import (
	"testing"

	"github.com/allanpk716/ChineseSubFinder/internal/pkg/global_value"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"

	"github.com/allanpk716/ChineseSubFinder/internal/pkg/random_auth_key"
)

func TestSubtitleBestApi_GetMediaInfo(t *testing.T) {

	my_util.ReadCustomAuthFile(log_helper.GetLogger4Tester())
	bapi := NewSubtitleBestApi(random_auth_key.AuthKey{
		BaseKey:  global_value.BaseKey(),
		AESKey16: global_value.AESKey16(),
		AESIv16:  global_value.AESIv16(),
	})

	mediaInfo, err := bapi.GetMediaInfo("tt7278862", "imdb", "series")
	if err != nil {
		t.Fatal(err)
	}

	println(mediaInfo.TitleCN)

	askForDownloadReply, err := bapi.AskDownloadSub("cd5e4bca49eea1f54f3eda5a38452b1c234075017857d010c76948124316cf2b",
		"haha123456", "")
	if err != nil {
		t.Fatal(err)
	}

	println("AskDownloadSub.Status:", askForDownloadReply.Status)
	println("AskDownloadSub.Message:", askForDownloadReply.Message)

	downloadSubReply, err := bapi.DownloadSub("cd5e4bca49eea1f54f3eda5a38452b1c234075017857d010c76948124316cf2b",
		"haha123456", "", "C:\\Tmp\\downloadhub\\123.srt")
	if err != nil {
		t.Fatal(err)
	}

	println("DownloadSub.Status", downloadSubReply.Status)
	println("DownloadSub.Message", downloadSubReply.Message)
}
