package imdb_helper

import (
	"testing"

	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
)

func TestGetVideoInfoFromIMDB(t *testing.T) {
	//imdbID := "tt12708542" // 星球大战：残次品
	//imdbID := "tt7016936" // 杀死伊芙
	//imdbID := "tt2990738" 	// 恐怖直播
	//imdbID := "tt3032476" 	// 风骚律师
	//imdbID := "tt6468322" 	// 纸钞屋
	//imdbID := "tt15299712" // 云南虫谷
	imdbID := "tt6856242" // The King`s Man
	imdbInfo, err := GetVideoInfoFromIMDBWeb(types.VideoNfoInfo{ImdbId: imdbID})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("\n\n Name:  %s\tYear:  %d\tID:  %s", imdbInfo.Name, imdbInfo.Year, imdbInfo.ID)
}

func TestIsChineseVideo(t *testing.T) {
	type args struct {
		imdbID string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{name: "杀死伊芙", args: args{
			imdbID: "tt7016936",
		}, want: false, wantErr: false,
		},
		{name: "云南虫谷", args: args{
			imdbID: "tt15299712",
		}, want: true, wantErr: false,
		},
		{name: "扫黑风暴", args: args{
			imdbID: "tt15199554",
		}, want: true, wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := IsChineseVideo(log_helper.GetLogger4Tester(), types.VideoNfoInfo{ImdbId: tt.args.imdbID})
			if (err != nil) != tt.wantErr {
				t.Errorf("IsChineseVideo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsChineseVideo() got = %v, want %v", got, tt.want)
			}
		})
	}
}
