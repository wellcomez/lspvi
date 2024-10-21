// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
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
	cmd.Call = backend_on_command
	if buf, er := msgpack.Marshal(cmd); er == nil {
		return cmd.write(buf)
	} else {
		return er
	}

}

const call_key = "key"
const call_term_stdout = "term"
const call_xterm_init = "init"
const call_resize = "resize"
const call_redraw = "call_redraw"
const call_paste_data = "call_paste_data"

const forward_call_refresh = "forward_call_refresh"
const lspvi_backend_start = "xterm_lspvi_start"

const backend_on_command = "call_term_command"
const backend_on_zoom = "zoom"
const backend_on_copy = "onselected"
const backend_on_openfile = "openfile"

type xterm_forward_cmd_refresh struct {
	Call string `json:"call"`
}
type xterm_forward_cmd_redraw struct {
	Call string `json:"call"`
}
type xterm_forward_cmd_paste struct {
	Call string `json:"call"`
	Data string `json:"data"`
}

type xterm_forward_cmd struct {
	Call string `json:"call"`
}
