package mainui

import (
	"fmt"
	// "log"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/reinhrst/fzf-lib"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	"zen108.com/lspvi/pkg/debug"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type GridClickCheck struct {
	*clickdetector
	target             tview.Primitive
	click              func(*tcell.EventMouse)
	dobule_click       func(*tcell.EventMouse)
	handle_mouse_event func(action tview.MouseAction, event *tcell.EventMouse)
}
type GridTreeClickCheck struct {
	*GridClickCheck
	tree *tview.TreeView
}
type GridListClickCheck struct {
	*GridClickCheck
	tree             *tview.List
	on_list_selected func()
	moveX            int
}

func NewFlexListClickCheck(grid *tview.Flex, list *tview.List, line int) *GridListClickCheck {
	return NewBoxListClickCheck(grid.Box, list, line)
}
func NewGridListClickCheck(grid *tview.Grid, list *tview.List, line int) *GridListClickCheck {
	return NewBoxListClickCheck(grid.Box, list, line)
}
func NewBoxListClickCheck(grid *tview.Box, list *tview.List, line int) *GridListClickCheck {
	ret := &GridListClickCheck{
		GridClickCheck: NewGridClickCheck(grid, list.Box),
		tree:           list,
	}
	ret.handle_mouse_event = func(action tview.MouseAction, event *tcell.EventMouse) {
		if action == tview.MouseMove {
			if !ret.has_click() {
				return
			}
			idx, err := get_grid_list_index(list, event, line)
			if err == nil {
				begin, _ := list.GetOffset()
				_, _, _, N := list.GetInnerRect()
				mouseX, _ := event.Position()
				if mouseX == ret.moveX {
					if begin <= idx && idx < begin+N {
						list.SetCurrentItem(idx)
					}
				}
				ret.moveX = mouseX
			}
		} else if action == tview.MouseScrollUp {
			list.MouseHandler()(action, event, nil)
		} else if action == tview.MouseScrollDown {
			list.MouseHandler()(action, event, nil)
		}
	}
	ret.click = func(em *tcell.EventMouse) {
		index, err := get_grid_list_index(list, em, line)
		if err != nil {
			return
		}
		list.SetCurrentItem(index)
		if ret.on_list_selected != nil {
			ret.on_list_selected()
		}
	}
	ret.dobule_click = func(event *tcell.EventMouse) {
		list.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, 0), nil)
	}
	return ret
}

func get_grid_list_index(list *tview.List, em *tcell.EventMouse, line int) (int, error) {
	_, y, _, _ := list.GetInnerRect()

	_, moustY := em.Position()
	offsetY, _ := list.GetOffset()
	index := (moustY-y)/line + offsetY
	if index > list.GetItemCount()-1 || index < 0 || list.GetItemCount() == 0 {
		return 0, fmt.Errorf("%d is out of range", index)
	}
	debug.TraceLog("mouseY", moustY, "listY=", y, "list offset", offsetY, "idnex", index)
	return index, nil
}
func NewTreeClickCheck(grid *tview.Box, tree *tview.TreeView) *GridTreeClickCheck {
	ret := &GridTreeClickCheck{
		GridClickCheck: NewGridClickCheck(grid, tree.Box),
		tree:           tree,
	}
	ret.handle_mouse_event = func(action tview.MouseAction, event *tcell.EventMouse) {
		if action == tview.MouseScrollUp {
			tree.MouseHandler()(action, event, nil)
		} else if action == tview.MouseScrollDown {
			tree.MouseHandler()(action, event, nil)
		}
	}
	return ret
}

