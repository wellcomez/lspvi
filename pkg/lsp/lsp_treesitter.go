package lspcore

import (
	"context"
	"embed"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	// "strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/tectiv3/go-lsp"

	ts_c "github.com/smacker/go-tree-sitter/c"
	ts_cpp "github.com/smacker/go-tree-sitter/cpp"
	ts_go "github.com/smacker/go-tree-sitter/golang"
	ts_js "github.com/smacker/go-tree-sitter/javascript"
	tree_sitter_markdown "github.com/smacker/go-tree-sitter/markdown/tree-sitter-markdown"
	ts_py "github.com/smacker/go-tree-sitter/python"
	ts_tsx "github.com/smacker/go-tree-sitter/typescript/tsx"
	ts_ts "github.com/smacker/go-tree-sitter/typescript/typescript"
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
	tsname     string
	lang       *sitter.Language
	sourceCode []byte
	// Symbols     []TreeSitterSymbol
	HlLine  t_symbol_line
	Outline []*Symbol
	tsdef   ts_lang_def
}

func TreesitterCheckIsSourceFile(filename string) bool {
	for _, v := range tree_sitter_lang_map {
		if v.filedetect.IsMe(filename) {
			return true
		}
	}
	return false
}

var ts_name_markdown = "markdown"
var ts_name_typescript = "typescript"
var ts_name_javascript = "javascript"
var ts_name_tsx = "tsx"

type ts_lang_def struct {
	name       string
	filedetect lsplang
	tslang     *sitter.Language
	def_ext    []string
	cb         func(*TreeSitter)
}

func (ts ts_lang_def) new_treesitter(name string) *TreeSitter {
	ret := NewTreeSitter(name)
	ret.lang = ts.tslang
	ret.tsname = ts.name
	return ret
}
func TsPtn(
	name string,
	filedetect lsplang,
	tslang *sitter.Language,
) ts_lang_def {
	return ts_lang_def{
		name,
		filedetect,
		tslang,
		[]string{},
		nil,
	}
}
func (s ts_lang_def) setcb(cb func(*TreeSitter)) ts_lang_def {
	s.cb = cb
	return s
}
func (s ts_lang_def) set_ext(file []string) ts_lang_def {
	s.def_ext = append(s.def_ext, file...)
	return s
}
func (s *ts_lang_def) get_ts_name(file string) string {
	if s.is_me(file) {
		return s.name
	}
	return ""
}
func (s *ts_lang_def) is_me(file string) bool {
	if len(s.def_ext) > 0 {
		if IsMe(file, s.def_ext) {
			return true
		}
	}
	if s.filedetect.IsMe(file) {
		return true
	}
	return false
}

type lsp_dummy struct {
}

// InitializeLsp implements lsplang.
func (l lsp_dummy) InitializeLsp(core *lspcore, wk WorkSpace) error {
	panic("unimplemented")
}

// IsMe implements lsplang.
func (l lsp_dummy) IsMe(filename string) bool {
	return false
}

// IsSource implements lsplang.
func (l lsp_dummy) IsSource(filename string) bool {
	panic("unimplemented")
}

// Launch_Lsp_Server implements lsplang.
func (l lsp_dummy) Launch_Lsp_Server(core *lspcore, wk WorkSpace) error {
	panic("unimplemented")
}

// Resolve implements lsplang.
func (l lsp_dummy) Resolve(sym lsp.SymbolInformation, symfile *Symbol_file) bool {
	panic("unimplemented")
}
func markdown_parser(ts *TreeSitter) {
	const head = "markup.heading"
	for _, line := range ts.HlLine {
		for _, s := range line {
			if strings.Index(s.SymobName, head) == 0 {
				ss := lsp.SymbolInformation{
					Name: s.Code,
					Kind: lsp.SymbolKindEnumMember,
					Location: lsp.Location{
						URI: lsp.NewDocumentURI(ts.filename),
						Range: lsp.Range{
							Start: lsp.Position{Line: int(s.Begin.Row), Character: int(s.Begin.Column)},
							End:   lsp.Position{Line: int(s.End.Row), Character: int(s.End.Column)},
						},
					},
				}
				aa := Symbol{
					SymInfo:   ss,
					classname: s.Code,
				}
				ts.Outline = append(ts.Outline, &aa)
			}
		}
	}
	 sort.Slice(ts.Outline, func(i, j int) bool {
		return ts.Outline[i].SymInfo.Location.Range.Start.Line < ts.Outline[j].SymInfo.Location.Range.Start.Line
	})
}

