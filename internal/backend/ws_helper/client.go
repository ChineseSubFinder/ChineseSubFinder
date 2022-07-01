package ws_helper

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/allanpk716/ChineseSubFinder/internal/types/backend/ws"
	"github.com/allanpk716/ChineseSubFinder/pkg/common"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 5 * 1024

	// 发送 chan 的队列长度
	bufSize = 5 * 1024

	upGraderReadBufferSize = 5 * 1024

	upGraderWriteBufferSize = 5 * 1024
)

var upGrader = websocket.Upgrader{
	ReadBufferSize:  upGraderReadBufferSize,
	WriteBufferSize: upGraderWriteBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	log              *logrus.Logger
	hub              *Hub
	conn             *websocket.Conn // 与服务器连接实例
	sendLogLineIndex int             // 日志发送到那个位置了
	authed           bool            // 是否已经通过认证
	send             chan []byte     // 发送给 client 的内容 bytes
	closeOnce        sync.Once
}

func NewClient(log *logrus.Logger, hub *Hub, conn *websocket.Conn, sendLogLineIndex int, authed bool, send chan []byte) *Client {
	return &Client{log: log, hub: hub, conn: conn, sendLogLineIndex: sendLogLineIndex, authed: authed, send: send}
}

func (c *Client) close() {
	c.closeOnce.Do(func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	})
}

// 接收 Client 发送来的消息
func (c *Client) readPump() {

	defer func() {
		if err := recover(); err != nil {
			c.log.Debugln("readPump.recover", err)
		}
	}()

	defer func() {
		// 触发移除 client 的逻辑
		//c.hub.unregister <- c
		c.close()
	}()
	var err error
	var message []byte
	c.conn.SetReadLimit(maxMessageSize)
	err = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		c.log.Debugln("readPump.SetReadDeadline", err)
		return
	}
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	// 收取 client 发送过来的消息
	for {
		_, message, err = c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.log.Debugln("readPump.IsUnexpectedCloseError", err)
			}
			return
		}

		revMessage := ws.BaseMessage{}
		err = json.Unmarshal(message, &revMessage)
		if err != nil {
			c.log.Debugln("readPump.BaseMessage.parse", err)
			return
		}

		if c.authed == false {
			// 如果没有经过认证，那么第一条一定需要判断是认证的消息
			if revMessage.Type != ws.Auth.String() {
				// 提掉线
				return
			}
			// 认证
			login := ws.Login{}
			err = json.Unmarshal([]byte(revMessage.Data), &login)
			if err != nil {
				c.log.Debugln("readPump.Login.parse", err)
				return
			}

			if login.Token != common.GetAccessToken() {
				// 登录 Token 不对
				// 发送 token 失败的消息
				outBytes, err := AuthReply(ws.AuthError)
				if err != nil {
					c.log.Debugln("readPump.AuthReply", err)
					return
				}
				c.send <- outBytes
				// 直接退出可能会导致发送的队列没有清空，这里单独判断一条特殊的命令，收到 Write 线程就退出
				c.send <- ws.CloseThisConnect

			} else {
				// Token 通过
				outBytes, err := AuthReply(ws.AuthOk)
				if err != nil {
					c.log.Debugln("readPump.AuthReply", err)
					return
				}
				c.send <- outBytes
				c.authed = true
			}

		} else {
			// 进过认证后的消息，无需再次带有 token 信息
		}
	}
}

// 向 Client 发送消息的队列
func (c *Client) writePump() {

	defer func() {
		if err := recover(); err != nil {
			c.log.Debugln("writePump.recover", err)
		}
	}()

	// 心跳计时器
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		c.close()
	}()

	for {
		select {
		case message, ok := <-c.send:

			if bytes.Equal(message, ws.CloseThisConnect) == true {
				return
			}

			// 这里是需要发送给 client 的消息
			// 当然首先还是得先把当前消息的发送超时，给确定下来
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				c.log.Debugln("writePump.SetWriteDeadline", err)
				return
			}
			if ok == false {
				// The hub closed the channel.
				err = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					c.log.Debugln("writePump close hub WriteMessage", err)
				}
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				c.log.Debugln("writePump.NextWriter", err)
				return
			}
			_, err = w.Write(message)
			if err != nil {
				c.log.Debugln("writePump.Write", err)
				return
			}

			if err := w.Close(); err != nil {
				c.log.Debugln("writePump.Close", err)
				return
			}
		case <-pingTicker.C:
			// 心跳相关，这里是定时器到了触发的间隔，设置发送下一条心跳的超时时间
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				c.log.Debugln("writePump.pingTicker.C.SetWriteDeadline", err)
				return
			}
			// 然后发送心跳
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.log.Debugln("writePump.pingTicker.C.WriteMessage", err)
				return
			}
		}
	}
}

// AuthReply 生成认证通过的回复数据
func AuthReply(inType ws.AuthMessage) ([]byte, error) {

	var err error
	var outData, outBytes []byte
	outData, err = json.Marshal(&ws.Reply{
		Message: inType.String(),
	})
	if err != nil {
		return nil, err
	}

	outBytes, err = ws.NewBaseMessage(ws.CommonReply.String(), string(outData)).Bytes()
	if err != nil {
		return nil, err
	}

	return outBytes, nil
}

// ServeWs 每个 Client 连接 ws 上线时触发
func ServeWs(log *logrus.Logger, hub *Hub, w http.ResponseWriter, r *http.Request) {

	conn, err := upGrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorln("ServeWs.Upgrade", err)
		return
	}

	client := NewClient(
		log,
		hub,
		conn,
		0,
		false,
		make(chan []byte, bufSize),
	)
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
