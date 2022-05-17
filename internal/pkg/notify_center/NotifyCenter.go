package notify_center

import (
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"sync"
)

type NotifyCenter struct {
	log        *logrus.Logger
	webhookUrl string
	infos      map[string]string
	mu         sync.Mutex
}

func NewNotifyCenter(log *logrus.Logger, webhookUrl string) *NotifyCenter {
	n := NotifyCenter{log: log, webhookUrl: webhookUrl}
	n.infos = make(map[string]string)
	return &n
}

func (n *NotifyCenter) Add(groupName, infoContent string) {
	if n == nil {
		return
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	n.infos[groupName] = infoContent
}

func (n *NotifyCenter) Send() {
	if n == nil || n.webhookUrl == "" {
		return
	}
	client := resty.New().SetTransport(&http.Transport{
		DisableKeepAlives:   true,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
	})
	for s, s2 := range n.infos {
		_, err := client.R().Get(n.webhookUrl + s + "/" + url.QueryEscape(s2))
		if err != nil {
			n.log.Errorln("NewNotifyCenter.Send", err)
			return
		}
	}
}

func (n *NotifyCenter) Clear() {
	if n == nil {
		return
	}
	n.infos = make(map[string]string)
}

var Notify *NotifyCenter
