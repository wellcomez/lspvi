package lspcore

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/tectiv3/go-lsp"
	"gopkg.in/yaml.v2"
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
var IconsRunne = map[int]rune{
	1: '󰈙', //-- File
	2: '', // -- Module
	3: '󰌗', // -- Namespace
	4: '', // -- Package
	5: '𝓒', //-- Class
	//5:   "󰌗 ", //-- Class
	6: '󰆧', //-- Method
	//6:  Method,
	7:  '', //-- Property
	8:  '', //-- Field
	9:  '', //-- Constructor
	10: '󰕘', //-- Enum
	//11: "󰕘 ", //-- Interface
	//11: '"' ,
	12: '󰊕', //-- Function
	13: '󰆧', //-- Variable
	14: '󰏿', //-- Constant
	15: '󰀬', //-- String
	16: '󰎠', //-- Number
	17: '◩', //-- Boolean
	18: '󰅪', //-- Array
	19: '󰅩', //-- Object
	20: '󰌋', //-- Key
	21: '󰟢', //-- Null
	//22: ' ', //-- EnumMember
	//23:  "󰌗 ", //-- Struct
	23:  '𝓢', //-- Struct
	24:  '', //-- Event
	25:  '󰆕', //-- Operator
	26:  '󰊄', //-- TypeParameter
	255: '󰉨', //-- Macro
}
var LspIcon = map[int]string{
	1: "󰈙 ",  //-- File
	2: " ",  // -- Module
	3: "󰌗 ",  // -- Namespace
	4: " ",  // -- Package
	5: Class, //-- Class
	//5:   "󰌗 ", //-- Class
	6: "󰆧 ", //-- Method
	//6:  Method,
	7:  " ", //-- Property
	8:  " ", //-- Field
	9:  " ", //-- Constructor
	10: "󰕘 ", //-- Enum
	//11: "󰕘 ", //-- Interface
	11: Interface,
	12: "󰊕 ", //-- Function
	13: "󰆧 ", //-- Variable
	14: "󰏿 ", //-- Constant
	15: "󰀬 ", //-- String
	16: "󰎠 ", //-- Number
	17: "◩ ", //-- Boolean
	18: "󰅪 ", //-- Array
	19: "󰅩 ", //-- Object
	20: "󰌋 ", //-- Key
	21: "󰟢 ", //-- Null
	22: " ", //-- EnumMember
	//23:  "󰌗 ", //-- Struct
	23:  Struct, //-- Struct
	24:  " ",   //-- Event
	25:  "󰆕 ",   //-- Operator
	26:  "󰊄 ",   //-- TypeParameter
	255: "󰉨 ",   //-- Macro
}

type Symbol struct {
	SymInfo   lsp.SymbolInformation
	Members   []Symbol
	Classname string
}

func inside_location(bigger lsp.Location, smaller lsp.Location) bool {
	return smaller.Range.Overlaps(bigger.Range)
	// if smaller.Range.Start.Line < bigger.Range.Start.Line {
	// 	return false
	// }
	// if smaller.Range.Start.Line == bigger.Range.Start.Line {
	// 	if smaller.Range.Start.Character < bigger.Range.Start.Character {
	// 		return false
	// 	}
	// }
	// if smaller.Range.End.Line == bigger.Range.End.Line {
	// 	if smaller.Range.End.Character > bigger.Range.End.Character {
	// 		return false
	// 	}
	// }
	// if smaller.Range.End.Line > bigger.Range.End.Line {
	// 	return false
	// }
	// return true

}
func (S Symbol) match(calr *CallStackEntry) bool {
	loc := lsp.Location{
		URI:   calr.Item.URI,
		Range: calr.Item.Range,
	}
	yes := strings.Contains(S.SymInfo.Name, calr.Name)
	irange := inside_location(S.SymInfo.Location, loc)
	if yes {

		log.Println("xxx", irange, S.SymInfo.Kind, calr.Item.Kind, calr.Name)
	}
	if S.SymInfo.Kind == lsp.SymbolKindMethod && calr.Item.Kind == lsp.SymbolKindFunction {
		calr.Item.Kind = lsp.SymbolKindMethod
	}
	if S.SymInfo.Kind == calr.Item.Kind {
		if irange {
			return true
		}
		if yes {
			info := ""
			if b1, err := NewBody(S.SymInfo.Location); err == nil {
				info = b1.Info()
			}
			info2 := ""
			if b1, err := NewBody(loc); err == nil {
				info2 = b1.Info()
			}

			log.Printf("Error Resovle failed %s %s \n>>>%s  \n>>>>%s", S.SymInfo.Name, calr.DisplayName(),
				info, info2)
		}
	}
	if irange {
		info := ""
		if b, err := NewBody(S.SymInfo.Location); err == nil {
			info = b.Info()
		}
		info2 := ""
		if b, err := NewBody(loc); err == nil {
			info2 = b.Info()
		}
		log.Printf("Error kind unmatch %s %s \n>>>%s  \n>>>>%s", S.SymInfo.Name, calr.DisplayName(), info, info2)
	}
	return false
}

