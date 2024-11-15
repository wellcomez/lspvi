package lspcore

import (
	"fmt"
	// "log"
	"os"
	"strings"
	"sync"

	"github.com/tectiv3/go-lsp"
	"gopkg.in/yaml.v2"
	"zen108.com/lspvi/pkg/debug"
)

var nf_cod_symbol_array = ''    // =\uea8a
var nf_cod_symbol_boolean = ''  // =\uea8f
var nf_cod_symbol_color = ''    // =\ueb5c
var nf_cod_symbol_constant = '' //=\ueb5d
// var      nf_cod_symbol_enum  =\uea95
var nf_cod_symbol_event = '' // =\uea86
var nf_cod_symbol_field = '' //=\ueb5f

var nf_cod_symbol_file = ''      //=\ueb60
var nf_cod_symbol_key = ''       // =\uea93
var nf_cod_symbol_misc = ''      // =\ueb63
var nf_cod_symbol_numeric = ''   //=\uea90
var nf_cod_symbol_parameter = '' // =\uea92
var nf_cod_symbol_property = ''  // =\ueb65
var nf_cod_symbol_ruler = ''     // =\uea96
var nf_cod_symbol_string = ''    // =\ueb8d

var nf_cod_symbol_enum_member = '' //=\ueb5e

var nf_cod_symbol_class = '' // =\ueb5b
var nf_cod_symbol_interface = ''
var nf_cod_symbol_variable = ''  //=\uea88
var nf_cod_symbol_method = ''    // =\uea8c
var nf_cod_symbol_snippet = ''   // =\ueb66
var nf_cod_symbol_structure = '' // =\uea91
var nf_cod_symbol_namespace = '' // =\uea8b
var nf_cod_symbol_enum = ''      // =\uea95
var nf_cod_symbol_keyword = ''   // =\ueb62
var nf_cod_symbol_operator = ''  // =\ueb64
var nf_cod_folder = ''           //=\uea83
// var FolderEmoji = "\U0001f4c1"
var FileIcon = "\U0001f4c4"
var Text = '󰉿'

// var Method = "ƒ"
var Method = nf_cod_symbol_method
var Function = ''
var Constructor = ""
var Field = "󰜢"
var Namespace = nf_cod_symbol_namespace

// var Variable = "󰀫"
// var Class = "𝓒"
// var  Interface =  '󰜰'
// var Enum = ""
// var Keyword = '󰌋'
// var Struct = '𝓢'
// var Operator = "󰆕"
// var Snippet = ''
// var Color = '󰏘'
// var Folder = '󰉋'
var Color = nf_cod_symbol_color
var Variable = nf_cod_symbol_variable
var Class = nf_cod_symbol_class

var Interface = nf_cod_symbol_interface

// var Interface = '\ueb61'
var Module = ''
var Property = '󰜢'
var Unit = '󰑭'
var Value = '󰎠'
var Enum = nf_cod_symbol_enum
var Keyword = nf_cod_symbol_keyword
var Snippet = nf_cod_symbol_snippet
var File = '󰈙'
var Reference = '󰈇'
var Folder = nf_cod_folder
var EnumMember = ''
var Constant = '󰏿'
var Struct = nf_cod_symbol_structure
var Event = ''
var Operator = nf_cod_symbol_operator

// var TypeParameter = '󰊄'
// var TypeParameter = '\uea92'
var TypeParameter = nf_cod_symbol_parameter
var IconsRunne = map[int]rune{
	1: File,      //-- File
	2: Module,    // -- Module
	3: Namespace, // -- Namespace
	4: '',       // -- Package
	5: Class,     //-- Class
	//5:   "󰌗 ", //-- Class
	6: Method, //-- Method
	//6:  Method,
	7:  nf_cod_symbol_property, //-- Property
	8:  '',                    //-- Field
	9:  '',                    //-- Constructor
	10: '󰕘',                    //-- Enum
	//11: "󰕘 ", //-- Interface
	11: Interface,
	12: Function,              //-- Function
	13: Variable,              //-- Variable
	14: Constant,              //-- Constant
	15: nf_cod_symbol_string,  //-- String
	16: '󰎠',                   //-- Number
	17: nf_cod_symbol_boolean, //-- Boolean
	18: nf_cod_symbol_array,   //-- Array
	19: '󰅩',                   //-- Object
	20: '󰌋',                   //-- Key
	21: '󰟢',                   //-- Null
	22: EnumMember,            //-- EnumMember
	// 23:  '󰌗', //-- Struct
	23:  Struct,        //-- Struct
	24:  Event,         //-- Event
	25:  '󰆕',           //-- Operator
	26:  TypeParameter, //-- TypeParameter
	255: '󰉨',           //-- Macro
}

