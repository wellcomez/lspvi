package mainui

import (
	"fmt"
	"strings"
	"sync/atomic"
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
	query_option  QueryOption
	last          QueryOption
	key           string
	fzf_on_result *fzf_on_listview
	quick         quick_view_data
	livekeydelay  keydelay
}
type quick_view_delegate struct {
	qf func(bool, ref_with_caller) bool
}
type livewgreppicker struct {
	*prev_picker_impl
	tmp_quick_data *quick_view_data
	grep_list_view *customlist
	main           MainService
	impl           *grep_impl
	quick_view     *quick_view_delegate
	not_live       bool
}

// name implements picker.
func (pk *livewgreppicker) name() string {
	return "Live grep"
}



func (pk livewgreppicker) update_preview_tree(index int, prev bool) {
	qk := &pk.impl.quick
	if data, err := qk.get_data(index); err == nil {
		if prev {
			pk.PrevOpen(data.Loc.URI.AsPath().String(), data.Loc.Range.Start.Line)
		} else {
			pk.main.OpenFileHistory(data.Loc.URI.AsPath().String(), &data.Loc)

		}
	}
}
func (pk livewgreppicker) open_view(index int, prev bool) {
	if pk.impl.quick.tree != nil {
		pk.update_preview_tree(index, prev)
	} else {
		pk.update_view_no_tree_at(index, prev)
	}
}

func (pk livewgreppicker) update_view_no_tree_at(cur int, prev bool) bool {
	if pk.impl.fzf_on_result != nil {
		cur = pk.impl.fzf_on_result.get_data_index(cur)
		if cur < 0 {
			return true
		}
	}
	if pk.impl.result == nil {
		return true
	}
	if cur < len(pk.impl.result.data) {
		item := pk.impl.result.data[cur]
		fpath := item.Loc.URI.AsPath().String()
		lineNumber := item.Loc.Range.Start.Line
		if prev {
			pk.PrevOpen(fpath, lineNumber-1)
		} else {
			pk.main.OpenFileHistory(fpath, &item.Loc)
		}
	}
	return false
}

// handle implements picker.
func (pk *livewgreppicker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyUp, tcell.KeyDown:
			pk.grep_list_view.InputHandler()(event, setFocus)
		}
	}
}

func (pk *livewgreppicker) grid(input *tview.InputField) *tview.Flex {
	layout := pk.prev_picker_impl.flex(input, 1)
	pk.listcustom.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		pk.open_view(index, true)
		pk.update_title()
	})
	x := tview.NewFlex()
	x.SetDirection(tview.FlexRow)
	file_include := tview.NewInputField()
	if style := global_theme.get_color("selection"); style != nil {
		fg, bg, _ := style.Decompose()
		file_include.SetFieldBackgroundColor(bg)
		file_include.SetFieldTextColor(fg)
	} else {
		file_include.SetFieldBackgroundColor(tcell.ColorDarkGrey)
	}
	file_include.SetPlaceholderStyle(tcell.StyleDefault.Background(tcell.ColorDarkGrey))
	// file_include.SetPlaceholder(global_prj_root)
	file_include.SetChangedFunc(func(text string) {
		pk.impl.query_option.include_pattern = text
		debug.DebugLog("dialog", text)
	})
	file_include.SetBackgroundColor(tcell.ColorBlack)
	var searchIcon = fmt.Sprintf("%c", '\ue68f')
	// searchIcon = "" // Search icon from Nerd Fonts
	// searchIcon = fmt.Sprintf("%c %c %c %c", '\uF15B','\ue731','\uf0b0','\uf15c')+fmt.Sprintf("%c",'\uea6d')
	// pk.listcustom.AddItem(searchIcon, "", nil)
	search_btn := tview.NewButton(searchIcon)
	set_color := func(btn *tview.Button) {
		btn.SetTitleAlign(tview.AlignCenter)
		btn.SetActivatedStyle(tcell.StyleDefault.Foreground(tview.Styles.ContrastBackgroundColor).Background(tview.Styles.PrimitiveBackgroundColor))
		btn.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorDarkGray).Background(tview.Styles.PrimitiveBackgroundColor))
	}
	set_color(search_btn)
	search_btn.SetSelectedFunc(func() {
		pk.impl.fzf_on_result = nil
		pk.parent.input.SetLabel(">")
		text := pk.parent.input.GetText()
		pk.UpdateQuery(text)
	})
	search_btn.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if InRect(event, search_btn) {
			switch action {
			case tview.MouseLeftClick, tview.MouseLeftDown:
				text := pk.parent.input.GetText()
				pk.UpdateQuery(text)
			case tview.MouseLeftDoubleClick:
				debug.DebugLog("dialog", "xxxxxxxxxxxxxxxx", "double")
				return tview.MouseConsumed, nil
			}
		}
		return action, event
	})

	var saveIcon = fmt.Sprintf("%c", '\uf0c7') // Floppy disk emoji
	save_btn := tview.NewButton(saveIcon)
	save_btn.SetSelectedFunc(func() {
		pk.Save()
	})
	set_color(save_btn)
	input_filter := tview.NewFlex()
	input_filter.SetDirection(tview.FlexColumn)
	x1 := tview.NewButton(fmt.Sprintf("%c", '\ueae5'))
	x1.SetTitleAlign(tview.AlignCenter)
	set_color(x1)
	input_filter.
		AddItem(x1, 3, 0, false).
		AddItem(file_include, 0, 9, false).
		AddItem(search_btn, 2, 0, false).
		AddItem(save_btn, 2, 0, false)

	x.AddItem(layout, 0, 10, false).AddItem(input_filter, 1, 1, false)
	return x
}

