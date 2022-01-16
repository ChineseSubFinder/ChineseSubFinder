package settings

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/strcut_json"
	"github.com/huandu/go-clone"
	"os"
	"sync"
)

type Settings struct {
	configFPath           string
	UserInfo              *UserInfo              `json:"user_info"`
	CommonSettings        *CommonSettings        `json:"common_settings"`
	AdvancedSettings      *AdvancedSettings      `json:"advanced_settings"`
	EmbySettings          *EmbySettings          `json:"emby_settings"`
	DeveloperSettings     *DeveloperSettings     `json:"developer_settings"`
	TimelineFixerSettings *TimelineFixerSettings `json:"timeline_fixer_settings"`
}

// GetSettings 获取 Settings 的实例
func GetSettings() *Settings {
	if settings == nil {

		settingsOnce.Do(func() {
			settings = NewSettings()
			if isFile(settings.configFPath) == false {
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
		configFPath:           nowConfigFPath,
		UserInfo:              &UserInfo{},
		CommonSettings:        NewCommonSettings(),
		AdvancedSettings:      NewAdvancedSettings(),
		EmbySettings:          NewEmbySettings(),
		DeveloperSettings:     NewDeveloperSettings(),
		TimelineFixerSettings: NewTimelineFixerSettings(),
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

// Check 检测，某些参数有范围限制
func (s *Settings) Check() {

	// 每个网站最多找 Top 几的字幕结果，评价系统成熟后，才有设计的意义
	if s.AdvancedSettings.Topic < 0 || s.AdvancedSettings.Topic > 3 {
		s.AdvancedSettings.Topic = 1
	}
	// 如果 Debug 模式开启了，强制设置线程数为1，方便定位问题
	if s.AdvancedSettings.DebugMode == true {
		s.CommonSettings.Threads = 1
	} else {
		// 并发线程的范围控制
		if s.CommonSettings.Threads <= 0 {
			s.CommonSettings.Threads = 1
		} else if s.CommonSettings.Threads >= 3 {
			s.CommonSettings.Threads = 3
		}
	}
}

// isDir 存在且是文件夹
func isDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// isFile 存在且是文件
func isFile(filePath string) bool {
	s, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return !s.IsDir()
}

var (
	settings     *Settings
	settingsOnce sync.Once
)

const (
	noPassword4Show = "******" // 填充使用
	configName      = "ChineseSubFinderSettings.json"
)
