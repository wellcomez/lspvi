package mainui

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	"zen108.com/lspvi/pkg/debug"
	lspcore "zen108.com/lspvi/pkg/lsp"
	"zen108.com/lspvi/pkg/ui/grep"
)

type grepresult struct {
	data []ref_with_caller
}
type grep_impl struct {
	result        *grepresult
	temp          *grepresult
	grep          *grep.Gorep
	taskid        int
	key           string
	fzf_on_result *fzf_on_listview
	quick         quick_view
}
type livewgreppicker struct {
	*prev_picker_impl
	grep_list_view *customlist
	main           MainService
	impl           *grep_impl
	qf             func(bool, ref_with_caller) bool
	not_live       bool
}

// name implements picker.
func (pk *livewgreppicker) name() string {
	return "live grep"
}

// greppicker
type greppicker struct {
	*livewgreppicker
	query string
}

// close implements picker.
func (g *greppicker) close() {
	if g.qf == nil {
		g.cq.CloseQueue()
	}
	g.livewgreppicker.close()
}

// UpdateQuery implements picker.
// Subtle: this method shadows the method (*livewgreppicker).UpdateQuery of greppicker.livewgreppicker.
func (g *greppicker) UpdateQuery(query string) {
	g.query = query
	if g.impl.fzf_on_result != nil {
		g.impl.fzf_on_result.OnSearch(query, true)
		g.grep_list_view.Key = query
	}
}

// handle implements picker.
// Subtle: this method shadows the method (*livewgreppicker).handle of greppicker.livewgreppicker.
func (g *greppicker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		focused := g.grep_list_view.HasFocus()
		var key = event.Key()
		if key == tcell.KeyEnter && !focused {
			if g.impl.key != g.query {
				RunQuery(g)
			} else {
				g.grep_list_view.InputHandler()(event, nil)
			}
		} else if key == tcell.KeyCtrlS {
			g.Save()
		} else {
			switch key {
			case tcell.KeyDown, tcell.KeyUp:
				g.grep_list_view.List.Focus(nil)
			case tcell.KeyEnter:
				g.grep_list_view.List.InputHandler()(event, nil)
			}
			g.livewgreppicker.handle_key_override(event, nil)
		}
	}
}

func RunQuery(g *greppicker) {
	g.impl.fzf_on_result = nil
	g.parent.input.SetLabel(">")
	g.livewgreppicker.UpdateQuery(g.query)
}

// name implements picker.
func (g *greppicker) name() string {
	return "grep word"
}

func (pk livewgreppicker) update_preview_tree() {
	qk := &pk.impl.quick
	ind := qk.view.GetCurrentItem()
	if data, err := qk.get_data(ind); err == nil {
		pk.PrevOpen(data.Loc.URI.AsPath().String(), data.Loc.Range.Start.Line)
	}
}
func (pk livewgreppicker) update_preview() {
	if pk.impl.quick.tree != nil {
		pk.update_preview_tree()
	} else {
		pk.update_preview_no_tree()
	}
}
func (pk livewgreppicker) update_preview_no_tree() {
	cur := pk.grep_list_view.GetCurrentItem()
	if pk.impl.fzf_on_result != nil {
		cur = pk.impl.fzf_on_result.get_data_index(cur)
		if cur < 0 {
			return
		}
	}
	if pk.impl.result == nil {
		return
	}
	if cur < len(pk.impl.result.data) {
		item := pk.impl.result.data[cur]
		fpath := item.Loc.URI.AsPath().String()
		lineNumber := item.Loc.Range.Start.Line
		pk.PrevOpen(fpath, lineNumber-1)
	}
}

func (pk livewgreppicker) handle_key_override(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	pk.update_preview()
	pk.update_title()
}

// handle implements picker.
func (pk *livewgreppicker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return pk.handle_key_override
}

