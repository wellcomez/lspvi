package mainui

import (
	"crypto/tls"
	"log"
	"os"

	"encoding/json"
	"github.com/gorilla/websocket"
	"strings"
)

func (proxy *backend_of_xterm) set_browser_selection(s string) {
	SendJsonMessage[Ws_on_selection](proxy.con,
		Ws_on_selection{SelectedString: s, Call: backend_on_copy})
}
func (proxy *backend_of_xterm) set_browser_font(zoom bool) {
	SendJsonMessage[Ws_font_size](proxy.con, Ws_font_size{Zoom: zoom, Call: backend_on_zoom})
}
func (proxy *backend_of_xterm) open_in_web(filename string) {
	buf, err := os.ReadFile(filename)
	if err == nil {
		SendJsonMessage[Ws_open_file](proxy.con, Ws_open_file{Filename: filename, Call: backend_on_openfile, Buf: buf})
	} else {
		log.Println(backend_on_openfile, filename, err)
	}
}

type backend_of_xterm struct {
	address string
	con     *websocket.Conn
}

func (p *backend_of_xterm) Request(t []byte) error {
	return p.con.WriteMessage(websocket.TextMessage, t)
}
func (proxy *backend_of_xterm) Open(listen bool) error {
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
		return err
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
								log.Printf("forward call %s", w.Call)
								switch w.Call {
								case forward_call_refresh:
									{

									}
								case call_paste_data:
									{
										var data xterm_forward_cmd_paste
										if err := json.Unmarshal(message, &data); err == nil {
											log.Println(data.Call, data.Data)
											paste := GlobalApp.GetFocus().PasteHandler()
											paste(data.Data, nil)
										}
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
			return err
		}
	}
	return nil
}

var proxy *backend_of_xterm

func start_lspvi_proxy(arg *Arguments, listen bool) {
	var p = &backend_of_xterm{address: arg.Ws}
	if err := p.Open(listen); err == nil {
		proxy = p
	} else {
		log.Println("start lspvi proxy failed", err)
	}
}
