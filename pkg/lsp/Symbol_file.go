package lspcore

import (
	// "crypto/sha256"
	// "encoding/hex"
	"errors"
	"fmt"
	"strconv"

	// "log"
	"os"
	"path/filepath"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/tectiv3/go-lsp"
	"zen108.com/lspvi/pkg/debug"
)

type Symbol_file struct {
	lsp          lspclient
	Filename     string
	Handle       lsp_data_changed
	Class_object []*Symbol
	Wk           *LspWorkspace
	tokens       *lsp.SemanticTokens
	verison      int
	Ts           *TreeSitter
	lspopen      bool
}

func member_is_added(m Symbol, class_symbol *Symbol) bool {
	for _, member := range class_symbol.Members {
		if member.SymInfo.Name == m.SymInfo.Name {
			return true
		}
	}
	return false
}
func find_in_outline(outline []*Symbol, class_symbol *Symbol) bool {
	for _, cls := range outline {
		if cls.SymInfo.Location.Range.Overlaps(class_symbol.SymInfo.Location.Range) {
			for _, m := range cls.Members {
				if m.SymInfo.Location.Range.Overlaps(class_symbol.SymInfo.Location.Range) {
					if !member_is_added(m, class_symbol) {
						class_symbol.Members = append(class_symbol.Members, m)
					}
				}
			}
		}
	}
	return false
}
func merge_ts_to_lsp(symbol *Symbol_file, Current *Symbol_file) {
	for _, v := range symbol.Class_object {
		if v.Is_class() {
			find_in_outline(Current.Class_object, v)
		}
	}
}
func MergeSymbol(ts *TreeSitter, symbol *Symbol_file) *Symbol_file {
	if ts != nil {
		var go_land lsp_lang_go
		if go_land.IsMe(ts.filename.Path()) {
			symbol = nil
		}
	}
	var Current *Symbol_file
	if ts != nil {
		Current = &Symbol_file{
			Class_object: ts.Outline,
		}
	}
	if symbol != nil && symbol.HasLsp() {
		if Current != nil {
			merge_ts_to_lsp(symbol, Current)
		}
		return symbol
	} else if Current != nil {
		return Current
	}
	return nil
}

type LspSignatureHelp struct {
	TriggerChar []string
	Document    func(v lsp.SignatureHelp, call SignatureHelp) (text []string)
}
type LspCompleteUtil struct {
	TriggerChar []string
	Document    func(v lsp.CompletionItem) (text []string)
}