func (s Symbol) Icon() string {
	if v, ok := IconsRunne[int(s.SymInfo.Kind)]; ok {
		return fmt.Sprintf("%c", v)
	}
	if v, ok := LspIcon[int(s.SymInfo.Kind)]; ok {
		return v
	}
	return ""
}

func (sym Symbol) SymbolListStrint() string {
	return sym.Icon() + " " + sym.SymInfo.Name
}
func (sym Symbol) Is_class() bool {
	return is_class(sym.SymInfo.Kind)
}

func is_class(kind lsp.SymbolKind) bool {
	switch kind {
	case lsp.SymbolKindClass, lsp.SymbolKindStruct, lsp.SymbolKindInterface, lsp.SymbolKindEnum:
		return true
	}
	return false
}
func is_memeber(kind lsp.SymbolKind) bool {
	switch kind {
	case lsp.SymbolKindMethod, lsp.SymbolKindField, lsp.SymbolKindConstructor, lsp.SymbolKindEnumMember:
		return true
	}
	return false
}

func (sym Symbol) contain(a Symbol) bool {
	return symbol_contain(sym.SymInfo, a.SymInfo)
}
func symbol_contain(a lsp.SymbolInformation, b lsp.SymbolInformation) bool {
	return b.Location.Range.Overlaps(a.Location.Range)
	// if a.Location.Range.End.Line > b.Location.Range.End.Line {
	// return true
	// }
	// if a.Location.Range.End.Line == b.Location.Range.End.Line {
	// 	if a.Location.Range.End.Character > b.Location.Range.End.Character {
	// 		return true
	// 	}
	// }
	// return false
}

type LspWorkspace struct {
	clients         []lspclient
	Wk              WorkSpace
	current         *Symbol_file
	filemap         map[string]*Symbol_file
	Handle          lsp_data_changed
	lock_symbol_map *sync.Mutex
}

