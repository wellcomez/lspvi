package lspcore

import (
	"fmt"
	"log"
	"strings"

	"github.com/tectiv3/go-lsp"
)

var FolderEmoji = "\U0001f4c1"
var FileIcon = "\U0001f4c4"
var Text = "󰉿"
var Method = "ƒ"
var Function = ""
var Constructor = ""
var Field = "󰜢"
var Variable = "󰀫"
var Class = "𝓒"
var Interface = ""
var Module = ""
var Property = "󰜢"
var Unit = "󰑭"
var Value = "󰎠"
var Enum = ""
var Keyword = "󰌋"
var Snippet = ""
var Color = "󰏘"
var File = "󰈙"
var Reference = "󰈇"
var Folder = "󰉋"
var EnumMember = ""
var Constant = "󰏿"
var Struct = "𝓢"
var Event = ""
var Operator = "󰆕"
var TypeParameter = ""

type Symbol struct {
	SymInfo   lsp.SymbolInformation
	Members   []Symbol
	classname string
}

func inside_location(bigger lsp.Location, smaller lsp.Location) bool {
	if smaller.Range.Start.Line < bigger.Range.Start.Line {
		return false
	}
	if smaller.Range.Start.Line == bigger.Range.Start.Line {
		if smaller.Range.Start.Character < bigger.Range.Start.Character {
			return false
		}
	}
	if smaller.Range.End.Line == bigger.Range.End.Line {
		if smaller.Range.End.Character > bigger.Range.End.Character {
			return false
		}
	}
	if smaller.Range.End.Line > bigger.Range.End.Line {
		return false
	}
	return true

}
func (S Symbol) match(calr *CallStackEntry) bool {
	loc := lsp.Location{
		URI:   calr.Item.URI,
		Range: calr.Item.Range,
	}
	yes := strings.Contains(S.SymInfo.Name, calr.Name)
	irange := inside_location(S.SymInfo.Location, loc)
	if yes {

		log.Printf("xxx", irange, S.SymInfo.Kind, calr.Item.Kind, calr.Name)
	}
	if S.SymInfo.Kind == lsp.SymbolKindMethod && calr.Item.Kind == lsp.SymbolKindFunction {
		calr.Item.Kind = lsp.SymbolKindMethod
	}
	if S.SymInfo.Kind == calr.Item.Kind {
		if irange {
			return true
		}
		if yes {
			log.Printf("Error Resovle failed %s %s \n>>>%s  \n>>>>%s", S.SymInfo.Name, calr.DisplayName(),
				NewBody(S.SymInfo.Location).Info(),
				NewBody(loc).Info())
		}
	}
	if irange {
		log.Printf("Error kind unmatch %s %s \n>>>%s  \n>>>>%s", S.SymInfo.Name, calr.DisplayName(),
			NewBody(S.SymInfo.Location).Info(),
			NewBody(loc).Info())
	}
	return false
}

func (s Symbol) Icon() string {
	switch s.SymInfo.Kind {
	case lsp.SymbolKindMethod:
		return Method
	case lsp.SymbolKindField:
		return Field
	case lsp.SymbolKindClass:
		return Class
	case lsp.SymbolKindFunction:
		return Function
	case lsp.SymbolKindConstructor:
		return Constructor
	case lsp.SymbolKindInterface:
		return Interface
	case lsp.SymbolKindVariable:
		return Variable
	case lsp.SymbolKindConstant:
		return Constant
	case lsp.SymbolKindEnum:
		return Enum
	case lsp.SymbolKindEnumMember:
		return EnumMember
	case lsp.SymbolKindOperator:
		return Operator
	case lsp.SymbolKindTypeParameter:
		return TypeParameter
	default:
		return ""
	}
}

func (sym Symbol) SymbolListStrint() string {
	return sym.Icon() + " " + sym.SymInfo.Name
}
func (sym Symbol) Is_class() bool {
	return is_class(sym.SymInfo.Kind)
}

func is_class(kind lsp.SymbolKind) bool {
	return kind == lsp.SymbolKindClass || kind == lsp.SymbolKindStruct
}
func is_memeber(kind lsp.SymbolKind) bool {
	return kind == lsp.SymbolKindMethod || kind == lsp.SymbolKindField || kind == lsp.SymbolKindConstructor
}

