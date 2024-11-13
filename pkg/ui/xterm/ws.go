// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package web 

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack/v5"
)

// "log"

// "github.com/gorilla/websocket"

// func (lspcon *lsp_ws_conn) SendWsData(t []byte, ws string) {
// 	// url := "ws://localhost:8080/ws"
// 	// url := "ws://" + ws + "/ws"
// 	// Skips certificate verification
// 	con := lspcon.con
// 	err := con.WriteMessage(websocket.TextMessage, t)
// 	if err != nil {
// 		log.Println("WebSocket发送消息失败:", err)
// 	} else {
// 		log.Println("send to ", ws, len(t))
// 	}
// }

func SendJsonMessage[T any](ws *websocket.Conn, data T) error {
	buf, err := json.Marshal(data)
	if err == nil {
		return ws.WriteMessage(websocket.TextMessage, buf)
	}
	return err
}

func SendMsgPackMessage[T any](ws *websocket.Conn, data T) error {
	buf, err := msgpack.Marshal(data)
	if err == nil {
		return ws.WriteMessage(websocket.TextMessage, buf)
	}
	return err
}