var tree_sitter_lang_map = []ts_lang_def{
	TsPtn("go", lsp_lang_go{}, ts_go.GetLanguage()),
	TsPtn("cpp", lsp_lang_cpp{}, ts_cpp.GetLanguage()).set_ext([]string{"h", "hpp", "cc", "cpp"}),
	TsPtn("c", lsp_lang_cpp{}, ts_c.GetLanguage()),
	TsPtn("python", lsp_lang_py{}, ts_py.GetLanguage()),
	TsPtn(ts_name_tsx, lsp_dummy{}, ts_tsx.GetLanguage()).set_ext([]string{"tsx"}),
	TsPtn(ts_name_javascript, lsp_ts{LanguageID: string(JAVASCRIPT)}, ts_js.GetLanguage()).set_ext([]string{"js"}),
	TsPtn(ts_name_typescript, lsp_ts{LanguageID: string(TYPE_SCRIPT)}, ts_ts.GetLanguage()).set_ext([]string{"ts"}),
	TsPtn(ts_name_markdown, lsp_md{}, tree_sitter_markdown.GetLanguage()).setcb(markdown_parser),
}

func (t *TreeSitter) Init(cb func(*TreeSitter)) error {
	for _, v := range tree_sitter_lang_map {
		if ts_name := v.get_ts_name(t.filename); len(ts_name) > 0 {
			t.tsname = ts_name
			t.tsdef = v
			t.Loadfile(v.tslang, cb)
			return nil
		}
	}
	return fmt.Errorf("not implemented")
}

func (t TreeSitter) query(queryname string) (t_symbol_line, error) {
	var SymbolsLine = make(t_symbol_line)
	path := filepath.Join("queries", t.tsname, queryname+".scm")
	buf, err := t.read_embbed(path)
	if err != nil {
		return SymbolsLine, err
	}
	ss := string(buf)
	heris := get_inherits(ss)
	log.Println(t.tsname, "heri", queryname, heris)
	ret, err := t.__query(queryname)
	if err != nil {
		return SymbolsLine, err
	}
	if len(ret) == 0 {
		for _, v := range heris {
			var ptr *ts_lang_def
			for i := range tree_sitter_lang_map {
				vv := tree_sitter_lang_map[i]
				if vv.name == v {
					ptr = &vv
					break
				}
			}
			if ptr != nil {
				ts := ptr.new_treesitter(t.filename)
				if err := ts._load_file(ptr.tslang); err == nil {
					if aaa, err := ts.__query(queryname); err == nil {
						for k, v := range aaa {
							SymbolsLine[k] = append(SymbolsLine[k], v...)
						}
					}
				}
			}
		}
		return SymbolsLine, nil
	}
	return ret, nil
}
func (t TreeSitter) __query(queryname string) (t_symbol_line, error) {
	var SymbolsLine = make(t_symbol_line)
	// pkg/lsp/queries/ada/highlights.scm
	// /home/z/dev/lsp/goui/pkg/lsp/queries/go/highlights.scm
	path := filepath.Join("queries", t.tsname, queryname+".scm")

	buf, err := t.read_embbed(path)
	if err != nil {
		return SymbolsLine, err
	}
	q, err := sitter.NewQuery(buf, t.lang)
	if err != nil {
		return SymbolsLine, err
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
	return SymbolsLine, nil
}

func get_inherits(ss string) []string {
	ind := strings.Index(ss, "\n")
	if ind > 0 {
		ss = ss[:ind]
		ss = strings.TrimPrefix(ss, "; inherits:")
		args := strings.Split(ss, ",")
		inerits := []string{}
		for _, v := range args {
			inerits = append(inerits, strings.TrimSpace(v))
		}
		return inerits
	}
	return []string{}
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

const query_highlights = "highlights"

func (ts *TreeSitter) Loadfile(lang *sitter.Language, cb func(*TreeSitter)) error {
	if err := ts._load_file(lang); err != nil {
		log.Println("fail to load treesitter", err)
		return err
	}
	go func() {
		ret, hlerr := ts.query(query_highlights)
		ts.HlLine = ret
		if hlerr != nil {
			log.Println("fail to load highlights", hlerr)
		}
		ret, local_err := ts.query("locals")
		if local_err != nil {
			log.Println("fail to load locals", local_err)
		} else {
			// symbols := get_ts_symbol(ret, ts)
			// ts.Outline = symbols
		}
		if ts.tsdef.cb != nil {
			ts.tsdef.cb(ts)
		}
		if cb != nil {
			cb(ts)
		}
	}()
	return nil
}

func get_ts_symbol(ret t_symbol_line, ts *TreeSitter) []lsp.SymbolInformation {
	prefix := "local.definition."
	symbols := []lsp.SymbolInformation{}
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
						kind := map[string]lsp.SymbolKind{
							"method": lsp.SymbolKindMethod,
						}
						log.Println("outline", s.Code, symboltype)
						s := lsp.SymbolInformation{
							Name: s.Code,
							Kind: kind[symboltype],
							Location: lsp.Location{
								URI: lsp.NewDocumentURI(ts.filename),
								Range: lsp.Range{
									Start: lsp.Position{Line: int(s.Begin.Row), Character: int(s.Begin.Column)},
									End:   lsp.Position{Line: int(s.End.Row), Character: int(s.End.Column)},
								},
							},
						}
						symbols = append(symbols, s)
					}
				case "var":
					continue
				default:
					log.Println("unhandled symbol type:", symboltype)
				}
			}
		}
	}
	return symbols
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
