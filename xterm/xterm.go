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
)

var sss = ptyout{&ptyout_impl{}}
var wg sync.WaitGroup
var need = true
var ptystdio *pty.Pty = nil

type wsize struct {
	Width  uint16 `json:"width"`
	Height uint16 `json:"height"`
}
type keycode struct {
	Key string `json:"key"`
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		// handle error
		return
	}
	defer conn.Close()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			// handle error
			break
		}
		log.Printf("Received: %s", message)
		err = conn.WriteMessage(messageType, message)
		if err != nil {
			break
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
	r.HandleFunc("/mouse", func(w http.ResponseWriter, r *http.Request) {
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
		for i, _ := range sss.imp.output {
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

func StartServer(root string, port int, cb func(port int)) {
	r := NewRouter(root)
	for i := port; i < 30000; i++ {
		if cb != nil {
			cb(i)
		}
		log.Printf("Server listening on http://localhost:%d\n", i)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", i), r); err != nil {
			log.Println(i, "Inused")
		}
	}
}

type ptyout_impl struct {
	output string
	pty    string
}
type ptyout struct {
	imp *ptyout_impl
}

// Write implements pty.ptyio.
func (p ptyout) Write(s []byte) (n int, err error) {
	p.imp.output += string(s)
	if need {
		wg.Done()
		need = false
	}
	return len(s), nil
}

// main
func main() {
	wg.Add(1)
	go func() {
		ptystdio = pty.Ptymain([]string{"/usr/bin/lspvi", "-tty"})
		io.Copy(sss, ptystdio.File)
		os.Exit(-1)
	}()
	StartServer(filepath.Dir(os.Args[0]), 13000, nil)
}
