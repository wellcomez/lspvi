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
	"sync"
	"time"

	// "time"
	// "github.com/tinylib/msgp/msgp"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack/v5"
	"zen108.com/lspvi/pkg/pty"
)

var wk *workdir
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
						ptystdio.File.Write([]byte("j"))
					}
					if w.Cols != 0 && w.Rows != 0 {
						ptystdio.UpdateSize(w.Rows, w.Cols)
					}
					sss.imp.ws = conn
					sss.imp._send(sss.imp.unsend)
				}
			case "resize":
				{
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
		// println(path)
		buf, _ := os.ReadFile(filepath.Join(wk.root, path))
		w.Write(buf)
	}).Methods("GET")
	r.HandleFunc("/ws", serveWs)
	r.HandleFunc("/term", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
	})
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// if !need {
		// 	wg.Add(1)
		// 	need = true
		// }
		p := filepath.Join("html", "index.html")
		buf, err := uiFS.Open(p)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		io.Copy(w, buf)
	}).Methods("GET")
	return r
}

var srv http.Server

func StartServer(root string, port int) {
	r := NewRouter(root)
	for i := port; i < 30000; i++ {
		// fmt.Printf("Server listening on http://localhost:%d\n", i)
		fmt.Println(i, "Check")
		x := fmt.Sprintf(":%d", i)
		srv = http.Server{Addr: x, Handler: r}
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
	// output string
	prev []byte
	// pty    string
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

var count = 0

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

func newFunction2(s []byte) {
	for i, v := range s {
		count++
		name := ""
		switch v {
		case 0xC:
			name = "newpage"
		case 0xD:
			name = "Enter"
		case 0xA:
			{
				name = "LF"
			}

		case 033:
			{
				var line, col string
				var err string
				var afterline = false
				for i, v := range s[i+1:] {
					err = string(v)
					if i == 0 {
						if v != '[' {
							break
						}
						continue
					}
					if !afterline {
						if v >= '0' && v <= '0' {
							line += string(v)
							continue
						}
						if v == ';' {
							afterline = true
						} else {
							break
						}
					} else {
						if v >= '0' && v <= '0' {
							col += string(v)
						} else {
							break
						}
					}
				}
				println("tttt", err, line, col)

			}
		case 0x1A:
			{
				name = "LF"
			}
		default:
			continue
			if v > 0x80 {
				name = fmt.Sprintf("0x%0x", v)
			} else {
				continue
			}
		}
		println(i, count, name, len(s))
	}
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
func StartWebUI() {
	argnew = []string{os.Args[0], "-tty"}
	args := os.Args[2:]
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