//	func (pk *livewgreppicker) grid2(input *tview.InputField) *tview.Flex {
//		layout := pk.prev_picker_impl.flex(input, 1)
//		pk.list_click_check.on_list_selected = func() {
//			pk.update_preview()
//			pk.update_title()
//		}
//		return layout
//	}

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
	grep.impl.livekeydelay.grepx = grep
	x.use_cusutom_list(grep.grep_list_view)
	impl.quick = quick_view_data{main: v.main}
	grep.grep_list_view.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		grep.open_view(i, false)
		v.hide()
	})
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
		x = grepx.grep_list_view.GetItemCount()
	}
	if x > 0 {
		index = index + 1
	}
	Type := "LiveGrep"
	if grepx.not_live {
		Type = "Search in Files"
	}
	s := fmt.Sprintf("%s %s %d/%d", Type, grepx.grep_list_view.Key, index, x)
	grepx.parent.update_dialog_title(s)
}

func (grepx *livewgreppicker) grep_to_list(end bool) {
	if !end {
		grepx.update_list_druring_grep()
	} else {
		grepx.update_list_druring_final()
	}
}
func (impl *livewgreppicker) update_preview() {
	impl.open_view(impl.grep_list_view.GetCurrentItem(), true)
}
func (grepx *livewgreppicker) update_list_druring_final() {
	grep := grepx.impl
	tmp := grep.temp
	if tmp != nil {
		grep.temp = nil
		grep.result.data = append(grep.result.data, tmp.data...)
	}
	if grepx.not_live {
		grepx.impl.fzf_on_result = new_fzf_on_list(grepx.grep_list_view, true)
		grepx.impl.fzf_on_result.selected = func(dataindex, listindex int) {
			var o *ref_with_caller
			qk := &grepx.impl.quick
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
	} else {
		task := grepx.impl.taskid
		sss := grepx.impl.last.query
		Refs := grep.result.data
		go func() {
			livegreptag := fmt.Sprint("LG-UI", sss, "-", len(Refs), task)
			debug.DebugLog(livegreptag, "start to trelist")
			defer debug.DebugLog(livegreptag, "end to treelist")
			main := grepx.main
			qk := new_quikview_data(main, data_grep_word, main.current_editor().Path(), Refs)
			if grepx.tmp_quick_data != nil {
				debug.DebugLog(livegreptag, grepx.tmp_quick_data)
				grepx.tmp_quick_data.abort = true
			}
			grepx.tmp_quick_data = qk
			debug.DebugLog(livegreptag, "treen-begin")
			data := qk.tree_to_listemitem(global_prj_root)
			if qk.abort {
				debug.DebugLog(livegreptag, "=======abort-1")
				return
			}
			debug.DebugLog(livegreptag, "treen-end")
			if task != grepx.impl.taskid {
				debug.DebugLog(livegreptag, "=======abort-2")
				return
			}
			grep.quick = *qk
			view := grepx.grep_list_view
			view.Clear()
			for i := range data {
				v := data[i]
				if task != grepx.impl.taskid {
					debug.DebugLog(livegreptag, "=======abort-3")
					return
				}
				view.AddItem(v.text, "", nil)
			}
			grepx.tmp_quick_data = nil
			main.App().QueueUpdateDraw(func() {
				grepx.update_preview()
				grepx.update_title()
			})
		}()
	}
}

type keydelay struct {
	grepx   *livewgreppicker
	waiting atomic.Bool
}

func (k *keydelay) OnKey(s string) {
	if k.waiting.Load() {
		debug.DebugLog("LG", "Exit", s, k.grepx.impl.query_option.query)
		return
	} else {
		debug.DebugLog("LG", "Start", s, k.grepx.impl.query_option.query)
	}
	go func() {
		k.waiting.Store(true)
		defer k.waiting.Store(false)
		if len(s) == 1 {
			<-time.After(time.Microsecond * 50)
		} else {
			<-time.After(time.Microsecond * 10)
		}
		if len(k.grepx.impl.query_option.query) == 0 {
			k.grepx.grep_list_view.Clear()
		} else {
			debug.DebugLog("LG", "run", s, k.grepx.impl.query_option.query)
			k.grepx.__updatequery(k.grepx.impl.query_option)
		}
	}()
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
		if grepx.quick_view != nil {
			grepx.quick_view.end_of_update_ui()
		} else {
			grepx.grep_to_list(true)
		}
	} else if grepx.quick_view != nil {
		grepx.quick_view.update_ui(o, grepx)
	} else {
		grepx.update_ui(o)
	}

}
func (v quick_view_delegate) end_of_update_ui() {
	v.qf(true, ref_with_caller{})
}
func (v quick_view_delegate) update_ui(o *grep.GrepOutput, grepx *livewgreppicker) {
	ref := to_ref_caller(grepx.impl.key, o)
	if !v.qf(false, ref) {
		grepx.stop_grep()
	}
}

