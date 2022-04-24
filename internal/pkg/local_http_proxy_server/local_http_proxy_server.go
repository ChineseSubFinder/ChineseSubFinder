package local_http_proxy_server

import (
	"encoding/base64"
	"fmt"
	"github.com/elazarl/goproxy"
	"golang.org/x/net/context"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"net/url"
)

// LocalHttpProxyServer see https://github.com/go-rod/rod/issues/305
type LocalHttpProxyServer struct {
	srv                      *http.Server
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

		go func() {
			if err := l.srv.ListenAndServe(); err != http.ErrServerClosed {
				panic(fmt.Sprintln("ListenAndServe() http proxy:", err))
			}
		}()

		l.isRunning = true
		l.LocalHttpProxyUrl = "http://127.0.0.1:" + l.LocalHttpProxyServerPort

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

		go func() {
			if err := l.srv.ListenAndServe(); err != http.ErrServerClosed {
				panic(fmt.Sprintln("ListenAndServe() socks5 proxy:", err))
			}
		}()

		l.isRunning = true
		l.LocalHttpProxyUrl = "http://127.0.0.1:" + l.LocalHttpProxyServerPort

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

	l.srv = nil
	l.isRunning = false
	l.LocalHttpProxyUrl = ""

	return nil
}

func (l *LocalHttpProxyServer) IsRunning() bool {
	return l.isRunning
}

const (
	LocalHttpProxyPort = "19036"
	ProxyAuthHeader    = "Proxy-Authorization"
)
