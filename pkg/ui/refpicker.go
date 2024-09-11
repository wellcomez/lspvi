package mainui

import (
	// "strings"

	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"

	// lsp "github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

func (grepx *prev_picker_impl) update_title(s string) {
	grepx.parent.Frame.SetTitle(s)
}
func (impl *prev_picker_impl) flex(input *tview.InputField, linenum int) *tview.Flex {
	code := impl.codeprev.view
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
	code := impl.codeprev.view
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
func (pk *refpicker) grid(input *tview.InputField) *tview.Grid {
	ret := pk.impl.grid(input, 2)
	pk.impl.codeprev.Load(pk.impl.file.Filename)
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
	codeprev         *CodeView
	parent           *fzfmain
	list_click_check *GridListClickCheck
	on_list_selected func()
	listdata         []ref_line
}

func (impl *prev_picker_impl) use_cusutom_list(l *customlist) {
	impl.listview = l.List
	impl.listcustom = l
}
func (impl *prev_picker_impl) update_preview() {
	cur := impl.listview.GetCurrentItem()
	if cur < len(impl.listdata) {
		item := impl.listdata[cur]
		impl.codeprev.Load(item.loc.URI.AsPath().String())
		impl.codeprev.gotoline(item.loc.Range.Start.Line)
	}
}

type refpicker_impl struct {
	*prev_picker_impl
	file *lspcore.Symbol_file
	refs []ref_with_caller
	fzf  *fzf_on_listview
	key  string
}
type refpicker struct {
	impl *refpicker_impl
}

// close implements picker.
func (pk refpicker) close() {
}

// name implements picker.
func (pk refpicker) name() string {
	return "refs"
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
func (pk refpicker) OnFileChange(file []lsp.Location) {
	panic("unimplemented")
}
func caller_to_listitem(caller *lspcore.CallStackEntry, root string) string {
	if caller == nil {
		return ""
	}
	callerstr := fmt.Sprintf("%s:%d **%-20s**",
		strings.TrimPrefix(
			caller.Item.URI.AsPath().String(), root),
		caller.Item.Range.Start.Line, caller.Name)
	return callerstr
}

type ref_line struct {
	caller string
	code   string
	loc    lsp.Location
	line   string
	path   string
}

func (ref ref_line) fzf_tring() string {
	return ref.String() + ref.caller
}
func (v ref_line) get_loc() (lsp.Location, error) {
	line, err := strconv.Atoi(v.line)
	if err != nil {
		return lsp.Location{}, err
	}
	loc := lsp.Location{
		URI: lsp.NewDocumentURI(v.path),
		Range: lsp.Range{Start: lsp.Position{Line: line},
			End: lsp.Position{Line: line},
		},
	}
	return loc, nil
}

func (ref ref_line) String() string {
	return fmt.Sprintf("%s %s:%d", ref.line, ref.path, ref.loc.Range.Start.Line)
}

type ref_with_caller struct {
	Loc    lsp.Location
	Caller *lspcore.CallStackEntry
	width  int
}

func (pk refpicker) OnLspRefenceChanged(key lspcore.SymolSearchKey, file []lsp.Location) {
	pk.impl.listview.Clear()
	listview := pk.impl.listview
	lsp := pk.impl.parent.main.lspmgr.Current
	pk.impl.refs = pk.impl.codeprev.main.get_loc_caller(file, lsp)
	for i := range pk.impl.refs {
		caller := pk.impl.refs[i]
		v := caller.Loc
		source_file_path := v.URI.AsPath().String()
		data, err := os.ReadFile(source_file_path)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		line := lines[v.Range.Start.Line]
		if len(line) == 0 {
			continue
		}
		gap := 40
		begin := max(0, v.Range.Start.Character-gap)
		end := min(len(line), v.Range.Start.Character+gap)
		path := strings.Replace(v.URI.AsPath().String(), pk.impl.codeprev.main.root, "", -1)
		callinfo := ""
		if caller.Caller != nil {
			callinfo = caller_to_listitem(caller.Caller, pk.impl.parent.main.root)
		}
		secondline := fmt.Sprintf("%s:%d%s", path, v.Range.Start.Line+1, callinfo)
		r := ref_line{
			caller: caller_to_listitem(caller.Caller, pk.impl.parent.main.root),
			loc:    v,
			line:   line,
			path:   path,
		}
		pk.impl.listdata = append(pk.impl.listdata, r)
		impl := pk.impl
		listview.AddItem(secondline, line[begin:end], 0, func() {
			impl.open_location(v)
		})
	}
	pk.impl.key = key.Key
	pk.loadlist()
	pk.update_preview()
}

func (impl *prev_picker_impl) open_location(v lsp.Location) {
	impl.codeprev.main.OpenFile(v.URI.AsPath().String(), &v)
	impl.parent.hide()
}

func (m *mainui) get_loc_caller(file []lsp.Location, lsp *lspcore.Symbol_file) []ref_with_caller {
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
		caller := m.lspmgr.GetCallEntry(v.URI.AsPath().String(), v.Range)
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
	sym := refpicker{
		impl: &refpicker_impl{
			prev_picker_impl: x,
			file:             &clone,
		},
	}
	sym.impl.use_cusutom_list(new_customlist(true))
	sym.impl.codeprev.view.SetBorder(true)
	return sym
}

func new_preview_picker(v *fzfmain) *prev_picker_impl {
	x := &prev_picker_impl{
		listview: tview.NewList(),
		codeprev: NewCodeView(v.main),
		parent:   v,
	}
	return x
}
func (pk *refpicker) load(ranges lsp.Range) {
	pk.impl.file.Handle = *pk
	pk.impl.file.Reference(ranges)
}
func (pk refpicker) handle_key_override(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := pk.impl.listview.InputHandler()
	handle(event, setFocus)
	pk.update_preview()
}

func (pk *refpicker) update_preview() {
	pk.impl.update_preview()
}

// handle implements picker.
func (pk refpicker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return pk.handle_key_override
}
func (pk refpicker) UpdateQuery(query string) {
	query = strings.ToLower(query)
	listview := pk.impl.listcustom
	listview.Clear()
	selected := func(data int, list int) {
		pk.onselected(data, list)
	}
	if fzf := pk.impl.fzf; fzf != nil {
		fzf.selected = selected
		oldkey := fzf.OnSearch(query, true)
		highlight_search_key(oldkey, listview, query)
		pk.update_preview()
	}
}

func (pk refpicker) onselected(data int, list int) {
	index := data
	v := pk.impl.listdata[index]

	pk.impl.codeprev.main.OpenFile(v.loc.URI.AsPath().String(), &v.loc)
	pk.impl.parent.hide()
}

func (pk *refpicker) loadlist() {
	listview := pk.impl.listcustom
	listview.Key = pk.impl.key
	for i := range pk.impl.listdata {
		v := pk.impl.listdata[i]
		listview.AddItem(v.path, v.line, func() {
			pk.onselected(i, i)
		})
	}
	pk.impl.fzf = new_fzf_on_list(listview, true)
}