func (grepx *livewgreppicker) update_ui(o *grep.GrepOutput) {
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
	if !grepx.parent.Visible {
		grepx.stop_grep()
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
}

// func convert_grep_info_location(o *grep.GrepOutput) lsp.Location {
// 	loc := lsp.Location{
// 		URI: lsp.NewDocumentURI(o.Fpath),
// 		Range: lsp.Range{
// 			Start: lsp.Position{Line: o.LineNumber - 1, Character: 0},
// 			End:   lsp.Position{Line: o.LineNumber - 1, Character: 0},
// 		},
// 	}
// 	return loc
// }

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
	if len(query) == 0 {
		pk.stop_grep()
		pk.grep_list_view.Clear()
	}
	pk.impl.query_option.query = query
	pk.impl.livekeydelay.OnKey(query)
}

// close implements picker.
func (pk *livewgreppicker) close() {
	if pk.quick_view == nil {
		pk.cq.CloseQueue()
	}
	pk.stop_grep()
}

type QueryOption struct {
	query           string
	include_pattern string
}

func (pk *livewgreppicker) __updatequery(query_option QueryOption) {
	pk.impl.last = query_option
	query := query_option.query
	if len(query) == 0 {
		pk.stop_grep()
		return
	}
	opt := grep.OptionSet{
		Grep_only:     true,
		G:             true,
		Wholeword:     true,
		IcludePattern: query_option.include_pattern,
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
		pk.stop_grep()
	}
	impl.grep = g
	impl.result = &grepresult{}
	g.CB = pk.end
	chans := g.Kick(global_prj_root)
	go g.Report(chans, false)

}

func (pk *livewgreppicker) stop_grep() {
	if pk.impl != nil && pk.impl.grep != nil {
		pk.impl.grep.Abort()
	}
}
