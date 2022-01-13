package config

import (
	"errors"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"github.com/spf13/viper"
	"os"
	"regexp"
	"strings"
)

// GetConfig 统一获取配置的接口
func GetConfig() *types.Config {
	return config
}

func LoadConfig(conf string) error {
	config = &types.Config{}
	v := viper.New()
	loadDefaults(v)

	v.SetConfigName(conf)       // 设置文件名称（无后缀）
	v.SetConfigType("yaml")     // 设置后缀名 {"1.6以后的版本可以不设置该后缀"}
	v.AddConfigPath("/config/") // 设置文件所在路径
	v.AddConfigPath(".")        // 设置文件所在路径

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			//fmt.Println("config file not found, use defaults")
		} else {
			return fmt.Errorf("error for reading config file: %w \n", err)
		}
	}

	v.AutomaticEnv()
	v.AllowEmptyEnv(true)

	if err := v.Unmarshal(config); err != nil {
		return fmt.Errorf("error for unmarshaling config file: %w \n", err)
	}

	config.SupportedVideoExts = map[string]int{"mp4": 1, "mkv": 1, "rmvb": 1, "iso": 1}
	// 读取用户自定义的视频后缀名列表
	for _, customExt := range strings.Split(config.CustomVideoExts, ",") {
		if customExt != "" {
			c := strings.ToLower(strings.TrimLeft(customExt, "."))
			config.SupportedVideoExts[c] = 1
		}
	}

	return nil
}

func CheckConfig() error {
	errMsg := ""
	f, err := os.Stat(config.MovieFolder)
	if err != nil {
		errMsg += fmt.Sprintf("MovieFolder %s not found\n", config.MovieFolder)
	} else if !f.IsDir() {
		errMsg += fmt.Sprintf("MovieFolder %s is not a directory\n", config.MovieFolder)
	}
	f, err = os.Stat(config.SeriesFolder)
	if err != nil {
		errMsg += fmt.Sprintf("SeriesFolder %s not found\n", config.SeriesFolder)
	} else if !f.IsDir() {
		errMsg += fmt.Sprintf("SeriesFolder %s is not a directory\n", config.SeriesFolder)
	}
	if config.HttpProxy != "" {
		re := regexp.MustCompile(`(http)://[\w\-_]+(\.[\w\-_]+)+([\w\-.,@?^=%&:/~+#]*[\w\-@?^=%&/~+#])?`)
		if result := re.FindAllStringSubmatch(config.HttpProxy, -1); result == nil {
			errMsg += fmt.Sprintf("proxy address %s is illegal, only support http://xx:xx", config.HttpProxy)
		}
	}
	// 这里进行 Default 值的判断
	config.SubTimelineFixerConfig.CheckDefault()
	if errMsg != "" {
		return errors.New(errMsg)
	}
	return nil
}

func loadDefaults(v *viper.Viper) {
	v.SetDefault("HttpProxy", "")
	v.SetDefault("EveryTime", "12h")
	v.SetDefault("Threads", 2)
	v.SetDefault("RunAtStartup", true)
	v.SetDefault("MovieFolder", "/media/电影")
	v.SetDefault("SeriesFolder", "/media/连续剧")
}

var (
	config = &types.Config{}
)
