package config

import (
	"errors"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/global_value"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"github.com/spf13/viper"
	"os"
	"runtime"
	"strings"
	"sync"
)

// GetConfig 统一获取配置的接口
func GetConfig() *types.Config {
	configOnce.Do(func() {
		configViper, err := initConfigure()
		if err != nil {
			panic("GetConfig - initConfigure " + err.Error())
		}
		config, err = readConfig(configViper)
		if err != nil {
			panic("GetConfig - readConfig " + err.Error())
		}
		// 读取用户自定义的视频后缀名列表
		for _, customExt := range strings.Split(config.CustomVideoExts, ",") {
			global_value.CustomVideoExts = append(global_value.CustomVideoExts, "."+customExt)
		}

		// 这里进行 Default 值的判断
		config.SubTimelineFixerConfig.CheckDefault()
	})
	return config
}

// initConfigure 初始化配置文件实例
func initConfigure() (*viper.Viper, error) {
	nowConfigDir := getConfigDir()
	if nowConfigDir == "" {
		fmt.Sprintf("initConfigure().getConfigDir()")
	}

	v := viper.New()
	v.SetConfigName("config")     // 设置文件名称（无后缀）
	v.SetConfigType("yaml")       // 设置后缀名 {"1.6以后的版本可以不设置该后缀"}
	v.AddConfigPath(nowConfigDir) // 设置文件所在路径

	err := v.ReadInConfig()
	if err != nil {
		return nil, errors.New("error reading config:" + err.Error())
	}

	return v, nil
}

// readConfig 读取配置文件
func readConfig(viper *viper.Viper) (*types.Config, error) {
	conf := &types.Config{}
	err := viper.Unmarshal(conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func getConfigDir() string {
	nowConfigDir := ""
	sysType := runtime.GOOS
	if sysType == "linux" {
		nowConfigDir = configDirLinux
	}
	if sysType == "windows" {
		nowConfigDir = configDirWindows
	}
	if sysType == "darwin" {
		home, _ := os.UserHomeDir()
		nowConfigDir = home + "/.config/chinesesubfinder/" + configDirDarwin
	}
	return nowConfigDir
}

var (
	config     *types.Config
	configOnce sync.Once
)

const (
	configDirLinux   = "/config/"
	configDirWindows = "."
	configDirDarwin  = "."
)
