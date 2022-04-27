package settings

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/local_http_proxy_server"
	"sync"
)

type ProxySettings struct {
	UseProxy                 bool   `json:"use_proxy"`                                    // 是否使用代理
	UseWhichProxyProtocol    string `json:"use_which_proxy_protocol"`                     // 是使用 socks5 还是 http 代理
	LocalHttpProxyServerPort string `json:"local_http_proxy_server_port" default:"19036"` // 本地代理服务器端口
	InputProxyAddress        string `json:"input_proxy_address"`                          // 输入的代理地址
	InputProxyPort           string `json:"input_proxy_port"`                             // 输入的代理端口
	NeedPWD                  bool   `json:"need_pwd"`                                     // 是否使用用户名密码
	InputProxyUsername       string `json:"input_proxy_username"`                         // 输入的代理用户名
	InputProxyPassword       string `json:"input_proxy_password"`                         // 输入的代理密码
	Referer                  string `json:"-"`                                            // 可能下载文件的时候需要设置

	localHttpProxyServer *local_http_proxy_server.LocalHttpProxyServer // 本地代理服务器
	locker               sync.Mutex
}

func NewProxySettings(useProxy bool, useWhichProxyProtocol string,
	localHttpProxyServerPort string,
	inputProxyAddress string, inputProxyPort string,
	inputProxyUsername string, inputProxyPassword string) *ProxySettings {

	set := ProxySettings{UseProxy: useProxy, UseWhichProxyProtocol: useWhichProxyProtocol,
		LocalHttpProxyServerPort: localHttpProxyServerPort,
		InputProxyAddress:        inputProxyAddress, InputProxyPort: inputProxyPort,
		InputProxyUsername: inputProxyUsername, InputProxyPassword: inputProxyPassword}

	if inputProxyUsername != "" && inputProxyPassword != "" {
		set.NeedPWD = true
	}

	return &set
}

func (p *ProxySettings) GetLocalHttpProxyUrl() string {
	defer p.locker.Unlock()
	p.locker.Lock()

	if p.UseProxy == false {
		return ""
	}

	if p.localHttpProxyServer == nil {
		p.localHttpProxyServer = local_http_proxy_server.NewLocalHttpProxyServer()
	}

	if p.localHttpProxyServer.IsRunning() == true {
		return p.localHttpProxyServer.LocalHttpProxyUrl
	}

	inputInfo := []string{
		p.UseWhichProxyProtocol,
		p.InputProxyAddress,
		p.InputProxyPort,
	}
	if p.InputProxyUsername != "" && p.InputProxyPassword != "" {
		inputInfo = append(inputInfo, p.InputProxyUsername, p.InputProxyPassword)
	}

	localHttpProxyUrl, err := p.localHttpProxyServer.Start(inputInfo, p.LocalHttpProxyServerPort)
	if err != nil {
		panic(fmt.Sprintln("start local http proxy server error:", err))
		return ""
	}

	return localHttpProxyUrl
}

func (p *ProxySettings) CloseLocalHttpProxyServer() error {
	defer p.locker.Unlock()
	p.locker.Lock()

	if p.localHttpProxyServer == nil {
		return nil
	}

	if p.localHttpProxyServer.IsRunning() == false {
		return nil
	}

	return p.localHttpProxyServer.Stop()
}