func (sym Symbol) contain(a Symbol) bool {
	return symbol_contain(sym.SymInfo, a.SymInfo)
}
func symbol_contain(a lsp.SymbolInformation, b lsp.SymbolInformation) bool {
	if a.Location.Range.End.Line > b.Location.Range.End.Line {
		return true
	}
	if a.Location.Range.End.Line == b.Location.Range.End.Line {
		if a.Location.Range.End.Character > b.Location.Range.End.Character {
			return true
		}
	}
	return false
}

type LspWorkspace struct {
	clients []lspclient
	Wk      WorkSpace
	Current *Symbol_file
	filemap map[string]*Symbol_file
	Handle  lsp_data_changed
}

func (wk LspWorkspace) find_from_stackentry(entry *CallStackEntry) (*Symbol, error) {
	filename := entry.Item.URI.AsPath().String()
	symbolfile, isnew, err := wk.open(filename)
	if err != nil {
		return nil, err
	}
	if isnew {
		symbolfile.__load_symbol_impl()
	}
	if symbolfile == nil {
		log.Printf("fail to loadd  %s\n", filename)
		return nil, fmt.Errorf("fail to loadd %s", filename)
	}
	return symbolfile.find_stack_symbol(entry)

}
func (wk LspWorkspace) Close() {
	for _, v := range wk.clients {
		v.Close()
	}
}
func (wk LspWorkspace) getClient(filename string) lspclient {
	for _, c := range wk.clients {
		ret := wk.new_client(c, filename)
		if ret != nil {
			return ret
		}

	}
	return nil
}

func (wk LspWorkspace) new_client(c lspclient, filename string) lspclient {
	if !c.IsMe(filename) {
		return nil
	}
	err := c.Launch_Lsp_Server()
	if err == nil {
		err = c.InitializeLsp(wk.Wk)
		if err == nil {
			return c
		}
	}
	return nil
}

func (wk *LspWorkspace) open(filename string) (*Symbol_file, bool, error) {
	val, ok := wk.filemap[filename]
	is_new := false
	if ok {
		wk.Current = val
		return val, is_new, nil
	}
	wk.filemap[filename] = &Symbol_file{
		Filename: filename,
		lsp:      wk.getClient(filename),
		Handle:   wk.Handle,
		Wk:       wk,
	}

	ret := wk.filemap[filename]
	if ret.lsp == nil {
		return nil, is_new, fmt.Errorf("fail to open %s", filename)
	}
	is_new = true
	err := ret.lsp.DidOpen(filename)
	if err != nil {
		return ret, is_new, err
	}
	token, err := ret.lsp.Semantictokens_full(filename)
	if err == nil {
		ret.tokens = token
	}
	return ret, is_new, err
}
func (wk *LspWorkspace) Open(filename string) (*Symbol_file, error) {
	ret, _, err := wk.open(filename)
	wk.Current = wk.filemap[filename]
	return ret, err

}
func NewLspWk(wk WorkSpace) *LspWorkspace {
	cpp := lsp_base{
		wk:   &wk,
		core: &lspcore{lang: lsp_lang_cpp{}, handle: wk.Callback, LanguageID: string(CPP)},
	}
	py := lsp_base{
		wk:   &wk,
		core: &lspcore{lang: lsp_lang_py{}, handle: wk.Callback, LanguageID: string(PYTHON)},
	}

	golang := lsp_base{wk: &wk, core: &lspcore{lang: lsp_lang_go{}, handle: wk.Callback, LanguageID: string(GO)}}
	ret := &LspWorkspace{
		clients: []lspclient{
			cpp, py, golang,
		},
		Wk: wk,
	}
	ret.filemap = make(map[string]*Symbol_file)
	return ret
}

type lsp_data_changed interface {
	OnSymbolistChanged(file *Symbol_file, err error)
	OnCodeViewChanged(file *Symbol_file)
	OnLspRefenceChanged(ranges lsp.Range, file []lsp.Location)
	OnFileChange(file []lsp.Location)
	OnLspCaller(search string, stacks []CallStack)
	OnLspCallTaskInViewChanged(stacks *CallInTask)
	OnCallTaskInViewResovled(stacks *CallInTask)
}
