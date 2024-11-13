package web

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/toc"
)

func MarkdownFileToHTMLString(md string) (ret []byte, err error) {
	if buf, e := os.ReadFile(md); e == nil {
		return MarkdownToHTMLStyle(buf)
	} else {
		err = e
	}
	return

}
func MarkdownToHTMLStyle(source []byte) (ret []byte, err error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
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
			),
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
func read_mark(r *http.Request, w http.ResponseWriter) {
	file := r.URL.Path
	file = strings.TrimPrefix(file, "/md/")
	if buf, err := MarkdownFileToHTMLString(filepath.Join(prj_root, file)); err == nil {
		w.Write(buf)
	} else {
		w.Write([]byte(err.Error()))
	}
}
