package settings

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/global_value"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/strcut_json"
	"github.com/huandu/go-clone"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Settings struct {
	SpeedDevMode          bool // 是否为开发模式，代码开启这个会跳过某些流程，加快测试速度
	configFPath           string
	UserInfo              *UserInfo              `json:"user_info"`
	CommonSettings        *CommonSettings        `json:"common_settings"`
	AdvancedSettings      *AdvancedSettings      `json:"advanced_settings"`
	EmbySettings          *EmbySettings          `json:"emby_settings"`
	DeveloperSettings     *DeveloperSettings     `json:"developer_settings"`
	TimelineFixerSettings *TimelineFixerSettings `json:"timeline_fixer_settings"`
	ExperimentalFunction  *ExperimentalFunction  `json:"experimental_function"`
}

// GetSettings 获取 Settings 的实例
func GetSettings(reloadSettings ...bool) *Settings {
	if _settings == nil {

		_settingsOnce.Do(func() {
			_settings = NewSettings()
			if isFile(_settings.configFPath) == false {
				// 配置文件不存在，新建一个空白的
				err := _settings.Save()
				if err != nil {
					panic("Can't Save Config File:" + configName + " Error: " + err.Error())
				}
			} else {
				// 读取存在的文件
				err := _settings.Read()
				if err != nil {
					panic("Can't Read Config File:" + configName + " Error: " + err.Error())
				}
			}
		})
		// 是否需要重新读取配置信息，这个可能在每次保存配置文件后需要操作
		if len(reloadSettings) >= 1 {
			if reloadSettings[0] == true {
				err := _settings.Read()
				if err != nil {
					panic("Can't Read Config File:" + configName + " Error: " + err.Error())
				}
			}
		}

	}
	return _settings
}

// SetFullNewSettings 从 Web 端传入新的 Settings 完整设置
func SetFullNewSettings(inSettings *Settings) error {

	nowConfigFPath := _settings.configFPath
	_settings = inSettings
	_settings.configFPath = nowConfigFPath

	return _settings.Save()
}

func NewSettings() *Settings {

	nowConfigFPath := filepath.Join(global_value.ConfigRootDirFPath(), configName)

	return &Settings{
		configFPath:           nowConfigFPath,
		UserInfo:              &UserInfo{},
		CommonSettings:        NewCommonSettings(),
		AdvancedSettings:      NewAdvancedSettings(),
		EmbySettings:          NewEmbySettings(),
		DeveloperSettings:     NewDeveloperSettings(),
		TimelineFixerSettings: NewTimelineFixerSettings(),
		ExperimentalFunction:  NewExperimentalFunction(),
	}
}

func (s *Settings) Read() error {

	// 需要检查 url 是否正确
	newEmbyAddressUrl := removeSuffixAddressSlash(s.EmbySettings.AddressUrl)
	_, err := url.Parse(newEmbyAddressUrl)
	if err != nil {
		return err
	}
	s.EmbySettings.AddressUrl = newEmbyAddressUrl

	return strcut_json.ToStruct(s.configFPath, s)
}

func (s *Settings) Save() error {

	// 需要检查 url 是否正确
	newEmbyAddressUrl := removeSuffixAddressSlash(s.EmbySettings.AddressUrl)
	_, err := url.Parse(newEmbyAddressUrl)
	if err != nil {
		return err
	}
	s.EmbySettings.AddressUrl = newEmbyAddressUrl

	return strcut_json.ToFile(s.configFPath, s)
}

func (s *Settings) GetNoPasswordSettings() *Settings {
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
		} else if s.CommonSettings.Threads >= 1 {
			s.CommonSettings.Threads = 1
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

// 将字符串后面最后一个字符，如果是 / 那么则替换掉，多个也会
func removeSuffixAddressSlash(orgAddressUrlString string) string {

	outString := orgAddressUrlString

	for {
		if strings.HasSuffix(outString, "/") == true {
			outString = outString[:len(outString)-1]
		} else {
			break
		}
	}
	return outString
}

var (
	_settings     *Settings
	_settingsOnce sync.Once
)

const (
	noPassword4Show = "******" // 填充使用
	configName      = "ChineseSubFinderSettings.json"
)