func icon_run_to_string() (ret map[int]string) {
	ret = make(map[int]string)
	for k, v := range IconsRunne {
		ret[k] = fmt.Sprintf("%c", v)
	}
	return
}

var LspIcon = icon_run_to_string()

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

		// log.Println(,"xxx", irange, S.SymInfo.Kind, calr.Item.Kind, calr.Name)
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

			debug.ErrorLogf(DebugTag, "Error Resovle failed %s %s \n>>>%s  \n>>>>%s", S.SymInfo.Name, calr.DisplayName(),
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
		debug.ErrorLogf(DebugTag, "Error kind unmatch %s %s \n>>>%s  \n>>>>%s", S.SymInfo.Name, calr.DisplayName(), info, info2)
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
	case lsp.SymbolKindMethod, lsp.SymbolKindField, lsp.SymbolKindConstructor, lsp.SymbolKindEnumMember, lsp.SymbolKindProperty:
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
	symbolfile, ok := wk.get(filename)
	if !ok {
		symbolfile = wk.OpenNoLsp(filename)
		symbolfile.LoadTreeSitter(true)
	}
	if sym, _ := symbolfile.find_stack_symbol(entry); sym != nil {
		return sym, nil
	}

	symbolfile, isnew, err := wk.open(filename)
	if err != nil {
		return nil, err
	}
	if isnew {
		symbolfile.__load_symbol_impl(false)
	}
	if symbolfile == nil {
		debug.ErrorLogf(DebugTag, "fail to loadd  %s\n", filename)
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
		if ret, err := wk.new_client(c, filename); ret != nil {
			return ret
		} else if err != nil {
			debug.ErrorLog(DebugTag, "getClient failed", err)
		}
	}
	return nil
}

func (wk LspWorkspace) new_client(c lspclient, filename string) (ret lspclient, err error) {
	if !c.IsMe(filename) {
		// err = fmt.Errorf("not match")
		return
	}
	if c.IsReady() {
		ret = c
		return
	}
	err = c.Launch_Lsp_Server()
	if err == nil {
		err = c.InitializeLsp(wk.Wk)
		if err == nil {
			ret = c
		} else {
			ret = c
		}
	}
	return
}

func (wk *LspWorkspace) open(filename string) (*Symbol_file, bool, error) {
	return wk.openbuffer(filename, "")
}
func (wk *LspWorkspace) openbuffer(filename string, content string) (*Symbol_file, bool, error) {
	is_new := false
	ret := wk.OpenNoLsp(filename)
	if ret.lspopen {
		wk.current = ret
		return ret, is_new, nil
	}
	if ret.lsp == nil {
		if lsp := wk.getClient(filename); lsp != nil {
			ret.lsp = lsp
		} else {
			err := fmt.Errorf("fail to open %s lsp is null", filename)
			debug.WarnLog(DebugTag, "openbuffer ", err)
			return nil, is_new, err
		}
	}
	is_new = true
	if err := ret.lsp.DidOpen(SourceCode{filename, content}, ret.verison); err == nil {
		ret.lspopen = true
	} else {
		return ret, is_new, err
	}
	if token, err := ret.lsp.Semantictokens_full(filename); err == nil {
		ret.tokens = token
	} else {
		return ret, is_new, err
	}
	return ret, is_new, nil
}

