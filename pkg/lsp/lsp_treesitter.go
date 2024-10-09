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
	"time"

	// "strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/tectiv3/go-lsp"

	ts_bash "github.com/smacker/go-tree-sitter/bash"
	ts_c "github.com/smacker/go-tree-sitter/c"
	ts_cpp "github.com/smacker/go-tree-sitter/cpp"
	ts_css "github.com/smacker/go-tree-sitter/css"
	ts_dockerfile "github.com/smacker/go-tree-sitter/dockerfile"
	ts_go "github.com/smacker/go-tree-sitter/golang"
	ts_html "github.com/smacker/go-tree-sitter/html"
	ts_java "github.com/smacker/go-tree-sitter/java"
	ts_js "github.com/smacker/go-tree-sitter/javascript"
	ts_lua "github.com/smacker/go-tree-sitter/lua"
	tree_sitter_markdown "github.com/smacker/go-tree-sitter/markdown/tree-sitter-markdown"
	ts_protobuf "github.com/smacker/go-tree-sitter/protobuf"
	ts_py "github.com/smacker/go-tree-sitter/python"
	ts_rust "github.com/smacker/go-tree-sitter/rust"
	ts_toml "github.com/smacker/go-tree-sitter/toml"
	ts_tsx "github.com/smacker/go-tree-sitter/typescript/tsx"
	ts_ts "github.com/smacker/go-tree-sitter/typescript/typescript"
	ts_yaml "github.com/smacker/go-tree-sitter/yaml"
	//"github.com/smacker/go-tree-sitter/markdown"
)

type TreesiterSymbolLine map[int][]TreeSitterSymbol
type Point struct {
	Row    uint32
	Column uint32
}
type TreeSitterSymbol struct {
	Begin, End Point
	SymbolName string
	Code       string
	Symbol     string
}
type TreeSitter struct {
	filename   SourceFile
	content    []byte
	parser     *sitter.Parser
	tree       *sitter.Tree
	sourceCode []byte
	HlLine     TreesiterSymbolLine
	Outline    []*Symbol
	tsdef      *ts_lang_def
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
	name            string
	filedetect      lsplang
	tslang          *sitter.Language
	def_ext         []string
	parser          func(*TreeSitter, outlinecb)
	hl              *sitter.Query
	local           *sitter.Query
	outline         *sitter.Query
	default_outline bool
}

const query_highlights = "highlights"
const query_locals = "locals"
const query_outline = "outline"

// const query_textobjects = "textobjects"
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
		nil,
		nil,
		true,
	}
	go func() {
		ret.load_scm()
	}()
	return ret
}

