package lspcore

import (
	"fmt"
	"log"
	"strings"

	"github.com/tectiv3/go-lsp"
)

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
	if S.SymInfo.Kind == calr.Item.Kind {
		if inside_location(S.SymInfo.Location, loc) {
			return true
		}
		yes := strings.Contains(S.SymInfo.Name, calr.Name)
		if yes {
			log.Printf("Error Resovle failed %s %s \n>>>%s  \n>>>>%s", S.SymInfo.Name, calr.DisplayName(),
				NewBody(S.SymInfo.Location).Info(),
				NewBody(loc).Info())
		}
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

type Symbol_file struct {
	lsp          lspclient
	Filename     string
	Handle       lsp_data_changed
	Class_object []*Symbol
	Wk           *LspWorkspace
}

func (sym Symbol_file) find_stack_symbol(call *CallStackEntry) (*Symbol, error) {
	for _, v := range sym.Class_object {
		if len(v.Members) > 0 {
			for _, v := range v.Members {
				if v.match(call) {
					return &v, nil
				}
			}
		}
		if v.match(call) {
			return v, nil
		}
	}
	return nil, nil
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
func (sym *Symbol_file) LoadSymbol() {
	sym.__load_symbol_impl()
	sym.Handle.OnSymbolistChanged(*sym)
}

func (sym *Symbol_file) __load_symbol_impl() error {
	if sym.lsp == nil {
		return fmt.Errorf("lsp is nil")
	}
	if len(sym.Class_object) > 0 {
		sym.Handle.OnSymbolistChanged(*sym)
		return nil
	}
	symbols, err := sym.lsp.GetDocumentSymbol(sym.Filename)
	if err != nil {
		return err
	}
	sym.build_class_symbol(symbols.SymbolInformation, 0, nil)
	return nil
}
func (sym *Symbol_file) CallinTask(loc lsp.Location) (*CallInTask, error) {
	task := NewCallInTask(loc, sym.lsp)
	task.run()
	sym.Handle.OnCallTaskInViewChanged(task)
	return task, nil
}

func (sym *Symbol_file) Async_resolve_stacksymbol(task *CallInTask, hanlde func()) {
	bin, binerr := NewPlanUmlBin()
	export_root, export_err := NewExportRoot(&sym.Wk.Wk)
	for _, s := range task.Allstack {
		var xx = class_resolve_task{
			wklsp:     sym.Wk,
			callstack: s,
		}
		xx.Run()
		if hanlde != nil {
			name := ""
			if len(s.Items) > 0 {
				name = s.Items[0].Name
			}
			if binerr == nil && export_err == nil && len(name) > 0 {
				content := s.Uml(true)
				export_root.SaveMD(task.Dir(), name, content)
				content = s.Uml(false)
				fileuml, err := export_root.SavePlanUml(task.Dir(), name, content)
				if err == nil {
					err = bin.Convert(fileuml)
					if err != nil {
						log.Println(err)
					}
				} else {
					log.Println(err)
				}
			}
			hanlde()
		}
	}
}
func (sym *Symbol_file) Callin(loc lsp.Location) ([]CallStack, error) {
	var ret []CallStack
	if sym.lsp == nil {
		return ret, fmt.Errorf("lsp is null")
	}
	c1, err := sym.lsp.PrepareCallHierarchy(loc)
	if err != nil {
		return ret, err
	}
	var stack CallStack
	if len(c1) > 0 {
		stack.Add(NewCallStackEntry(c1[0]))
		c2, err := sym.lsp.CallHierarchyIncomingCalls(c1[0])
		if err == nil && len(c2) > 0 {
			stack.Add(NewCallStackEntry(c2[0].From))
		}
	}
	ret = append(ret, stack)
	sym.Handle.OnCallInViewChanged(ret)
	return ret, nil
}
func (sym *Symbol_file) Reference(ranges lsp.Range) {
	if sym.lsp == nil {
		return
	}
	loc, err := sym.lsp.GetReferences(sym.Filename, ranges.Start)
	if err != nil {
		return
	}
	sym.Handle.OnRefenceChanged(loc)
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
func (sym *Symbol_file) build_class_symbol(symbols []lsp.SymbolInformation, begin int, parent *Symbol) int {
	var i = begin
	for i = begin; i < len(symbols); {
		v := symbols[i]
		s := Symbol{
			SymInfo: v,
		}
		if is_class(v.Kind) {
			sym.Class_object = append(sym.Class_object, &s)

			i = sym.build_class_symbol(symbols, i+1, &s)
			continue
		}
		if parent != nil {
			if parent.contain(s) {
				if is_memeber(v.Kind) {
					s.classname = parent.SymInfo.Name
					parent.Members = append(parent.Members, s)
				}
			} else {
				yes := sym.lsp.Resolve(v, sym)
				if !yes {
					sym.Class_object = append(sym.Class_object, &s)
				}
				return i + 1
			}
		} else {
			yes := sym.lsp.Resolve(v, sym)
			if !yes {
				sym.Class_object = append(sym.Class_object, &s)
			}
		}
		i = i + 1
	}
	return i
}

type LspWorkspace struct {
	cpp     lsp_cpp
	py      lsp_py
	Wk      WorkSpace
	Current *Symbol_file
	filemap map[string]*Symbol_file
	Handle  lsp_data_changed
	cppcore *lspcore
	pycore  *lspcore
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
	wk.cpp.Close()
	wk.py.Close()
}
func (wk LspWorkspace) getClient(filename string) lspclient {
	if wk.cpp.IsMe(filename) {
		err := wk.cpp.Launch_Lsp_Server()
		if err == nil {
			wk.cpp.InitializeLsp(wk.Wk)
		}
		return wk.cpp
	}
	if wk.py.IsMe(filename) {
		return wk.py
	}
	return nil
}

func (wk *LspWorkspace) open(filename string) (*Symbol_file, bool, error) {
	val, ok := wk.filemap[filename]
	if ok {
		wk.Current = val
		return val, false, nil
	}
	wk.filemap[filename] = &Symbol_file{
		Filename: filename,
		lsp:      wk.getClient(filename),
		Handle:   wk.Handle,
		Wk:       wk,
	}

	ret := wk.filemap[filename]
	if ret.lsp == nil {
		return nil, false, fmt.Errorf("fail to open %s", filename)
	}
	err := ret.lsp.DidOpen(filename)
	return ret, true, err
}
func (wk *LspWorkspace) Open(filename string) (*Symbol_file, error) {
	ret, _, err := wk.open(filename)
	wk.Current = wk.filemap[filename]
	return ret, err

}
func NewLspWk(wk WorkSpace) *LspWorkspace {
	cppcore := &lspcore{}
	pycore := &lspcore{}
	ret := &LspWorkspace{
		cpp:     new_lsp_cpp(wk, cppcore),
		py:      lsp_py{new_lsp_base(wk, pycore)},
		Wk:      wk,
		pycore:  pycore,
		cppcore: cppcore,
	}
	ret.filemap = make(map[string]*Symbol_file)
	return ret
}

type lsp_data_changed interface {
	OnSymbolistChanged(file Symbol_file)
	OnCodeViewChanged(file Symbol_file)
	OnRefenceChanged(file []lsp.Location)
	OnCallInViewChanged(stacks []CallStack)
	OnCallTaskInViewChanged(stacks *CallInTask)
	OnCallTaskInViewResovled(stacks *CallInTask)
}
