package web

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"zen108.com/lspvi/pkg/ui/common"
)

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
