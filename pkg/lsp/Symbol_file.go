package lspcore

import (
	// "crypto/sha256"
	// "encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/tectiv3/go-lsp"
)

type Symbol_file struct {
	lsp          lspclient
	Filename     string
	Handle       lsp_data_changed
	Class_object []*Symbol
	Wk           *LspWorkspace
	tokens       *lsp.SemanticTokens
}

func (sym *Symbol_file) HasLsp() bool {
	return sym.lsp != nil
}
func (sym *Symbol_file) Find(rang lsp.Range) *Symbol {
	for _, v := range sym.Class_object {
		if v.Is_class() {
			for i := range v.Members {
				f := &v.Members[i]
				f = sym.newMethod(f, rang)
				if f != nil {
					return f
				}
			}
		}
		ret := sym.newMethod(v, rang)
		if ret != nil {
			return ret
		}
	}
	return nil
}

func (*Symbol_file) newMethod(v *Symbol, rang lsp.Range) *Symbol {
	if v.SymInfo.Kind == lsp.SymbolKindFunction || v.SymInfo.Kind == lsp.SymbolKindMethod {
		if rang.Start.Line > v.SymInfo.Location.Range.Start.Line && rang.End.Line <= v.SymInfo.Location.Range.End.Line {
			return v
		}
	}
	return nil

}
func (sym *Symbol_file) Filter(key string) *Symbol_file {
	if len(key) == 0 {
		return sym
	}
	ret := []*Symbol{}
	for _, v := range sym.Class_object {
		member := []Symbol{}
		for i, vv := range v.Members {
			if strings.Contains(strings.ToLower(vv.SymInfo.Name), key) {
				member = append(member, v.Members[i])
			}
		}
		var sss = *v
		root := &sss
		if strings.Contains(strings.ToLower(v.SymInfo.Name), key) {
			root.Members = member
			ret = append(ret, root)
		} else if len(member) > 0 {
			root.Members = member
			ret = append(ret, root)
		}

	}
	return &Symbol_file{
		Class_object: ret,
	}
}
func (sym *Symbol_file) build_class_symbol(symbols []lsp.SymbolInformation, begin int, parent *Symbol) int {
	var i = begin
	for i = begin; i < len(symbols); {
		v := symbols[i]
		s := Symbol{
			SymInfo: v,
		}
		if is_class(v.Kind) {
			var found = false
			for _, c := range sym.Class_object {
				if s.SymInfo.Name == c.SymInfo.Name {
					i = sym.build_class_symbol(symbols, i+1, &s)
					c.Members = append(c.Members, s.Members...)
					found = true
					break
				}
			}
			if !found {
				sym.Class_object = append(sym.Class_object, &s)
				i = sym.build_class_symbol(symbols, i+1, &s)
			}
			continue
		}
		if parent != nil {
			if parent.contain(s) {
				if is_memeber(v.Kind) {
					s.Classname = parent.SymInfo.Name
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
func (sym *Symbol_file) WorkspaceQuery(query string) ([]lsp.SymbolInformation, error) {
	if sym.lsp == nil {
		return nil, errors.New("lsp is nil")
	}
	return sym.lsp.WorkSpaceSymbol(query)
}
func (sym *Symbol_file) Reference(ranges lsp.Range) {
	if sym.lsp == nil {
		return
	}
	loc, err := sym.lsp.GetReferences(sym.Filename, ranges.Start)
	if err != nil {
		return
	}
	body, err := NewBody(lsp.Location{URI: lsp.NewDocumentURI(sym.Filename), Range: ranges})
	if err != nil {
		log.Println(err)
		return
	}
	key := body.String()
	sym.Handle.OnLspRefenceChanged(SymolSearchKey{Ranges: ranges, File: sym.Filename, Key: key}, loc)
}
func (sym *Symbol_file) Declare(ranges lsp.Range, line *OpenOption) {
	if sym.lsp == nil {
		return
	}
	loc, err := sym.lsp.GetDeclare(sym.Filename, ranges.Start)
	if err != nil {
		return
	}
	sym.Handle.OnFileChange(loc, line)
}
type OpenOption struct{
	Line int
}
func (sym *Symbol_file) GotoDefine(ranges lsp.Range, line *OpenOption) {
	if sym.lsp == nil {
		return
	}
	loc, err := sym.lsp.GetDefine(sym.Filename, ranges.Start)
	if err != nil {
		return
	}
	if len(loc) > 0 {
		sym.Handle.OnFileChange(loc, line)
	}
}

func (sym *Symbol_file) Caller(loc lsp.Location, cb bool) ([]CallStack, error) {
	var ret []CallStack
	if sym.lsp == nil {
		return ret, fmt.Errorf("lsp is null")
	}
	c1, err := sym.lsp.PrepareCallHierarchy(loc)
	if err != nil {
		return ret, err
	}
	for _, v := range c1 {
		log.Println("prepare", v.Name, v.URI.AsPath().String(), v.Range.Start.Line, v.Kind.String())
	}
	// for _, prepare := range c1 {
	// body, err := NewBody(loc)
	// if err != nil {
	// 	log.Println(err)
	// 	return ret, err
	// }
	if len(c1) > 0 {
		prepare := c1[0]
		body, err := NewBody(lsp.Location{URI: prepare.URI, Range: prepare.Range})
		if err != nil {
			return ret, err
		}
		search_txt :=
			body.String()
		c2, err := sym.lsp.CallHierarchyIncomingCalls(prepare)
		if err == nil {
			for _, f := range c2 {
				var stack CallStack
				v := f.From
				log.Println("caller ", v.Name, v.URI.AsPath().String(), v.Range.Start.Line, v.Kind.String())
				stack.Add(NewCallStackEntry(f.From, f.FromRanges, []lsp.Location{}))
				ret = append(ret, stack)
			}
		}
		if cb {
			sym.Handle.OnLspCaller(search_txt, c1[0], ret)
		}
	}
	return ret, nil
}
func (sym *Symbol_file) CallinTask(loc lsp.Location) (*CallInTask, error) {
	task := NewCallInTask(loc, sym.lsp)
	task.run()
	sym.Handle.OnLspCallTaskInViewChanged(task)
	return task, nil
}

func (sym *Symbol_file) Async_resolve_stacksymbol(task *CallInTask, hanlde func()) {
	bin, binerr := NewPlanUmlBin()
	export_root, export_err := NewExportRoot(&sym.Wk.Wk)
	dir_to_remvoe := filepath.Join(export_root.Dir, task.Dir())
	os.RemoveAll(dir_to_remvoe)
	rename := map[string]int{}
	for _, s := range task.Allstack {
		var xx = class_resolve_task{
			wklsp:     sym.Wk,
			callstack: s,
		}
		xx.Run()
		if hanlde != nil {
			name := "callin"
			if len(s.Items) > 0 {
				name = s.Items[0].Name
				if d, ok := rename[name]; ok {
					rename[name] = d + 1
					name = fmt.Sprintf("%d_%s", d, name)
				} else {
					rename[name] = 1
				}
				// for i := range s.Items {
				// 	index := len(s.Items)
				// 	index = index - 1 - i
				// 	name += "." + s.Items[index].Name
				// }
				// if len(name) > 1024 {
				// 	buf := sha256.Sum256([]byte(name))
				// 	name = hex.EncodeToString(buf[:])
				// }
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
			task.Save(export_root.Dir)
			hanlde()
		}
	}
	task.Save(export_root.Dir)
}
func (sym *Symbol_file) __load_symbol_impl(reload bool) error {
	if sym.lsp == nil {
		return fmt.Errorf("lsp is nil")
	}
	if reload {
		sym.Class_object = []*Symbol{}
	}
	if len(sym.Class_object) > 0 {
		sym.Handle.OnSymbolistChanged(sym, nil)
		return nil
	}
	symbols, err := sym.lsp.GetDocumentSymbol(sym.Filename)
	if err != nil {
		return err
	}
	sym.build_class_symbol(symbols.SymbolInformation, 0, nil)
	return nil
}
func (sym *Symbol_file) DidSave() {
	if sym.lsp != nil {
		buf, err := os.ReadFile(sym.Filename)
		var text string
		if err == nil {
			text = string(buf)
		}
		sym.lsp.DidSave(sym.Filename, text)
	}
}
func (sym *Symbol_file) LoadSymbol(reload bool) {
	err := sym.__load_symbol_impl(reload)
	sym.Handle.OnSymbolistChanged(sym, err)
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