func (wk *LspWorkspace) OpenNoLsp(filename string) (ret *Symbol_file) {
	ret, _ = wk.get(filename)
	if ret == nil {
		ret = &Symbol_file{
			Filename: filename,
			lsp:      wk.getClient(filename),
			Handle:   wk.Handle,
			Wk:       wk,
		}
		wk.set(filename, ret)
	}
	return ret
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
			Range: s.SymInfo.Location.Range,
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
func (wk *LspWorkspace) CallHierarchyIncomingCalls(param lsp.CallHierarchyItem) (ret []lsp.CallHierarchyIncomingCall, err error) {
	filename := param.URI.AsPath().String()
	if lsp := wk.getClient(filename); lsp != nil {
		return lsp.CallHierarchyIncomingCalls(param)
	}
	return nil, err
}
func (wk *LspWorkspace) PrepareCallHierarchy(loc lsp.Location) (ret []lsp.CallHierarchyItem, err error) {
	filename := loc.URI.AsPath().String()
	if lsp := wk.getClient(filename); lsp != nil {
		return lsp.PrepareCallHierarchy(loc)
	}
	return nil, err
}
func (wk *LspWorkspace) GetReference(loc lsp.Location) (ret []lsp.Location, err error) {
	filename := loc.URI.AsPath().String()
	if lsp := wk.getClient(filename); lsp != nil {
		return lsp.GetReferences(filename, loc.Range.Start)
	}
	return nil, err
}

func (wk *LspWorkspace) Open(filename string) (*Symbol_file, error) {
	ret, _, err := wk.open(filename)
	wk.current = ret
	debug.InfoLog(DebugTag, "open", filename, "len", len(wk.filemap))
	return ret, err

}

type LangConfig struct {
	Cmd string `yaml:"cmd"`
	Log string `yaml:"log"`
}

type ConfigLspPart struct {
	Lsp LspConfig `yaml:"lsp"`
}
type LspConfig struct {
	C          LangConfig `yaml:"c"`
	Golang     LangConfig `yaml:"go"`
	Py         LangConfig `yaml:"py"`
	Java       LangConfig `yaml:"java"`
	Javascript LangConfig `yaml:"javascript"`
	Typescript LangConfig `yaml:"typescript"`
}

func NewLspWk(wk WorkSpace) *LspWorkspace {
	buf, lsp_config_err := os.ReadFile(wk.ConfigFile)
	var lsp_config LspConfig
	if lsp_config_err == nil {
		var config ConfigLspPart
		yaml.Unmarshal(buf, &config)
		lsp_config = config.Lsp
	}

	cpp := create_lang_lsp(wk, CPP, lsp_lang_cpp{}, lsp_config.C)
	golang := create_lang_lsp(wk, GO, lsp_lang_go{}, lsp_config.Golang)
	rs := create_lang_lsp(wk, RUST, lsp_lang_rs{}, lsp_config.Golang)
	py := create_lang_lsp(wk, PYTHON, lsp_lang_py{}, lsp_config.Py)
	java := create_lang_lsp(wk, JAVA, lsp_lang_jedi{}, lsp_config.Java)
	ts := create_lang_lsp(wk, TYPE_SCRIPT, lsp_ts{LanguageID: string(TYPE_SCRIPT)}, lsp_config.Typescript)
	js := create_lang_lsp(wk, JAVASCRIPT, lsp_ts{LanguageID: string(JAVASCRIPT)}, lsp_config.Javascript)
	ret := &LspWorkspace{
		clients: []lspclient{
			cpp, py, golang, ts, js, rs, java,
		},
		Wk:              wk,
		lock_symbol_map: &sync.Mutex{},
	}
	ret.filemap = make(map[string]*Symbol_file)
	return ret
}

func create_lang_lsp(wk WorkSpace, lang LanguageIdentifier, l lsplang, config LangConfig) lsp_base {
	cpp := lsp_base{
		wk: &wk,
		core: &lspcore{
			lsp_stderr: lsp_server_errorlog{lsp_log: wk.Callback, lang: string(lang)},
			lang:       l,
			config:     config,
			handle:     wk,
			LanguageID: string(lang)},
	}
	return cpp
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