func (pk *livewgreppicker) grid(input *tview.InputField) *tview.Flex {
	layout := pk.prev_picker_impl.flex(input, 1)
	pk.listview.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		pk.update_preview()
		pk.update_title()
	})
	x := tview.NewFlex()
	x.SetDirection(tview.FlexRow)

	file_include := tview.NewInputField()
	file_include.SetFieldBackgroundColor(tcell.ColorDarkGrey)
	file_include.SetPlaceholderStyle(tcell.StyleDefault)
	file_include.SetLabel("include path ")
	file_include.SetPlaceholder(global_prj_root)
	file_include.SetChangedFunc(func(text string) {
		debug.DebugLog("dialog", text)
	})
	file_include.SetBackgroundColor(tcell.ColorBlack)
	btn := tview.NewButton("Find")
	btn.SetSelectedFunc(func() {
		pk.impl.fzf_on_result = nil
		pk.parent.input.SetLabel(">")
		text := pk.parent.input.GetText()
		pk.UpdateQuery(text)
	})
	btn.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if InRect(event, btn) {
			switch action {
			case tview.MouseLeftClick:
				debug.DebugLog("dialog", "xxxxxxxxxxxxxxxx", "lef click")
			case tview.MouseLeftDown:
				text := pk.parent.input.GetText()
				pk.UpdateQuery(text)
				debug.DebugLog("dialog", "xxxxxxxxxxxxxxxx", "down")
			case tview.MouseLeftDoubleClick:
				debug.DebugLog("dialog", "xxxxxxxxxxxxxxxx", "double")
				return tview.MouseConsumed, nil
			}
		}
		return action, event
	})
	input_filter := tview.NewGrid()
	input_filter.
		AddItem(file_include, 0, 0, 1, 20, 1, 20, false).
		AddItem(btn, 0, 21, 1, 4, 1, 4, false)

	x.AddItem(layout, 0, 10, false).AddItem(input_filter, 1, 1, false)
	return x
}

func (pk *livewgreppicker) grid2(input *tview.InputField) *tview.Flex {
	layout := pk.prev_picker_impl.flex(input, 1)
	pk.list_click_check.on_list_selected = func() {
		pk.update_preview()
		pk.update_title()
	}
	return layout
}
func new_grep_picker(v *fzfmain, code CodeEditor) *greppicker {
	grep := &greppicker{
		livewgreppicker: new_live_grep_picker(v, code),
	}
	grep.not_live = true
	grep.listview.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		grep.impl.quick.selection_handle(i, s1, s2, r)
		v.hide()
	})
	return grep
}
func new_live_grep_picker(v *fzfmain, code CodeEditor) *livewgreppicker {
	main := v.main
	x := new_preview_picker(v)
	impl := &grep_impl{taskid: int(time.Now().Second()) * 100}
	grep := &livewgreppicker{
		prev_picker_impl: x,
		grep_list_view:   new_customlist(false),
		main:             main,
		impl:             impl,
		not_live:         false,
	}
	x.use_cusutom_list(grep.grep_list_view)
	impl.quick.view = grep.grep_list_view
	impl.quick.main = v.main
	impl.quick.cq = NewCodeOpenQueue(code, v.main)
	impl.quick.quickview = new_quick_preview()
	v.Visible = true
	return grep
}
func (grepx *livewgreppicker) update_title() {
	if grepx.impl.result == nil {
		return
	}
	index := grepx.grep_list_view.GetCurrentItem()
	x := len(grepx.impl.result.data)
	if grepx.impl.quick.tree != nil {
		x = grepx.impl.quick.view.GetItemCount()
	}
	if x > 0 {
		index = index + 1
	}
	s := fmt.Sprintf("Grep %s %d/%d", grepx.grep_list_view.Key, index, x)
	grepx.parent.update_dialog_title(s)
}

func (parent *fzfmain) update_dialog_title(s string) {
	UpdateTitleAndColor(parent.Frame.Box, s)
}
func (grepx *livewgreppicker) grep_to_list(end bool) {
	if !end {
		grepx.update_list_druring_grep()
	} else {
		grepx.update_list_druring_final()
	}
}
func (grepx *livewgreppicker) update_list_druring_final() {
	grep := grepx.impl
	tmp := grep.temp
	if tmp != nil {
		grep.temp = nil
		grep.result.data = append(grep.result.data, tmp.data...)
	}
	Refs := grep.result.data
	qk := &grep.quick
	qk.view.Clear()
	qk.Refs.Refs = grep.result.data
	qk.tree = &list_view_tree_extend{filename: qk.main.current_editor().Path()}
	qk.tree.build_tree(Refs)
	data := qk.tree.BuildListStringGroup(qk, global_prj_root, qk.main.Lspmgr())
	for i := range data {
		v := data[i]
		if v.ref_index > 0 {

		}
		qk.view.AddItem(v.text, "", func() {
			grepx.parent.hide()
		})
	}
	grepx.update_preview()
	grepx.update_title()
	grepx.main.App().ForceDraw()
}