func (ret *ts_lang_def) load_scm() {
	if h, er := ret.query(query_highlights); er == nil {
		ret.hl = h
	} else {
		log.Println("fail to load highlights ", ret.name, er)
	}
	if h, er := ret.query(query_locals); er == nil {
		ret.local = h
	} else {
		log.Println("fail to load local ", ret.name, er)
	}
	if h, er := ret.query(query_outline); er == nil {
		ret.outline = h
	} else {
		log.Println("fail to load outline ", ret.name, er)
	}
}
func (tsdef *ts_lang_def) create_treesitter(file string) *TreeSitter {
	ret := NewTreeSitter(file, []byte{})
	ret.tsdef = tsdef
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
	// log.Println(t.name, "heri", queryname, heris)
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
func (s *ts_lang_def) set_default_outline() *ts_lang_def {
	s.default_outline = true
	return s
}
func (s *ts_lang_def) setparser(parser func(*TreeSitter, outlinecb)) *ts_lang_def {
	s.parser = parser
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
	return nil
}

// IsMe implements lsplang.
func (l lsp_dummy) IsMe(filename string) bool {
	return false
}

// IsSource implements lsplang.
func (l lsp_dummy) IsSource(filename string) bool {
	return false
}

// Launch_Lsp_Server implements lsplang.
func (l lsp_dummy) Launch_Lsp_Server(core *lspcore, wk WorkSpace) error {
	return nil
}

// Resolve implements lsplang.
func (l lsp_dummy) Resolve(sym lsp.SymbolInformation, symfile *Symbol_file) bool {
	return false
}
func markdown_parser(ts *TreeSitter) {
	if len(ts.Outline) > 0 {
		return
	}
	const head = "markup.heading"
	for _, line := range ts.HlLine {
		for _, s := range line {
			if strings.Index(s.SymbolName, head) == 0 {
				ss := ts_to_symbol(s, ts)
				aa := Symbol{
					SymInfo:   ss,
					Classname: s.Code,
				}
				ts.Outline = append(ts.Outline, &aa)
			}
		}
	}
	sort.Slice(ts.Outline, func(i, j int) bool {
		return ts.Outline[i].SymInfo.Location.Range.Start.Line < ts.Outline[j].SymInfo.Location.Range.Start.Line
	})
}

func ts_to_symbol(s TreeSitterSymbol, ts *TreeSitter) lsp.SymbolInformation {
	ss := lsp.SymbolInformation{
		Name: s.Code,
		Kind: lsp.SymbolKindVariable,
		Location: lsp.Location{
			URI: lsp.NewDocumentURI(ts.filename.Path()),
			Range: lsp.Range{
				Start: lsp.Position{Line: int(s.Begin.Row), Character: int(s.Begin.Column)},
				End:   lsp.Position{Line: int(s.End.Row), Character: int(s.End.Column)},
			},
		},
	}
	return ss
}
func java_outline(ts *TreeSitter, cb outlinecb) {
	if len(ts.Outline) > 0 {
		return
	}
	items := []*lsp.SymbolInformation{}
	if ts.tsdef.outline != nil {
		ret, err := ts.query_buf(ts.tsdef.outline)
		if err != nil {
			return
		}
		keys := []int{}
		for i := range ret {
			keys = append(keys, i)
		}
		sort.Ints(keys)
		for _, lineno := range keys {
			line, err := ret[lineno]
			if !err {
				continue
			}
			for _, item := range line {
				if item.Symbol == "line_comment" {
					continue
				}
				code := item.Code
				Range := item.lsprange()
				switch item.SymbolName {
				case "definition.method", "definition.class", "definition.interface":
					{
						c := ts_to_symbol(item, ts)
						switch item.Symbol {
						case "interface_declaration":
							c.Kind = lsp.SymbolKindInterface
						case "type_declaration":
							c.Kind = lsp.SymbolKindClass
						case "field_declaration":
							c.Kind = lsp.SymbolKindField
						case "enum_specifier":
							c.Kind = lsp.SymbolKindEnum
						case "method_declaration":
							c.Kind = lsp.SymbolKindMethod
						case "struct_specifier":
							c.Kind = lsp.SymbolKindStruct
						case "class_specifier", "class_declaration":
							c.Kind = lsp.SymbolKindClass
						case "function_definition", "function_declaration", "function_item":
							c.Kind = lsp.SymbolKindFunction
						default:
							log.Printf("query_result:%s| symbol:%20s    | code:%20s", item.SymbolName, item.Symbol, item.Code)
						}
						items = append(items, &c)
					}
				case "name":
					{
						foreach_check(items, Range, &item, func(v *lsp.SymbolInformation, tss *TreeSitterSymbol) bool {
							v.Name = tss.Code
							return true
						})
					}
				case "variable.member":
					{
						c := ts_to_symbol(item, ts)
						c.Kind = lsp.SymbolKindField
						c.Name = item.Code
						items = append(items, &c)
					}
				default:
					log.Printf("-----| %20s | %20s | %20s  |%s", item.PositionInfo(), item.SymbolName, item.Symbol, code)
				}
			}
		}
	}
	var s = Symbol_file{lsp: lsp_base{core: &lspcore{lang: lsp_dummy{}}}}
	document_symbol := []lsp.SymbolInformation{}
	for _, v := range items {
		document_symbol = append(document_symbol, *v)
	}
	s.build_class_symbol(document_symbol, 0, nil)
	ts.Outline = s.Class_object
}

type outlinecb func([]*lsp.SymbolInformation) []*lsp.SymbolInformation

func rs_outline(ts *TreeSitter, cb outlinecb) {
	if len(ts.Outline) > 0 {
		return
	}
	items := []*lsp.SymbolInformation{}
	if ts.tsdef.outline != nil {
		ret, err := ts.query_buf(ts.tsdef.outline)
		if err != nil {
			return
		}
		keys := []int{}
		for i := range ret {
			keys = append(keys, i)
		}
		sort.Ints(keys)
		for _, lineno := range keys {
			line, err := ret[lineno]
			if !err {
				continue
			}
			for _, item := range line {
				if item.Symbol == "line_comment" {
					continue
				}
				code := item.Code
				Range := item.lsprange()
				switch item.SymbolName {
				case "item":
					{
						c := ts_to_symbol(item, ts)
						switch item.Symbol {
						case "interface_declaration":
							c.Kind = lsp.SymbolKindInterface
						case "type_declaration":
							c.Kind = lsp.SymbolKindClass
						case "field_declaration":
							c.Kind = lsp.SymbolKindField
						case "enum_specifier":
							c.Kind = lsp.SymbolKindEnum
						case "method_elem", "method_declaration":
							c.Kind = lsp.SymbolKindMethod
						case "struct_specifier":
							c.Kind = lsp.SymbolKindStruct
						case "class_specifier":
							c.Kind = lsp.SymbolKindClass
						case "function_definition", "function_declaration", "function_item":
							c.Kind = lsp.SymbolKindFunction
						default:
							log.Printf("query_result:%s| symbol:%20s    | code:%20s", item.SymbolName, item.Symbol, item.Code)
						}
						items = append(items, &c)
					}
				case "context":
					{
						foreach_check(items, Range, &item, func(v *lsp.SymbolInformation, tss *TreeSitterSymbol) bool {
							switch item.Symbol {
							case "fn", "func":
								{
									v.Kind = lsp.SymbolKindFunction
									return true
								}
							case "class":
								{
									v.Kind = lsp.SymbolKindClass
									return true
								}
							case "field_declaration":
								{
									v.Kind = lsp.SymbolKindField
								}
							default:
							}
							return false
						})
					}
				case "name":
					{
						foreach_check(items, Range, &item, func(v *lsp.SymbolInformation, tss *TreeSitterSymbol) bool {
							v.Name = tss.Code
							return true
						})
					}
				default:
					log.Printf("-----| %20s | %20s | %20s  |%s", item.PositionInfo(), item.SymbolName, item.Symbol, code)
				}
			}
		}
	}
	var s = Symbol_file{lsp: lsp_base{core: &lspcore{lang: lsp_dummy{}}}}
	document_symbol := []lsp.SymbolInformation{}
	if cb != nil {
		items = cb(items)
	}
	for _, v := range items {
		document_symbol = append(document_symbol, *v)
	}
	s.build_class_symbol(document_symbol, 0, nil)
	ts.Outline = s.Class_object
}

func foreach_check(items []*lsp.SymbolInformation, Range lsp.Range, item *TreeSitterSymbol, check func(*lsp.SymbolInformation, *TreeSitterSymbol) bool) {
	var matched *lsp.SymbolInformation
	for i := range items {
		v := items[i]
		r := v.Location.Range
		if Range.Overlaps(r) {
			if matched == nil {
				matched = v
			} else {
				prev_range := matched.Location.Range
				new_range := v.Location.Range
				if new_range.Overlaps(prev_range) {
					matched = v
				}
			}
		}
	}
	if matched == nil {
		return
	}
	check(matched, item)
}
func bash_parser(ts *TreeSitter, cb outlinecb) {
	if len(ts.Outline) > 0 {
		return
	}
	Outline := []*Symbol{}
	if ts.tsdef.local != nil {
		lines, err := ts.query_buf(ts.tsdef.local)
		if err != nil {
			return
		}
		ss := get_ts_symbol(lines, ts)
		for _, v := range ss {
			Outline = append(Outline, &Symbol{
				SymInfo:   v,
				Members:   []Symbol{},
				Classname: "",
			})
		}
	}
	ts.Outline = Outline
	sort.Slice(ts.Outline, func(i, j int) bool {
		return ts.Outline[i].SymInfo.Location.Range.Start.Line < ts.Outline[j].SymInfo.Location.Range.Start.Line
	})
}

var tree_sitter_lang_map = []*ts_lang_def{
	new_tsdef("lua", lsp_dummy{}, ts_lua.GetLanguage()).set_ext([]string{"lua"}).setparser(rs_outline),
	new_tsdef("rust", lsp_dummy{}, ts_rust.GetLanguage()).set_ext([]string{"rs"}).setparser(rs_outline),
	new_tsdef("yaml", lsp_dummy{}, ts_yaml.GetLanguage()).set_ext([]string{"yaml", "yml"}).setparser(func(ts *TreeSitter, o outlinecb) {
		rs_outline(ts, func(si []*lsp.SymbolInformation) []*lsp.SymbolInformation {
			items := si
			sort.Slice(items, func(i, j int) bool {
				return items[i].Location.Range.Start.Line < items[j].Location.Range.Start.Line
			})
			ret := []*lsp.SymbolInformation{}
			for _, newone := range items {
				var find *lsp.SymbolInformation
				for _, prev := range ret {
					if newone.Location.Range.Overlaps(prev.Location.Range) {
						if find == nil {
							find = prev
						} else {
							if prev.Location.Range.Overlaps(find.Location.Range) {
								find = prev
							}
						}
					}
				}
				if find == nil {
					ret = append(ret, newone)
				} else {
					find.Kind = lsp.SymbolKindInterface
					newone.Kind = lsp.SymbolKindField
					ret = append(ret, newone)
				}
			}
			return ret
		})
	}),
	new_tsdef("proto", lsp_dummy{}, ts_protobuf.GetLanguage()).set_ext([]string{"proto"}).setparser(rs_outline),
	new_tsdef("css", lsp_dummy{}, ts_css.GetLanguage()).set_ext([]string{"css"}).setparser(rs_outline),
	new_tsdef("dockerfile", lsp_dummy{}, ts_dockerfile.GetLanguage()).set_ext([]string{"dockfile"}).setparser(rs_outline),
	new_tsdef("html", lsp_dummy{}, ts_html.GetLanguage()).set_ext([]string{"html"}).setparser(rs_outline),
	new_tsdef("toml", lsp_dummy{}, ts_toml.GetLanguage()).set_ext([]string{"toml"}).setparser(rs_outline),
	new_tsdef("java", lsp_dummy{}, ts_java.GetLanguage()).set_ext([]string{"java"}).setparser(java_outline),
	new_tsdef("bash", lsp_dummy{}, ts_bash.GetLanguage()).set_ext([]string{"sh"}).setparser(bash_parser),
	new_tsdef("go", lsp_lang_go{}, ts_go.GetLanguage()).setparser(func(ts *TreeSitter, o outlinecb) {
		rs_outline(ts, func(si []*lsp.SymbolInformation) []*lsp.SymbolInformation {
			if len(si) > 0 {
				if content, err := os.ReadFile(si[0].Location.URI.AsPath().String()); err == nil {
					lines := strings.Split(string(content), "\n")
					ret := si
					for _, v := range ret {
						if is_class(v.Kind) {
							line := lines[v.Location.Range.Start.Line]
							if strings.Index(line, "interface") > 0 {
								v.Kind = lsp.SymbolKindInterface
							}else if strings.Index(line, "struct") > 0 {
								v.Kind = lsp.SymbolKindStruct
							}
							continue
						}
					}
					return ret
				}
			}
			return si
		})
	}).set_default_outline(),
	new_tsdef("cpp", lsp_lang_cpp{}, ts_cpp.GetLanguage()).set_ext([]string{"h", "hpp", "cc", "cpp"}).setparser(rs_outline),
	new_tsdef("c", lsp_lang_cpp{}, ts_c.GetLanguage()).setparser(rs_outline),
	new_tsdef("python", lsp_lang_py{}, ts_py.GetLanguage()).setparser(rs_outline),
	new_tsdef(ts_name_tsx, lsp_dummy{}, ts_tsx.GetLanguage()).set_ext([]string{"tsx"}).setparser(rs_outline).set_default_outline(),
	new_tsdef(ts_name_javascript, lsp_ts{LanguageID: string(JAVASCRIPT)}, ts_js.GetLanguage()).set_ext([]string{"js"}).setparser(rs_outline),
	new_tsdef(ts_name_typescript, lsp_ts{LanguageID: string(TYPE_SCRIPT)}, ts_ts.GetLanguage()).set_ext([]string{"ts"}).setparser(rs_outline),
	new_tsdef(ts_name_markdown, lsp_md{}, tree_sitter_markdown.GetLanguage()).setparser(rs_outline),
}

func (t *TreeSitter) DefaultOutline() bool {
	return t.tsdef.default_outline
}
func (t *TreeSitter) Init(cb func(*TreeSitter)) error {
	if t.tsdef != nil {
		t.Loadfile(t.tsdef.tslang, cb)
		return nil
	}
	for i := range tree_sitter_lang_map {
		v := tree_sitter_lang_map[i]
		if ts_name := v.get_ts_name(t.filename.Path()); len(ts_name) > 0 {
			t.tsdef = v
			t.Loadfile(v.tslang, cb)
			return nil
		}
	}
	return fmt.Errorf("not implemented")
}

func (t TreeSitter) query(queryname string) (TreesiterSymbolLine, error) {
	if queryname == query_highlights {
		return t.query_buf(t.tsdef.hl)
	}
	return make(TreesiterSymbolLine), nil
}

func (t *TreeSitter) query_buf(q *sitter.Query) (TreesiterSymbolLine, error) {
	var SymbolsLine TreesiterSymbolLine = make(TreesiterSymbolLine)
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
			s := TreeSitterSymbol{Point(start), Point(end), hlname, name, t.tsdef.tslang.SymbolName(c.Node.Symbol())}

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
		_, err := io.Copy(d, file)
		// log.Println(n)
		if err != nil {
			return []byte{}, err
		} else {
			return imp.buf, nil
		}
	}
	return []byte{}, err
}

