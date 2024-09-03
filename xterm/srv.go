package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"zen108.com/lspvi/pkg/pty"
)

var sss = ptyout{&ptyout_impl{}}

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
	r.HandleFunc("/term", func(w http.ResponseWriter, r *http.Request) {
		if sss.imp == nil {
			w.Write([]byte("xx"))
		}
		w.Write([]byte(sss.imp.output))
	}).Methods("GET")
	// 处理根路径
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
}
type ptyout struct {
	imp *ptyout_impl
}

// Write implements pty.ptyio.
func (p ptyout) Write(s string) {
	p.imp.output += s
}

func main() {
	go func() {
		pty.Ptymain([]string{"/usr/bin/lspvi"}, sss)
	}()
	StartServer(filepath.Dir(os.Args[0]), 13000, nil)
}
