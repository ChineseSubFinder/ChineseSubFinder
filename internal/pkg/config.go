package pkg

import (
	"errors"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"github.com/spf13/viper"
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
	})
	return config
}

// initConfigure 初始化配置文件实例
func initConfigure() (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigName("config") // 设置文件名称（无后缀）
	v.SetConfigType("yaml")   // 设置后缀名 {"1.6以后的版本可以不设置该后缀"}
	v.AddConfigPath(".")      // 设置文件所在路径

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

var (
	config     *types.Config
	configOnce sync.Once
)
