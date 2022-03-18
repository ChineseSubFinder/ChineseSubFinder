package emby_helper

import (
	"reflect"
	"testing"
)

// TODO 暂不方便在其他环境进行单元测试
func TestEmbyHelper_GetRecentlyAddVideoList(t *testing.T) {

	//embyConfig := settings.NewEmbySettings()
	//embyConfig.Enable = true
	//embyConfig.AddressUrl = "http://123:8096"
	//embyConfig.APIKey = "123"
	//embyConfig.SkipWatched = true
	//embyConfig.MoviePathsMapping["X:\\电影"] = "/mnt/share1/电影"
	//embyConfig.MoviePathsMapping["X:\\连续剧"] = "/mnt/share1/连续剧"
	//
	//em := NewEmbyHelper(*embyConfig)
	//movieList, seriesList, err := em.GetRecentlyAddVideoListWithNoChineseSubtitle()
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//println(len(movieList), len(seriesList))
}

// TODO 暂不方便在其他环境进行单元测试
func TestEmbyHelper_RefreshEmbySubList(t *testing.T) {
	//config := config.GetConfig()
	//em := NewEmbyHelper(config.EmbyConfig)
	//bok, err := em.RefreshEmbySubList()
	//if err != nil {
	//	t.Fatal(err)
	//}
	//println(bok)
}

// TODO 暂不方便在其他环境进行单元测试
func TestEmbyHelper_GetInternalEngSubAndExSub(t *testing.T) {
	//config := config.GetConfig()
	//em := NewEmbyHelper(config.EmbyConfig)
	//// 81873 -- R&M - S05E01
	//// R&M S05E10  2 org english, 5 简英 	145499
	//// 基地 S01E03 							166840
	//found, internalEngSub, exCh_EngSub, err := em.GetInternalEngSubAndExChineseEnglishSub("166840")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//if found == false {
	//	t.Fatal("need found sub")
	//}
	//
	//println(internalEngSub[0].FileName, exCh_EngSub[0].FileName)
}

func Test_sortStringSliceByLength(t *testing.T) {
	type args struct {
		m []string
	}
	tests := []struct {
		name string
		args args
		want PathSlices
	}{
		{
			name: "00",
			args: args{
				[]string{"/aa/bb/cc", "/aa", "/aa/bb"},
			},
			want: []PathSlice{{
				Path: "/aa/bb/cc",
			}, {
				Path: "/aa/bb",
			}, {
				Path: "/aa",
			},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sortStringSliceByLength(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sortStringSliceByLength() = %v, want %v", got, tt.want)
			}
		})
	}
}
