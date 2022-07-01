package ws

import (
	"encoding/json"
)

// BaseMessage 基础的消息结构，附带的信息会在其中的 data 字段需要二次解析
type BaseMessage struct {
	Type string `json:"type"`
	Data string `json:"data"` // 收到具体的消息需要从这里二次解析，判断类型由 Type 决定
}

func NewBaseMessage(typeInfo, dataInfo string) *BaseMessage {

	return &BaseMessage{
		typeInfo, dataInfo,
	}
}

func (b *BaseMessage) Bytes() ([]byte, error) {
	return json.Marshal(b)
}
