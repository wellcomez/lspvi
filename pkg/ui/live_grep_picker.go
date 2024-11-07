// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

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
	result        grepresult
	temp          grepresult
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
	file_include   *tview.InputField
	not_live       bool
	grepword       bool
	filecounter    int
	grep_progress  func(p grep.GrepProgress)
}

// name implements picker.
func (pk *livewgreppicker) name() string {
	return "Live grep"
}

func (pk livewgreppicker) open_view_from_tree(index int, prev bool) {
	qk := &pk.impl.quick
	if data, err := qk.get_data(index); err == nil {
		if prev {
			pk.PrevOpen(data.Loc.URI.AsPath().String(), data.Loc.Range.Start.Line)
		} else {
			pk.parent.open_in_edior(data.Loc)
		}
	}
}
func (pk livewgreppicker) open_view(index int, prev bool) {
	if pk.impl.quick.tree != nil {
		pk.open_view_from_tree(index, prev)
	} else {
		pk.open_view_from_normal_list(index, prev)
	}
}

func (pk livewgreppicker) open_view_from_normal_list(cur int, prev bool) bool {
	if pk.impl.fzf_on_result != nil {
		cur = pk.impl.fzf_on_result.get_data_index(cur)
		if cur < 0 {
			return true
		}
	}
	if cur < len(pk.impl.result.data) {
		item := pk.impl.result.data[cur]
		fpath := item.Loc.URI.AsPath().String()
		lineNumber := item.Loc.Range.Start.Line
		debug.DebugLog("live grep", "open normal file", item.Loc)
		if prev {
			pk.PrevOpen(fpath, lineNumber)
		} else {
			pk.parent.open_in_edior(item.Loc)
		}
	}
	return false
}

// handle implements picker.
func (pk *livewgreppicker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyUp, tcell.KeyDown, tcell.KeyEnter:
			pk.grep_list_view.InputHandler()(event, setFocus)
		}
	}
}

func (pk *livewgreppicker) grid(input *tview.InputField) *tview.Flex {
	layout := pk.prev_picker_impl.flex(input, 1)
	x := tview.NewFlex()
	x.SetDirection(tview.FlexRow)
	file_include := tview.NewInputField()
	if style := global_theme.select_style(); style != nil {
		fg, bg, _ := style.Decompose()
		file_include.SetFieldBackgroundColor(bg)
		file_include.SetFieldTextColor(fg)
	} else {
		file_include.SetFieldBackgroundColor(tcell.ColorDarkGrey)
	}
	file_include.SetPlaceholderStyle(tcell.StyleDefault.Background(tcell.ColorDarkGrey))
	// file_include.SetPlaceholder(global_prj_root)
	file_include.SetChangedFunc(func(text string) {
		pk.impl.query_option.PathPattern = text
		pk.__updatequery(pk.impl.query_option)
		debug.DebugLog("dialog", text)
	})
	file_include.SetBackgroundColor(tcell.ColorBlack)
	pk.file_include = file_include
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
	save_btn.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if InRect(event, save_btn) {
			switch action {
			case tview.MouseLeftClick, tview.MouseLeftDown, tview.MouseLeftDoubleClick:
				pk.Save()
				return tview.MouseConsumed, nil
			}
		}
		return action, event
	})
	set_color(save_btn)

	input_filter := tview.NewFlex()
	input_filter.SetDirection(tview.FlexColumn)

	exclude := NewIconButton('\ueae5')
	exclude.selected = pk.impl.query_option.Exclude
	exclude.click = func(b bool) {
		pk.impl.query_option.Exclude = b
		pk.__updatequery(pk.impl.query_option)
	}

	cap := NewIconButton('\ueab1')
	cap.selected = !pk.impl.query_option.Ignorecase
	cap.click = func(b bool) {
		pk.impl.query_option.Ignorecase = !b
		pk.__updatequery(pk.impl.query_option)
	}

	word := NewIconButton('\ueb7e')
	word.selected = pk.impl.query_option.Wholeword
	word.click = func(b bool) {
		pk.impl.query_option.Wholeword = b
		pk.__updatequery(pk.impl.query_option)
	}
	input_filter.
		AddItem(exclude, 2, 0, false).
		AddItem(cap, 2, 0, false).
		AddItem(word, 2, 0, false).
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