//	func NewGridTreeClickCheck(grid *tview.Grid, tree *tview.TreeView) *GridTreeClickCheck {
//		ret := &GridTreeClickCheck{
//			GridClickCheck: NewGridClickCheck(grid.Box, tree.Box),
//			tree:           tree,
//		}
//		ret.handle_mouse_event = func(action tview.MouseAction, event *tcell.EventMouse) {
//			if action == tview.MouseScrollUp {
//				tree.MouseHandler()(action, event, nil)
//			} else if action == tview.MouseScrollDown {
//				tree.MouseHandler()(action, event, nil)
//			}
//		}
//		return ret
//	}
func NewGridClickCheck(grid *tview.Box, target tview.Primitive) *GridClickCheck {
	ret := &GridClickCheck{
		clickdetector: &clickdetector{lastMouseClick: time.Time{}},
		target:        target,
	}
	grid.SetMouseCapture(ret.handle_mouse)
	return ret
}
func (pk *GridClickCheck) handle_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	if InRect(event, pk.target) {
		a1, _ := pk.clickdetector.handle(action, event)
		if a1 == tview.MouseLeftClick {
			if pk.click != nil {
				pk.click(event)
			}
		} else if a1 == tview.MouseLeftDoubleClick {
			if pk.dobule_click != nil {
				pk.dobule_click(event)
			}
		} else {
			if pk.handle_mouse_event != nil {
				pk.handle_mouse_event(action, event)
			}
		}
		return tview.MouseConsumed, event
	}
	return action, event
}
func (picker *symbolpicker) layout(input *tview.InputField, isflex bool) (row *tview.Flex, col *tview.Grid) {
	list := picker.impl.symview.view
	list.SetBorder(true)
	code := picker.impl.codeprev.Primitive()
	picker.impl.codeprev.LoadFileNoLsp(picker.impl.filename, 0)
	if isflex {
		layout := layout_list_row_edit(list, code, input)
		picker.impl.click = NewTreeClickCheck(layout.Box, picker.impl.symview.view)
		row = layout
	} else {
		layout := layout_list_edit(list, code, input)
		picker.impl.click = NewTreeClickCheck(layout.Box, picker.impl.symview.view)
		col = layout
	}
	picker.impl.click.click = func(event *tcell.EventMouse) {
		_, y := event.Position()
		t := picker.impl.symview.view
		_, rectY, _, _ := t.GetInnerRect()
		y += t.GetScrollOffset() - rectY
		nodes := picker.impl.symview.nodes()
		if y >= len(nodes) || len(nodes) == 0 {
			return
		}
		if y < 0 {
			return
		}
		node := nodes[y]
		t.SetCurrentNode(node)
		picker.update_preview()
	}
	picker.impl.click.dobule_click = func(event *tcell.EventMouse) {
		picker.impl.click.tree.MouseHandler()(tview.MouseLeftClick, event, nil)
		// log.Println("dobule")

	}

	return row, col
}

func new_outline_picker(v *fzfmain, code CodeEditor) symbolpicker {

	sym := symbolpicker{
		impl: &SymbolWalkImpl{
			filename: code.Path(),
			codeprev: NewCodeView(v.main),
		},
	}
	symbolview := &SymbolTreeViewExt{}
	symbolview.SymbolTreeView = NewSymbolTreeView(v.main, code)
	symbolview.parent = v
	symbolview.SymbolTreeView.view.SetSelectedFunc(symbolview.OnClickSymobolNode)
	symbolview.collapse_children = false
	sym.impl.symbol = symbolview.merge_symbol(code.TreeSitter(), code.LspSymbol())
	if sym.impl.symbol != nil {
		symbolview.update_in_main_sync(sym.impl.symbol)
		symbolview.view.GetRoot().ExpandAll()
	}
	sym.impl.symview = symbolview
	sym.impl.fzf = NewSymbolListFzf(sym.impl.symbol)
	return sym
}

type SymbolTreeViewExt struct {
	*SymbolTreeView
	parent *fzfmain
}

func (v SymbolTreeViewExt) OnClickSymobolNode(node *tview.TreeNode) {
	v.SymbolTreeView.OnClickSymobolNode(node)
	v.parent.hide()
	v.main.set_viewid_focus(v.SymbolTreeView.editor.vid())
	v.main.CmdLine().Vim.EnterEscape()
}

type SymbolWalkImpl struct {
	symbol   *lspcore.Symbol_file
	filename string
	symview  *SymbolTreeViewExt
	gs       *GenericSearch
	codeprev CodeEditor
	click    *GridTreeClickCheck
	fzf      *FzfSymbolFilter
}

type symbolpicker struct {
	impl *SymbolWalkImpl
}

// close implements picker.
func (sym symbolpicker) close() {
}

// name implements picker.
func (sym symbolpicker) name() string {
	return "Document symbol " + sym.impl.symview.editor.FileName()
}

func (wk symbolpicker) handle_key_override(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := wk.impl.symview.view.InputHandler()
	handle(event, setFocus)
	wk.update_preview()
}

func (wk symbolpicker) update_preview() {
	cur := wk.impl.symview.view.GetCurrentNode()
	if cur != nil {
		value := cur.GetReference()
		if value != nil {
			if sym, ok := value.(lsp.SymbolInformation); ok {
				wk.impl.codeprev.goto_location_no_history(sym.Location.Range, false, nil)
			}
		}
	}
}

