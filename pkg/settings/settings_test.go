package settings

import (
	"reflect"
	"testing"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/strcut_json"
)

func TestNewSettings(t *testing.T) {

	inSettings := Settings{
		UserInfo: &UserInfo{
			Username: "abcd",
			Password: "123456",
		},
		CommonSettings: &CommonSettings{
			ScanInterval:     "12h",
			Threads:          12,
			RunScanAtStartUp: true,
			MoviePaths:       []string{"aaa", "bbb"},
			SeriesPaths:      []string{"ccc", "ddd"},
		},
		AdvancedSettings: &AdvancedSettings{
			ProxySettings: &ProxySettings{
				UseProxy:                 true,
				LocalHttpProxyServerPort: "123",
			},
			DebugMode:                  true,
			SaveFullSeasonTmpSubtitles: true,
			SubTypePriority:            1,
			SubNameFormatter:           1,
			SaveMultiSub:               true,
			CustomVideoExts:            []string{"aaa", "bbb"},
			FixTimeLine:                true,
		},
		EmbySettings: &EmbySettings{
			Enable:                true,
			AddressUrl:            "123456",
			APIKey:                "api123",
			MaxRequestVideoNumber: 1000,
			SkipWatched:           true,
			MoviePathsMapping:     map[string]string{"aa": "123", "bb": "456"},
			SeriesPathsMapping:    map[string]string{"aab": "123", "bbc": "456"},
		},
		DeveloperSettings: &DeveloperSettings{
			BarkServerAddress: "bark",
		},
	}

	err := strcut_json.ToFile(fileName, inSettings)
	if err != nil {
		t.Fatal(err)
	}

	outSettings := NewSettings(pkg.ConfigRootDirFPath())
	err = strcut_json.ToStruct(fileName, &outSettings)
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(inSettings.UserInfo, outSettings.UserInfo) == false {
		t.Fatal("inSettings Write And Read Not The Same")
	}
}

const fileName = "testfile.json"
