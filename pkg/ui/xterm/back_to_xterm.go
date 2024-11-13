// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package web

import (
	"crypto/tls"

	// "log"
	"os"

	"encoding/json"
	"strings"

	"github.com/gorilla/websocket"
	"zen108.com/lspvi/pkg/debug"
	"zen108.com/lspvi/pkg/ui/common"
)

const xtermtag = "xterm"

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
		debug.InfoLog(xtermtag, backend_on_openfile, filename)
	} else {
		debug.ErrorLog(xtermtag, backend_on_openfile, filename, err)
	}
}
func (proxy *backend_of_xterm) open_in_prj(global_prj_root string) {
	if err := SendJsonMessage[Ws_open_prj](proxy.con, Ws_open_prj{PrjRoot: global_prj_root, Call: backend_on_open_prj}); err != nil {
		debug.ErrorLog(xtermtag, backend_on_open_prj, err)
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
		debug.ErrorLog("xterm", "WebSocket error ", err)
		return err
	}
	proxy.con = con
	if listen {
		proxy.start_listen_xterm_comand(con)
	}
	return nil
}

func (proxy *backend_of_xterm) start_listen_xterm_comand(con *websocket.Conn) error {
	if err := SendJsonMessage(con, xterm_forward_cmd{Call: lspvi_backend_start}); err == nil {

		go func() bool {
			for {
				msg_type, message, err := proxy.con.ReadMessage()
				if err != nil {
					if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
						debug.DebugLog(xtermtag, "WebSocket proxy connection close", err)
						return true
					}
					debug.DebugLog(xtermtag, "WebSocket proxy connection read e:", err)
					continue
				}
				switch msg_type {
				case websocket.TextMessage:
					{
						var w init_call
						err = json.Unmarshal(message, &w)
						if err == nil {
							debug.InfoLog(xtermtag, "forward call ", w.Call)
							proxy.process_xterm_command(w, message)
						} else {
							debug.ErrorLog(xtermtag, "recv", err, "msg len=", len(message))
						}

					}
				case websocket.BinaryMessage:
					debug.InfoLog(xtermtag, "recv binary message", len(message))
				}
			}
		}()
	} else {
		debug.ErrorLog(xtermtag, "SendJsonMessage", err)
		return err
	}
	return nil
}

func (*backend_of_xterm) process_xterm_command(w init_call, message []byte) {
	switch w.Call {
	case call_redraw:
		{
			GlobalApp.QueueUpdateDraw(func() {

			})
		}
	case forward_call_refresh:
		{

		}
	case call_paste_data:
		{
			var data xterm_forward_cmd_paste
			if err := json.Unmarshal(message, &data); err == nil {
				debug.InfoLog(xtermtag, data.Call, data.Data)
				paste := GlobalApp.GetFocus().PasteHandler()
				paste(data.Data, nil)
			}
		}
	}
}

var proxy *backend_of_xterm

func Start_lspvi_proxy(arg *common.Arguments, listen bool) {
	var p = &backend_of_xterm{address: arg.Ws}
	if err := p.Open(listen); err == nil {
		proxy = p
	} else {
		debug.ErrorLog(xtermtag, "start lspvi proxy failed", err)
	}
}
func SetBrowserFont(zoom bool) {
	if proxy != nil {
		proxy.set_browser_font(zoom)
	}
}
func OpenInPrj(file string) (yes bool) {
	if yes = proxy != nil; yes {
		proxy.open_in_prj(file)
	} else {
		debug.DebugLog(xtermtag, "proxy is nil")
	}
	return
}
func OpenInWeb(file string) (yes bool) {
	yes = proxy != nil
	if !yes {
		return
	}
	if common.Is_open_as_md(file) || common.Is_image(file) {
		proxy.open_in_web(file)
	} else {
		yes = false
	}
	return
}
func Set_browser_selection(s string) {
	if proxy != nil {
		proxy.set_browser_selection(s)
	}
}
