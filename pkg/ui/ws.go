package mainui

import (
	"log"

	"github.com/gorilla/websocket"
)

var con *websocket.Conn

func SendWsData(t []byte) {
	// url := "ws://localhost:8080/ws"
	if con == nil {
		url := "ws://localhost:13000/ws"
		c, _, err := websocket.DefaultDialer.Dial(url, nil)
		con = c
		if err != nil {
			log.Fatal("WebSocket连接失败:", err)
		}
	}
	// defer c.Close()
	err := con.WriteMessage(websocket.TextMessage, t)
	log.Println("ws errr", err)
}
