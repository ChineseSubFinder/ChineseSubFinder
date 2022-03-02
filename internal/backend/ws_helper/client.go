package ws_helper

import (
	"bytes"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512

	// send buffer size
	bufSize = 256

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
	hub              *Hub
	conn             *websocket.Conn // 与服务器连接实例
	sendLogLineIndex int             // 日志发送到那个位置了
	authed           bool            // 是否已经通过认证
	send             chan []byte     // 发送给 client 的内容 bytes
}

// 接收 Client 发送来的消息
func (c *Client) readPump() {

	defer func() {
		// 触发移除 client 的逻辑
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		log_helper.GetLogger().Errorln("readPump.SetReadDeadline", err)
		return
	}
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	// 收取 client 发送过来的消息
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log_helper.GetLogger().Errorln("readPump.IsUnexpectedCloseError", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, []byte{}, []byte{}, -1))
		c.hub.broadcast <- message
	}
}

// 向 Client 发送消息的队列
func (c *Client) writePump() {

	// 心跳计时器
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			// 这里是需要发送给 client 的消息
			// 当然首先还是得先把当前消息的发送超时，给确定下来
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				log_helper.GetLogger().Errorln("writePump.SetWriteDeadline", err)
				return
			}
			if ok == false {
				// The hub closed the channel.
				err = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					log_helper.GetLogger().Errorln("writePump close hub WriteMessage", err)
				}
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log_helper.GetLogger().Errorln("writePump.NextWriter", err)
				return
			}
			_, err = w.Write(message)
			if err != nil {
				log_helper.GetLogger().Errorln("writePump.Write", err)
				return
			}
			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {

				_, err = w.Write(<-c.send)
				if err != nil {
					log_helper.GetLogger().Errorln("writePump.Write", err)
					return
				}
			}

			if err := w.Close(); err != nil {
				log_helper.GetLogger().Errorln("writePump.Close", err)
				return
			}
		case <-ticker.C:
			// 心跳相关，这里是定时器到了触发的间隔，设置发送下一条心跳的超时时间
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log_helper.GetLogger().Errorln("writePump.ticker.C.SetWriteDeadline", err)
				return
			}
			// 然后发送心跳
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log_helper.GetLogger().Errorln("writePump.ticker.C.WriteMessage", err)
				return
			}
		}
	}
}

// ServeWs 每个 Client 连接 ws 上线时触发
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {

	conn, err := upGrader.Upgrade(w, r, nil)
	if err != nil {
		log_helper.GetLogger().Errorln("ServeWs.Upgrade", err)
		return
	}

	client := &Client{
		hub:              hub,
		conn:             conn,
		sendLogLineIndex: 0,
		authed:           false,
		send:             make(chan []byte, bufSize),
	}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