func (wk LspWorkspace) IsSource(filename string) bool {
	for _, v := range wk.clients {
		if v.IsMe(filename) {
			return true
		}
	}
	return false
}
func (wk LspWorkspace) find_from_stackentry(entry *CallStackEntry) (*Symbol, error) {
	filename := entry.Item.URI.AsPath().String()
	symbolfile, isnew, err := wk.open(filename)
	if err != nil {
		return nil, err
	}
	if isnew {
		symbolfile.__load_symbol_impl(false)
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
	if c.IsReady() {
		return c
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
	return wk.openbuffer(filename, "")
}
func (wk *LspWorkspace) openbuffer(filename string, content string) (*Symbol_file, bool, error) {
	val, ok := wk.get(filename)
	is_new := false
	if ok {
		wk.current = val
		return val, is_new, nil
	}
	ret := &Symbol_file{
		Filename: filename,
		lsp:      wk.getClient(filename),
		Handle:   wk.Handle,
		Wk:       wk,
	}
	wk.set(filename, ret)
	if ret.lsp == nil {
		return nil, is_new, fmt.Errorf("fail to open %s", filename)
	}
	is_new = true
	err := ret.lsp.DidOpen(SourceCode{filename, content}, ret.verison)
	if err != nil {
		return ret, is_new, err
	}
	token, err := ret.lsp.Semantictokens_full(filename)
	if err == nil {
		ret.tokens = token
	}
	return ret, is_new, err
}

func (wk *LspWorkspace) set(filename string, ret *Symbol_file) {
	wk.lock_symbol_map.Lock()
	defer wk.lock_symbol_map.Unlock()
	wk.filemap[filename] = ret
}

func (wk *LspWorkspace) get(filename string) (*Symbol_file, bool) {
	wk.lock_symbol_map.Lock()
	defer wk.lock_symbol_map.Unlock()
	val, ok := wk.filemap[filename]
	return val, ok
}
func (wk *LspWorkspace) GetCallEntry(filename string, r lsp.Range) (*CallStackEntry, *Symbol_file) {
	sym, _ := wk.Get(filename)
	if sym == nil {
		return nil, nil
	}
	s := sym.Find(r)
	if s == nil {
		return nil, sym
	}
	return &CallStackEntry{
		Item: lsp.CallHierarchyItem{
			Name:  s.SymInfo.Name,
			Range: r,
			URI:   s.SymInfo.Location.URI,
			Kind:  s.SymInfo.Kind,
		},
		Name:      s.SymInfo.Name,
		ClassName: s.Classname,
	}, sym
}
func (wk *LspWorkspace) Get(filename string) (*Symbol_file, error) {
	ret, ok := wk.get(filename)
	if ok {
		return ret, nil
	}
	return ret, fmt.Errorf("not loaded")

}
func (wk *LspWorkspace) CloseSymbolFile(sym *Symbol_file) error {
	err := sym.lsp.DidClose(sym.Filename)
	wk.lock_symbol_map.Lock()
	delete(wk.filemap, sym.Filename)
	wk.lock_symbol_map.Unlock()
	return err
}

func (wk *LspWorkspace) OpenBuffer(filename string, buffer string) (*Symbol_file, error) {
	ret, _, err := wk.openbuffer(filename, buffer)
	wk.current = ret
	return ret, err

}

func (wk *LspWorkspace) Open(filename string) (*Symbol_file, error) {
	ret, _, err := wk.open(filename)
	wk.current = ret
	return ret, err

}

type LangConfig struct {
	Cmd string `yaml:"cmd"`
}

type ConfigLspPart struct {
	Lsp LspConfig `yaml:"lsp"`
}
type LspConfig struct {
	C          LangConfig `yaml:"c"`
	Golang     LangConfig `yaml:"go"`
	Py         LangConfig `yaml:"py"`
	Javascript LangConfig `yaml:"javascript"`
	Typescript LangConfig `yaml:"typescript"`
}

func (c LangConfig) is_cmd_ok() bool {
	_, err := os.Stat(c.Cmd)
	return err == nil
}

func NewLspWk(wk WorkSpace) *LspWorkspace {
	buf, lsp_config_err := os.ReadFile(wk.ConfigFile)
	var lsp_config LspConfig
	if lsp_config_err == nil {
		var config ConfigLspPart
		yaml.Unmarshal(buf, &config)
		lsp_config = config.Lsp
	}
	cpp := lsp_base{
		wk:   &wk,
		core: &lspcore{lang: lsp_lang_cpp{lsp_config.C}, handle: wk, LanguageID: string(CPP)},
	}
	py := lsp_base{
		wk:   &wk,
		core: &lspcore{lang: lsp_lang_py{lsp_config.Py}, handle: wk, LanguageID: string(PYTHON)},
	}

	golang := lsp_base{wk: &wk, core: &lspcore{lang: lsp_lang_go{lsp_config.Golang}, handle: wk, LanguageID: string(GO)}}

	ts := lsp_base{wk: &wk, core: &lspcore{lang: lsp_ts{LanguageID: string(TYPE_SCRIPT), config: lsp_config.Javascript}, handle: wk, LanguageID: string(TYPE_SCRIPT)}}
	js := lsp_base{wk: &wk, core: &lspcore{lang: lsp_ts{LanguageID: string(JAVASCRIPT), config: lsp_config.Typescript}, handle: wk, LanguageID: string(JAVASCRIPT)}}
	ret := &LspWorkspace{
		clients: []lspclient{
			cpp, py, golang, ts, js,
		},
		Wk:              wk,
		lock_symbol_map: &sync.Mutex{},
	}
	ret.filemap = make(map[string]*Symbol_file)
	return ret
}

type SymolSearchKey struct {
	File   string
	Ranges lsp.Range
	Key    string
	sym    *Symbol_file
}

func (key SymolSearchKey) Symbol() *Symbol_file {
	return key.sym
}

type lsp_data_changed interface {
	OnSymbolistChanged(file *Symbol_file, err error)
	OnCodeViewChanged(file *Symbol_file)
	OnLspRefenceChanged(ranges SymolSearchKey, file []lsp.Location, err error)
	OnGetImplement(SymolSearchKey, ImplementationResult, error, *OpenOption)
	OnFileChange(file []lsp.Location, line *OpenOption)
	OnLspCaller(search string, c lsp.CallHierarchyItem, stacks []CallStack)
	OnLspCallTaskInViewChanged(stacks *CallInTask)
	OnLspCallTaskInViewResovled(stacks *CallInTask)
}
