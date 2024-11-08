package lspcore

import (
	"fmt"

	ts_csharp "github.com/smacker/go-tree-sitter/csharp"
	ts_protobuf "github.com/smacker/go-tree-sitter/protobuf"
	ts_swift "github.com/smacker/go-tree-sitter/swift"
	"github.com/tectiv3/go-lsp"
	ts_lua "github.com/tree-sitter-grammars/tree-sitter-lua/bindings/go"
	ts_toml "github.com/tree-sitter-grammars/tree-sitter-toml"
	ts_yaml "github.com/tree-sitter-grammars/tree-sitter-yaml"
	sitter "github.com/tree-sitter/go-tree-sitter"
	ts_bash "github.com/tree-sitter/tree-sitter-bash/bindings/go"
	ts_c "github.com/tree-sitter/tree-sitter-c/bindings/go"
	ts_cpp "github.com/tree-sitter/tree-sitter-cpp/bindings/go"
	ts_css "github.com/tree-sitter/tree-sitter-css/bindings/go"
	ts_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
	ts_html "github.com/tree-sitter/tree-sitter-html/bindings/go"
	ts_java "github.com/tree-sitter/tree-sitter-java/bindings/go"
	ts_js "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	tree_sitter_json "github.com/tree-sitter/tree-sitter-json/bindings/go"
	tree_sitter_markdown "github.com/tree-sitter/tree-sitter-markdown/bindings/go"
	ts_py "github.com/tree-sitter/tree-sitter-python/bindings/go"
	tree_sitter_ruby "github.com/tree-sitter/tree-sitter-ruby/bindings/go"
	ts_rust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
	tree_sitter_typescript "github.com/tree-sitter/tree-sitter-typescript/bindings/go"

	"strings"
	"unsafe"

	tree_sitter_zig "github.com/tree-sitter-grammars/tree-sitter-zig/bindings/go"
	"zen108.com/lspvi/pkg/debug"
)

type ts_lang_def struct {
	name            string
	filedetect      lsplang
	tslang          *sitter.Language
	def_ext         []string
	parser          func(*TreeSitter, outlinecb)
	hl              *sitter.Query
	inject          *sitter.Query
	local           *sitter.Query
	outline         *sitter.Query
	default_outline bool
	scm_loaded      bool

	intiqueue *TreesitterInit
}

// const query_textobjects = "textobjects"
func new_tsdef(
	name string,
	filedetect lsplang,
	tslang unsafe.Pointer,
) *ts_lang_def {
	ret := &ts_lang_def{
		name,
		filedetect,
		sitter.NewLanguage(tslang),
		[]string{},
		nil,
		nil,
		nil,
		nil,
		nil,
		true,
		false,
		&TreesitterInit{make(chan ts_init_call, 20), false},
	}
	// ret.load_scm()
	return ret
}

