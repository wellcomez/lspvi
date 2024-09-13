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
	sourceCode []byte
	HlLine     t_symbol_line
	Outline    []*Symbol
	tsdef      *ts_lang_def
	inited     bool
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
	hl         *sitter.Query
}

const query_highlights = "highlights"

func new_tsdef(
	name string,
	filedetect lsplang,
	tslang *sitter.Language,
) *ts_lang_def {
	ret := &ts_lang_def{
		name,
		filedetect,
		tslang,
		[]string{},
		nil,
		nil,
	}
	go func() {
		if h, er := ret.query(query_highlights); er == nil {
			ret.hl = h
		}
	}()
	return ret
}
func (t *ts_lang_def) create_query_buffer(lang string, queryname string) ([]byte, error) {
	path := filepath.Join("queries", lang, queryname+".scm")
	buf, err := t.read_embbed(path)
	if err != nil {
		return nil, err
	}
	ss := string(buf)
	heris := get_inherits(ss)
	log.Println(t.name, "heri", queryname, heris)
	var merge_buf = []byte{}
	if len(heris) > 0 {
		for _, v := range heris {
			if b, err := t.create_query_buffer(v, queryname); err == nil {
				merge_buf = append(merge_buf, b...)
			}
		}
		merge_buf = append(merge_buf, buf...)
	} else {
		merge_buf = append(merge_buf, buf...)
	}
	return merge_buf, nil
}
func (t *ts_lang_def) query(queryname string) (*sitter.Query, error) {
	if buf, err := t.create_query_buffer(t.name, queryname); err == nil {
		return t.create_query(buf)
	} else {
		return nil, err
	}
}
func (s *ts_lang_def) setcb(cb func(*TreeSitter)) *ts_lang_def {
	s.cb = cb
	return s
}
func (s *ts_lang_def) set_ext(file []string) *ts_lang_def {
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
	if ts.inited {
		return
	}
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

var tree_sitter_lang_map = []*ts_lang_def{
	new_tsdef("go", lsp_lang_go{}, ts_go.GetLanguage()),
	new_tsdef("cpp", lsp_lang_cpp{}, ts_cpp.GetLanguage()).set_ext([]string{"h", "hpp", "cc", "cpp"}),
	new_tsdef("c", lsp_lang_cpp{}, ts_c.GetLanguage()),
	new_tsdef("python", lsp_lang_py{}, ts_py.GetLanguage()),
	new_tsdef(ts_name_tsx, lsp_dummy{}, ts_tsx.GetLanguage()).set_ext([]string{"tsx"}),
	new_tsdef(ts_name_javascript, lsp_ts{LanguageID: string(JAVASCRIPT)}, ts_js.GetLanguage()).set_ext([]string{"js"}),
	new_tsdef(ts_name_typescript, lsp_ts{LanguageID: string(TYPE_SCRIPT)}, ts_ts.GetLanguage()).set_ext([]string{"ts"}),
	new_tsdef(ts_name_markdown, lsp_md{}, tree_sitter_markdown.GetLanguage()).setcb(markdown_parser),
}

func (t *TreeSitter) Init(cb func(*TreeSitter)) error {
	if t.tsdef != nil {
		go func() {
			if t.HlLine != nil {
				t.callback_to_ui(cb)
			}
		}()
		return nil
	}
	for _, v := range tree_sitter_lang_map {
		if ts_name := v.get_ts_name(t.filename); len(ts_name) > 0 && v.hl != nil {
			t.tsdef = v
			t.Loadfile(v.tslang, cb)
			return nil
		}
	}
	return fmt.Errorf("not implemented")
}

func (t TreeSitter) query(queryname string) (t_symbol_line, error) {
	if queryname == query_highlights {
		return t.query_buf(t.tsdef.hl)
	}
	return make(t_symbol_line), nil
}

func (t *TreeSitter) query_buf(q *sitter.Query) (t_symbol_line, error) {
	var SymbolsLine t_symbol_line = make(t_symbol_line)
	if q == nil {
		return SymbolsLine, fmt.Errorf("query not found")
	}
	qc := sitter.NewQueryCursor()
	qc.Exec(q, t.tree.RootNode())
	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}

		for i := range m.Captures {
			c := m.Captures[i]

			captureName := q.CaptureNameForId(c.Index)

			start := c.Node.StartPoint()
			end := c.Node.EndPoint()
			name := c.Node.Content(t.sourceCode)

			hlname := captureName
			s := TreeSitterSymbol{Point(start), Point(end), hlname, name}

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

func (t *ts_lang_def) create_query(buf []byte) (*sitter.Query, error) {
	q, err := sitter.NewQuery(buf, t.tslang)
	return q, err
}

func get_inherits(ss string) []string {
	ind := strings.Index(ss, "\n")
	if ind > 0 {
		ss = ss[:ind]
		const x = "; inherits:"
		if strings.HasPrefix(ss, x) {
			ss = strings.TrimPrefix(ss, x)
			args := strings.Split(ss, ",")
			inerits := []string{}
			for _, v := range args {
				inerits = append(inerits, strings.TrimSpace(v))
			}
			return inerits
		}
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

func (ts ts_lang_def) read_embbed(p string) ([]byte, error) {
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

var loaded_files = make(map[string]*TreeSitter)

func GetNewTreeSitter(name string) *TreeSitter {
	if len(name) == 0 {
		return nil
	}
	if ts, ok := loaded_files[name]; ok {
		return ts
	}
	v := NewTreeSitter(name)
	loaded_files[name] = v
	return v
}
func NewTreeSitter(name string) *TreeSitter {
	ret := &TreeSitter{
		parser:   sitter.NewParser(),
		filename: name,
		inited:   false,
	}
	ret.HlLine = make(map[int][]TreeSitterSymbol)
	return ret
}

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
		ts.callback_to_ui(cb)
		ts.inited = true
	}()
	return nil
}

func (ts *TreeSitter) callback_to_ui(cb func(*TreeSitter)) {
	if ts.tsdef.cb != nil {
		ts.tsdef.cb(ts)
	}
	if cb != nil {
		cb(ts)
	}
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
