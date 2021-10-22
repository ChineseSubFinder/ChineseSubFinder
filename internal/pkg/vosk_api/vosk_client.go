package vosk_api

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"io"
	"net/url"
	"os"
)

const Host = "192.168.50.135"
const Port = "2700"
const buffsize = 8000

type Message struct {
	Result []struct {
		Conf  float64
		End   float64
		Start float64
		Word  string
	}
	Text string
}

var m Message

func GetResult(audioFileFullPath string) error {
	u := url.URL{Scheme: "ws", Host: Host + ":" + Port, Path: ""}
	println("connecting to ", u.String())

	// Opening websocket connection
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	defer c.Close()

	f, err := os.Open(audioFileFullPath)
	if err != nil {
		return err
	}

	for {
		buf := make([]byte, buffsize)
		dat, err := f.Read(buf)

		if dat == 0 && err == io.EOF {
			err = c.WriteMessage(websocket.TextMessage, []byte("{\"eof\" : 1}"))
			if err != nil {
				return err
			}
			break
		}
		if err != nil {
			return err
		}

		err = c.WriteMessage(websocket.BinaryMessage, buf)
		if err != nil {
			return err
		}

		// Read message from server
		_, _, err = c.ReadMessage()
		if err != nil {
			return err
		}
	}

	// Read final message from server
	_, msg, err := c.ReadMessage()
	if err != nil {
		return err
	}

	// Closing websocket connection
	err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return err
	}

	// Unmarshalling received message
	err = json.Unmarshal(msg, &m)
	if err != nil {
		return err
	}
	println(m.Text)

	return nil
}
