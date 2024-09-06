package mainui

import (
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
	"zen108.com/lspvi/pkg/pty"
)

var start_process func(int, string)
var wk *workdir
var httpport = 0
var sss = ptyout{&ptyout_impl{unsend: []byte{}}}
var wg sync.WaitGroup
var ptystdio *pty.Pty = nil

type init_call struct {
	Call string `json:"call"`
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
	Host string `json:"host"`
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

func serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// handle error
		return
	}
	defer conn.Close()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(messageType, message, err)
			continue
		}
		var w init_call
		err = json.Unmarshal(message, &w)
		if err == nil {
			switch w.Call {
			case "init":
				{
					if start_process != nil {
						start_process(httpport, w.Host)
					}
					if start_process == nil {
						if ptystdio == nil {
							newFunction1(w.Host)
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
						sss.imp._send(sss.imp.unsend)
						sss.imp.unsend = []byte{}
					}
				}
			case "resize":
				{
					if ptystdio == nil {
						return
					}
					var res resize
					err = json.Unmarshal(message, &res)

					if err == nil {
						if res.Rows != ptystdio.Rows || res.Cols != ptystdio.Cols {
							ptystdio.UpdateSize(res.Rows, res.Cols)
						}
						// ptystdio.File.Write([]byte(key.Data))
						continue
					}
				}
			case "openfile":
				{
					var file Ws_open_file
					err = json.Unmarshal(message, &file)
					if err == nil && wk != nil {
						name := filepath.Base(file.Filename)
						x := "__" + name
						tempfile := filepath.Join(wk.temp, x)
						err := os.WriteFile(tempfile, file.Buf, 0666)
						if err != nil {
							fmt.Println(err)
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
			case "key":
				{
					if ptystdio == nil {
						return
					}
					var key keydata
					err = json.Unmarshal(message, &key)
					if err == nil {
						if key.Cols != 0 && key.Rows != 0 {
							ptystdio.UpdateSize(key.Rows, key.Cols)
						}
						ptystdio.File.Write([]byte(key.Data))
						continue
					}
				}
			default:
				fmt.Println("unknown call", w.Call)
			}
			continue
		}
	}
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
		println(path)
		buf, _ := os.ReadFile(filepath.Join(".", path))
		w.Write(buf)
	}).Methods("GET")
	r.HandleFunc("/temp/{path:.*}", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		buf, _ := os.ReadFile(filepath.Join(wk.root, path))
		w.Write(buf)
	}).Methods("GET")
	r.HandleFunc("/ws", serveWs)
	// r.HandleFunc("/term", func(w http.ResponseWriter, r *http.Request) {
	// 	conn, err := upgrader.Upgrade(w, r, nil)
	// 	if err != nil {
	// 		return
	// 	}
	// 	defer conn.Close()
	// })
	r.HandleFunc("/static/{path:.*}", func(w http.ResponseWriter, r *http.Request) {
		read_embbed(r, w)
	})
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		read_embbed(r, w)
	}).Methods("GET")
	return r
}

func read_embbed(r *http.Request, w http.ResponseWriter) {
	file := r.URL.Path
	if file == "/" {
		file = "index.html"
	}
	if s,ok:=strings.CutPrefix(file, "/static/");ok{
		file=s
	}
	sss.imp.ws = nil
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
		imp._send(s)
	}
}

func (imp *ptyout_impl) write_ws(s []byte) error {
	err := imp.ws.WriteMessage(websocket.BinaryMessage, s)
	return err
}

type ptyout_data struct {
	Call   string
	Output []byte
}

func (imp *ptyout_impl) _send(s []byte) bool {
	fmt.Println("_send", len(s))
	// printf("\033[5;10HHello, World!\n"); // 将光标移动到第5行第10列，然后打印 "Hello, World!"
	// newFunction2(s)
	data := ptyout_data{
		Output: s,
		Call:   "term",
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
	// fmt.Println("xxx",len(s),string(s))
	// p.imp.output += string(s)
	// if need {
	// 	wg.Done()
	// 	need = false
	// }
	return len(s), nil
}

var argnew []string

// main
func StartWebUI(cb func(int, string)) {
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

func newFunction1(host string) {
	go func() {
		argnew = append(argnew, "-ws", host)
		ptystdio = pty.Ptymain(argnew)
		io.Copy(sss, ptystdio.File)
		os.Exit(-1)
	}()
}