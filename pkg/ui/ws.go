package mainui

import (
	"crypto/tls"
	"log"
	"strings"

	"github.com/gorilla/websocket"
)

func (lspcon *lsp_ws_conn) SendWsData(t []byte, ws string) {
	// url := "ws://localhost:8080/ws"
	// url := "ws://" + ws + "/ws"
	// Skips certificate verification
	con := lspcon.con
	err := con.WriteMessage(websocket.TextMessage, t)
	if err != nil {
		log.Println("WebSocket发送消息失败:", err)
	} else {
		log.Println("send to ", ws, len(t))
	}
}
