package mainui

import (
	// "strings"

	"fmt"
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"

	// lsp "github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

func (grepx *prev_picker_impl) update_title(s string) {
	grepx.parent.update_dialog_title(s)
}
func (impl *prev_picker_impl) flex(input *tview.InputField, linenum int) *tview.Flex {
	code := impl.codeprev.Primitive()
	if impl.listcustom != nil {
		list := impl.listcustom
		list.SetBorder(true)
		layout := layout_list_row_edit(list, code, input)
		impl.list_click_check = NewFlexListClickCheck(layout, list.List, linenum)
		impl.list_click_check.on_list_selected = func() {
			if impl.on_list_selected != nil {
				impl.on_list_selected()
			} else {
				impl.update_preview()
			}
		}
		return layout
	} else {
		list := impl.listview
		list.SetBorder(true)
		layout := layout_list_row_edit(list, code, input)
		impl.list_click_check = NewFlexListClickCheck(layout, list, linenum)
		impl.list_click_check.on_list_selected = func() {
			if impl.on_list_selected != nil {
				impl.on_list_selected()
			} else {
				impl.update_preview()
			}
		}
		return layout
	}
}
func (impl *prev_picker_impl) grid(input *tview.InputField, linenum int) *tview.Grid {
	list := impl.listview
	list.SetBorder(true)
	code := impl.codeprev.Primitive()
	var layout *tview.Grid
	if impl.listcustom != nil {
		layout = layout_list_edit(impl.listcustom, code, input)
		list = impl.listcustom.List
	} else {
		layout = layout_list_edit(list, code, input)
	}
	impl.list_click_check = NewGridListClickCheck(layout, list, linenum)
	impl.list_click_check.on_list_selected = func() {
		if impl.on_list_selected != nil {
			impl.on_list_selected()
		} else {
			impl.update_preview()
		}
	}
	return layout
}
func (pk *refpicker) row(input *tview.InputField) *tview.Flex {
	ret := pk.impl.flex(input, 2)
	pk.impl.PrevOpen(pk.impl.file.Filename, -1)
	return ret
}
func (pk *refpicker) grid(input *tview.InputField) *tview.Grid {
	ret := pk.impl.grid(input, 2)
	pk.impl.PrevOpen(pk.impl.file.Filename, -1)
	return ret
}
func layout_list_row_edit(list tview.Primitive, code tview.Primitive, input *tview.InputField) *tview.Flex {
	layout := tview.NewFlex()
	layout.SetDirection(tview.FlexRow)
	layout.AddItem(list, 0, 6, false).AddItem(code, 0, 9, false).AddItem(input, 1, 1, true)
	return layout
}

func layout_list_edit(list tview.Primitive, code tview.Primitive, input *tview.InputField) *tview.Grid {
	layout := tview.NewGrid().
		SetColumns(-1, 24, 16, -1).
		SetRows(-1, 3, 3, 2).
		AddItem(list, 0, 0, 3, 2, 0, 0, false).
		AddItem(code, 0, 2, 3, 2, 0, 0, false).
		AddItem(input, 3, 0, 1, 4, 0, 0, false)
	return layout
}

type prev_picker_impl struct {
	listview         *tview.List
	listcustom       *customlist
	codeprev         CodeEditor
	cq               *CodeOpenQueue
	parent           *fzfmain
	list_click_check *GridListClickCheck
	on_list_selected func()
	listdata         []ref_line
	// editor           CodeEditor
}

func (imp *prev_picker_impl) PrevOpen(filename string, line int) {
	imp.cq.enqueue(EditorOpenArgument{filename: filename, line: line})
}
func (impl *prev_picker_impl) use_cusutom_list(l *customlist) {
	impl.listview = l.List
	impl.listcustom = l
}
func (impl *prev_picker_impl) update_preview() {
	cur := impl.listview.GetCurrentItem()
	if cur < len(impl.listdata) {
		item := impl.listdata[cur]
		impl.PrevOpen(item.loc.URI.AsPath().String(), item.loc.Range.Start.Line)
	}
}

type refpicker_impl struct {
	*prev_picker_impl
	file *lspcore.Symbol_file
	refs []ref_with_caller
	fzf  *fzf_on_listview
	key  string
	qk   quick_view
}
type refpicker struct {
	impl *refpicker_impl
}

// OnGetImplement implements lspcore.lsp_data_changed.
func (pk refpicker) OnGetImplement(ranges lspcore.SymolSearchKey, file lspcore.ImplementationResult, err error, option *lspcore.OpenOption) {
	panic("unimplemented")
}

// close implements picker.
func (pk refpicker) close() {
}

// name implements picker.
func (pk refpicker) name() string {
	if pk.impl.listcustom.GetItemCount() == 0 {
		return fmt.Sprint("Reference", " ", "0/0")
	}
	return fmt.Sprint("Reference", " ", pk.impl.listcustom.GetCurrentItem()+1, "/", pk.impl.listcustom.GetItemCount())
}

// OnLspCaller implements lspcore.lsp_data_changed.
func (pk refpicker) OnLspCaller(txt string, c lsp.CallHierarchyItem, stacks []lspcore.CallStack) {
	panic("unimplemented")
}

// OnLspCallTaskInViewChanged implements lspcore.lsp_data_changed.
func (pk refpicker) OnLspCallTaskInViewChanged(stacks *lspcore.CallInTask) {
	panic("unimplemented")
}

// OnLspCallTaskInViewResovled implements lspcore.lsp_data_changed.
func (pk refpicker) OnLspCallTaskInViewResovled(stacks *lspcore.CallInTask) {
	panic("unimplemented")
}

// OnCodeViewChanged implements lspcore.lsp_data_changed.
func (pk refpicker) OnCodeViewChanged(file *lspcore.Symbol_file) {
	panic("unimplemented")
}

// OnFileChange implements lspcore.lsp_data_changed.
func (pk refpicker) OnFileChange([]lsp.Location, *lspcore.OpenOption) {
	panic("unimplemented")
}
func (pk refpicker) update_title() {
	s := pk.name()
	pk.impl.parent.update_dialog_title(s)
}

// func caller_to_listitem(caller *lspcore.CallStackEntry, root string) string {
// 	if caller == nil {
// 		return ""
// 	}
// 	caller_color := global_theme.search_highlight_color()
// 	if c, err := global_theme.get_lsp_color(lsp.SymbolKindFunction); err == nil {
// 		f, _, _ := c.Decompose()
// 		caller_color = f
// 	}
// 	callerstr := fmt.Sprintf("%s:%d %-20s",
// 		trim_project_filename(
// 			caller.Item.URI.AsPath().String(), root),
// 		caller.Item.Range.Start.Line+1, fmt_color_string(caller.Name, caller_color))
// 	return callerstr
// }

type ref_line struct {
	caller string
	code   string
	loc    lsp.Location
	line   string
	path   string
}

// func (ref ref_line) fzf_tring() string {
// 	return ref.String() + ref.caller
// }
// func (v ref_line) get_loc() (lsp.Location, error) {
// 	line, err := strconv.Atoi(v.line)
// 	if err != nil {
// 		return lsp.Location{}, err
// 	}
// 	loc := lsp.Location{
// 		URI: lsp.NewDocumentURI(v.path),
// 		Range: lsp.Range{Start: lsp.Position{Line: line},
// 			End: lsp.Position{Line: line},
// 		},
// 	}
// 	return loc, nil
// }

func (ref ref_line) String() string {
	return fmt.Sprintf("%s %s:%d", ref.line, ref.path, ref.loc.Range.Start.Line)
}

type ref_with_caller struct {
	Loc      lsp.Location
	Caller   *lspcore.CallStackEntry
	CodeLine string
}

func (pk refpicker) OnLspRefenceChanged(key lspcore.SymolSearchKey, file []lsp.Location) {
	listview := pk.impl.listview
	listview.Clear()
	refs := get_loc_caller(pk.impl.parent.main, file, key.Symbol())
	pk.impl.refs = refs

	qk := &pk.impl.qk
	qk.main = pk.impl.parent.main

	qk.view = pk.impl.listcustom
	qk.Refs.Refs = refs
	tree := list_view_tree_extend{}
	tree.build_tree(pk.impl.refs)
	qk.tree = &tree
	data := tree.BuildListStringGroup(qk, global_prj_root, pk.impl.parent.main.Lspmgr())

	pk.impl.key = key.Key
	pk.loadlist(data)
	pk.update_preview()
}

func (impl *prev_picker_impl) open_location(v lsp.Location) {
	impl.parent.main.OpenFileHistory(v.URI.AsPath().String(), &v)
	impl.parent.hide()
}

func get_loc_caller(m MainService, file []lsp.Location, lsp *lspcore.Symbol_file) []ref_with_caller {
	ref_call_in := []ref_with_caller{}
	for _, v := range file {
		stacks, err := lsp.Caller(v, false)
		if err == nil {
			a := newFunction(stacks, v)
			if a != nil {
				ref_call_in = append(ref_call_in, *a)
				continue
			}

		}
		caller, _ := m.Lspmgr().GetCallEntry(v.URI.AsPath().String(), v.Range)
		ref_call_in = append(ref_call_in, ref_with_caller{Loc: v, Caller: caller})

	}
	return ref_call_in
}

func newFunction(stacks []lspcore.CallStack, v lsp.Location) *ref_with_caller {
	for _, s := range stacks {
		for _, item := range s.Items {
			for _, r := range item.FromRanges {
				if r.Start.Line == v.Range.Start.Line {
					return &ref_with_caller{
						Loc:    v,
						Caller: item,
					}
				}
			}
		}
	}
	return nil
}

// OnSymbolistChanged implements lspcore.lsp_data_changed.
func (ref refpicker) OnSymbolistChanged(file *lspcore.Symbol_file, err error) {
	panic("unimplemented")
}

func new_refer_picker(clone lspcore.Symbol_file, v *fzfmain) refpicker {
	x := new_preview_picker(v)
	impl := refpicker_impl{
		prev_picker_impl: x,
		file:             &clone,
	}
	sym := refpicker{
		impl: &impl,
	}
	x1 := new_customlist(true)
	x1.SetSelectedFunc(func(index_list int, s1, s2 string, r rune) {
		log.Println(index_list, s1, s2, r)
		data_index := impl.fzf.get_data_index(index_list)
		loc := impl.qk.get_data(data_index)
		v.main.OpenFileHistory(loc.Loc.URI.AsPath().String(), &loc.Loc)
		v.hide()
	})
	sym.impl.use_cusutom_list(x1)
	return sym
}

func new_preview_picker(v *fzfmain) *prev_picker_impl {
	x := &prev_picker_impl{
		listview: tview.NewList(),
		codeprev: NewCodeView(v.main),
		parent:   v,
		// editor:   editor,
	}
	x.cq = NewCodeOpenQueue(x.codeprev)
	return x
}
func (pk *refpicker) load(ranges lsp.Range) {
	pk.impl.file.Handle = *pk
	pk.impl.file.Reference(ranges)
}
func (pk refpicker) handle_key_override(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := pk.impl.listcustom.InputHandler()

	handle(event, setFocus)
	switch event.Key() {
	case tcell.KeyUp, tcell.KeyDown:
		data := pk.impl.qk.get_data(pk.impl.listcustom.GetCurrentItem())
		pk.impl.PrevOpen(data.Loc.URI.AsPath().String(), data.Loc.Range.Start.Line)
		pk.update_title()
	default:
		break
	}
}

func (pk *refpicker) update_preview() {
	pk.impl.update_preview()
}

func remove_hl(mc []colortext, new_query string) string {
	maintext := ""
	for i := range mc {
		item := mc[i]
		maintext += item.text
	}
	if new_query != "" {
		maintext = strings.ReplaceAll(maintext, new_query, fmt_bold_string(new_query))
	}
	return maintext
}
func highlight_listitem_search_key(old_query string, view *customlist, new_query string) {
	sss := [][2]string{}
	for i := 0; i < view.GetItemCount(); i++ {
		m, s := view.GetItemText(i)
		mc := GetColorText(m, []colortext{})
		m = remove_hl(mc, new_query)

		mc = GetColorText(s, []colortext{})
		s = remove_hl(mc, new_query)

		sss = append(sss, [2]string{m, s})
	}
	view.Clear()
	for _, v := range sss {
		view.AddItem(v[0], v[1], nil)
	}
}

// handle implements picker.
func (pk refpicker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return pk.handle_key_override
}
func (pk refpicker) UpdateQuery(query string) {
	query = strings.ToLower(query)
	listview := pk.impl.listcustom
	listview.Clear()
	if fzf := pk.impl.fzf; fzf != nil {
		oldkey := fzf.OnSearch(query, true)
		highlight_listitem_search_key(oldkey, listview, query)
		pk.update_preview()
	}
}

// func (pk refpicker) onselected(data_index int, list int) {
// 	v := pk.impl.qk.get_data(data_index)
// 	pk.impl.parent.main.OpenFileHistory(v.Loc.URI.AsPath().String(), &v.Loc)
// 	pk.impl.parent.hide()
// }

func (pk *refpicker) loadlist(data []*list_tree_node) {
	listview := pk.impl.listcustom
	listview.Key = pk.impl.key
	for i := range data {
		v := data[i]
		listview.AddItem(v.text, "", nil)
	}
	pk.impl.fzf = new_fzf_on_list(listview, true)
}