type SourceFile struct {
	filepathname string
	modTiem      time.Time
}

func (s SourceFile) Path() string {
	return s.filepathname
}

func NewFile(filename string) SourceFile {
	fileInfo, err := os.Stat(filename)
	modTime := time.Time{}
	if err == nil {
		modTime = fileInfo.ModTime()
	}
	return SourceFile{filepathname: filename, modTiem: modTime}
}
func (s SourceFile) Same(s1 SourceFile) bool {
	return s == s1
}

var loaded_files = make(map[string]*TreeSitter)

func GetNewTreeSitter(name string, content []byte) *TreeSitter {
	if len(name) == 0 {
		return nil
	}
	if len(content) == 0 {
		if ts, ok := loaded_files[name]; ok {
			if ts.filename.Same(NewFile(name)) {
				return ts
			}
		}
	}
	v := NewTreeSitter(name, content)
	loaded_files[name] = v
	return v
}
func NewTreeSitter(name string, content []byte) *TreeSitter {
	ret := &TreeSitter{
		parser:   sitter.NewParser(),
		filename: NewFile(name),
		content:  content,
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
	}()
	return nil
}

func (ts *TreeSitter) callback_to_ui(cb func(*TreeSitter)) {
	if ts.tsdef.parser != nil {
		ts.tsdef.parser(ts, nil)
	}
	if cb != nil {
		cb(ts)
	}
}

