package mainui

import (
	"crypto/tls"
	"log"
	"strings"

	"github.com/gorilla/websocket"
)

var con *websocket.Conn

func SendWsData(t []byte, ws string) {
	// url := "ws://localhost:8080/ws"
	if con == nil {
		// url := "ws://" + ws + "/ws"
		url := ws
		dial := websocket.DefaultDialer
		if strings.Index(ws, "wss://") == 0 {
			dial = &websocket.Dialer{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // Skips certificate verification
				},
			}
		}
		c, _, err := dial.Dial(url, nil)
		con = c
		if err != nil {
			log.Printf("WebSocket连接失败:", err)
			return
		}
	}
	err := con.WriteMessage(websocket.TextMessage, t)
	if err != nil {
		log.Println("WebSocket发送消息失败:", err)
	} else {
		log.Println("send to ", ws, len(t))
	}
}
