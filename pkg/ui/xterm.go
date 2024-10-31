// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"crypto/tls"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	// "time"
	// "github.com/tinylib/msgp/msgp"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack/v5"
	"zen108.com/lspvi/pkg/debug"
	"zen108.com/lspvi/pkg/pty"
)

var use_https = false
var start_process func(int, string)
var wk *workdir
var httpport = 0
var sss = ptyout{&ptyout_impl{unsend: []byte{}}}
var wg sync.WaitGroup
var ptystdio *pty.Pty = nil

type init_call struct {
	Call    string `json:"call"`
	Cols    uint16 `json:"cols"`
	Rows    uint16 `json:"rows"`
	Host    string `json:"host"`
	Cmdline string `json:"cmdline"`
}
type keydata struct {
	Call string `json:"call"`
	Data string `json:"data"`
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}
type resize struct {
	Call string `json:"call"`
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var is_chan_start = false

type xterm_request struct {
	backend *lspvi_backend
}
type lspvi_backend struct {
	forward lspvi_command_forward
	ws      *websocket.Conn
}

func (term *lspvi_backend) process(method string, message []byte) bool {
	if term.forward.process(method, message) {
		return true
	}
	switch method {
	case forward_call_refresh:
		{
			ForwardFromXterm[xterm_forward_cmd_refresh](message, term)
		}
	default:
		return false
	}
	return true
}

func ForwardFromXterm[T any](message []byte, term *lspvi_backend) {
	var a T
	if err := json.Unmarshal(message, &a); err == nil {
		err = SendJsonMessage(term.ws, a)
		if err != nil {
			log.Println("error sending message to websocket", err)
		}
	}
}

type lspvi_command_forward struct {
}

func (term lspvi_command_forward) process(method string, message []byte) bool {
	switch method {
	case backend_on_copy:
		{
			ForwardFromLspvi[Ws_on_selection](sss.imp, message)
		}
	case backend_on_zoom:
		{
			ForwardFromLspvi[Ws_font_size](sss.imp, message)
		}
	case backend_on_openfile:
		{
			var file Ws_open_file
			err := json.Unmarshal(message, &file)
			if err == nil && wk != nil {
				name := filepath.Base(file.Filename)
				x := "__" + name
				tempfile := filepath.Join(wk.temp, x)
				err := os.WriteFile(tempfile, file.Buf, 0666)
				if err != nil {
					debug.DebugLog("xterm ", err)
				} else {
					buf, err := msgpack.Marshal(Ws_open_file{
						Filename: filepath.Join("/temp", x),
						Call:     "openfile",
					})
					if err == nil {
						sss.imp.write_ws(buf)
					}
				}
			}

		}
	default:
		return false
	}
	return true
}

var g_xterm *xterm_request

func serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// handle error
		return
	}
	defer conn.Close()
	var xterm *xterm_request
	var _lspvi_backend *lspvi_backend
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(messageType, message, err)
			continue
		}
		var w init_call
		err = json.Unmarshal(message, &w)
		if err == nil {
			method := w.Call
			log.Println("received call message", w.Call)
			switch w.Call {
			case call_key:
				{
					if xterm != nil {
						xterm.process(method, message)
					}
				}
			case lspvi_backend_start:
				{
					_lspvi_backend = &lspvi_backend{ws: conn}
					g_xterm.backend = _lspvi_backend
				}
			case call_xterm_init:
				{
					xterm = new_xterm_init(w, conn)
					g_xterm = xterm
				}
			default:
				if xterm != nil {
					xterm.process(method, message)
				} else if _lspvi_backend != nil {
					_lspvi_backend.process(method, message)
				}
			}
			continue
		}
	}
}

func (srv *xterm_request) process(method string, message []byte) {
	switch method {
	case call_redraw:
		{
			ForwardFromXterm[xterm_forward_cmd_redraw](message, srv.backend)
		}
	case call_key:
		{
			srv.handle_xterm_input(message)
		}
	case call_resize:
		{
			srv.handle_xterm_resize(message)
		}
	case call_paste_data:
		{
			ForwardFromXterm[xterm_forward_cmd_paste](message, srv.backend)
		}
	}
}

func ForwardFromLspvi[T any](imp *ptyout_impl, message []byte) error {
	var file T
	err := json.Unmarshal(message, &file)
	if err == nil {
		if buf, err := msgpack.Marshal(file); err == nil {
			return imp.write_ws(buf)
		}
	}
	return err
}

