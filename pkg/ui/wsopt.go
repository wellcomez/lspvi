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

func (cmd Ws_term_command) sendmsgpack() error {
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
const xterm_request_forward_refresh = "xterm_request_forward_refresh"
const lspvi_backend_start = "xterm_lspvi_start"

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

type xterm_forward_cmd struct {
	Call string
}

func (proxy *ws_to_xterm_proxy) set_browser_selection(s string) {
	SendJsonMessage[Ws_on_selection](proxy.con,
		Ws_on_selection{SelectedString: s, Call: call_on_copy})
}
func (proxy *ws_to_xterm_proxy) set_browser_font(zoom bool) {
	SendJsonMessage[Ws_font_size](proxy.con, Ws_font_size{Zoom: zoom, Call: call_zoom})
}
func (proxy *ws_to_xterm_proxy) open_in_web(filename string) {
	buf, err := os.ReadFile(filename)
	if err == nil {
		SendJsonMessage[Ws_open_file](proxy.con, Ws_open_file{Filename: filename, Call: call_openfile, Buf: buf})
	} else {
		log.Println(call_openfile, filename, err)
	}
}

type ws_to_xterm_proxy struct {
	address string
	con     *websocket.Conn
}

var proxy *ws_to_xterm_proxy

func start_lspvi_proxy(arg *Arguments, listen bool) {
	proxy = &ws_to_xterm_proxy{address: arg.Ws}
	proxy.Open(listen)
}
func (p *ws_to_xterm_proxy) Request(t []byte) error {
	return p.con.WriteMessage(websocket.TextMessage, t)
}
func (proxy *ws_to_xterm_proxy) Open(listen bool) {
	ws := proxy.address
	// url := "ws://localhost:8080/ws"
	dial := websocket.DefaultDialer
	if strings.Index(ws, "wss://") == 0 {
		dial = &websocket.Dialer{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // Skips certificate verification
			},
		}
	}
	con, _, err := dial.Dial(ws, nil)
	if err != nil {
		log.Printf("WebSocket连接失败:%v", err)
		return
	}
	proxy.con = con
	if listen {

		if err := SendJsonMessage(con, xterm_forward_cmd{Call: lspvi_backend_start}); err == nil {

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
						{
							var w init_call
							err = json.Unmarshal(message, &w)
							if err == nil {
								switch w.Call {
								case xterm_request_forward_refresh:
									{

									}
								}
							} else {
								log.Println("recv", err, "msg len=", len(message))
							}

						}
					case websocket.BinaryMessage:
						log.Println("recv binary message", len(message))
					}
				}
			}()
		} else {
			log.Println("SendJsonMessage", err)
		}
	}
}