func (sym Symbol_file) LspHelp() (ret LspUtil, err error) {
	if sym.lsp == nil {
		err = fmt.Errorf("lsp is nil")
		return
	}
	return sym.lsp.LspHelp()
}
func (sym Symbol_file) LspClient() lspclient {
	return sym.lsp
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
func (s *Symbol_file) SignatureHelp(arg SignatureHelp) (ret lsp.SignatureHelp, err error) {
	arg.File = s.Filename
	if s.lsp == nil {
		err = errors.New("lsp is nil")
		if arg.HelpCb != nil {
			arg.HelpCb(lsp.SignatureHelp{}, arg, err)
		}
		return
	}
	debug.DebugLog("help", "lsp signature help", "pos", arg.Pos.String(), strconv.Quote(arg.TriggerCharacter), filepath.Base(arg.File))
	ret, err = s.lsp.SignatureHelp(arg)
	s.on_error(err)
	return
}

func (s *Symbol_file) on_error(err error) {
	if err == nil {
		return
	}
	s.Wk.Wk.Callback.LspLogOutput("LSPSERVER_API_ERROR", fmt.Sprintf("%v", err))
	v := err.(*jsonrpc2.Error)
	if v != nil {
		if v.Code == -32602 {
			s.verison++
			s.lspopen = false
			if err := s.lsp.DidOpen(SourceCode{Path: s.Filename}, s.verison); err != nil {
				debug.DebugLog(DebugTag, "lsp reopen failed", s.Filename, err)
			}
			return
		}
	}
}
func (s *Symbol_file) Format(opt FormatOption) (ret []lsp.TextEdit, err error) {
	if s.lsp == nil {
		err = errors.New("lsp is nil")
		if opt.Format != nil {
			opt.Format(ret, err)
		}
		return
	}
	if opt.Filename == "" {
		opt.Filename = s.Filename
	}
	ret, err = s.lsp.Format(opt)
	s.on_error(err)
	return
}
func (s Symbol_file) IsTrigger(param string) (TriggerChar, error) {
	return s.lsp.IsTrigger(param)
}
func (s Symbol_file) CompletionItemResolve(param *lsp.CompletionItem) (ret *lsp.CompletionItem, err error) {
	if s.lsp == nil {
		err = errors.New("lsp is nil")
		return
	}
	ret, err = s.lsp.CompletionItemResolve(param)
	s.on_error(err)
	return
}
func (s *Symbol_file) DidComplete(param Complete) (ret lsp.CompletionList, err error) {
	param.File = s.Filename
	if s.lsp == nil {
		err = errors.New("lsp is nil")
		if param.CompleteHelpCallback != nil {
			param.CompleteHelpCallback(lsp.CompletionList{}, param, err)
		}
		return
	}
	ret, err = s.lsp.DidComplete(param)
	s.on_error(err)
	return
}
func (*Symbol_file) newMethod(v *Symbol, rang lsp.Range) *Symbol {
	if v.SymInfo.Kind == lsp.SymbolKindFunction || v.SymInfo.Kind == lsp.SymbolKindMethod {
		if rang.Overlaps(v.SymInfo.Location.Range) {
			return v
		}
	}
	return nil

}
func (sym *Symbol_file) build_class_symbol(symbols []lsp.SymbolInformation, begin int, parent *Symbol) int {
	var i = begin
	for i = begin; i < len(symbols); {
		v := symbols[i]
		s := Symbol{
			SymInfo: v,
		}
		if is_class(v.Kind) {
			inparent := false
			if parent != nil {
				if !parent.contain(s) {
					return i
				} else {
					inparent = true
				}
			}
			if !inparent {
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
					if i+1 < len(symbols) {
						next := symbols[i]
						if next.Location.Range.Overlaps(s.SymInfo.Location.Range) {
							i = sym.build_class_symbol(symbols, i+1, &s)
						}
					}
				}
			} else {
				if i+1 < len(symbols) {
					next := symbols[i]
					if next.Location.Range.Overlaps(s.SymInfo.Location.Range) {
						i = sym.build_class_symbol(symbols, i+1, &s)
					}
				}
				parent.Members = append(parent.Members, s)
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
		if s := GetLangTreeSitterSymbol(sym.Filename); s != nil {
			return s.WorkspaceQuery(query)
		}
		return nil, errors.New("lsp is nil")
	}
	return sym.lsp.WorkSpaceSymbol(query)
}
func (sym *Symbol_file) GetImplement(param SymolParam, option *OpenOption) {
	var ranges lsp.Range = param.Ranges
	var loc ImplementationResult
	var err error
	var key = param.Key
	if sym.lsp != nil {
		loc, err = sym.lsp.GetImplement(sym.Filename, ranges.Start)
	} else {
		err = fmt.Errorf("lsp is nil")
	}
	sym.Handle.OnGetImplement(
		SymolSearchKey{Ranges: ranges, File: sym.Filename, Key: key, sym: sym},
		loc,
		err,
		option)
}
func (sym *Symbol_file) Reference(req SymolParam) {
	ranges := req.Ranges
	key := req.Key
	var loc []lsp.Location
	var err error
	if sym.lsp == nil {
		err = fmt.Errorf("lsp is nil")
	} else {
		loc, err = sym.lsp.GetReferences(sym.Filename, req.Ranges.Start)
		sym.on_error(err)
	}
	sym.Handle.OnLspRefenceChanged(SymolSearchKey{Ranges: ranges, File: sym.Filename, Key: key, sym: sym}, loc, err)
}
func (sym *Symbol_file) Declare(ranges lsp.Range, line *OpenOption) {
	if sym.lsp == nil {
		return
	}
	loc, err := sym.lsp.GetDeclare(sym.Filename, ranges.Start)
	if err != nil {
		sym.on_error(err)
		return
	}
	if len(loc) > 0 {
		sym.Handle.OnFileChange(loc, line)
	}
}

type SymolParam struct {
	Ranges lsp.Range
	Key    string
	Line   *OpenOption
	File   string
}
type OpenTabOption int

const (
	OpenTabOption_CurrentTab OpenTabOption = iota
	OpenTabOption_NewTab
	OpenTabOption_Below
)

type OpenOption struct {
	LineNumber int
	Offset     int
	Newtab     OpenTabOption
	Openner    int
}

func NewOpenOption(line, offset int) *OpenOption {
	return &OpenOption{LineNumber: line, Offset: offset, Newtab: OpenTabOption_CurrentTab, Openner: -1}
}

//	func (z OpenOption) SetOffset(s int) OpenOption {
//		z.Offset = s
//		return z
//	}
func (z *OpenOption) SetNewTab(s OpenTabOption) *OpenOption {
	z.Newtab = s
	return z
}

func (z *OpenOption) SetOpenner(s int) *OpenOption {
	z.Openner = s
	return z
}

//	func (z OpenOption) SetLine(s int) OpenOption {
//		z.LineNumber = s
//		return z
//	}
func (sym *Symbol_file) GotoDefine(ranges lsp.Range, line *OpenOption) {
	if sym.lsp == nil {
		return
	}
	loc, err := sym.lsp.GetDefine(sym.Filename, ranges.Start)
	if err != nil {
		sym.on_error(err)
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
		debug.DebugLog(DebugTag, "prepare", v.Name, v.URI.AsPath().String(), v.Range.Start.Line, v.Kind.String())
	}
	// for _, prepare := range c1 {
	// body, err := NewBody(loc)
	// if err != nil {
	// 	log.Println(err)
	// 	return ret, err
	// }
	if len(c1) > 0 {
		prepare := c1[0]

		c2, err := sym.lsp.CallHierarchyIncomingCalls(prepare)
		if err == nil {
			for _, f := range c2 {
				var stack CallStack
				v := f.From
				debug.DebugLog("caller ", v.Name, v.URI.AsPath().String(), v.Range.Start.Line, v.Kind.String())
				stack.Add(NewCallStackEntry(f.From, f.FromRanges, []lsp.Location{}))
				ret = append(ret, stack)
			}
		}
		if cb {
			body, err := NewBody(lsp.Location{URI: prepare.URI, Range: prepare.Range})
			if err != nil {
				return ret, err
			}
			search_txt :=
				body.String()
			sym.Handle.OnLspCaller(search_txt, c1[0], ret)
		}
	}
	return ret, nil
}
func (sym *Symbol_file) CallHierarchyOutcomingCall(callitem lsp.CallHierarchyItem) ([]lsp.CallHierarchyOutgoingCall, error) {
	if sym.lsp == nil {
		return nil, fmt.Errorf("lsp is null")
	}
	return sym.lsp.CallHierarchyOutcomingCalls(callitem)
}
func (sym *Symbol_file) CallHierarchyIncomingCall(callitem lsp.CallHierarchyItem) (ret []lsp.CallHierarchyIncomingCall, err error) {
	if sym.lsp == nil {
		err = fmt.Errorf("lsp is null")
		return
	}
	ret, err = sym.lsp.CallHierarchyIncomingCalls(callitem)
	sym.on_error(err)
	return
}
func (sym *Symbol_file) LoadTreeSitter(justempty bool) {
	if justempty {
		if len(sym.Class_object) > 0 {
			return
		}
	}
	sym.Ts = GetNewTreeSitter(sym.Filename, CodeChangeEvent{File: sym.Filename, Full: true})
	sym.Ts.initblock()
	sym.Class_object = sym.Ts.Outline
}
func (sym *Symbol_file) PrepareCallHierarchy(loc lsp.Location) (ret []lsp.CallHierarchyItem, err error) {
	if sym.lsp == nil {
		err = fmt.Errorf("lsp is null")
		return
	}
	ret, err = sym.lsp.PrepareCallHierarchy(loc)
	sym.on_error(err)
	return
}
func (sym *Symbol_file) CallinTask(loc lsp.Location, level int) (*CallInTask, error) {
	task := NewCallInTask(loc, sym.lsp, level)
	task.sym = sym
	task.Run()
	sym.Handle.OnLspCallTaskInViewChanged(task)
	return task, nil
}

type Rename_record struct {
	rename map[string]int
}

func NewRenameRecord() *Rename_record {
	return &Rename_record{
		rename: make(map[string]int),
	}
}
func (sym *Symbol_file) Async_resolve_stacksymbol(task *CallInTask, hanlde func()) {
	export_root, _ := NewExportRoot(&sym.Wk.Wk)
	dir_to_remvoe := filepath.Join(export_root.Dir, task.Dir())
	os.RemoveAll(dir_to_remvoe)
	rename := Rename_record{rename: make(map[string]int)}
	for _, s := range task.Allstack {
		// for i := range s.Items {
		// 	index := len(s.Items)
		// 	index = index - 1 - i
		// 	name += "." + s.Items[index].Name
		// }
		// if len(name) > 1024 {
		// 	buf := sha256.Sum256([]byte(name))
		// 	name = hex.EncodeToString(buf[:])
		// }
		s.Resolve(sym, hanlde, &rename, task)
	}
	task.Save(export_root.Dir)
}

func (s *CallStack) Resolve(sym *Symbol_file, hanlde func(), rename *Rename_record, task *CallInTask) {
	index := 0
	for ind, call := range task.Allstack {
		if call == s {
			index = ind
			break
		}
	}
	os.Remove(s.MdName)
	os.Remove(s.UmlPngName)
	os.Remove(s.UtxtName)
	os.Remove(s.UmlName)
	var xx = class_resolve_task{
		wklsp:     sym.Wk,
		callstack: s,
	}
	xx.Run()
	if hanlde != nil {
		bin, binerr := NewPlanUmlBin()
		export_root, export_err := NewExportRoot(&sym.Wk.Wk)
		name := "callin"
		if len(s.Items) > 0 {
			if rename != nil {
				name = fmt.Sprintf("%d.%s", index, s.Items[0].DirName())
				if d, ok := rename.rename[name]; ok {
					rename.rename[name] = d + 1
					name = fmt.Sprintf("%d_%s", d, name)
				} else {
					rename.rename[name] = 1
				}
			}

		}
		if binerr == nil && export_err == nil && len(name) > 0 {
			content := s.Uml(true)
			if name, err := export_root.SaveMD(task.Dir(), name, content); err == nil {
				s.MdName = name
			}
			content = s.Uml(false)
			fileuml, err := export_root.SavePlanUml(task.Dir(), name, content)
			if err == nil {
				s.UmlName = fileuml
				s.UtxtName, s.UmlPngName, err = bin.Convert(fileuml)
				if err != nil {
					debug.ErrorLog(DebugTag, err)
				}
			} else {
				debug.ErrorLog(DebugTag, err)
			}
		}
		task.Save(export_root.Dir)
		hanlde()
	}
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
	return sym.LspLoadSymbol()
}

const LSP_DEBUG_TAG = "LSPDEBUG"

func (sym *Symbol_file) LspLoadSymbol() error {
	if sym.lsp == nil {
		return fmt.Errorf("lsp empty")
	}
	symbols, err := sym.lsp.GetDocumentSymbol(sym.Filename)
	if err != nil {
		return err
	}
	sym.Class_object = []*Symbol{}
	sym.build_class_symbol(symbols.SymbolInformation, 0, nil)
	return nil
}
func (sym *Symbol_file) NotifyCodeChange(event CodeChangeEvent) error {
	if sym.lsp != nil {
		if opt := sym.lsp.syncOption(); opt != nil {
			// var Start, End lsp.Position
			changeevents := []lsp.TextDocumentContentChangeEvent{}
			if event.Full {
				var data = event.Data
				var err error
				// endline := len(event.Data)
				if len(data) == 0 {
					if data, err = os.ReadFile(sym.Filename); err != nil {
						return err
					}
				}
				// Start = lsp.Position{
				// 	Line:      0,
				// 	Character: 0,
				// }
				// End = lsp.Position{
				// 	Line:      endline - 1,
				// 	Character: 0,
				// }
				// if End.Line == 0 {
				// 	debug.ErrorLog(DebugTag, "notify_code_change", "empty data", event)
				// }
				changeevents = []lsp.TextDocumentContentChangeEvent{{
					// Range: &lsp.Range{
					// 	Start: Start,
					// 	End:   End,
					// },
					Text: string(data),
				}}
			} else {
				for _, v := range event.Events {
					// if v.Range.End.Line == 0 {
					// 	log.Println("")
					// }
					e := lsp.TextDocumentContentChangeEvent{
						Range: &v.Range,
						Text:  v.Text,
					}
					changeevents = append(changeevents, e)
				}
			}
			if opt.Change != lsp.TextDocumentSyncKindNone {
				sym.verison++
				debug.TraceLog("cqdebug", "didchange", changeevents[0].String())
				return sym.lsp.DidChange(sym.Filename, sym.verison, changeevents)
			} else {
				return fmt.Errorf("TextDocumentSyncKindNone is None")
			}
		}
		return fmt.Errorf("sync option is None")
	}
	return fmt.Errorf("sym lsp is nil")
}

// func newFunction1(DidSave bool, sym *Symbol_file, OpenClose bool) {
// 	if DidSave {
// 		buf, err := os.ReadFile(sym.Filename)
// 		var text string
// 		if err == nil {
// 			text = string(buf)
// 		}
// 		sym.lsp.DidSave(sym.Filename, text)
// 	} else if OpenClose {
// 		wk := sym.Wk
// 		if err := wk.CloseSymbolFile(sym); err != nil {
// 			debug.ErrorLog(LSP_DEBUG_TAG, "OpenClose close symbol file failed", err)
// 		} else if sym, err := sym.Wk.Open(sym.Filename); sym != nil {
// 			if err != nil {
// 				debug.ErrorLog(LSP_DEBUG_TAG, "OpenClose open symbol file failed", err)
// 			}
// 			if err := sym.LspLoadSymbol(); err != nil {
// 				debug.ErrorLog(LSP_DEBUG_TAG, "OpenClose load symbol failed", err, sym.Filename)
// 			} else {
// 				debug.InfoLog(LSP_DEBUG_TAG, "OpenClose load symbol success", sym.Filename)
// 			}
// 		}

//		}
//	}
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
	sym.on_error(err)
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
