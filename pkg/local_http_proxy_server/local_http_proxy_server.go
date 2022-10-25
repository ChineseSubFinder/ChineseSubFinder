package local_http_proxy_server

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/elazarl/goproxy"
	"golang.org/x/net/context"
	"golang.org/x/net/proxy"
)

// SetProxyInfo 设置代理信息，设置之前会停止现在运行的代理
func SetProxyInfo(UseProxy bool, iInfos []string, iLocalPort string) error {

	locker.Lock()
	defer locker.Unlock()

	useProxy = UseProxy

	if localHttpProxyServer != nil && localHttpProxyServer.IsRunning() == true {
		// 需要关闭代理
		err := localHttpProxyServer.Stop()
		if err != nil {
			return err
		}
	}

	proxyInfos = iInfos
	localHttpProxyServerPort = iLocalPort

	return nil
}

// GetProxyUrl 获取代理地址，同时启动实例
func GetProxyUrl() string {

	locker.Lock()
	defer locker.Unlock()

	if useProxy == false {

		if localHttpProxyServer != nil && localHttpProxyServer.IsRunning() == true {
			// 需要关闭代理
			err := localHttpProxyServer.Stop()
			if err != nil {
				println("localHttpProxyServer.Stop() Error:", err.Error())
			}
		}
		return ""
	}

	if localHttpProxyServer == nil {
		localHttpProxyServer = NewLocalHttpProxyServer()
	}
	if localHttpProxyServer.IsRunning() == true {
		return localHttpProxyServer.LocalHttpProxyUrl
	}

	localHttpProxyUrl, err := localHttpProxyServer.Start(proxyInfos, localHttpProxyServerPort)
	if err != nil {
		panic(fmt.Sprintln("start local http proxy server error:", err))
		return ""
	}

	return localHttpProxyUrl
}

// LocalHttpProxyServer see https://github.com/go-rod/rod/issues/305
type LocalHttpProxyServer struct {
	srv                      *http.Server
	locker                   sync.Mutex
	isRunning                bool
	LocalHttpProxyServerPort string // 本地开启的 Http 代理服务器端口
	LocalHttpProxyUrl        string // 本地开启的 Http 代理服务器地址包含端口
}

func NewLocalHttpProxyServer() *LocalHttpProxyServer {
	return &LocalHttpProxyServer{}
}

func setBasicAuth(username, password string, req *http.Request) {
	req.Header.Set(ProxyAuthHeader, fmt.Sprintf("Basic %s", basicAuth(username, password)))
}

func basicAuth(username, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}

