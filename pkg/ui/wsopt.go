package mainui

import (
	"encoding/json"
	"log"
	"os"

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
