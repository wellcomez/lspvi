package lspcore

import (
	"context"
	"embed"
	"fmt"
	"github.com/reinhrst/fzf-lib"
	"github.com/tectiv3/go-lsp"
	sitter "github.com/tree-sitter/go-tree-sitter"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"zen108.com/lspvi/pkg/debug"
)

// import "github.com/tree-sitter-grammars/tree-sitter-toml"

// import ts_toml "github.com/tree-sitter-grammars/tree-sitter-toml/bindings/go"

type TreesiterSymbolLine map[int][]TreeSitterSymbol
type Point struct {
	Row    uint32
	Column uint32
}
type TreeSitterSymbol struct {
	Begin, End  Point
	CaptureName string
	Code        string
	Symbol      string
}
type ts_symbol_parser interface {
	resolve([]TreeSitterSymbol, *lsp.SymbolInformation, string) (bool, error)
}
type TreeSitter struct {
	filename       SourceFile
	content        []byte
	parser         *sitter.Parser
	tree           *sitter.Tree
	sourceCode     []byte
	HlLine         TreesiterSymbolLine
	Outline        []*Symbol
	tsdef          *ts_lang_def
	symbol_resolve ts_symbol_parser

	InjectOutline []*Symbol
}

// tree.Edit(sitter.EditInput{
//     StartIndex:  32,  // Byte offset where the change starts
//     OldEndIndex: 37,  // Byte offset where the old text ends
//     NewEndIndex: 34,  // Byte offset where the new text ends
// })

