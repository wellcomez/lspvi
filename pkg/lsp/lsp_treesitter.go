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

	treec "github.com/smacker/go-tree-sitter/c"
	treego "github.com/smacker/go-tree-sitter/golang"
	ts "github.com/smacker/go-tree-sitter/javascript"
	treepyt "github.com/smacker/go-tree-sitter/python"
)

type Point struct {
	Row    uint32
	Column uint32
}
type TreeSitterSymbol struct {
	Begin, End Point
	SymobName  string
}
type TreeSitter struct {
	filename   string
	parser     *sitter.Parser
	tree       *sitter.Tree
	langname   string
	lang       *sitter.Language
	sourceCode []byte
	Symbols    []TreeSitterSymbol
}

func (t *TreeSitter) Init() error {
	langgo := lsp_lang_go{}
	if langgo.IsMe(t.filename) {
		t.langname = "go"
		t.Loadfile(treego.GetLanguage())
		return nil
	}
	c := lsp_lang_cpp{}
	if c.IsMe(t.filename) {
		t.langname = "c"
		t.Loadfile(treec.GetLanguage())
		return nil
	}
	py := lsp_lang_py{}
	if py.IsMe(t.filename) {
		t.langname = "py"
		t.Loadfile(treepyt.GetLanguage())
		return nil
	}
	_ts := lsp_ts{}
	if _ts.IsMe(t.filename) {
		t.langname = "js"
		t.Loadfile(ts.GetLanguage())
	}
	return fmt.Errorf("not implemented")
}
func (t *TreeSitter) load_hightlight() error {
	// pkg/lsp/queries/ada/highlights.scm
	// /home/z/dev/lsp/goui/pkg/lsp/queries/go/highlights.scm
	path := filepath.Join("queries", t.langname, "highlights.scm")

	buf, err := t.read_embbed(path)
	if err != nil {
		return err
	}
	q, err := sitter.NewQuery(buf, t.lang)
	if err != nil {
		return err
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
			start := c.Node.StartPoint()
			end := c.Node.EndPoint()
			log.Println("symbolname", t.lang.SymbolName(x), "symbotype:", t.lang.SymbolType(x), start, end, c.Node.Content(t.sourceCode))
			s := TreeSitterSymbol{Point(start), Point(end), t.lang.SymbolName(x)}
			t.Symbols = append(t.Symbols, s)
		}
	}
	return nil
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
		imp := &imp_copydata{}
		var d = copydata{imp}
		n, err := io.Copy(d, file)
		log.Println(n)
		if err != nil {
			return []byte{}, err
		} else {
			return imp.buf, nil
		}
	}
	return []byte{}, err
}
func NewTreeSitter(name string) *TreeSitter {
	ret := &TreeSitter{
		parser:   sitter.NewParser(),
		filename: name,
	}
	return ret
}

func (ts *TreeSitter) Loadfile(lang *sitter.Language) error {
	if err := ts._load_file(lang); err != nil {
		return err
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