func get_ts_symbol(ret TreesiterSymbolLine, ts *TreeSitter) []lsp.SymbolInformation {
	prefix := "local.definition."
	symbols := []lsp.SymbolInformation{}
	scopes := []TreeSitterSymbol{}
	for lineno := range ret {
		line := ret[lineno]
		for i := 0; i < len(line); i++ {
			s := line[i]
			pos := s.PositionInfo()
			if s.SymbolName == "local.scope" {
				if strings.Index(s.Symbol, "expression") > 0 {
					continue
				}
				switch s.Symbol {
				case "method_declaration", "function_definition", "if_expression", "function_item", "closure_expression", "block":
					{
						scopes = append(scopes, s)
					}
				default:
					log.Println("=====================", s.SymbolName, s.Symbol, pos)
				}
			}
		}
	}
	for lineno := range ret {
		line := ret[lineno]
		for i := 0; i < len(line); i++ {
			s := line[i]
			pos := fmt.Sprint(s.Begin.Row, ":", s.Begin.Column, s.End.Row, ":", s.End.Column)
			Range := s.lsprange()
			if strings.Index(s.SymbolName, prefix) == 0 {
				symboltype := strings.Replace(s.SymbolName, prefix, "", 1)
				symbol_kind := map[string]lsp.SymbolKind{
					"method":      lsp.SymbolKindMethod,
					"function":    lsp.SymbolKindFunction,
					"namespace":   lsp.SymbolKindNamespace,
					"field":       lsp.SymbolKindField,
					"var":         lsp.SymbolKindVariable,
					"constructor": lsp.SymbolKindConstructor,
					"type.class":  lsp.SymbolKindClass,
					"type":        lsp.SymbolKindClass,
				}
				if kind, ok := symbol_kind[symboltype]; ok {
					log.Println("outline", s.Code, symboltype, pos)
					add := true
					switch kind {
					case lsp.SymbolKindVariable:
						{
							add = newFunction(scopes, Range, add)
						}
					}
					if !add {
						log.Println("unhandled skip symboltype:", symboltype, s.Code, pos, s.Symbol)
						continue
					}

					s := lsp.SymbolInformation{
						Name: s.Code,
						Kind: kind,
						Location: lsp.Location{
							URI:   lsp.NewDocumentURI(ts.filename.Path()),
							Range: Range,
						},
					}
					symbols = append(symbols, s)
				} else {
					log.Println("unhandled-symboltype:", symboltype, s.Code, pos, s.Symbol)
				}
			} else if s.SymbolName == "local.scope" {
				continue
			} else {
				// add := newFunction(scopes, Range, true)
				// if s.Symbol != "word" {
				// 	if add {
				// 		s := lsp.SymbolInformation{
				// 			Name: s.Code,
				// 			Kind: lsp.SymbolKindClass,
				// 			Location: lsp.Location{
				// 				URI:   lsp.NewDocumentURI(ts.filename),
				// 				Range: Range,
				// 			},
				// 		}
				// 		symbols = append(symbols, s)
				// 		continue
				// 	}
				// }
				// log.Println("unhandled symbol-name:", s.SymobName, s.Code, pos, s.Symbol)
			}
		}
	}
	return symbols
}

