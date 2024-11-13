package web

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"zen108.com/lspvi/pkg/ui/common"
)

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
func NewRouter(root string) *mux.Router {
	r := mux.NewRouter()
	ss := common.NewWorkdir(root)
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
		buf, _ := os.ReadFile(filepath.Join(wk.Root, path))
		w.Write(buf)
	}).Methods("GET")
	r.HandleFunc("/ws", serveWs)
	r.HandleFunc("/md/{path:.*}", func(w http.ResponseWriter, r *http.Request) {
		read_mark(r, w)
	})
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
