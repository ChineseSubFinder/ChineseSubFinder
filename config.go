package main

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/spf13/viper"
)
import "errors"

// InitConfigure 初始化配置文件实例
func InitConfigure() (*viper.Viper, error) {
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
// ReadConfig 读取配置文件
func ReadConfig(viper *viper.Viper) (*common.Config, error) {
	conf := &common.Config{}
	err := viper.Unmarshal(conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}