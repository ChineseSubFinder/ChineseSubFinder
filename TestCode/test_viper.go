package TestCode

type CommonSettings struct {
	UseHttpProxy     bool     `json:"use_http_proxy"`       // 是否使用 http 代理
	HttpProxyAddress string   `json:"http_proxy_address"`   // Http 代理地址，内网
	ScanInterval     string   `json:"scan_interval""`       // 一轮字幕扫描的间隔
	Threads          int      `json:"threads"`              // 同时扫描的并发数
	RunScanAtStartUp bool     `json:"run_scan_at_start_up"` // 完成引导设置后，下次运行程序就开始扫描
	MoviePaths       []string `json:"movie_paths"`          // 电影的目录
	SeriesPaths      []string `json:"series_paths"`         // 连续剧的目录
}

//// initConfigure 初始化配置文件实例
//func initConfigure() (*viper.Viper, error) {
//
//	v := viper.New()
//	v.SetConfigName("ChineseSubFinderConfig") // 设置文件名称（无后缀）
//	v.SetConfigType("yaml")                   // 设置后缀名 {"1.6以后的版本可以不设置该后缀"}
//	v.AddConfigPath(".")                      // 设置文件所在路径
//
//	err := v.ReadInConfig()
//	if err != nil {
//		return nil, errors.New("error reading config:" + err.Error())
//	}
//
//	return v, nil
//}
//
//func writeConfig(viper *viper.Viper, settings *CommonSettings) error {
//	viper.WriteConfig()
//}
//
//// readConfig 读取配置文件
//func readConfig(viper *viper.Viper) (*CommonSettings, error) {
//	conf := &CommonSettings{}
//	err := viper.Unmarshal(conf)
//	if err != nil {
//		return nil, err
//	}
//	return conf, nil
//}
