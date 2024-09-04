package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"zen108.com/lspvi/pkg/pty"
	mainui "zen108.com/lspvi/pkg/ui"
)

var sss = ptyout{&ptyout_impl{}}
var wg sync.WaitGroup
var need = true
var ptystdio *pty.Pty = nil

type wsize struct {
	Width  uint16 `json:"width"`
	Height uint16 `json:"height"`
}
type base_type interface{}
type keycode struct {
	Key string `json:"key"`
}
type wsbase struct {
	Call string `json:"call"`
}
type init_call struct {
	Call string `json:"call"`
}
type keydata struct {
	Call string `json:"call"`
	Data string `json:"data"`
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
					var file mainui.Ws_open_file
					err = json.Unmarshal(message, &file)
					if err == nil {
						name := filepath.Base(file.Filename)
						dir := filepath.Dir(os.Args[0])
						dir = filepath.Join(dir, "temp")
						os.MkdirAll(dir, 0755)
						x := "__" + name
						tempfile := filepath.Join(dir, x)
						err := os.WriteFile(tempfile, file.Buf, 0666)
						if err != nil {
							fmt.Println(err)
						} else {
							file.Filename = filepath.Join("/temp", x)
							buf, err := json.Marshal(file)
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
		println(path)
		buf, _ := os.ReadFile(filepath.Join(".", path))
		w.Write(buf)
	}).Methods("GET")
	r.HandleFunc("/ws", serveWs)
	r.HandleFunc("/key", func(w http.ResponseWriter, r *http.Request) {
		var k keycode
		buf, err := io.ReadAll(r.Body)
		if err == nil {
			err = json.Unmarshal(buf, &k)
			if err == nil {
				if ptystdio != nil {
					ptystdio.File.Write([]byte(k.Key))
				}
			}
		}
	})
	r.HandleFunc("/term", func(w http.ResponseWriter, r *http.Request) {
		var k wsize
		buf, err := io.ReadAll(r.Body)
		if err == nil {
			if json.Unmarshal(buf, &k) == nil {
				if k.Height != ptystdio.Rows || k.Width != ptystdio.Cols {
					ptystdio.UpdateSize(k.Height, k.Width)
				}
			}
		}
		if len(sss.imp.output) == 0 {
			wg.Wait()
		}

		if sss.imp == nil {
			w.Write([]byte("xx"))
		}
		oldlen := len(sss.imp.pty)
		different := false
		for i := range sss.imp.output {
			if i < oldlen {
				if sss.imp.output[i] != sss.imp.pty[i] {
					fmt.Println(time.Now(), "output changed", i, sss.imp.output[i], sss.imp.pty[i])
					different = true
				}
			} else {
				fmt.Println(time.Now(), "output longger", len(sss.imp.output), len(sss.imp.pty))
				different = true
			}
		}
		if !different {
			w.Write([]byte("0000"))
			return
		}
		sss.imp.pty = sss.imp.output
		w.Write([]byte(sss.imp.pty))
		wg.Add(1)
		sss.imp.output = ""
		need = true
	})
	// 处理根路径
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// if !need {
		// 	wg.Add(1)
		// 	need = true
		// }
		buf, _ := os.ReadFile("index.html")
		w.Write(buf)
	}).Methods("GET")
	return r
}

func StartServer(root string, port int) {
	r := NewRouter(root)
	for i := port; i < 30000; i++ {
		// fmt.Printf("Server listening on http://localhost:%d\n", i)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", i), r); err != nil {
			fmt.Println(i, "Inused")
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
	output string
	prev   string
	pty    string
	ws     *websocket.Conn
	unsend string
	files  open_files
}

func (imp *ptyout_impl) send(s string) {
	if imp.prev != s {
		imp._send(s)
	}
}

func (imp *ptyout_impl) write_ws(s []byte) error {
	err := imp.ws.WriteMessage(websocket.TextMessage, s)
	return err
}
func (imp *ptyout_impl) _send(s string) bool {
	fmt.Println("_send", len(s))
	data := map[string]string{
		"output": s,
		"call":   "term",
	}
	buf, err := json.Marshal(data)
	if imp.ws == nil {
		imp.unsend += s
		return true
	}
	if err == nil {
		imp.write_ws(buf)
		imp.prev = s
	}
	return false
}

type ptyout struct {
	imp *ptyout_impl
}

// Write implements pty.ptyio.
func (p ptyout) Write(s []byte) (n int, err error) {
	p.imp.send(string(s))
	// fmt.Println("xxx",len(s),string(s))
	// p.imp.output += string(s)
	// if need {
	// 	wg.Done()
	// 	need = false
	// }
	return len(s), nil
}

// main
func main() {
	sss.imp.files.Files = []open_file{}
	wg.Add(1)
	go func() {
		ptystdio = pty.Ptymain([]string{"/usr/bin/lspvi", "-tty"})
		io.Copy(sss, ptystdio.File)
		os.Exit(-1)
	}()
	StartServer(filepath.Dir(os.Args[0]), 13000)
}