func (ret *ts_lang_def) load_scm() (err error) {
	if ret.scm_loaded {
		return
	}
	ret.scm_loaded = true
	ret.hl, err = ret.query(query_highlights)
	if err != nil {
		debug.ErrorLog(DebugTag, "fail to load highlights ", ret.name, err)
	}
	ret.inject, err = ret.query(inject_query)
	if err != nil {
		debug.ErrorLog(DebugTag, "fail to load ", inject_query, ret.name, err)
	}
	ret.local, err = ret.query(query_locals)
	if err != nil {
		debug.ErrorLog(DebugTag, "fail to load local ", ret.name, err)
	}
	ret.outline, err = ret.query(query_outline)
	if err != nil {
		debug.ErrorLog(DebugTag, "fail to load outline ", ret.name, err)
	}
	return
}
func (tsdef *ts_lang_def) create_treesitter(file string) *TreeSitter {
	ret := NewTreeSitter(file, []byte{})
	ret.tsdef = tsdef
	return ret
}
func (t *ts_lang_def) create_query_buffer(lang string, queryname string) ([]byte, error) {
	path := fmt.Sprintf("queries/%s/%s", lang, queryname+".scm")
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

var tree_sitter_lang_map = []*ts_lang_def{
	new_tsdef("go", lsp_lang_go{}, ts_go.Language()).setparser(func(ts *TreeSitter, o outlinecb) {
		rs_outline(ts, func(ts *TreeSitter, si *OutlineSymolList) {
			if len(si.items) > 0 {
				lines := []string{}
				if len(ts.sourceCode) > 0 {
					lines = strings.Split(string(ts.sourceCode), "\n")
				}
				ret := append([]*lsp.SymbolInformation{}, si.items...)
				for _, v := range ret {
					if is_class(v.Kind) {
						line := lines[v.Location.Range.Start.Line]
						if strings.Index(line, "interface") > 0 {
							v.Kind = lsp.SymbolKindInterface
						} else if strings.Index(line, "struct") > 0 {
							v.Kind = lsp.SymbolKindStruct
						}
						continue
					}
					if is_memeber(v.Kind) {
						if strings.Index(v.Name, "(().") == 0 {
							v.Name = strings.Replace(v.Name, "(().", "", 1)
						}
					}
				}
				si.items = ret
			}
		})
	}).set_default_outline(),
	new_tsdef("cpp", lsp_dummy{}, ts_cpp.Language()).set_ext([]string{"hpp", "cc", "cpp"}).setparser(rs_outline),
	new_tsdef("c", lsp_dummy{}, ts_c.Language()).setparser(rs_outline).set_ext([]string{"c", "h"}),
	new_tsdef("python", lsp_lang_py{}, ts_py.Language()).setparser(rs_outline),
	new_tsdef("lua", lsp_dummy{}, ts_lua.Language()).set_ext([]string{"lua"}).setparser(rs_outline),
	new_tsdef("zig", lsp_dummy{}, tree_sitter_zig.Language()).set_ext([]string{"zig"}).setparser(rs_outline),
	new_tsdef("rust", lsp_dummy{}, ts_rust.Language()).set_ext([]string{"rs"}).setparser(rs_outline),
	new_tsdef("yaml", lsp_dummy{}, ts_yaml.Language()).set_ext([]string{"yaml", "yml"}).setparser(func(ts *TreeSitter, o outlinecb) {
		rs_outline(ts, yaml_group)
	}),
	new_tsdef("csharp", lsp_dummy{}, ts_csharp.Language()).set_ext([]string{"cs"}).setparser(rs_outline),
	new_tsdef("swift", lsp_dummy{}, ts_swift.Language()).set_ext([]string{"swift"}).setparser(rs_outline),
	new_tsdef("proto", lsp_dummy{}, ts_protobuf.Language()).set_ext([]string{"proto"}).setparser(rs_outline),
	new_tsdef("css", lsp_dummy{}, ts_css.Language()).set_ext([]string{"css"}).setparser(rs_outline),
	// new_tsdef("dockerfile", lsp_dummy{}, ts_dockerfile.GetLanguage()).set_ext([]string{"dockfile"}).setparser(rs_outline),
	new_tsdef("html", lsp_dummy{}, ts_html.Language()).set_ext([]string{"html"}).setparser(rs_outline),
	new_tsdef("toml", lsp_dummy{}, ts_toml.Language()).set_ext([]string{"toml"}).setparser(func(ts *TreeSitter, o outlinecb) {
		rs_outline(ts, yaml_group)
	}),
	new_tsdef("json", lsp_dummy{}, tree_sitter_json.Language()).set_ext([]string{"json"}).setparser(func(ts *TreeSitter, o outlinecb) {
		rs_outline(ts, yaml_group)
	}),
	new_tsdef("java", lsp_dummy{}, ts_java.Language()).set_ext([]string{"java"}).setparser(java_outline),
	new_tsdef("bash", lsp_dummy{}, ts_bash.Language()).set_ext([]string{"sh"}).setparser(bash_parser),
	new_tsdef("ruby", lsp_dummy{}, tree_sitter_ruby.Language()).set_ext([]string{"ruby"}).setparser(bash_parser),
	new_tsdef(ts_name_tsx, lsp_dummy{}, tree_sitter_typescript.LanguageTSX()).set_ext([]string{"tsx"}).setparser(rs_outline).set_default_outline(),
	new_tsdef(ts_name_javascript, lsp_ts{LanguageID: string(JAVASCRIPT)}, ts_js.Language()).set_ext([]string{"js"}).setparser(rs_outline),
	new_tsdef(ts_name_typescript, lsp_ts{LanguageID: string(TYPE_SCRIPT)}, tree_sitter_typescript.LanguageTypescript()).set_ext([]string{"ts"}).setparser(rs_outline),
	new_tsdef(ts_name_markdown, lsp_md{}, tree_sitter_markdown.Language()).setparser(rs_outline),
}
