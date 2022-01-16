package strcut_json

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"reflect"
	"testing"
)

func TestToFile(t *testing.T) {

	inSettings := settings.Settings{
		UserInfo: &settings.UserInfo{
			Username: "abcd",
			Password: "123456",
		},
		CommonSettings: &settings.CommonSettings{
			UseHttpProxy:     true,
			HttpProxyAddress: "123",
			ScanInterval:     "12h",
			Threads:          12,
			RunScanAtStartUp: true,
			MoviePaths:       []string{"aaa", "bbb"},
			SeriesPaths:      []string{"ccc", "ddd"},
		},
		AdvancedSettings: &settings.AdvancedSettings{
			DebugMode:                  true,
			SaveFullSeasonTmpSubtitles: true,
			SubTypePriority:            1,
			SubNameFormatter:           1,
			SaveMultiSub:               true,
			CustomVideoExts:            []string{"aaa", "bbb"},
			FixTimeLine:                true,
		},
		EmbySettings: &settings.EmbySettings{
			Enable:                 true,
			AddressUrl:             "123456",
			APIKey:                 "api123",
			MaxRequestVideoNumber:  1000,
			SkipWatched:            true,
			MovieDirectoryMapping:  map[string]string{"aa": "123", "bb": "456"},
			SeriesDirectoryMapping: map[string]string{"aab": "123", "bbc": "456"},
		},
		DeveloperSettings: &settings.DeveloperSettings{
			BarkServerUrl: "bark",
		},
	}

	err := ToFile(fileName, inSettings)
	if err != nil {
		t.Fatal(err)
	}

	outSettings := settings.NewSettings()
	err = ToStruct(fileName, &outSettings)
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(inSettings.UserInfo, outSettings.UserInfo) == false {
		t.Fatal("inSettings Write And Read Not The Same")
	}
}

const fileName = "testfile.json"
