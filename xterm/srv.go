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

	"github.com/gorilla/mux"
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

		sss.imp.pty = sss.imp.output
		if sss.imp == nil {
			w.Write([]byte("xx"))
		}
		w.Write([]byte(sss.imp.pty))
		need = true
		wg.Add(1)
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

func main() {
	wg.Add(1)
	go func() {
		ptystdio = pty.Ptymain([]string{"/usr/bin/lspvi"})
		io.Copy(sss, ptystdio.File)
	}()
	StartServer(filepath.Dir(os.Args[0]), 13000, nil)
}
