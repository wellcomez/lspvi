package lspcore

import (
	"context"
	"embed"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/tectiv3/go-lsp"
)

type TreeSitter struct {
	filename string
	kind     []lsp.SymbolKind
	parser   *sitter.Parser
	tree     *sitter.Tree
	langname string
	lang     *sitter.Language
	sourceCode []byte
}

func (t *TreeSitter) load_hightlight() ([]lsp.SymbolInformation, error) {
	// pkg/lsp/queries/ada/highlights.scm
	// /home/z/dev/lsp/goui/pkg/lsp/queries/go/highlights.scm
	path := filepath.Join("queries", t.langname, "highlights.scm")

	buf, err := t.read_embbed(path)
	if err != nil {
		return []lsp.SymbolInformation{}, err
	}
	q, err := sitter.NewQuery(buf, t.lang)
	if err != nil {
		return []lsp.SymbolInformation{}, err
	}
	qc := sitter.NewQueryCursor()
	qc.Exec(q, t.tree.RootNode())

		for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}
		// Apply predicates filtering
		m = qc.FilterPredicates(m, t.sourceCode)
		for _, c := range m.Captures {
			x := c.Node.Symbol()
			start:=c.Node.StartPoint()
			end:=c.Node.EndPoint()
			fmt.Println(t.lang.SymbolName(x),x ,start,end, c.Node.Content(t.sourceCode))
		}
	}
	return []lsp.SymbolInformation{}, nil
}

//go:embed  queries
var query_fs embed.FS

type imp_copydata struct {
	buf []byte
}
type copydata struct {
	impl *imp_copydata
}

// Write implements io.Writer.
func (c copydata) Write(p []byte) (n int, err error) {
	c.impl.buf = p
	return len(p), nil
}

func (ts TreeSitter) read_embbed(p string) ([]byte, error) {
	file, err := query_fs.Open(p)
	if err == nil {
		imp:=&imp_copydata{
		}
		var d =copydata{imp}
		n,err:=io.Copy(d, file)
		log.Println(n)
		if err != nil {
			return []byte{}, err
		}else{
			return imp.buf,nil
		}
	}
	return []byte{}, err
}
func NewTreeSitter(name string) *TreeSitter {
	ret := &TreeSitter{
		parser: sitter.NewParser(),
		filename:name,
	}
	return ret
}

func (ts *TreeSitter) Loadfile(lang *sitter.Language) ([]lsp.SymbolInformation, error) {
	if err := ts._load_file(lang); err != nil {
		return []lsp.SymbolInformation{}, err
	}
	return ts.load_hightlight()
}

func (ts *TreeSitter) _load_file(lang *sitter.Language) error {
	ts.lang = lang
	ts.parser.SetLanguage(lang)
	buf, err := os.ReadFile(ts.filename)
	if err != nil {
		return err
	}
	ts.sourceCode = buf
	tree, err := ts.parser.ParseCtx(context.Background(), nil, buf)
	ts.tree = tree
	return err
}
