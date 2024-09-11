package lspcore

import (
	"context"
	"embed"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	// "strings"

	sitter "github.com/smacker/go-tree-sitter"

	ts_c "github.com/smacker/go-tree-sitter/c"
	ts_go "github.com/smacker/go-tree-sitter/golang"
	ts_js "github.com/smacker/go-tree-sitter/javascript"
	ts_py "github.com/smacker/go-tree-sitter/python"
	//tree_sitter_markdown "github.com/smacker/go-tree-sitter/markdown/tree-sitter-markdown"
	tree_sitter_markdown_inline "github.com/smacker/go-tree-sitter/markdown/tree-sitter-markdown-inline"
	//"github.com/smacker/go-tree-sitter/markdown"
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
	// Symbols     []TreeSitterSymbol
	SymbolsLine map[int][]TreeSitterSymbol
}

func TreesitterCheckIsSourceFile(filename string) bool {
	for _, v := range lsp_lang_map {
		if v.IsMe(filename) {
			return true
		}
	}
	return false
}




var lsp_lang_map = map[string]lsplang{
	"go": lsp_lang_go{},
	"c":  lsp_lang_cpp{},
	"py": lsp_lang_py{},
	"js": lsp_ts{},
	"md": lsp_md{},
}
var ts_lang_map = map[string]*sitter.Language{
	"go": ts_go.GetLanguage(),
	"c":  ts_c.GetLanguage(),
	"py": ts_py.GetLanguage(),
	"js": ts_js.GetLanguage(),
	"md": tree_sitter_markdown_inline.GetLanguage(),
}

func (t *TreeSitter) Init() error {
	for k, v := range lsp_lang_map {
		if v.IsMe(t.filename) {
			t.langname = k
			t.Loadfile(ts_lang_map[k])
			return nil
		}
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
		// m = qc.FilterPredicates(m, t.sourceCode)
		for i := range m.Captures {
			c := m.Captures[i]

			captureName := q.CaptureNameForId(c.Index)
			// symbol := c.Node.Symbol()
			start := c.Node.StartPoint()
			end := c.Node.EndPoint()
			// name := c.Node.Content(t.sourceCode)
			// symbolname := t.lang.SymbolName(symbol)
			// log.Println(strings.Join([]string{symbolname, captureName}, "."), symbolname, symbol, c.Node.Type(), start, end, name)
			// log.Println(captureName, symbolname, start, end, name)
			hlname := captureName
			s := TreeSitterSymbol{Point(start), Point(end), hlname}
			// t.Symbols = append(t.Symbols, s)
			row := int(s.Begin.Row)
			if _, ok := t.SymbolsLine[row]; !ok {
				t.SymbolsLine[row] = []TreeSitterSymbol{s}
			} else {
				t.SymbolsLine[row] =
					append(t.SymbolsLine[row], s)
			}
		}
		// match, a, b := qc.NextCapture()
		// println(match, a, b)
		// for _, v := range match.Captures {
		// println(v)
		// }

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
	ret.SymbolsLine = make(map[int][]TreeSitterSymbol)
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
