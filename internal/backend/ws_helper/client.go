package ws_helper

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend/ws"
	"github.com/allanpk716/ChineseSubFinder/internal/types/log_hub"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
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
	maxMessageSize = 5 * 1024

	// 发送 chan 的队列长度
	bufSize = 5 * 1024

	upGraderReadBufferSize = 5 * 1024

	upGraderWriteBufferSize = 5 * 1024
	// 字幕扫描任务执行状态
	subScanJobStatusInterval = 5 * time.Second
	// 字幕扫描运行中任务日志信息
	runningLogInterval = 5 * time.Second
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
	closeOnce        sync.Once
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
			log_helper.GetLogger().Warningln("readPump.recover", err)
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
		log_helper.GetLogger().Errorln("readPump.SetReadDeadline", err)
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
				log_helper.GetLogger().Errorln("readPump.IsUnexpectedCloseError", err)
			}
			return
		}

		revMessage := ws.BaseMessage{}
		err = json.Unmarshal(message, &revMessage)
		if err != nil {
			log_helper.GetLogger().Errorln("readPump.BaseMessage.parse", err)
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
				log_helper.GetLogger().Errorln("readPump.Login.parse", err)
				return
			}

			if login.Token != common.GetAccessToken() {
				// 登录 Token 不对
				// 发送 token 失败的消息
				outBytes, err := AuthReply(ws.AuthError)
				if err != nil {
					log_helper.GetLogger().Errorln("readPump.AuthReply", err)
					return
				}
				c.send <- outBytes
				// 直接退出可能会导致发送的队列没有清空，这里单独判断一条特殊的命令，收到 Write 线程就退出
				c.send <- ws.CloseThisConnect

			} else {
				// Token 通过
				outBytes, err := AuthReply(ws.AuthOk)
				if err != nil {
					log_helper.GetLogger().Errorln("readPump.AuthReply", err)
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
			log_helper.GetLogger().Warningln("writePump.recover", err)
		}
	}()

	// 心跳计时器
	pingTicker := time.NewTicker(pingPeriod)
	// 字幕扫描任务状态计时器
	subScanJobStatusTicker := time.NewTicker(subScanJobStatusInterval)
	// 正在运行扫描器的日志
	runningLogTicker := time.NewTicker(runningLogInterval)
	defer func() {
		pingTicker.Stop()
		subScanJobStatusTicker.Stop()
		runningLogTicker.Stop()
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
				log_helper.GetLogger().Errorln("writePump.SetWriteDeadline", err)
				return
			}
			if ok == false {
				// The hub closed the channel.
				err = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					log_helper.GetLogger().Warningln("writePump close hub WriteMessage", err)
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

			if err := w.Close(); err != nil {
				log_helper.GetLogger().Errorln("writePump.Close", err)
				return
			}
		case <-pingTicker.C:
			// 心跳相关，这里是定时器到了触发的间隔，设置发送下一条心跳的超时时间
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log_helper.GetLogger().Errorln("writePump.pingTicker.C.SetWriteDeadline", err)
				return
			}
			// 然后发送心跳
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log_helper.GetLogger().Errorln("writePump.pingTicker.C.WriteMessage", err)
				return
			}
		case <-subScanJobStatusTicker.C:
			// 字幕扫描任务状态
			if c.authed == false {
				// 没有认证通过，就无需处理次定时器时间
				continue
			}
			// 如果没有开启总任务，或者停止总任务了，那么这里获取到的应该是 nil，不应该继续往下
			info := common.GetSubScanJobStatus()
			if info == nil {
				continue
			}
			// 统一丢到 send 里面得了
			outLogsBytes, err := SubScanJobStatusReply(info)
			if err != nil {
				log_helper.GetLogger().Errorln("writePump.SubScanJobStatusReply", err)
				return
			}
			c.send <- outLogsBytes

		case <-runningLogTicker.C:
			// 正在运行扫描日志
			if c.authed == false {
				// 没有认证通过，就无需处理次定时器时间
				continue
			}
			nowRunningLog := log_helper.GetOnceLog4Running()
			if nowRunningLog == nil {
				continue
			}
			// 找到日志，把当前已有的日志发送出去，然后记录发送到哪里了
			// 这里需要考虑一次性的信息太多，超过发送的缓冲区，所以需要拆分发送
			outLogsBytes, err := RunningLogReply(nowRunningLog, c.sendLogLineIndex)
			if err != nil {
				log_helper.GetLogger().Errorln("writePump.RunningLogReply", err)
				return
			}
			// 拆分到一条日志来发送
			for _, logsByte := range outLogsBytes {
				c.send <- logsByte
				c.sendLogLineIndex += 1
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

// RunningLogReply 发送的 Running Log 数据，iPreSendLines 之前俺发送到第几条数据，则不发发送过的
func RunningLogReply(log *log_hub.OnceLog, iPreSendLines ...int) ([][]byte, error) {

	if log == nil {
		return nil, errors.New("RunningLogReply input log is nil")
	}

	var outLogBytes = make([][]byte, 0)
	var err error
	var preSendLines = 0
	if len(iPreSendLines) > 0 {
		preSendLines = iPreSendLines[0]
		if preSendLines < 0 {
			preSendLines = 0
		}
		log.LogLines = log.LogLines[preSendLines:]
	}

	logs := log_helper.GetSpiltOnceLog(log)
	for _, onceLog := range logs {
		var outData, outBytes []byte
		outData, err = json.Marshal(onceLog)
		if err != nil {
			return nil, err
		}

		outBytes, err = ws.NewBaseMessage(ws.RunningLog.String(), string(outData)).Bytes()
		if err != nil {
			return nil, err
		}

		outLogBytes = append(outLogBytes, outBytes)
	}

	return outLogBytes, nil
}

// SubScanJobStatusReply 当前字幕扫描的进度信息
func SubScanJobStatusReply(info *ws.SubDownloadJobInfo) ([]byte, error) {

	var err error
	var outData, outBytes []byte
	outData, err = json.Marshal(info)
	if err != nil {
		return nil, err
	}

	outBytes, err = ws.NewBaseMessage(ws.SubDownloadJobsStatus.String(), string(outData)).Bytes()
	if err != nil {
		return nil, err
	}

	return outBytes, nil
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