func (s TreeSitterSymbol) PositionInfo() string {
	pos := fmt.Sprint(s.Begin.Row, ":", s.Begin.Column, s.End.Row, ":", s.End.Column)
	return pos
}

func newFunction(scopes []TreeSitterSymbol, Range lsp.Range, add bool) bool {
	for _, v := range scopes {
		if Range.Overlaps(v.lsprange()) {
			// if v.Symbol == "method_declaration" || v.Symbol == "function_definition" {
			// 	add = false
			// }
			return false
		}
	}
	return add
}

func (s TreeSitterSymbol) lsprange() lsp.Range {
	x := lsp.Range{
		Start: lsp.Position{Line: int(s.Begin.Row), Character: int(s.Begin.Column)},
		End:   lsp.Position{Line: int(s.End.Row), Character: int(s.End.Column)},
	}
	return x
}

func (ts *TreeSitter) _load_file(lang *sitter.Language) error {
	ts.parser.SetLanguage(lang)
	if len(ts.content) == 0 {
		buf, err := os.ReadFile(ts.filename.Path())
		if err != nil {
			return err
		}
		ts.sourceCode = buf
	} else {
		ts.sourceCode = ts.content
	}
	tree, err := ts.parser.ParseCtx(context.Background(), nil, ts.sourceCode)
	ts.tree = tree
	return err
}
