package mainui

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack/v5"
)

type Ws_on_selection struct {
	Call           string
	SelectedString string
}
type Ws_font_size struct {
	Call string
	Zoom bool
}
type Ws_open_file struct {
	Call     string
	Filename string
	Buf      []byte
}
type wsresp struct {
	imp *ptyout_impl
}

func (resp wsresp) write(buf []byte) error {
	return resp.imp.write_ws(buf)
}

type Ws_term_command struct {
	wsresp
	Call    string
	Command string
}
type Ws_call_lspvi_start struct {
	ws      *websocket.Conn
	Call    string
	Command string
}

func (cmd Ws_call_lspvi_start) Request(proxy *ws_to_xterm_proxy) error {
	cmd.Call = call_lspvi_start
	return newFunction2[Ws_call_lspvi_start](proxy, cmd)
}

func newFunction2[T any](proxy *ws_to_xterm_proxy, data T) error {
	if buf, er := msgpack.Marshal(data); er == nil {
		return proxy.Request(buf)
	} else {
		return er
	}
}

func (cmd Ws_term_command) resp() error {
	cmd.Call = call_term_command
	if buf, er := msgpack.Marshal(cmd); er == nil {
		return cmd.write(buf)
	} else {
		return er
	}

}

const call_zoom = "zoom"
const call_term_command = "call_term_command"
const call_on_copy = "onselected"
const call_term_stdout = "term"
const call_openfile = "openfile"
const call_lspvi_start = "lspvi_start"

func set_browser_selection(s string, ws string) {
	if buf, err := json.Marshal(&Ws_on_selection{SelectedString: s, Call: call_on_copy}); err == nil {
		SendWsData(buf, ws)
	} else {
		log.Println("selected", len(s), err)
	}
}
func set_browser_font(zoom bool, ws string) {
	if buf, err := json.Marshal(&Ws_font_size{Zoom: zoom, Call: call_zoom}); err == nil {
		SendWsData(buf, ws)
	} else {
		log.Println("zoom", zoom, err)
	}
}
func open_in_web(filename, ws string) {
	buf, err := os.ReadFile(filename)
	if err == nil {
		buf, err = json.Marshal(&Ws_open_file{Filename: filename, Call: call_openfile, Buf: buf})
		if err == nil {
			SendWsData(buf, ws)
		}
	}
	if err != nil {
		log.Println(call_openfile, filename, err)
	}
}

type ws_to_xterm_proxy struct {
	address string
	con     *websocket.Conn
}

func (p *ws_to_xterm_proxy) Request(t []byte) error {
	return con.WriteMessage(websocket.TextMessage, t)
}
func (proxy *ws_to_xterm_proxy) Open() {
	ws := proxy.address
	// url := "ws://localhost:8080/ws"
	if con == nil {
		dial := websocket.DefaultDialer
		if strings.Index(ws, "wss://") == 0 {
			dial = &websocket.Dialer{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // Skips certificate verification
				},
			}
		}
		c, _, err := dial.Dial(ws, nil)
		con = c
		if err != nil {
			log.Printf("WebSocket连接失败:%v", err)
			return
		}
	}
	proxy.con = con
	go func() {
		for {
			msg_type, message, err := proxy.con.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					log.Println("WebSocket proxy connection close")
					return
				}
				log.Println("WebSocket proxy connection read e:", err)
				continue
			}
			switch msg_type {
			case websocket.TextMessage:
				log.Println("recv", len(message))
			case websocket.BinaryMessage:
				log.Println("recv", len(message))
			}
		}
	}()
	Ws_call_lspvi_start{}.Request(proxy)
}
