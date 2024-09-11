package lspcore

import (
	"context"
	"embed"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	// "strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/tectiv3/go-lsp"

	ts_c "github.com/smacker/go-tree-sitter/c"
	ts_go "github.com/smacker/go-tree-sitter/golang"
	ts_js "github.com/smacker/go-tree-sitter/javascript"
	tree_sitter_markdown "github.com/smacker/go-tree-sitter/markdown/tree-sitter-markdown"
	ts_py "github.com/smacker/go-tree-sitter/python"
	//"github.com/smacker/go-tree-sitter/markdown"
)

type t_symbol_line map[int][]TreeSitterSymbol
type Point struct {
	Row    uint32
	Column uint32
}
type TreeSitterSymbol struct {
	Begin, End Point
	SymobName  string
	Code       string
}
type TreeSitter struct {
	filename   string
	parser     *sitter.Parser
	tree       *sitter.Tree
	langname   string
	lang       *sitter.Language
	sourceCode []byte
	// Symbols     []TreeSitterSymbol
	HlLine  t_symbol_line
	Outline t_symbol_line
}

func TreesitterCheckIsSourceFile(filename string) bool {
	for _, v := range lsp_lang_map {
		if v.IsMe(filename) {
			return true
		}
	}
	return false
}

var ts_name_markdown = "markdown"
var lsp_lang_map = map[string]lsplang{
	"go":             lsp_lang_go{},
	"c":              lsp_lang_cpp{},
	"py":             lsp_lang_py{},
	"js":             lsp_ts{},
	ts_name_markdown: lsp_md{},
}
var ts_lang_map = map[string]*sitter.Language{
	"go":             ts_go.GetLanguage(),
	"c":              ts_c.GetLanguage(),
	"py":             ts_py.GetLanguage(),
	"js":             ts_js.GetLanguage(),
	ts_name_markdown: tree_sitter_markdown.GetLanguage(),
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

func (t TreeSitter) query(queryname string) (error, t_symbol_line) {
	var SymbolsLine = make(t_symbol_line)
	// pkg/lsp/queries/ada/highlights.scm
	// /home/z/dev/lsp/goui/pkg/lsp/queries/go/highlights.scm
	path := filepath.Join("queries", t.langname, queryname+".scm")

	buf, err := t.read_embbed(path)
	if err != nil {
		return err, SymbolsLine
	}
	q, err := sitter.NewQuery(buf, t.lang)
	if err != nil {
		return err, SymbolsLine
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
			name := c.Node.Content(t.sourceCode)
			// symbolname := t.lang.SymbolName(symbol)
			// log.Println(strings.Join([]string{symbolname, captureName}, "."), symbolname, symbol, c.Node.Type(), start, end, name)
			// log.Println(captureName, symbolname, start, end, name)
			hlname := captureName
			s := TreeSitterSymbol{Point(start), Point(end), hlname, name}
			// t.Symbols = append(t.Symbols, s)
			row := int(s.Begin.Row)
			if _, ok := SymbolsLine[row]; !ok {
				SymbolsLine[row] = []TreeSitterSymbol{s}
			} else {
				SymbolsLine[row] =
					append(SymbolsLine[row], s)
			}
		}
	}
	return nil, SymbolsLine
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
	ret.HlLine = make(map[int][]TreeSitterSymbol)
	return ret
}

func (ts *TreeSitter) Loadfile(lang *sitter.Language) error {
	if err := ts._load_file(lang); err != nil {
		log.Println("fail to load treesitter", err)
		return err
	}
	hlerr, ret := ts.query("highlights")
	ts.HlLine = ret
	if hlerr != nil {
		log.Println("fail to load highlights", hlerr)
	}
	local_err, ret := ts.query("locals")
	if local_err != nil {
		log.Println("fail to load locals", local_err)
	} else {
		lsp.NewDocumentURI(ts.filename)
		prefix := "local.definition."
		for i := range ret {
			v := ret[i]
			for i := 0; i < len(v); i++ {
				s := v[i]
				if strings.Index(s.SymobName, prefix) == 0 {
					symboltype := strings.Replace(s.SymobName, prefix, "", 1)
					switch symboltype {
					case "type":
						{
							log.Println("outline", s.Code, symboltype)
						}
					case "method", "function", "namespace":
						{
							log.Println("outline", s.Code, symboltype)
						}
					case "var":
						continue
					default:
						log.Println("unhandled symbol type:", symboltype)
					}
				}
			}
		}
		ts.Outline = ret
	}
	return hlerr
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