// - `StartIndex` is 32, marking the start of the change.
// - `OldEndIndex` is 37, marking the end of the original text `"Hello"`.
// - `NewEndIndex` is 34, marking the end of the new text `"Hi"`.
func (t *TreeSitter) EditChange(event CodeChangeEvent) {
	// for _, v := range event.TsEvents {
	// 	t.tree.Edit(sitter.EditInput{
	// 		StartIndex:  uint32(v.StartIndex),
	// 		OldEndIndex: uint32(v.OldEndIndex),
	// 		NewEndIndex: uint32(v.NewEndIndex),
	// 		StartPoint: sitter.Point{
	// 			Row:    v.StartPoint.Row,
	// 			Column: v.StartPoint.Column,
	// 		},
	// 		NewEndPoint: sitter.Point{
	// 			Row:    v.NewEndPoint.Row,
	// 			Column: v.NewEndPoint.Column,
	// 		},
	// 		OldEndPoint: sitter.Point{
	// 			Row:    v.OldEndPoint.Row,
	// 			Column: v.OldEndPoint.Column,
	// 		},
	// 	})
	// }
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

const inject_query = "injections"
const query_highlights = "highlights"
const query_locals = "locals"
const query_outline = "outline"

func markdown_parser(ts *TreeSitter) {
	if len(ts.Outline) > 0 {
		return
	}
	const head = "markup.heading"
	for _, line := range ts.HlLine {
		for _, s := range line {
			if strings.Index(s.CaptureName, head) == 0 {
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
func SymbolSwift(o Outline, ts *TreeSitter) (ret *lsp.SymbolInformation, done bool) {
	name := ""
	for _, v := range o {
		if v.CaptureName == "name" {
			if v.Symbol == "pattern" || v.Symbol == "type_identifier" {
				name = v.Code
				break
			}
		}
	}
	if ind := o.find("item"); ind != -1 {

		item := o[ind]
		c := ts_to_symbol(item, ts)
		ret = &c
		switch item.Symbol {
		case "property_identifier", "property_declaration":
			c.Kind = lsp.SymbolKindProperty
			// if strings.Contains(item.Code, "{") {
			// 	c.Kind = lsp.SymbolKindMethod
			// }
		case "interface_declaration", "protocol_declaration":
			c.Kind = lsp.SymbolKindInterface
		case "type_declaration":
			c.Kind = lsp.SymbolKindClass
		case "field_declaration", "public_field_definition":
			c.Kind = lsp.SymbolKindField
		case "enum_specifier":
			c.Kind = lsp.SymbolKindEnum
		case "method_elem", "method_declaration", "method_definition":
			c.Kind = lsp.SymbolKindMethod
		case "struct_specifier":
			c.Kind = lsp.SymbolKindStruct
		case "init_declaration", "deinit_declaration":
			c.Kind = lsp.SymbolKindConstructor
		case "class_specifier", "class_declaration":
			c.Kind = lsp.SymbolKindClass
		case "function_definition", "function_declaration", "function_item":
			c.Kind = lsp.SymbolKindFunction
		default:
			debug.TraceLogf("query_result:%s| symbol:%20s    | code:%20s", item.CaptureName, item.Symbol, item.Code)
		}
		code := item.Code
		c.Name = name
		if len(code) > 0 {
			if ts.symbol_resolve != nil {
				if yes, _ := ts.symbol_resolve.resolve(o, ret, code); yes {
					done = true
				}
			}
		}

	}
	return
}
func (o Outline) Symbol(ts *TreeSitter) (ret *lsp.SymbolInformation, done bool) {
	name := ""
	if i := o.find("name"); i != -1 {
		name = o[i].Code
	}
	if ind := o.find("item"); ind != -1 {

		item := o[ind]
		c := ts_to_symbol(item, ts)
		ret = &c
		switch item.Symbol {
		case "property_identifier":
			c.Kind = lsp.SymbolKindProperty
		case "interface_declaration":
			c.Kind = lsp.SymbolKindInterface
		case "type_declaration":
			c.Kind = lsp.SymbolKindClass
		case "field_declaration", "public_field_definition":
			c.Kind = lsp.SymbolKindField
		case "enum_specifier":
			c.Kind = lsp.SymbolKindEnum
		case "method_elem", "method_declaration", "method_definition":
			c.Kind = lsp.SymbolKindMethod
		case "struct_specifier":
			c.Kind = lsp.SymbolKindStruct
		case "class_specifier", "class_declaration":
			c.Kind = lsp.SymbolKindClass
		case "function_definition", "function_declaration", "function_item":
			c.Kind = lsp.SymbolKindFunction
		default:
			debug.TraceLogf("query_result:%s| symbol:%20s    | code:%20s", item.CaptureName, item.Symbol, item.Code)
		}
		code := item.Code
		c.Name = name
		if len(code) > 0 {
			if ts.symbol_resolve != nil {
				if yes, _ := ts.symbol_resolve.resolve(o, ret, code); yes {
					done = true
				}
			}
		}

	}
	return
}

func (o Outline) find(a string) (ret int) {
	ret = -1
	for i, item := range o {
		if item.CaptureName == a {
			ret = i
			return
		}
	}
	return
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
func (o Outline) JavaSymbol(ts *TreeSitter) (ret *lsp.SymbolInformation, done bool) {
	name := ""
	if i := o.find("name"); i != -1 {
		name = o[i].Code
	}
	var item TreeSitterSymbol
	for _, v := range []string{"definition.method", "definition.class", "definition.interface"} {
		if ind := o.find(v); ind != -1 {
			item = o[ind]
			c := ts_to_symbol(item, ts)
			c.Name = name
			ret = &c
			break
		}
	}
	if ret != nil {
		c := ret
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
			debug.TraceLogf("query_result:%s| symbol:%20s    | code:%20s", item.CaptureName, item.Symbol, item.Code)
		}
	} else {
		if idx := o.find("variable.member"); idx != -1 {
			a := ts_to_symbol(o[idx], ts)
			a.Kind = lsp.SymbolKindField
			ret = &a
			ret.Name = o[idx].Code
		}
	}
	return
}
func java_outline(ts *TreeSitter, cb outlinecb) {
	if len(ts.Outline) > 0 {
		return
	}
	items := []*lsp.SymbolInformation{}
	if ts.tsdef.outline != nil {
		ret, _ := ts.query_buf_outline(ts.tsdef.outline)
		for _, line := range ret {
			c, _ := line.JavaSymbol(ts)
			if c != nil {
				items = append(items, c)
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

type outlinecb func(*TreeSitter, *OutlineSymolList)
type OutlineSymolList struct {
	items []*lsp.SymbolInformation
}

func (o *OutlineSymolList) Add(item *lsp.SymbolInformation) {
	o.items = append(o.items, item)
}
func swift_outline(ts *TreeSitter, cb outlinecb) {
	if len(ts.Outline) > 0 {
		return
	}
	items := OutlineSymolList{}
	var ret []Outline
	if ts.tsdef.outline != nil {
		if r, err := ts.query_buf_outline(ts.tsdef.outline); err != nil {
			return
		} else {
			ret = r
		}
	}
	for _, line := range ret {
		sym, _ := SymbolSwift(line, ts)
		if sym == nil {
			continue
		}
		// for _, v := range line {
		// 	if v.Symbol == "pattern" || v.Symbol == "type_identifier" {
		// 		sym.Name = v.Code
		// 		break
		// 	}
		// }
		items.Add(sym)
	}
	ss := items.items
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Location.Range.Start.Line < ss[j].Location.Range.Start.Line
	})
	items.items = ss
	lang := lsp_dummy{}
	core := &lspcore{lang: lang}
	var s = Symbol_file{lsp: lsp_base{core: core}}
	document_symbol := []lsp.SymbolInformation{}
	if cb != nil {
		cb(ts, &items)
	}
	for _, v := range items.items {
		document_symbol = append(document_symbol, *v)
	}
	s.build_class_symbol(document_symbol, 0, nil)
	ts.Outline = s.Class_object
}
func rs_outline(ts *TreeSitter, cb outlinecb) {
	if len(ts.Outline) > 0 {
		return
	}
	items := OutlineSymolList{}
	var ret []Outline
	if ts.tsdef.outline != nil {
		if r, err := ts.query_buf_outline(ts.tsdef.outline); err != nil {
			return
		} else {
			ret = r
		}
	}
	for _, line := range ret {
		sym, done := line.Symbol(ts)
		if sym == nil {
			continue
		}
		if !done {
			for _, item := range line {
				if item.Symbol == "line_comment" || item.Symbol == "comment" {
					sym = nil
					break
				}
				code := item.Code
				switch item.CaptureName {
				case "context":
					{
						var v = sym
						switch item.Symbol {
						case "fn", "func":
							{
								if v.Kind != lsp.SymbolKindMethod {
									v.Kind = lsp.SymbolKindFunction
								}
							}
						case "class":
							{
								v.Kind = lsp.SymbolKindClass
							}
						case "field_declaration":
							{
								v.Kind = lsp.SymbolKindField
							}
						default:
						}
					}

				default:
					debug.TraceLogf("-----| %20s | %20s | %20s  |%s", item.PositionInfo(), item.CaptureName, item.Symbol, code)
				}
			}
		}
		if sym != nil {
			items.Add(sym)
		}
	}
	lang := lsp_dummy{}
	core := &lspcore{lang: lang}
	lsp_lang_go := lsp_lang_go{}
	lsp_cpp := lsp_lang_cpp{}
	if lsp_lang_go.IsMe(ts.filename.Path()) {
		core = &lspcore{lang: lsp_lang_go}
	} else if lsp_cpp.IsMe(ts.filename.Path()) {
		core = &lspcore{lang: lsp_cpp}
	}
	var s = Symbol_file{lsp: lsp_base{core: core}}
	document_symbol := []lsp.SymbolInformation{}
	if cb != nil {
		cb(ts, &items)
	}
	for _, v := range items.items {
		document_symbol = append(document_symbol, *v)
	}
	sort.Slice(document_symbol, func(i, j int) bool {
		return document_symbol[i].Location.Range.Start.Line < document_symbol[j].Location.Range.Start.Line
	})
	s.build_class_symbol(document_symbol, 0, nil)
	ts.Outline = s.Class_object
}

func bash_parser(ts *TreeSitter, cb outlinecb) {
	if len(ts.Outline) > 0 {
		return
	}
	Outline := []*Symbol{}
	if ts.tsdef.local != nil {
		lines, err := ts.query_buf_outline(ts.tsdef.local)
		if err != nil {
			return
		}
		ss := treesitter_local(lines, ts)
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
func yaml_group(ts *TreeSitter, si *OutlineSymolList) {
	items := si.items
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
	si.items = ret
}

type ts_call int

const (
	ts_load_call ts_call = iota
)

type block_call struct {
	done chan bool
}
type ts_init_call struct {
	t     *TreeSitter
	cb    func(*TreeSitter)
	call  ts_call
	block *block_call
}
type TreesitterInit struct {
	t     chan ts_init_call
	start bool
}

// var ts_init = &TreesitterInit{t: make(chan ts_init_call, 10), start: false}

func (ts_int *TreesitterInit) Run(t ts_init_call) {
	if !ts_int.start {
		ts_int.start = true
		go func() {
			for {
				select {
				case call := <-ts_int.t:
					// t.t.init(t.cb)
					t := call.t
					cb := call.cb
					switch call.call {
					case ts_load_call:
						t.Loadfile(t.tsdef.tslang, cb)
						if call.block != nil {
							call.block.done <- true
						}
					default:
					}
				}
			}
		}()
	}
	ts_int.t <- t
}

func (t *TreeSitter) DefaultOutline() bool {
	return t.tsdef.default_outline
}
func (t *TreeSitter) Init(cb func(*TreeSitter)) error {
	return t.init(cb)
}
func (t *TreeSitter) initblock() error {
	var b = &block_call{
		done: make(chan bool),
	}
	if t.tsdef != nil {

		t.tsdef.intiqueue.Run(ts_init_call{t, nil, ts_load_call, b})
		<-b.done
		return nil
	}
	for i := range tree_sitter_lang_map {
		v := tree_sitter_lang_map[i]
		if ts_name := v.get_ts_name(t.filename.Path()); len(ts_name) > 0 {
			v.load_scm()
			t.tsdef = v
			// t.Loadfile(v.tslang, cb)
			t.tsdef.intiqueue.Run(ts_init_call{t, nil, ts_load_call, b})
			<-b.done
			return nil
		}
	}
	return fmt.Errorf("not implemented")
}
func (t *TreeSitter) init(cb func(*TreeSitter)) error {
	if t.tsdef != nil {
		t.tsdef.intiqueue.Run(ts_init_call{t, cb, ts_load_call, nil})
		return nil
	}
	// t.Loadfile(v.tslang, cb)
	if t.load_ts_def() {
		t.tsdef.intiqueue.Run(ts_init_call{t, cb, ts_load_call, nil})
		return nil
	}
	return fmt.Errorf("not implemented")
}

func (t *TreeSitter) load_ts_def() bool {
	for i := range tree_sitter_lang_map {
		v := tree_sitter_lang_map[i]
		if ts_name := v.get_ts_name(t.filename.Path()); len(ts_name) > 0 {
			v.load_scm()
			t.tsdef = v
			return true
		}
	}
	return false
}

type ts_inject struct {
	lang     string
	content  []TreeSitterSymbol
	hline    TreesiterSymbolLine
	Outline  []*Symbol
	hostfile string
}

func offset_lsp_range(r lsp.Range, offset int) (ret lsp.Range) {
	r.Start.Line = r.Start.Line + offset
	r.End.Line = r.End.Line + offset
	ret = r
	return
}
func (inject *ts_inject) hl() {
	for _, v := range inject.content {
		t := NewTreeSitterParse(inject.lang, v.Code)
		if t.load_ts_def() {
			t.Loadfile(t.tsdef.tslang, nil)

			var Outline []*Symbol
			var HlLine = make(TreesiterSymbolLine)
			for k, l := range t.HlLine {
				for i := range l {
					l[i].Begin.Row = v.Begin.Row + l[i].Begin.Row
					l[i].End.Row = v.Begin.Row + l[i].End.Row
				}
				HlLine[k+int(v.Begin.Row)] = l
			}
			for _, sym := range t.Outline {
				sym.SymInfo.Location.Range = offset_lsp_range(sym.SymInfo.Location.Range, int(v.Begin.Row))
				sym.SymInfo.Location.URI = lsp.NewDocumentURI(inject.hostfile)
				var Members []Symbol
				for _, m := range sym.Members {
					m.SymInfo.Location.Range = offset_lsp_range(m.SymInfo.Location.Range, int(v.Begin.Row))
					Members = append(Members, m)
				}
				sym.Members = Members
				Outline = append(Outline, sym)
			}
			inject.Outline = Outline
			for k, l := range HlLine {
				if line, ok := inject.hline[k]; ok {
					line = append(line, l...)
					inject.hline[k] = line
				} else {
					inject.hline[k] = l
				}
			}
		}
		// ts.HlLine
	}
}

func (ts *TreeSitter) get_higlight(queryname string) (ret TreesiterSymbolLine, err error) {
	if queryname == query_highlights {
		ret, err = ts.query_highlight_buff(ts.tsdef.hl)
		if ts.tsdef.inject != nil {
			inejcts := []*ts_inject{}
			if injected, err := ts.query_buf_outline(ts.tsdef.inject); err == nil && len(injected) > 0 {
				debug.DebugLog(DebugTag, "inject", len(injected))
				var d *ts_inject
				for _, aaa := range injected {
					for _, v := range aaa {
						if v.CaptureName == "_lang" {
							d = &ts_inject{}
							d.lang = "inject." + v.Code
							d.content = nil
							d.hline = make(TreesiterSymbolLine)
							d.hostfile = ts.filename.Path()
							inejcts = append(inejcts, d)
						} else {
							if d != nil {
								d.content = append(d.content, v)
							}
						}
					}
				}
			}
			for _, v := range inejcts {
				v.hl()
				ts.InjectOutline = append(ts.InjectOutline, v.Outline...)
			}
			for _, inj := range inejcts {
				for lienno, v := range inj.hline {
					if line, ok := ret[lienno]; ok {
						line = append(line, v...)
						ret[lienno] = line
					} else {
						ret[lienno] = v
					}
				}
			}
		}
		return
	}
	return make(TreesiterSymbolLine), nil
}

type Outline []TreeSitterSymbol

func (t *TreeSitter) query_buf_outline(query *sitter.Query) (ret []Outline, err error) {
	if query == nil {
		err = fmt.Errorf("query not found")
		return
	}
	qc := sitter.NewQueryCursor()
	matches := qc.Matches(query, t.tree.RootNode(), t.sourceCode)

	for match := matches.Next(); match != nil; match = matches.Next() {
		var outline Outline
		for _, capture := range match.Captures {
			// debug.DebugLogf(
			// 	DebugTag,
			// 	"Match %d, Capture %d (%s): %s\n",
			// 	match.PatternIndex,
			// 	capture.Index,
			// 	query.CaptureNames()[capture.Index],
			// 	capture.Node.Utf8Text(t.sourceCode),
			// )
			start := capture.Node.StartPosition()
			end := capture.Node.EndPosition()
			CaptureName := query.CaptureNames()[capture.Index]
			// t.tsdef.tslang.NodeKindIsNamed(bb)
			s := TreeSitterSymbol{
				Begin:       Point{Column: uint32(start.Column), Row: uint32(start.Row)},
				End:         Point{Column: uint32(end.Column), Row: uint32(end.Row)},
				Code:        capture.Node.Utf8Text(t.sourceCode),
				CaptureName: CaptureName,
				Symbol:      capture.Node.Kind(),
			}
			outline = append(outline, s)
		}
		ret = append(ret, outline)
	}
	// for {
	// 	m, ok := qc.NextMatch()
	// 	if !ok {
	// 		break
	// 	}
	// 	var outline Outline
	// 	for i := 0; i < len(m.Captures); i++ {
	// 		c := m.Captures[i]

	// 		captureName := q.CaptureNameForId(c.Index)

	// 		start := c.Node.StartPoint()
	// 		end := c.Node.EndPoint()
	// 		name := c.Node.Content(t.sourceCode)

	// 		hlname := captureName
	// 		s := TreeSitterSymbol{Point(start), Point(end), hlname, name, t.tsdef.tslang.SymbolName(c.Node.Symbol())}
	// 		outline = append(outline, s)
	// 	}
	// 	ret = append(ret, outline)
	// }
	return
}
func (t *TreeSitter) query_highlight_buff(query *sitter.Query) (TreesiterSymbolLine, error) {
	var SymbolsLine TreesiterSymbolLine = make(TreesiterSymbolLine)
	if query == nil {
		return SymbolsLine, fmt.Errorf("query not found")
	}
	qc := sitter.NewQueryCursor()
	matches := qc.Matches(query, t.tree.RootNode(), t.sourceCode)
	for match := matches.Next(); match != nil; match = matches.Next() {
		for _, capture := range match.Captures {
			start := capture.Node.StartPosition()
			end := capture.Node.EndPosition()
			s := TreeSitterSymbol{
				Begin: Point{Column: uint32(start.Column), Row: uint32(start.Row)},
				End:   Point{Column: uint32(end.Column), Row: uint32(end.Row)},
				// Code:   capture.Node.Utf8Text(t.sourceCode),
				CaptureName: query.CaptureNames()[capture.Index],
				Symbol:      capture.Node.Kind(),
			}
			// idf:=query.CaptureQuantifiers(uint(match.PatternIndex))[capture.Index]
			// debug.DebugLogf(
			// 	DebugTag,
			// 	"Match %d, Capture %d (%s): %d code=%s\n",
			// 	match.PatternIndex,
			// 	capture.Index,
			// 	// query.CaptureNames()[capture.Index],
			// 	s.Symbol,
			// 	-1,
			// 	s.Code,
			// 	//  TreeSitterSymbol{Point(start), Point(end), hlname, name, t.tsdef.tslang.SymbolName(c.Node.Symbol())}
			// )
			row := int(s.Begin.Row)
			if _, ok := SymbolsLine[row]; !ok {
				SymbolsLine[row] = []TreeSitterSymbol{s}
			} else {
				SymbolsLine[row] =
					append(SymbolsLine[row], s)
			}
		}
	}
	// for {
	// 	m, ok := qc.NextMatch()
	// 	if !ok {
	// 		break
	// 	}

	// 	for i := range m.Captures {
	// 		c := m.Captures[i]

	// 		captureName := q.CaptureNameForId(c.Index)

	// 		start := c.Node.StartPoint()
	// 		end := c.Node.EndPoint()
	// 		name := c.Node.Content(t.sourceCode)

	// 		hlname := captureName
	// 		s := TreeSitterSymbol{Point(start), Point(end), hlname, name, t.tsdef.tslang.SymbolName(c.Node.Symbol())}

	// 		row := int(s.Begin.Row)
	// 		if _, ok := SymbolsLine[row]; !ok {
	// 			SymbolsLine[row] = []TreeSitterSymbol{s}
	// 		} else {
	// 			SymbolsLine[row] =
	// 				append(SymbolsLine[row], s)
	// 		}
	// 	}
	// }
	return SymbolsLine, nil
}

func (t *ts_lang_def) create_query(buf []byte) (*sitter.Query, error) {
	q, err := sitter.NewQuery(t.tslang, string(buf))
	if q != nil {
		return q, nil
	}
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
var lang_ts_symbol = make(map[string]*LangTreesitterSymbol)

type LangTreesitterSymbol struct {
	fzf    *fzf.Fzf
	ext    string
	files  []string
	symbol []Symbol
}

func GetLangTreeSitterSymbol(name string) (s *LangTreesitterSymbol) {
	if s, ok := lang_ts_symbol[name]; ok {
		return s
	} else {
		s := NewLangTreeSitterSymbol(name)
		lang_ts_symbol[name] = s
		return s
	}
}
func (sym *LangTreesitterSymbol) WorkspaceQuery(query string) (ret []lsp.SymbolInformation, err error) {
	debug.DebugLog(LSP_DEBUG_TAG, "workquery-begin", query)
	// sym.fzf.Search(query)
	// result := <-sym.fzf.GetResultChannel()
	// for _, m := range result.Matches {
	// 	// if m.Score < 50 {
	// 	// 	continue
	// 	// }
	// 	s := sym.symbol[m.HayIndex].SymInfo
	// 	ret = append(ret, s)
	// }
	for i := range sym.symbol {
		ret = append(ret, sym.symbol[i].SymInfo)
	}
	debug.DebugLog(LSP_DEBUG_TAG, "workquery-end", query, len(ret))
	return
}
func child_symbol(s Symbol, prefix string) (ret []Symbol) {
	if len(prefix) > 0 {
		s.SymInfo.Name = strings.Join([]string{prefix, s.SymInfo.Name}, ".")
	}
	ret = append(ret, s)
	prefix = s.SymInfo.Name
	for _, v := range s.Members {
		ret = append(ret, child_symbol(v, prefix)...)
	}
	return ret
}
func NewLangTreeSitterSymbol(name string) (s *LangTreesitterSymbol) {
	s = &LangTreesitterSymbol{}
	var ret []Symbol
	ext := filepath.Ext(name)
	s.ext = ext
	for file, ts := range loaded_files {
		if filepath.Ext(file) == ext {
			for _, v := range ts.Outline {
				ret = append(ret, child_symbol(*v, "")...)
			}
		}
		s.files = append(s.files, file)
	}

	keys := []string{}
	for _, v := range ret {
		keys = append(keys, v.SymInfo.Name)
	}
	for _, v := range keys {
		debug.DebugLog(LSP_DEBUG_TAG, "NewLangTreeSitterSymbol", v)
	}
	s.symbol = ret
	opt := fzf.DefaultOptions()
	opt.Fuzzy = true
	s.fzf = fzf.New(keys, opt)
	return
}
func NewTreeSitterParse(name string, data string) *TreeSitter {
	if len(name) == 0 {
		return nil
	}
	v := NewTreeSitter(name, []byte(data))
	return v
}
func GetNewTreeSitter(name string, event CodeChangeEvent) *TreeSitter {
	if len(name) == 0 {
		return nil
	}
	if len(event.Data) == 0 {
		if ts, ok := loaded_files[name]; ok {
			if !event.Full && len(event.TsEvents) > 0 {
			} else {
				if ts.filename.Same(NewFile(name)) {
					return ts
				}
			}
		}
	}
	v := NewTreeSitter(name, event.Data)
	loaded_files[name] = v
	return v
}

type cpp_go_symbol_resolve struct {
	class_obbject []*lsp.SymbolInformation
}

func (t *cpp_go_symbol_resolve) trim_symbol_name(sym *lsp.SymbolInformation) {
	if strings.Contains(sym.Name, ":") {
		debug.DebugLog("cpp_go_symbol_resolve", "trim_symbol_name", "sym.Name", sym.Name)
	}
	if idx := strings.Index(sym.Name, "\n"); idx > 0 {
		name := sym.Name[:idx]
		nam2 := sym.Name[idx+1:]
		nam2 = strings.TrimLeft(nam2, " ")
		sym.Name = name + nam2
	}
}
func (t *cpp_go_symbol_resolve) resolve(line []TreeSitterSymbol, sym *lsp.SymbolInformation, code string) (done bool, err error) {
	if strings.Index(code, "PrerenderActivationNavigationState") > 0 {
		debug.DebugLog("ts", code)
	}
	done = false
	if sym.Kind == lsp.SymbolKindClass || sym.Kind == lsp.SymbolKindInterface || sym.Kind == lsp.SymbolKindStruct {
		t.class_obbject = append(t.class_obbject, sym)
		return
	}
	t.trim_symbol_name(sym)
	if sym.Kind == lsp.SymbolKindField {
		for _, v := range line {
			if v.Symbol == ")" && v.CaptureName == "context" {
				sym.Kind = lsp.SymbolKindMethod
			}
			if v.CaptureName == "name" {
				sym.Name = v.Code
				t.trim_symbol_name(sym)
			}
		}
		if sym.Kind == lsp.SymbolKindMethod {
			t.change_to_method(sym)
			done = true
			return
		}

	} else if sym.Kind == lsp.SymbolKindFunction || sym.Kind == lsp.SymbolKindMethod {

		if strings.Index(sym.Name, "::") > 0 {
			sym.Kind = lsp.SymbolKindMethod
			done = true
			return
		}
		t.change_to_method(sym)
	}
	return
}

func (t *cpp_go_symbol_resolve) change_to_method(sym *lsp.SymbolInformation) bool {
	for _, v := range t.class_obbject {
		if sym.Location.Range.Start.AfterOrEq(v.Location.Range.Start) && sym.Location.Range.End.BeforeOrEq(v.Location.Range.End) {
			sym.Kind = lsp.SymbolKindMethod
			return true
		}

	}
	return false
}

type go_const_spec struct {
	data        []*lsp.SymbolInformation
	name        string
	classsymbol lsp.SymbolInformation
	context     TreeSitterSymbol
}

func (spec *go_const_spec) Is(s *lsp.SymbolInformation, context *TreeSitterSymbol) bool {
	v := spec.data[len(spec.data)-1]
	yes := false
	if context != nil {
		yes = spec.context == *context

	}
	if yes || v.Location.Range.End.Line+1 == s.Location.Range.End.Line {
		spec.data = append(spec.data, s)
		spec.classsymbol.Location.Range.End = s.Location.Range.End
		return true
	}
	return false
}

type ts_go_symbol_resolve struct {
	class_object    []*lsp.SymbolInformation
	enum_const      []*go_const_spec
	last_enum_const *go_const_spec
}

// resolve implements ts_symbol_parser.
func (t *ts_go_symbol_resolve) resolve(line []TreeSitterSymbol, sym *lsp.SymbolInformation, code string) (done bool, err error) {
	is_const_spec := false
	for _, v := range line {
		if v.CaptureName == "item" {
			if v.Symbol == "const_spec" {
				var context *TreeSitterSymbol
				for i := range line {
					c := line[i]
					if c.CaptureName == "context" {
						context = &c
						break
					}
				}
				found := false
				is_const_spec = true
				for i := range t.class_object {
					c := t.class_object[i]
					if strings.Contains(v.Code, c.Name) {
						sym.Kind = lsp.SymbolKindEnumMember
						t.last_enum_const = &go_const_spec{
							name:        c.Name,
							data:        []*lsp.SymbolInformation{sym},
							classsymbol: *c,
						}
						classsymbol := &t.last_enum_const.classsymbol
						classsymbol.Name = classsymbol.Name+ " (const enum)" 
						classsymbol.Location.Range.Start.Line = sym.Location.Range.Start.Line - 1
						if context != nil {
							t.last_enum_const.context = *context
							classsymbol.Location.Range.Start = lsp.Position{Line: int(context.Begin.Row), Character: int(context.Begin.Column)}
						}
						classsymbol.Location.Range.End = sym.Location.Range.End
						classsymbol.Kind = lsp.SymbolKindEnum
						t.enum_const = append(t.enum_const, t.last_enum_const)
						found = true
						break
					}
				}
				if !found {
					if t.last_enum_const != nil {
						if t.last_enum_const.Is(sym, context) {
							sym.Kind = lsp.SymbolKindEnumMember
						}
					}
				}
			}
			break
		}
	}
	if !is_const_spec {
		t.last_enum_const = nil
	}
	if is_class(sym.Kind) {
		t.class_object = append(t.class_object, sym)
	}
	done = false
	if sym.Kind == lsp.SymbolKindMethod {
		method_name := ""
		class_name := ""
		for i, v := range line {
			if v.Symbol == "field_identifier" {
				method_name = v.Code
				continue
			}
			if v.Symbol == ")" {
				if i-1 > 0 && line[i-1].CaptureName == "context" {
					class_name = line[i-1].Code
				}
			}
		}
		if len(method_name) > 0 && len(class_name) > 0 {
			sym.ContainerName = class_name
			sym.Name = fmt.Sprintf("(%s).%s", class_name, method_name)
			done = true
		}
		return
	}
	return
}

func NewTreeSitter(name string, content []byte) *TreeSitter {

	ret := &TreeSitter{
		parser:   sitter.NewParser(),
		filename: NewFile(name),
		content:  content,
	}
	if lsp_lang_go.IsMe(lsp_lang_go{}, name) {
		var golang ts_go_symbol_resolve
		ret.symbol_resolve = &golang
	} else if lsp_lang_cpp.IsMe(lsp_lang_cpp{}, name) {
		var golang cpp_go_symbol_resolve
		ret.symbol_resolve = &golang
	}
	ret.HlLine = make(map[int][]TreeSitterSymbol)
	return ret
}
func (ts *TreeSitter) IsMe(filename string) bool {
	return ts.filename.filepathname == filename
}
func (ts *TreeSitter) Loadfile(lang *sitter.Language, cb func(*TreeSitter)) error {
	if err := ts._load_file(lang); err != nil {
		debug.ErrorLog("fail to load treesitter", err)
		return err
	}
	// go func() {
	ret, hlerr := ts.get_higlight(query_highlights)
	ts.HlLine = ret
	if hlerr != nil {
		debug.ErrorLog("fail to load highlights", hlerr)
	}
	if ts.tsdef.parser != nil {
		ts.tsdef.parser(ts, nil)
	}
	if _, yes := lang_ts_symbol[filepath.Ext(ts.filename.filepathname)]; yes {
		ss := NewLangTreeSitterSymbol(ts.filename.filepathname)
		lang_ts_symbol[ss.ext] = ss
	}
	if cb != nil {
		go ts.callback_to_ui(cb)
		return nil
	}
	// }()
	return nil
}

func (ts *TreeSitter) callback_to_ui(cb func(*TreeSitter)) {

	if cb != nil {
		cb(ts)
	}
}

type ts_scope struct {
	*TreeSitterSymbol
	elements []lsp.SymbolInformation
	symbol   *lsp.SymbolInformation
}

func treesitter_local(ret []Outline, ts *TreeSitter) (result []lsp.SymbolInformation) {
	prefix := "local.definition."
	symbol_kind := map[string]lsp.SymbolKind{
		prefix + "method":      lsp.SymbolKindMethod,
		prefix + "function":    lsp.SymbolKindFunction,
		prefix + "namespace":   lsp.SymbolKindNamespace,
		prefix + "field":       lsp.SymbolKindField,
		prefix + "var":         lsp.SymbolKindVariable,
		prefix + "constructor": lsp.SymbolKindConstructor,
		prefix + "type.class":  lsp.SymbolKindClass,
		prefix + "type":        lsp.SymbolKindClass,
	}
	tsscope := []ts_scope{}
	for i := range ret {
		line := ret[i]
		for i := 0; i < len(line); i++ {
			s := line[i]
			if s.CaptureName == "local.scope" {
				tsscope = append(tsscope, ts_scope{
					&TreeSitterSymbol{
						Begin: s.Begin, End: s.End,
						Code:        s.Code,
						CaptureName: s.CaptureName,
						Symbol:      s.Symbol},
					nil, nil,
				})
			} else {
				for i := range tsscope {
					v := tsscope[i]
					p := lsp.Range{
						Start: lsp.Position{Line: int(s.Begin.Row), Character: int(s.Begin.Column)},
						End:   lsp.Position{Line: int(s.End.Row), Character: int(s.End.Column)},
					}
					if v.lsprange().Overlaps(p) {
						switch s.CaptureName {
						case "local.definition.function":
							if v.Symbol == "function_definition" {
								symbol := lsp.SymbolInformation{
									Name: s.Code,
									Kind: lsp.SymbolKindFunction,
									Location: lsp.Location{
										URI:   lsp.NewDocumentURI(ts.filename.Path()),
										Range: v.lsprange(),
									},
								}
								v.symbol = &symbol
								result = append(result, symbol)
							}
						}
						// pos := fmt.Sprint(s.Begin.Row, ":", s.Begin.Column, s.End.Row, ":", s.End.Column)
						Range := s.lsprange()
						if kind, ok := symbol_kind[s.CaptureName]; ok {
							s := lsp.SymbolInformation{
								Name: s.Code,
								Kind: kind,
								Location: lsp.Location{
									URI:   lsp.NewDocumentURI(ts.filename.Path()),
									Range: Range,
								},
							}
							v.elements = append(v.elements, s)
						}
						tsscope[i] = v
						break
					}
				}
			}
		}
	}
	return
}
func get_ts_symbol(ret []Outline, ts *TreeSitter) []lsp.SymbolInformation {
	prefix := "local.definition."
	symbols := []lsp.SymbolInformation{}
	scopes := []TreeSitterSymbol{}
	for lineno := range ret {
		line := ret[lineno]
		for i := 0; i < len(line); i++ {
			s := line[i]
			pos := s.PositionInfo()
			if s.CaptureName == "local.scope" {
				if strings.Index(s.Symbol, "expression") > 0 {
					continue
				}
				switch s.Symbol {
				case "method_declaration", "function_definition", "if_expression", "function_item", "closure_expression", "block":
					{
						scopes = append(scopes, s)
					}
				default:
					debug.TraceLog("=====================", s.CaptureName, s.Symbol, pos)
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
			if strings.Index(s.CaptureName, prefix) == 0 {
				symboltype := strings.Replace(s.CaptureName, prefix, "", 1)
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
					debug.TraceLog("outline", s.Code, symboltype, pos)
					add := true
					switch kind {
					case lsp.SymbolKindVariable:
						{
							add = in_scope_range(scopes, Range, add)
						}
					}
					if !add {
						debug.DebugLog("unhandled skip symboltype:", symboltype, s.Code, pos, s.Symbol)
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
					debug.DebugLog("unhandled-symboltype:", symboltype, s.Code, pos, s.Symbol)
				}
			} else if s.CaptureName == "local.scope" {
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

func in_scope_range(scopes []TreeSitterSymbol, Range lsp.Range, add bool) bool {
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
	tree := ts.parser.ParseCtx(context.Background(), ts.sourceCode, nil)
	if tree == nil {
		return fmt.Errorf("tree is nil")
	}
	ts.tree = tree
	return nil
}