func new_live_grep_picker(v *fzfmain, q QueryOption) *livewgreppicker {
	main := v.main
	x := new_preview_picker(v)
	impl := &grep_impl{taskid: int(time.Now().Second()) * 100, query_option: q}
	// impl.query_option.Ignorecase = q.Ignorecase
	grep := &livewgreppicker{
		prev_picker_impl: x,
		grep_list_view:   new_customlist(false),
		main:             main,
		impl:             impl,
		not_live:         false,
	}
	v.input.SetText(q.Query)
	grep.impl.livekeydelay.grepx = grep
	x.use_cusutom_list(grep.grep_list_view)
	impl.quick = quick_view_data{main: v.main, ignore_symbol_resolv: true}
	grep.set_list_handle()
	v.Visible = true
	return grep
}
func (grepx *livewgreppicker) update_title() {
	index := grepx.grep_list_view.GetCurrentItem()
	update_title_with_index(grepx, index)
}

func update_title_with_index(grepx *livewgreppicker, index int) {
	x := grepx.grep_list_view.GetItemCount()
	if x > 0 {
		index = index + 1
	}

	x1 := grepx.parent.input.GetText()
	Type := fmt.Sprintf("LiveGrep %d %s", grepx.filecounter, x1)
	if grepx.not_live {
		Type = fmt.Sprintf("Search %s in Files %d", x1, grepx.filecounter)
	}
	s := fmt.Sprintf("%s %d/%d", Type, index, x)
	grepx.parent.update_dialog_title(s)
}

func (impl *livewgreppicker) update_preview() {
	impl.open_view(impl.grep_list_view.GetCurrentItem(), true)
}
func (grepx *livewgreppicker) update_list_druring_final() {
	if !grepx.not_live || grepx.grepword {
		grepx.end_of_livegrep()
	} else {
		grepx.end_of_grep()
	}
}

func (grepx *livewgreppicker) end_of_grep() {
	grep := grepx.impl
	grep.get_grep_new_data()

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
			grepx.parent.open_in_edior(o.Loc)
		}
	}
}

func (grepx *livewgreppicker) end_of_livegrep() {
	grep := grepx.impl
	grep.get_grep_new_data()

	task := grepx.impl.taskid
	sss := grepx.impl.last.Query
	Refs := grep.result.data
	go func() {
		livegreptag := fmt.Sprint("LG-UI", sss, "-", len(Refs), task)
		debug.DebugLog(livegreptag, "start to trelist")
		defer debug.DebugLog(livegreptag, "end to treelist")
		main := grepx.main
		save := grepx.grepword
		qk := new_quikview_data(main,
			data_grep_word,
			main.current_editor().Path(),
			&SearchKey{
				&lspcore.SymolSearchKey{Key: grep.key},
				&grep.query_option,
			},
			Refs,
			save)
		if grepx.tmp_quick_data != nil {
			debug.DebugLog(livegreptag, "tmp_quick_data", grepx.tmp_quick_data.abort)
			grepx.tmp_quick_data.abort = true
		}
		grepx.tmp_quick_data = qk
		debug.DebugLog(livegreptag, "treen-begin")
		qk.tree_to_listemitem()
		if qk.abort {
			debug.DebugLog(livegreptag, "=======abort-1")
			return
		}
		tree := qk.build_flextree_data(5)
		data := tree.ListColorString()
		debug.DebugLog(livegreptag, "treen-end")
		if task != grepx.impl.taskid {
			debug.DebugLog(livegreptag, "=======abort-2")
			return
		}
		grep.quick = *qk
		if save {
			qk.Save()
		}
		view := grepx.grep_list_view
		view.Clear()
		use_color := true
		for i := range data {
			v := data[i]
			if task != grepx.impl.taskid {
				debug.DebugLog(livegreptag, "=======abort-3")
				return
			}
			view.AddColorItem(v.line, nil, nil)
		}
		lastindex := -1
		view.SetSelectedFunc(func(index int, s1, s2 string, r rune) {
			_, pos, _, parent := tree.GetNodeIndex(index)
			if pos == NodePostion_LastChild {
				if !parent.HasMore() {
					pos = NodePostion_Child
				}

			}
			switch pos {
			case NodePostion_Root:
				{
					if lastindex == index || tree.loadcount == 0 {
						tree.Toggle(parent, use_color)
					}
					go RefreshTreeList(view, tree, index, use_color)
					grepx.update_title()
					return
				}
			case NodePostion_LastChild:
				{
					if parent.HasMore() {
						tree.LoadMore(parent, use_color)
						go RefreshTreeList(view, tree, index, use_color)
					}
				}
			case NodePostion_Child:
				{
					if lastindex == index {
						if caller, err := tree.GetCaller(index); err == nil {
							loc := caller.Loc
							grepx.parent.open_in_edior(loc)
						}
					}
				}
			}

		})
		view.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
			lastindex = index
			if caller, err := tree.GetCaller(index); err == nil {
				loc := caller.Loc
				grepx.PrevOpen(loc.URI.AsPath().String(), loc.Range.Start.Line)
			}
			if _, pos, _, parent := tree.GetNodeIndex(index); pos == NodePostion_LastChild {
				if parent.HasMore() {
					tree.LoadMore(parent, use_color)
					go RefreshTreeList(view, tree, index, use_color)
				}
			}
			grepx.update_title()
		})

		grepx.tmp_quick_data = nil
		main.App().QueueUpdateDraw(func() {
			grepx.update_preview()
			grepx.update_title()
		})
	}()
}

