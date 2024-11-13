package web

import (
	"bytes"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/toc"
	"zen108.com/lspvi/pkg/ui/common"
)

func MarkdownFileToHTMLString(md string, roots string) (ret []byte, err error) {
	var buf []byte
	if buf, err = os.ReadFile(md); err == nil {
		// if buf, err = ChangeLink(buf, "/md/"); err == nil {
		return MarkdownToHTMLStyle(buf, roots)
		// }
	}
	return

}
func MarkdownToHTMLStyle(source []byte, root string) (ret []byte, err error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			create_link_parser(root),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
		goldmark.WithExtensions(
			&toc.Extender{},
		),
		goldmark.WithExtensions(
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"),
				highlighting.WithFormatOptions(
					chromahtml.WithLineNumbers(true),
				),
			),
			css_extentsion(),
		),
	)
	var buf bytes.Buffer
	if e := md.Convert(source, &buf); e != nil {
		err = e
	} else {
		ret = buf.Bytes()
	}
	return
}
func MarkdownToHTML(md []byte) (ret []byte, err error) {
	var buf bytes.Buffer
	err = goldmark.Convert(md, &buf)
	if err == nil {
		ret = buf.Bytes()
	}
	return
}
func read_mark_index(r *http.Request, w http.ResponseWriter) {
	file := r.URL.Path
	file = strings.TrimPrefix(file, "/md/")
	p := "md.html"
	tmpl := template.Must(template.ParseFS(uiFS, "html/"+p))
	err := tmpl.Execute(w, map[string]interface{}{
		"MDFILE": file,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func read_mark(r *http.Request, w http.ResponseWriter) {
	file := r.URL.Path
	if common.Is_open_as_md(file) {
		if buf, err := MarkdownFileToHTMLString(filepath.Join(project_root, file), "/md"); err == nil {
			w.Write(buf)
		} else {
			w.Write([]byte(err.Error()))
		}
	} else {
		filename := filepath.Join(project_root, file)
		if buf, err := os.ReadFile(filename); err == nil {
			w.Write(buf)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
