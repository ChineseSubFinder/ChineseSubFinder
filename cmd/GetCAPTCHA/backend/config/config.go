package config

import (
	"errors"
	"github.com/spf13/viper"
)

type Config struct {
	EveryTime                     string
	SSHKeyFullPath                string
	SSHKeyPwd                     string
	CloneProjectDesSaveDir        string
	GitProjectUrl                 string
	DesURL                        string
	WhenSubSupplierInvalidWebHook string
	PostUrl                       string
	AuthToken                     string

	UseProxy                 bool   // 是否使用代理
	UseWhichProxyProtocol    string // 是使用 socks5 还是 http 代理
	LocalHttpProxyServerPort string // 本地代理服务器端口
	InputProxyAddress        string // 输入的代理地址
	InputProxyPort           string // 输入的代理端口
	NeedPWD                  bool   // 是否使用用户名密码
	InputProxyUsername       string // 输入的代理用户名
	InputProxyPassword       string // 输入的代理密码
}

// GetConfig 统一获取配置的接口
func GetConfig() *Config {

	configViper, err := initConfigure()
	if err != nil {
		panic("GetConfig - initConfigure something " + err.Error())
	}
	config, err = readConfig(configViper)
	if err != nil {
		panic("GetConfig - readConfig something " + err.Error())
	}

	return config
}

// initConfigure 初始化配置文件实例
func initConfigure() (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigName("something") // 设置文件名称（无后缀）
	v.SetConfigType("yaml")      // 设置后缀名 {"1.6以后的版本可以不设置该后缀"}
	v.AddConfigPath(".")         // 设置文件所在路径

	err := v.ReadInConfig()
	if err != nil {
		return nil, errors.New("error reading something config:" + err.Error())
	}

	return v, nil
}

// readConfig 读取配置文件
func readConfig(viper *viper.Viper) (*Config, error) {
	conf := &Config{}
	err := viper.Unmarshal(conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

var (
	config *Config
)