func RefreshTreeList(view *customlist, tree *FlexTreeNodeRoot, index int, usecolor bool) {
	view.Clear()
	if usecolor {
		for _, v := range tree.ColorstringItem {
			view.AddColorItem(v.line, nil, nil)
		}
	} else {
		for _, v := range tree.ListItem {
			view.AddItem(v, "", nil)
		}
	}
	view.SetCurrentItem(index)
}

type keydelay struct {
	grepx   *livewgreppicker
	waiting atomic.Bool
}

// OnKey
func (k *keydelay) OnKey(query string) {
	k.grepx.impl.query_option.Query = query
	if k.waiting.Load() {
		debug.DebugLog("LG", "Exit", query, query)
		return
	} else {
		debug.DebugLog("LG", "Start", query, query)
	}
	go func() {
		k.waiting.Store(true)
		defer k.waiting.Store(false)
		if len(query) == 1 {
			<-time.After(time.Microsecond * 50)
		} else {
			<-time.After(time.Microsecond * 10)
		}
		debug.DebugLog("LG", "run", query, query)
		k.grepx.__updatequery(k.grepx.impl.query_option)
	}()
}
func (impl *grep_impl) get_grep_new_data() (draw bool, data []ref_with_caller) {
	grep := impl
	tmp := grep.temp
	draw = len(tmp.data) > 0
	data = tmp.data
	grep.result.data = append(grep.result.data, tmp.data...)
	grep.temp = grepresult{}
	return
}