// Start 传入参数为 [0] protocol [1] Address [2] Port [3] Username [4] Password，如果没得账号密码，则后面两个可以不传入
func (l *LocalHttpProxyServer) Start(settings []string, localHttpProxyServerPort string) (string, error) {

	if len(settings) < 3 {
		return "", nil
	}

	l.LocalHttpProxyServerPort = localHttpProxyServerPort

	protocol := settings[0]
	InputProxyAddress := settings[1]
	InputProxyPort := settings[2]

	InputProxyUsername := ""
	InputProxyPassword := ""
	if len(settings) >= 5 {
		InputProxyUsername = settings[3]
		InputProxyPassword = settings[4]
	}

	proxyAddress := InputProxyAddress + ":" + InputProxyPort

	switch protocol {
	case "http":
		middleProxy := goproxy.NewProxyHttpServer()
		middleProxy.Tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + proxyAddress)
		}
		//middleProxy.Logger = httpLogger()
		connectReqHandler := func(req *http.Request) {
			setBasicAuth(InputProxyUsername, InputProxyPassword, req)
		}
		middleProxy.ConnectDial = middleProxy.NewConnectDialToProxyWithHandler("http://"+proxyAddress, connectReqHandler)

		middleProxy.OnRequest().Do(goproxy.FuncReqHandler(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			setBasicAuth(InputProxyUsername, InputProxyPassword, req)
			return req, nil
		}))

		l.srv = &http.Server{Addr: ":" + l.LocalHttpProxyServerPort, Handler: middleProxy}
		l.locker.Lock()
		l.isRunning = true
		l.locker.Unlock()
		go func() {

			println("Try Start Local Proxy Server at :", l.LocalHttpProxyServerPort)

			if err := l.srv.ListenAndServe(); err != http.ErrServerClosed {
				println(fmt.Sprintln("ListenAndServe() http proxy:", err))
			}
			l.locker.Lock()
			l.srv = nil
			l.isRunning = false
			l.LocalHttpProxyUrl = ""
			l.locker.Unlock()

			println("http proxy closed")
		}()

		time.Sleep(3 * time.Second)

		l.locker.Lock()
		l.LocalHttpProxyUrl = "http://127.0.0.1:" + l.LocalHttpProxyServerPort
		l.locker.Unlock()

		return l.LocalHttpProxyUrl, nil

	case "socks5":

		var dialer proxy.Dialer
		var err error
		if len(settings) >= 5 {
			auth := proxy.Auth{
				User:     InputProxyUsername,
				Password: InputProxyPassword,
			}
			dialer, err = proxy.SOCKS5("tcp", proxyAddress, &auth, proxy.Direct)
			if err != nil {
				return "", err
			}
		} else {
			dialer, err = proxy.SOCKS5("tcp", proxyAddress, nil, proxy.Direct)
			if err != nil {
				return "", err
			}
		}
		dialContext := func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialer.Dial(network, address)
		}
		dialContextOld := func(network, address string) (net.Conn, error) {
			return dialer.Dial(network, address)
		}
		transport := &http.Transport{DialContext: dialContext, DisableKeepAlives: true}

		middleProxy := goproxy.NewProxyHttpServer()
		middleProxy.Tr = transport
		//middleProxy.Logger = httpLogger()
		middleProxy.ConnectDial = dialContextOld
		middleProxy.OnRequest().Do(goproxy.FuncReqHandler(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			setBasicAuth(InputProxyUsername, InputProxyPassword, req)
			return req, nil
		}))

		l.srv = &http.Server{Addr: ":" + l.LocalHttpProxyServerPort, Handler: middleProxy}
		l.locker.Lock()
		l.isRunning = true
		l.locker.Unlock()
		go func() {

			println("Try Start Local Proxy Server at :", l.LocalHttpProxyServerPort)
			if err := l.srv.ListenAndServe(); err != http.ErrServerClosed {
				println(fmt.Sprintln("ListenAndServe() socks5 proxy:", err))
			}

			l.locker.Lock()
			l.srv = nil
			l.isRunning = false
			l.LocalHttpProxyUrl = ""
			l.locker.Lock()

			println("socks5 proxy closed")
		}()

		time.Sleep(3 * time.Second)

		l.locker.Lock()
		l.LocalHttpProxyUrl = "http://127.0.0.1:" + l.LocalHttpProxyServerPort
		l.locker.Unlock()

		return l.LocalHttpProxyUrl, nil
	}
	return "", fmt.Errorf("proxy type invalid, not http or socks5")
}

func (l *LocalHttpProxyServer) Stop() error {
	if l.srv != nil {
		err := l.srv.Close()
		if err != nil {
			return err
		}
	}

	//
	//l.isRunning = false
	l.locker.Lock()
	l.LocalHttpProxyUrl = ""
	l.locker.Unlock()
	return nil
}

func (l *LocalHttpProxyServer) IsRunning() bool {

	l.locker.Lock()
	defer l.locker.Unlock()
	return l.isRunning
}

const (
	LocalHttpProxyPort = "19036"
	ProxyAuthHeader    = "Proxy-Authorization"
)

var (
	locker                   sync.Mutex
	localHttpProxyServer     *LocalHttpProxyServer
	useProxy                 bool
	proxyInfos               = make([]string, 0)
	localHttpProxyServerPort string
)
