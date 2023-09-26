package imdb_helper

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/local_http_proxy_server"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/media_info_dealers"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/random_auth_key"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/subtitle_best_api"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/tmdb_api"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types"
	"testing"
)

func TestGetVideoInfoFromIMDB(t *testing.T) {
	//imdbID := "tt12708542" // 星球大战：残次品
	//imdbID := "tt7016936" // 杀死伊芙
	//imdbID := "tt2990738" 	// 恐怖直播
	//imdbID := "tt3032476" 	// 风骚律师
	//imdbID := "tt6468322" 	// 纸钞屋
	//imdbID := "tt15299712" // 云南虫谷
	//imdbID := "tt6856242" // The King`s Man
	//imdbInfo, err := getVideoInfoFromIMDBWeb(types.VideoNfoInfo{ImdbId: imdbID})
	//if err != nil {
	//	t.Fatal(err)
	//}
	//t.Logf("\n\n Name:  %s\tYear:  %d\tID:  %s", imdbInfo.Name, imdbInfo.Year, imdbInfo.ID)
}

func TestIsChineseVideo(t *testing.T) {
	type args struct {
		imdbID  string
		isMovie bool
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{name: "杀死伊芙", args: args{
			imdbID: "tt7016936", isMovie: false,
		}, want: false, wantErr: false,
		},
		{name: "云南虫谷", args: args{
			imdbID: "tt15299712", isMovie: false,
		}, want: true, wantErr: false,
		},
		{name: "扫黑风暴", args: args{
			imdbID: "tt15199554", isMovie: true,
		}, want: true, wantErr: false,
		},
		{name: "倚天屠龙记", args: args{
			imdbID: "tt1471140", isMovie: false,
		}, want: true, wantErr: false,
		},
		{name: "Only Murders in the Building", args: args{
			imdbID: "tt12851524", isMovie: false,
		}, want: false, wantErr: false,
		},
	}

	defInstance()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, imdbInfo, err := IsChineseVideo(dealers, types.VideoNfoInfo{ImdbId: tt.args.imdbID})
			if (err != nil) != tt.wantErr {
				t.Errorf("IsChineseVideo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsChineseVideo() got = %v, want %v", got, tt.want)
			}
			println("IsChineseVideo:", got)
			println(imdbInfo.Name)
			println(imdbInfo.Year)
			println(imdbInfo.Languages[0])
		})
	}
}

func defInstance() {

	settings.SetConfigRootPath(".")
	pkg.ReadCustomAuthFile(log_helper.GetLogger4Tester())

	authKey := random_auth_key.AuthKey{
		BaseKey:  pkg.BaseKey(),
		AESKey16: pkg.AESKey16(),
		AESIv16:  pkg.AESIv16(),
	}

	err := local_http_proxy_server.SetProxyInfo(settings.Get().AdvancedSettings.ProxySettings.GetInfos())
	if err != nil {
		panic(err)
	}
	dealers = media_info_dealers.NewDealers(log_helper.GetLogger4Tester(),
		subtitle_best_api.NewSubtitleBestApi(log_helper.GetLogger4Tester(), authKey))

	tmdbApi, err := tmdb_api.NewTmdbHelper(log_helper.GetLogger4Tester(), settings.Get().AdvancedSettings.TmdbApiSettings.ApiKey, true)
	if err != nil {
		panic(err)
	}
	dealers.SetTmdbHelperInstance(tmdbApi)
}

var (
	dealers *media_info_dealers.Dealers
)

func TestGetIMDBInfoFromVideoFile(t *testing.T) {

	defInstance()

	imdbInfo, err := GetIMDBInfoFromVideoFile(dealers, "X:\\电影\\西虹市首富 (2018)\\西虹市首富 (2018) 720p AAC.mkv", true)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("\n\n Name:  %s\tYear:  %d\tID:  %s", imdbInfo.Name, imdbInfo.Year, imdbInfo.Languages[0])

	imdbInfo, err = GetIMDBInfoFromVideoFile(dealers, "X:\\连续剧\\黄石 (2018)\\Season 1\\Yellowstone (2018) - S01E01 - Daybreak Bluray-480p.mkv", false)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("\n\n Name:  %s\tYear:  %d\tID:  %s", imdbInfo.Name, imdbInfo.Year, imdbInfo.Languages[0])
}