type SymbolFzf struct {
	sym    lspcore.Symbol
	index  int
	Member []SymbolFzf
}
type FzfSymbolFilter struct {
	ClassObject      []SymbolFzf
	names            []string
	filter           map[int]bool
	selected_postion map[int][]int
	fzf              *fzf.Fzf
}

func NewSymbolListFzf(sym *lspcore.Symbol_file) *FzfSymbolFilter {
	ret := &FzfSymbolFilter{[]SymbolFzf{}, []string{}, nil, nil, nil}
	ret.Convert(sym)
	ret.fzf = fzf.New(ret.names, fzf.Options{Fuzzy: true, CaseMode: fzf.CaseIgnore})
	return ret
}
func (f *FzfSymbolFilter) Query(q string) {
	f.fzf.Search(q)
	result := <-f.fzf.GetResultChannel()
	f.filter = make(map[int]bool)
	f.selected_postion = make(map[int][]int)
	q=strings.ToLower(q)
	for _, v := range result.Matches {
		if v.Score<50{
			continue
		}
		pos := fzf_color_pos(v.Positions, v.Key)
		l:=0
		maxlen:=0
		for _, e := range pos {
			s:=(e.Y -e.X)
			l+=s
			maxlen=max(s,maxlen) 
		}
		if len(q)>l{
			continue
		}
		if len(q)>1 && maxlen==1{
			continue
		}
		for _, e := range pos {
			debug.DebugLog("symboltree",v.Key, "color",v.Key [e.X:e.Y],v.Score,l )	
		}
		f.filter[int(v.HayIndex)] = true
		f.selected_postion[int(v.HayIndex)] = v.Positions
	}
}
func (f *FzfSymbolFilter) GetSymbolFile(key string) *lspcore.Symbol_file {
	ret := []*lspcore.Symbol{}
	for _, v := range f.ClassObject {
		member := []lspcore.Symbol{}
		for _, vv := range v.Member {
			if _, ok := f.filter[vv.index]; ok {
				member = append(member, v.sym)
			}
		}
		var sss = v.sym
		root := &sss
		if _, yes := f.filter[v.index]; yes {
			root.Members = member
			ret = append(ret, root)
		} else if len(member) > 0 {
			root.Members = member
			ret = append(ret, root)
		}
	}
	return &lspcore.Symbol_file{
		Class_object: ret,
	}
}
func (f *FzfSymbolFilter) Convert(symbol *lspcore.Symbol_file) {
	if symbol == nil {
		return
	}
	id := 0
	ret := []SymbolFzf{}
	names := []string{}
	classobject := symbol.Class_object
	for _, v := range classobject {
		member := []SymbolFzf{}
		for i := range v.Members {
			sym := SymbolFzf{v.Members[i], id, nil}
			id++
			names = append(names, sym.sym.SymInfo.Name)
			member = append(member, sym)
		}
		var sss = SymbolFzf{*v, id, nil}
		names = append(names, sss.sym.SymInfo.Name)
		id++
		sss.Member = member
		ret = append(ret, sss)
	}
	f.ClassObject = ret
	f.names = names
}
func (sym symbolpicker) Filter(key string) *lspcore.Symbol_file {
	return sym.StringMatch(key)
}

func (sym symbolpicker) StringMatch(key string) *lspcore.Symbol_file {
	if len(key) == 0 || sym.impl.symbol == nil {
		return nil
	}
	ret := []*lspcore.Symbol{}
	for _, v := range sym.impl.symbol.Class_object {
		member := []lspcore.Symbol{}
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
	return &lspcore.Symbol_file{
		Class_object: ret,
	}
}

// handle implements picker.
func (wk symbolpicker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return wk.handle_key_override
}
func (wk symbolpicker) Updatequeryold(query string) {
	wk.impl.gs = NewGenericSearch(view_outline_list, query)
	ret := wk.impl.symview.OnSearch(query)
	if len(ret) > 0 {
		wk.impl.symview.movetonode(ret[0].Y)
	}
}
func (picker symbolpicker) UpdateQuery(query string) {
	file := picker.Filter(strings.ToLower(query))
	picker.impl.symview.update_in_main_sync(file)
	root := picker.impl.symview.view.GetRoot()
	if root != nil {
		root.ExpandAll()
		for _, v := range root.GetChildren() {
			v.Expand()
		}
		children := root.GetChildren()
		if len(children) > 0 {
			picker.impl.symview.view.SetCurrentNode(children[0])
			picker.update_preview()
		}
	}
}