func handle_lspvi_on_copy(message []byte, w init_call) {
	var err error
	var file Ws_on_selection
	err = json.Unmarshal(message, &file)
	if err == nil {
		if buf, err := msgpack.Marshal(Ws_on_selection{
			SelectedString: file.SelectedString,
			Call:           w.Call,
		}); err == nil {
			sss.imp.write_ws(buf)
		}
	}
}

func (xterm_request) handle_xterm_input(message []byte) {
	if ptystdio == nil {
		return
	}
	var key keydata
	err := json.Unmarshal(message, &key)
	if err == nil {
		if key.Cols != 0 && key.Rows != 0 {
			ptystdio.UpdateSize(key.Rows, key.Cols)
		}
		ptystdio.File.Write([]byte(key.Data))
	}
}

func (xterm_request) handle_xterm_resize(message []byte) {
	if ptystdio == nil {
		return
	}
	var res resize
	err := json.Unmarshal(message, &res)

	if err == nil {
		if res.Rows != ptystdio.Rows || res.Cols != ptystdio.Cols {
			ptystdio.UpdateSize(res.Rows, res.Cols)
		}
	}
}

func new_xterm_init(w init_call, conn *websocket.Conn) *xterm_request {
	if start_process != nil {
		start_process(httpport, w.Host)
	}
	if start_process == nil {
		if ptystdio == nil {
			url := "ws://" + w.Host + "/ws"
			if use_https {
				url = "wss://" + w.Host + "/ws"
			}
			create_lspvi_backend(url, w.Cmdline)
			var i = 0
			for {
				if ptystdio == nil {
					time.Sleep(time.Millisecond * 10)
					i++
					if i > 100 {
						os.Exit(1)
					}
				} else {
					break
				}
			}
		} else {
			ptystdio.File.Write([]byte{27})
			ptystdio.File.Write([]byte{12})
		}
		if w.Cols != 0 && w.Rows != 0 {
			ptystdio.UpdateSize(w.Rows, w.Cols)
		}
	}
	sss.imp.ws = conn
	if len(sss.imp.unsend) > 0 {
		sss.imp.send_term_stdout(sss.imp.unsend)
		sss.imp.unsend = []byte{}
	}
	go func() {
		if is_chan_start {
			return
		}
		is_chan_start = true
		for {
			data := <-ws_buffer_chan
			sss.imp.ws.WriteMessage(websocket.BinaryMessage, data.Buf)
		}
	}()
	return &xterm_request{}
}
func NewRouter(root string) *mux.Router {
	r := mux.NewRouter()
	ss := new_workdir(root)
	wk = &ss
	// staticDir := "./node_modules"
	// fileServer := http.FileServer(http.Dir(staticDir))
	// r.Handle("/static/", http.StripPrefix("/static/", fileServer))
	r.HandleFunc("/node_modules/{path:.*}", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// println(path)
		buf, _ := os.ReadFile(filepath.Join(".", path))
		w.Write(buf)
	}).Methods("GET")
	r.HandleFunc("/temp/{path:.*}", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		buf, _ := os.ReadFile(filepath.Join(wk.root, path))
		w.Write(buf)
	}).Methods("GET")
	r.HandleFunc("/ws", serveWs)
	r.HandleFunc("/static/{path:.*}", func(w http.ResponseWriter, r *http.Request) {
		read_embbed(r, w)
	})
	r.HandleFunc("/{path:.*}", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			reset_lsp_backend()
			read_embbed(r, w)
		} else {
			var root = project_root
			if len(root) == 0 {
				root, _ = filepath.Abs(".")
			}
			buf, _ := os.ReadFile(filepath.Join(root, path))
			w.Write(buf)
		}
	}).Methods("GET")
	return r
}

func reset_lsp_backend() {
	sss.imp.ws = nil
	if start_process == nil {
		ptystdio = nil
	}
}

func read_embbed(r *http.Request, w http.ResponseWriter) {
	file := r.URL.Path
	if file == "/" {
		file = "index.html"
	}
	if s, ok := strings.CutPrefix(file, "/static/"); ok {
		file = s
	}
	if devroot, err := filepath.Abs("."); err == nil {
		var file_under_dev = filepath.Join(devroot, "pkg", "ui", "html", file)
		if _, err := os.Stat(file_under_dev); err == nil {
			buf, err := os.ReadFile(file_under_dev)
			if err == nil {
				w.Write(buf)
				return
			}
		}

	}
	p := filepath.Join("html", file)
	buf, err := uiFS.Open(p)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	io.Copy(w, buf)
}

