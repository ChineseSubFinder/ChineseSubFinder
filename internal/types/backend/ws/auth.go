package ws

type Login struct {
	Type  WSType `json:"type"`
	Token string `json:"token"`
}