// update_list_druring_grep
func (grepx *livewgreppicker) update_list_druring_grep() {
	grep := grepx.impl
	openpreview := len(grep.result.data) == 0
	yes, data := grep.get_grep_new_data()
	if yes {
		for _, o := range data {
			fpath := o.Loc.URI.AsPath().String()
			style := global_theme.search_highlight_color_style()
			line := o.get_code(style)
			lineNumber := o.Loc.Range.Start.Line
			path := trim_project_filename(fpath, global_prj_root)
			data := colorstring{}
			data.add_color_text_list(line.line).prepend(fmt.Sprintf("%s:%d ", path, lineNumber+1), 0)
			if list := grepx.grep_list_view; list != nil {
				list.AddColorItem(data.line, nil, nil)
			}
		}
		if openpreview {
			grepx.update_preview()
		}
		grepx.update_title()
		grepx.main.App().ForceDraw()
	}
}
func (grepx *livewgreppicker) grep_callback(task int, o *grep.GrepOutput) {
	if grepx.quick_view == nil {
		if grepx.parent != nil && !grepx.parent.Visible {
			grepx.stop_grep()
			return
		}
	}
	if task != grepx.impl.taskid {
		return
	}
	if o == nil {
		if grepx.quick_view != nil {
			grepx.quick_view.end_of_update_ui()
		} else {
			grepx.update_list_druring_final()
		}
	} else if grepx.quick_view != nil {
		grepx.quick_view.update_ui(o, grepx)
	} else {
		grep := grepx.impl
		if grep.AddData(o) {
			grepx.main.App().QueueUpdate(func() {
				grepx.update_list_druring_grep()
			})
		}
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

func (grep *grep_impl) AddData(o *grep.GrepOutput) bool {
	grep.temp.data = append(grep.temp.data, to_ref_caller(grep.key, o))
	if len(grep.result.data) > 50 {
		if len(grep.temp.data) < 50 {
			return false
		}
	}
	return true
}

//	func convert_grep_info_location(o *grep.GrepOutput) lsp.Location {
//		loc := lsp.Location{
//			URI: lsp.NewDocumentURI(o.Fpath),
//			Range: lsp.Range{
//				Start: lsp.Position{Line: o.LineNumber - 1, Character: 0},
//				End:   lsp.Position{Line: o.LineNumber - 1, Character: 0},
//			},
//		}
//		return loc
//	}
func DefaultQuery(word string) QueryOption {
	x := QueryOption{grep.OptionSet{Query: word, Wholeword: true, Ignorecase: true}}
	return x
}
func (a QueryOption) Whole(b bool) QueryOption {
	a.Wholeword = b
	return a
}
func (a QueryOption) SetPathPattern(b string) QueryOption {
	a.PathPattern = b
	return a
}
func (a QueryOption) Key(b string) QueryOption {
	a.Query = b
	return a
}
func (a QueryOption) Cap(b bool) QueryOption {
	a.Ignorecase = !b
	return a
}

func to_ref_caller(key string, o *grep.GrepOutput) ref_with_caller {
	b := 0
	if o.X == -1 {
		if ind := strings.Index(o.Line, key); ind >= 0 {
			b = ind
		}
	} else {
		b = o.X
	}
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
		Key:  SearchKey{&lspcore.SymolSearchKey{Key: pk.impl.key}, &pk.impl.last},
		Date: time.Now().Unix(),
	}
	Result.Refs = append(Result.Refs, pk.impl.result.data...)
	data.Result = Result
	main := pk.main
	main.save_qf_uirefresh(data)
}
func (pk livewgreppicker) UpdateQuery(query string) {
	pk.set_list_handle()
	pk.impl.livekeydelay.OnKey(query)
}

func (pk *livewgreppicker) set_list_handle() {
	lastindex := -1
	pk.grep_list_view.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		lastindex = index
		pk.open_view_from_normal_list(index, true)
		update_title_with_index(pk, index)
	})
	pk.grep_list_view.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		if lastindex == i {
			idx := pk.grep_list_view.GetCurrentItem()
			pk.open_view_from_normal_list(idx, false)
		}
	})
}

// close implements picker.
func (pk *livewgreppicker) close() {
	if pk.quick_view == nil {
		// pk.cq.CloseQueue()
	}
	pk.stop_grep()
}

type QueryOption struct {
	grep.OptionSet
}

func (pk *livewgreppicker) __updatequery(query_option QueryOption) {
	if query_option == pk.impl.last {
		if grep := pk.impl.grep; grep != nil {
			if grep.IsRunning() {
				return
			}
		}
	}

	pk.stop_grep()
	var clean_ui = func() {
		pk.grep_list_view.Clear()
		pk.codeprev.Clear()
		pk.grep_list_view.SetCurrentItem(-1)
		pk.update_title()
	}

	var new_query_stack = func() {
		pk.impl.last = query_option
		pk.impl.taskid++
		query := pk.impl.last.Query
		pk.impl.key = query
		pk.grep_list_view.Key = query
		pk.grep_list_view.Clear()
		if query == "" {
			return
		}

		if g, err := grep.NewGorep(pk.impl.taskid, query, pk.impl.last.OptionSet); err == nil {
			impl := pk.impl
			if impl.grep != nil {
				pk.stop_grep()
			}
			impl.grep = g
			impl.result = grepresult{}
			g.CB = pk.grep_callback
			g.GrepProgress(func(p grep.GrepProgress) {
				pk.filecounter = p.FileCount
				if pk.grep_progress != nil {
					pk.grep_progress(p)
				} else {
					go pk.main.App().QueueUpdateDraw(func() {
						pk.update_title()
					})
				}
			})
			g.Kick(global_prj_root)
		}
	}
	go pk.main.App().QueueUpdateDraw(func() {
		clean_ui()
		go new_query_stack()
	})

}

func (pk *livewgreppicker) stop_grep() {
	if pk.impl != nil && pk.impl.grep != nil {
		pk.impl.grep.Abort()
	} else {
		debug.DebugLog("stop_grep grep is nil")
	}
}