var srv http.Server

func StartServer(root string, port int) {
	r := NewRouter(root)
	cert := NewCert()
	if cert != nil {
		if cert.Getcert() == nil {
			use_https = true
		}
	}
	if use_https {
		for i := port; i < 30000; i++ {

			tlsConfig := &tls.Config{}

			// 加载证书
			creds, err := tls.LoadX509KeyPair(cert.servercrt, cert.serverkey)
			if err != nil {
				log.Fatalf("Failed to load X509 key pair: %v", err)
				break
			}
			tlsConfig.Certificates = []tls.Certificate{creds}

			// 创建 HTTPS 服务器
			for i := port; i < 30000; i++ {
				x := fmt.Sprintf(":%d", i)
				httpport = port
				if start_process != nil {
					start_process(i, "")
				}
				// 启动 HTTPS 服务器
				log.Println("Starting HTTPS server on ", x)
				err = http.ListenAndServeTLS(x, cert.servercrt, cert.serverkey, r)
				if err != nil {
					log.Printf("Failed to start HTTPS server: %v", err)
					continue
				}
			}
		}
	}
	for i := port; i < 30000; i++ {
		// fmt.Printf("Server listening on http://localhost:%d\n", i)
		fmt.Println(i, "Check")
		x := fmt.Sprintf(":%d", i)
		srv = http.Server{Addr: x, Handler: r}
		httpport = port
		if start_process != nil {
			start_process(i, "")
		}
		if err := srv.ListenAndServe(); err != nil {
			continue
		}
	}
}

type open_file struct {
	Name string
	Path string
}
type open_files struct {
	Files []open_file
}
type ptyout_impl struct {
	prev   []byte
	ws     *websocket.Conn
	unsend []byte
	files  open_files
}

func (imp *ptyout_impl) send(s []byte) {
	if len(imp.prev) != len(s) {
		imp.send_term_stdout(s)
	}
}

func (imp *ptyout_impl) write_ws(s []byte) error {
	ws_buffer_chan <- buffer_to_send{s}
	return nil
}

type ptyout_data struct {
	Call   string
	Output []byte
}
type buffer_to_send struct {
	Buf []byte
}

var ws_buffer_chan = make(chan buffer_to_send)

func (imp *ptyout_impl) send_term_stdout(s []byte) bool {
	log.Println("_send", len(s))
	// log.Println("_send", len(s), string(s))
	// printf("\033[5;10HHello, World!\n"); // 将光标移动到第5行第10列，然后打印 "Hello, World!"
	// newFunction2(s)
	data := ptyout_data{
		Output: s,
		Call:   call_term_stdout,
	}
	buf, err := msgpack.Marshal(data)
	if imp.ws == nil {
		imp.unsend = append(imp.unsend, s...)
		return true
	}
	if err == nil {
		imp.write_ws(buf)
		imp.prev = append(imp.prev, s...)
	}
	return false
}

type ptyout struct {
	imp *ptyout_impl
}

//go:embed  html
var uiFS embed.FS

// Write implements pty.ptyio.
func (p ptyout) Write(s []byte) (n int, err error) {
	p.imp.send(s)
	return len(s), nil
}

var argnew []string
var project_root string

// main
func StartWebUI(arg Arguments, cb func(int, string)) {
	project_root = arg.Root
	start_process = cb
	argnew = []string{os.Args[0], "-tty"}

	args := []string{}
	if len(os.Args) > 2 {
		args = os.Args[2:]
	}
	argnew = append(argnew, args...)
	sss.imp.files.Files = []open_file{}
	wg.Add(1)
	StartServer(filepath.Dir(os.Args[0]), 13000)
}

func create_lspvi_backend(host string, cmdline string) {
	go func() {
		if len(cmdline) > 0 {
			argnew = []string{argnew[0]}
			args := strings.Split(cmdline, " ")
			if args[0] == "lspvi" {
				args = args[1:]
				argnew = append(argnew, "-ws", host)
				argnew = append(argnew, args...)
			} else {
				argnew = args
			}
		} else {
			argnew = append(argnew, "-ws", host)
		}
		ptystdio = pty.Ptymain(argnew)
		io.Copy(sss, ptystdio.File)
		// sss.imp.send_term_stdout([]byte("F5#"))
		ptystdio = nil
		impl := sss.imp
		Ws_term_command{Command: "quit", wsresp: wsresp{imp: impl}}.sendmsgpack()
	}()
}