func (grepx *livewgreppicker) update_list_druring_grep() {
	grep := grepx.impl
	openpreview := len(grep.result.data) == 0
	tmp := grep.temp
	if tmp == nil {
		return
	}
	grep.temp = nil
	grep.result.data = append(grep.result.data, tmp.data...)
	if len(grep.result.data) > 500 {
		return
	}
	for _, o := range tmp.data {
		fpath := o.Loc.URI.AsPath().String()
		line := o.get_code(0)
		lineNumber := o.Loc.Range.Start.Line
		path := trim_project_filename(fpath, global_prj_root)
		data := fmt.Sprintf("%s:%d %s", path, lineNumber, line)
		grepx.grep_list_view.AddItem(data, "", func() {
			grepx.main.OpenFileHistory(path, &o.Loc)
			grepx.parent.hide()
		})
	}
	if openpreview {
		grepx.update_preview()
	}
	grepx.update_title()
	grepx.main.App().ForceDraw()
}
func (grepx *livewgreppicker) end(task int, o *grep.GrepOutput) {
	if task != grepx.impl.taskid {
		return
	}
	if o == nil {
		if grepx.qf != nil {
			grepx.qf(true, ref_with_caller{})
		} else if grepx.not_live {
			grepx.main.App().QueueUpdate(func() {
				grepx.grep_to_list(true)
				grepx.impl.fzf_on_result = new_fzf_on_list(grepx.grep_list_view, true)
				grepx.impl.fzf_on_result.selected = func(dataindex, listindex int) {
					var o *ref_with_caller
					qk := grepx.impl.quick
					if qk.tree != nil {
						if s, err := qk.get_data(dataindex); err == nil {
							o = s
						}
					} else {
						o = &grepx.impl.result.data[dataindex]
					}
					if o != nil {
						grepx.main.OpenFileHistory(o.Loc.URI.AsPath().String(), &o.Loc)
					}
					grepx.parent.hide()
				}
			})
		}
		return
	}
	grep := grepx.impl
	if grep.result == nil {
		grep.result = &grepresult{
			data: []ref_with_caller{},
		}
	}
	if grep.temp == nil {
		grep.temp = &grepresult{
			data: []ref_with_caller{},
		}
	}
	// log.Printf("end %d %s", task, o.destor)
	if grepx.qf == nil {
		if !grepx.parent.Visible {
			grep.grep.Abort()
			return
		}
		grep.temp.data = append(grep.temp.data, to_ref_caller(grepx.impl.key, o))
		if len(grep.result.data) > 10 {
			if len(grep.temp.data) < 50 {
				return
			}
		}
		grepx.main.App().QueueUpdate(func() {
			grepx.grep_to_list(false)
		})
	} else {
		ref := to_ref_caller(grepx.impl.key, o)
		if !grepx.qf(false, ref) {
			grep.grep.Abort()
		}
	}

}

func convert_grep_info_location(o *grep.GrepOutput) lsp.Location {
	loc := lsp.Location{
		URI: lsp.NewDocumentURI(o.Fpath),
		Range: lsp.Range{
			Start: lsp.Position{Line: o.LineNumber - 1, Character: 0},
			End:   lsp.Position{Line: o.LineNumber - 1, Character: 0},
		},
	}
	return loc
}

func to_ref_caller(key string, o *grep.GrepOutput) ref_with_caller {
	b := strings.Index(o.Line, key)
	e := b + len(key)
	// sss := o.Line[b:e]
	// log.Println(sss)
	start := lsp.Position{Line: o.LineNumber - 1, Character: b}
	end := start
	end.Character = e
	ref := ref_with_caller{
		Loc: lsp.Location{
			URI: lsp.NewDocumentURI(o.Fpath),
			Range: lsp.Range{
				Start: start,
				End:   end,
			},
		},
		Grep:   *o.GrepInfo,
		IsGrep: true,
	}
	return ref
}

func (pk livewgreppicker) Save() {
	Result := search_reference_result{}
	data := qf_history_data{Type: data_grep_word,
		Key:  lspcore.SymolSearchKey{Key: pk.impl.key},
		Date: time.Now().Unix(),
	}
	Result.Refs = append(Result.Refs, pk.impl.result.data...)
	data.Result = Result
	main := pk.main
	main.save_qf_uirefresh(data)
}

func (pk livewgreppicker) UpdateQuery(query string) {
	// pk.defer_input.start(query)
	pk.__updatequery(query)
}

// close implements picker.
func (pk *livewgreppicker) close() {
	if pk.qf == nil {
		pk.cq.CloseQueue()
	}
	pk.stop_grep()
}

func (pk *livewgreppicker) __updatequery(query string) {
	if len(query) == 0 {
		pk.stop_grep()
		return
	}
	opt := grep.OptionSet{
		Grep_only: true,
		G:         true,
		Wholeword: true,
	}
	pk.impl.taskid++
	pk.impl.key = query
	pk.grep_list_view.Key = query
	pk.grep_list_view.Clear()
	g, err := grep.NewGorep(pk.impl.taskid, query, &opt)
	if err != nil {
		return
	}
	impl := pk.impl
	if impl.grep != nil {
		impl.grep.Abort()
	}
	impl.grep = g
	impl.result = &grepresult{}
	g.CB = pk.end
	chans := g.Kick(global_prj_root)
	go g.Report(chans, false)

}

func (pk livewgreppicker) stop_grep() {
	if pk.impl != nil && pk.impl.grep != nil {
		pk.impl.grep.Abort()
	}
}
