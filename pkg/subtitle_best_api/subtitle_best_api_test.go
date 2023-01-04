package subtitle_best_api

import (
	"testing"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/random_auth_key"
)

func TestSubtitleBestApi_GetMediaInfo(t *testing.T) {

	pkg.ReadCustomAuthFile(log_helper.GetLogger4Tester())
	bapi := NewSubtitleBestApi(
		log_helper.GetLogger4Tester(),
		random_auth_key.AuthKey{
			BaseKey:  pkg.BaseKey(),
			AESKey16: pkg.AESKey16(),
			AESIv16:  pkg.AESIv16(),
		}, settings.Get().AdvancedSettings.ProxySettings)

	downloadTaskReply, err := bapi.AskDownloadTask("123")
	if err != nil {
		t.Fatal(err)
	}
	println(downloadTaskReply.Status, downloadTaskReply.Message)

	feedReply, err := bapi.FeedBack(pkg.RandStringBytesMaskImprSrcSB(64), "1.0.0", "None", true, true)
	if err != nil {
		t.Fatal(err)
	}
	println("FeedBack:", feedReply.Status, feedReply.Message)

	mediaInfo, err := bapi.GetMediaInfo("tt7278862", "imdb", "series")
	if err != nil {
		t.Fatal(err)
	}
	println(mediaInfo.TitleCN)

	convertIDResult, err := bapi.ConvertId("438148", "tmdb", "movie")
	if err != nil {
		t.Fatal(err)
	}
	println("IMDBId", convertIDResult.IMDBId)

	askFindSubReply, err := bapi.AskFindSub("0053b934afa0285e4de140e148c1c3768de73cfaad4170825c698308f8485c19",
		"tt4236770", "73586", "4", "1", "haha123456", "")
	if err != nil {
		t.Fatal(err)
	}
	println("AskFindSub.Status:", askFindSubReply.Status)
	println("AskFindSub.Message:", askFindSubReply.Message)

	findSubReply, err := bapi.FindSub("0053b934afa0285e4de140e148c1c3768de73cfaad4170825c698308f8485c19",
		"tt4236770", "73586", "4", "1", "haha123456", "")
	if err != nil {
		t.Fatal(err)
	}
	println("FindSub.Status:", findSubReply.Status)
	println("FindSub.Message:", findSubReply.Message)

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
