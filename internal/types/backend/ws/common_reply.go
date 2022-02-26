package ws

type Reply struct {
	Type    WSType `json:"type"`
	Message string `json:"message"`
}
