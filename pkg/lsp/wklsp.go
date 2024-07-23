package lspcore

import "github.com/tectiv3/go-lsp"

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
	SymInfo lsp.SymbolInformation
	Members []Symbol
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
	if len(sym.Class_object) > 0 {
		sym.Handle.OnSymbolistChanged(*sym)
		return
	}
	symbols, err := sym.lsp.GetDocumentSymbol(sym.Filename)
	if err != nil {
		return
	}
	sym.build_class_symbol(symbols.SymbolInformation, 0, nil)
	sym.Handle.OnSymbolistChanged(*sym)
}
func (sym *Symbol_file) CallinTask(loc lsp.Location) (*CallInTask, error) {
	task := NewCallInTask(loc, sym.lsp)
	task.run()
	sym.Handle.OnCallTaskInViewChanged(task)
	return task, nil
}
func (sym *Symbol_file) Callin(loc lsp.Location) ([]CallStack, error) {
	var ret []CallStack
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
	wk      WorkSpace
	Current *Symbol_file
	filemap map[string]*Symbol_file
	Handle  lsp_data_changed
}

func (wk LspWorkspace) getClient(filename string) lspclient {
	if wk.cpp.IsMe(filename) {
		err := wk.cpp.Launch_Lsp_Server()
		if err == nil {
			wk.cpp.InitializeLsp(wk.wk)
		}
		return wk.cpp
	}
	if wk.py.IsMe(filename) {
		return wk.py
	}
	return nil
}

func (wk *LspWorkspace) Open(filename string) (*Symbol_file, error) {
	val, ok := wk.filemap[filename]
	if ok {
		wk.Current = val
		return val, nil
	}
	wk.filemap[filename] = &Symbol_file{
		Filename: filename,
		lsp:      wk.getClient(filename),
		Handle:   wk.Handle,
	}

	wk.Current = wk.filemap[filename]
	ret := wk.filemap[filename]
	err := ret.lsp.DidOpen(filename)
	return ret, err
}
func NewLspWk(wk WorkSpace) *LspWorkspace {
	ret := &LspWorkspace{
		cpp: new_lsp_cpp(wk),
		py:  lsp_py{new_lsp_base(wk)},
		wk:  wk,
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
}
