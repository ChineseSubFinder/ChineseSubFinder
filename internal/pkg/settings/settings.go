package settings

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/strcut_json"
	"github.com/huandu/go-clone"
	"sync"
)

type Settings struct {
	configFPath       string
	UserInfo          *UserInfo          `json:"user_info"`
	CommonSettings    *CommonSettings    `json:"common_settings"`
	AdvancedSettings  *AdvancedSettings  `json:"advanced_settings"`
	EmbySettings      *EmbySettings      `json:"emby_settings"`
	DeveloperSettings *DeveloperSettings `json:"developer_settings"`
}

// GetSettings 获取 Settings 的实例
func GetSettings() *Settings {
	if settings == nil {

		settingsOnce.Do(func() {
			settings = NewSettings()
			if my_util.IsFile(settings.configFPath) == false {
				// 配置文件不存在，新建一个空白的
				err := settings.Save()
				if err != nil {
					panic("Can't Save Config File:" + configName + " Error: " + err.Error())
				}
			} else {
				// 读取存在的文件
				err := settings.Read()
				if err != nil {
					panic("Can't Read Config File:" + configName + " Error: " + err.Error())
				}
			}
		})
	}
	return settings
}

// SetFullNewSettings 从 Web 端传入新的 Settings 完整设置
func SetFullNewSettings(inSettings *Settings) error {
	settings = inSettings
	return settings.Save()
}

func NewSettings() *Settings {

	nowConfigFPath := ""

	return &Settings{
		configFPath:       nowConfigFPath,
		UserInfo:          &UserInfo{},
		CommonSettings:    NewCommonSettings(),
		AdvancedSettings:  NewAdvancedSettings(),
		EmbySettings:      NewEmbySettings(),
		DeveloperSettings: NewDeveloperSettings(),
	}
}

func (s *Settings) Read() error {
	return strcut_json.ToStruct(s.configFPath, s)
}

func (s *Settings) Save() error {
	return strcut_json.ToFile(s.configFPath, s)
}

func (s Settings) GetNoPasswordSettings() *Settings {
	nowSettings := clone.Clone(s).(*Settings)
	nowSettings.UserInfo.Password = noPassword4Show
	return nowSettings
}

var (
	settings     *Settings
	settingsOnce sync.Once
)

const (
	noPassword4Show = "******" // 填充使用
	configName      = "ChineseSubFinderSettings.json"
)
