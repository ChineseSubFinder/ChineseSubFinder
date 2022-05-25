package subtitle_best_api

import (
	"testing"

	"github.com/allanpk716/ChineseSubFinder/internal/pkg/random_auth_key"
)

func TestSubtitleBestApi_GetMediaInfo(t *testing.T) {

	bapi := NewSubtitleBestApi(random_auth_key.AuthKey{
		BaseKey:  "xx",
		AESKey16: "xx",
		AESIv16:  "xx",
	})

	mediaInfo, err := bapi.GetMediaInfo("tt7278862", "imdb", "series")
	if err != nil {
		t.Fatal(err)
	}

	println(mediaInfo.TitleCN)
}
