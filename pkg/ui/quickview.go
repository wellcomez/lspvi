// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	// str "github.com/boyter/go-string"
	"github.com/gdamore/tcell/v2"
	// "github.com/reinhrst/fzf-lib"
	fzflib "github.com/reinhrst/fzf-lib"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"

	// "zen108.com/lspvi/pkg/debug"
	"zen108.com/lspvi/pkg/debug"
	lspcore "zen108.com/lspvi/pkg/lsp"
	"zen108.com/lspvi/pkg/ui/grep"
)

type quick_preview struct {
	codeprev CodeEditor
	frame    *tview.Frame
	visisble bool
	// cq       *CodeOpenQueue
	delay opendelay
}

// update_preview
func (pk *quick_preview) update_preview(loc lsp.Location) {
	pk.visisble = true
	title := fmt.Sprintf("%s:%d", loc.URI.AsPath().String(), loc.Range.End.Line+1)
	UpdateTitleAndColor(pk.frame.Box, title)
	// pk.cq.LoadFileNoLsp(, loc.Range.Start.Line)
	pk.delay.OnKey(loc.URI.AsPath().String(), loc.Range.Start.Line)
}
func new_quick_preview(main MainService) *quick_preview {
	codeprev := NewCodeView(main)
	frame := tview.NewFrame(codeprev.view)
	frame.SetBorder(true)
	return &quick_preview{
		codeprev: codeprev,
		frame:    frame,
		delay:    opendelay{code: codeprev, app: main.App()},
	}
}

type SearchKey struct {
	*lspcore.SymolSearchKey
	SearchOption *QueryOption
}

// quick_view
type quick_view struct {
	Type DateType

	data quick_view_data
	// *tview.Flex
	*view_link
	quickview *quick_preview
	view      *customlist
	Name      string
	main      MainService
	// menu         *contextmenu
	menuitem  []context_menu_item
	searchkey SearchKey

	cmd_search_key string
	grep           *greppicker
	sel            *list_multi_select
	flex_tree      *FlexTreeNodeRoot
	cq             *CodeOpenQueue
	currentstate   string
}
type list_view_tree_extend struct {
	root           []list_tree_node
	tree_data_item []*list_tree_node
	filename       string
}

func (l list_view_tree_extend) NeedCreate() bool {
	return len(l.root) == 0
}

type qf_history_data struct {
	Type   DateType
	Key    SearchKey
	Result search_reference_result
	Date   int64
	UID    string
	// SearchOption QueryOption
}

func (main *mainui) save_qf_uirefresh(data qf_history_data) error {
	h, err := new_qf_history(main.Lspmgr().Wk)
	if err != nil {
		return err
	}
	err = h.save_history(global_prj_root, data, true)
	if err == nil {
		main.console_index_list.SetCurrentItem(0)
	}
	main.reload_index_list()
	return err
}
func (h *qf_history_data) ListItem() string {

	return ""
}

func (qk quick_view) save() error {
	date := time.Now().Unix()
	qk.main.save_qf_uirefresh(
		qf_history_data{qk.Type, qk.searchkey, qk.data.Refs, date, ""})
	return nil
}

func (qk *quick_view) RestoreLast() {
	data, _ := qf_history_ReadLast(qk.main.Lspmgr().Wk)
	if data != nil {
		qk.UpdateListView(data.Type, data.Result.Refs, data.Key)
	}
}

type quick_view_context struct {
	qk *quick_view
}

func (menu quick_view_context) on_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	return action, event
}

// getbox implements context_menu_handle.
func (menu quick_view_context) getbox() *tview.Box {
	yes := menu.qk.main.Tab().activate_tab_id == view_quickview
	if yes {
		return menu.qk.view.Box
	}
	return nil
}

// menuitem implements context_menu_handle.
func (menu quick_view_context) menuitem() []context_menu_item {
	return menu.qk.menuitem
}

type fzf_list_item struct {
	maintext, secondText string
}
type fzf_on_listview struct {
	selected_index []int
	// selected_text    []string
	selected_postion [][]int
	fzf              *fzflib.Fzf
	listview         *customlist
	fuzz             bool
	list_data        []fzf_list_item
	selected         func(dataindex int, listindex int)
	query            string
	data             []string
}

func new_fzf_on_list_data(list *customlist, data []string, fuzz bool) *fzf_on_listview {
	ret := &fzf_on_listview{
		listview:       list,
		fuzz:           fuzz,
		selected_index: []int{},
		data:           data,
	}
	opt := fzflib.DefaultOptions()
	opt.Fuzzy = fuzz
	key := []string{}
	for i := 0; i < len(data); i++ {
		a := data[i]
		ret.list_data = append(ret.list_data, fzf_list_item{a, ""})
		key = append(key, a)
		ret.selected_index = append(ret.selected_index, i)
	}
	if len(key) > 0 {
		ret.fzf = fzflib.New(key, opt)
	}
	return ret
}
func new_fzf_on_list(list *customlist, fuzz bool) *fzf_on_listview {
	ret := &fzf_on_listview{
		listview:       list,
		fuzz:           fuzz,
		selected_index: []int{},
	}
	opt := fzflib.DefaultOptions()
	opt.Fuzzy = fuzz
	key := []string{}
	for i := 0; i < list.GetItemCount(); i++ {
		a, b := list.GetItemText(i)
		ret.list_data = append(ret.list_data, fzf_list_item{a, b})
		key = append(key, a+b)
		ret.selected_index = append(ret.selected_index, i)
	}
	if len(key) > 0 {
		ret.fzf = fzflib.New(key, opt)
	}
	return ret
}
func (fzfdata *fzf_on_listview) OnSearchSortScore(txt string, score int) {
	var result fzflib.SearchResult
	fzf := fzfdata.fzf
	if len(txt) > 0 && fzf != nil {
		fzf.Search(txt)
		fzfdata.query = txt
		result = <-fzf.GetResultChannel()
		fzfdata.selected_index = []int{}
		fzfdata.selected_postion = [][]int{}
		// ss := []string{}

		var Matches []fzflib.MatchResult
		for _, v := range result.Matches {
			if v.Score < score {
				continue
			}
			Matches = append(Matches, v)
		}
		sort.Slice(Matches, func(i, j int) bool {
			return Matches[i].Score > Matches[j].Score
		})
		for _, v := range Matches {
			fzfdata.selected_index = append(fzfdata.selected_index, int(v.HayIndex))
			fzfdata.selected_postion = append(fzfdata.selected_postion, v.Positions)
			// ss = append(ss, v.Key)
		}
		// fzfdata.selected_text = ss
	} else {
		fzfdata.reset_selection_index()

	}
}
func (fzfdata *fzf_on_listview) OnSearch(txt string, update bool) string {
	var result fzflib.SearchResult
	old := fzfdata.query
	fzf := fzfdata.fzf
	if len(txt) > 0 && fzf != nil {
		fzf.Search(txt)
		fzfdata.query = txt
		result = <-fzf.GetResultChannel()
		fzfdata.selected_index = []int{}
		fzfdata.selected_postion = [][]int{}
		ss := []string{}
		for _, v := range result.Matches {
			fzfdata.selected_index = append(fzfdata.selected_index, int(v.HayIndex))
			fzfdata.selected_postion = append(fzfdata.selected_postion, v.Positions)
			ss = append(ss, v.Key)
		}
		// fzfdata.selected_text = ss
	} else {
		fzfdata.reset_selection_index()

	}
	if update {
		fzfdata.refresh_list()
	}
	return old
}

func (fzf *fzf_on_listview) reset_selection_index() {
	fzf.selected_index = []int{}
	fzf.selected_postion = [][]int{}
	for i := 0; i < len(fzf.list_data); i++ {
		fzf.selected_index = append(fzf.selected_index, i)
		fzf.selected_postion = append(fzf.selected_postion, []int{})
	}
	// fzf.selected_text = make([]string, len(fzf.selected_index))
}
func (fzf *fzf_on_listview) get_data_index(index int) int {
	if len(fzf.selected_index) == 0 {
		return -1
	}
	if index == -1 {
		return fzf.selected_index[fzf.listview.GetCurrentItem()]
	}
	return fzf.selected_index[index]
}
func (fzf *fzf_on_listview) refresh_list() {
	fzf.listview.Clear()
	if fzf.fuzz {
		fzf.listview.Key = ""
	}
	for i, v := range fzf.selected_index {
		list := i
		data := v
		colors := fzf.selected_postion[i]
		a := fzf.list_data[data]
		s := a.maintext
		s = fzf_color(colors, s)
		fzf.listview.AddItem(s, a.secondText, func() {
			if fzf.selected != nil {
				fzf.listview.SetCurrentItem(list)
				fzf.selected(data, list)
			}
		})
	}
	fzf.listview.SetCurrentItem(0)
}
func fzf_color_pos(colors []int) []Pos {
	sort.Slice(colors, func(i, j int) bool {
		return colors[i] < colors[j]
	})
	var colors2 = []Pos{}
	for _, v := range colors {
		if len(colors2) == 0 {
			colors2 = append(colors2, Pos{v, v + 1})
		} else {
			last := colors2[len(colors2)-1]
			if last.Y == v {
				last.Y = v + 1
				colors2[len(colors2)-1] = last
			} else if v > last.Y {
				colors2 = append(colors2, Pos{v, v + 1})
			}
		}
	}
	return colors2
}
func convert_string_colortext(colors []int, ssss string, normal tcell.Color, hl tcell.Color) (ss []colortext) {
	if hl == 0 {
		hl = tcell.ColorYellow
	}
	s := []rune(ssss)
	if len(colors) <= len(s) {
		var colors2 = fzf_color_pos(colors)
		begin := 0
		for _, v := range colors2 {
			normal_text := s[begin:v.X]
			ss = append(ss, colortext{string(normal_text), normal, 0})
			x := s[v.X:v.Y]
			ss = append(ss, colortext{string(x), hl, 0})
			begin = v.Y
		}
		if begin < len(s) {
			ss = append(ss, colortext{text: string(s[begin:])})
		}
	}
	return
}

//	func fzf_color_with_color(colors []int, s string, normal tcell.Color, hl tcell.Color) string {
//		if hl == 0 {
//			hl = tcell.ColorYellow
//		}
//		if len(colors) < len(s) {
//			ss := []string{}
//			var colors2 = fzf_color_pos(colors, s)
//			begin := 0
//			for _, v := range colors2 {
//				normal_text := s[begin:v.X]
//				if normal != 0 {
//					normal_text = fmt_color_string(normal_text, normal)
//				}
//				ss = append(ss, normal_text)
//				x := s[v.X:v.Y]
//				x = fmt_color_string(x, hl)
//				ss = append(ss, x)
//				begin = v.Y
//			}
//			if begin < len(s) {
//				ss = append(ss, s[begin:])
//			}
//			if len(ss) > 0 {
//				s = strings.Join(ss, "")
//			}
//		}
//		return s
//	}
func fzf_color(colors []int, args string) (ret string) {
	ss := []string{}
	ret = args
	xxx := []rune(args)
	var colors2 = fzf_color_pos(colors)
	begin := 0
	for _, v := range colors2 {
		ss = append(ss, string(xxx[begin:v.X]))
		x := fmt_color_string(string(xxx[v.X:v.Y]), tcell.ColorYellow)
		ss = append(ss, string([]rune(x)))
		begin = v.Y
	}
	if begin < len(args) {
		ss = append(ss, string(xxx[begin:]))
	}
	if len(ss) > 0 {
		ret = strings.Join(ss, "")
	}
	return
}

// new_quikview
func new_quikview(main *mainui) *quick_view {
	view := new_customlist(false)
	view.default_color = tcell.ColorGreen
	view.List.SetMainTextStyle(tcell.StyleDefault.Normal()).ShowSecondaryText(false)
	var vid view_id = view_quickview

	ret := &quick_view{
		view_link: &view_link{id: view_quickview, up: view_code, right: view_callin},
		Name:      vid.getname(),
		view:      view,
		main:      main,
		quickview: new_quick_preview(main),
		cq:        NewCodeOpenQueue(nil, main),
	}

	var right_context quick_view_context
	right_context.qk = ret

	ret.menuitem = []context_menu_item{
		{item: cmditem{Cmd: cmdactor{desc: "Open "}}, handle: func() {
			if view.GetItemCount() > 0 {
				// ret.selection_handle_impl(view.GetCurrentItem(), true)
			}
		}},
		{item: cmditem{Cmd: cmdactor{desc: "Save "}}, handle: func() {
			ret.save()
		}},
		{item: cmditem{Cmd: cmdactor{desc: "Copy"}}, handle: func() {
			ss := ret.sel.list.selected
			var data []*colorstring
			if ret.data.tree == nil {
				data = ret.data.BuildListString("")
			} else {
				x := ret.data.tree.tree_data_item
				for _, v := range x {
					s := colorstring{}
					s.a(v.color_string.plaintext())
					data = append(data, &s)
				}
			}
			if len(ss) > 0 {
				sss := data[ss[0]:ss[1]]
				aa := []string{}
				for _, v := range sss {
					aa = append(aa, v.plaintext())
				}
				data := strings.Join(aa, "\n")
				main.CopyToClipboard(data)
				ret.sel.clear()
				main.app.ForceDraw()
			} else {
				main.CopyToClipboard(data[view.GetCurrentItem()].plaintext())
			}
		}},
	}
	ret.sel = &list_multi_select{list: view}
	main.sel.Add(ret.sel)
	view.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if a, e := main.Right_context_menu().handle_menu_mouse_action(action, event, right_context, view.Box); a == tview.MouseConsumed {
			return a, e
		}
		return action, event
	})
	return ret
}

func remove_color(sss []string) []string {
	aa := []string{}
	for i := range sss {
		r := GetColorText(sss[i], []colortext{})
		text := ""
		for _, v := range r {
			for _, v := range lspcore.IconsRunne {
				text = strings.ReplaceAll(text, fmt.Sprintf("%c", v), "")
			}
			text = text + v.text
		}
		aa = append(aa, text)
	}
	return aa
}

func checkDirExists(dirPath string) bool {
	_, err := os.Stat(dirPath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	// 其他类型的错误
	return false
}

// func (qk *quick_view) DrawPreview(screen tcell.Screen, left, top, width, height int) bool {
// 	qk.quickview.draw(left, top, width, height, screen)
// 	return false
// }

func (qk *quick_preview) draw(left, top, width, height int, screen tcell.Screen) {
	if !qk.visisble {
		return
	}
	frame := qk.frame
	frame.SetRect(left, top, width, height)
	frame.Draw(screen)
}

func (qk *quick_view) OnSearch(txt string) {
	// old_query := qk.cmd_search_key
	// view := qk.view
	// highlight_search_key(old_query, view, txt)
	qk.cmd_search_key = txt
}

// String
func (qk *quick_view) String() string {
	return qk.currentstate
}

func (qk *quick_view) get_current_index_string(index int) string {
	scanned := ""
	if qk.Type == data_grep_word && qk.grep != nil {
		scanned = fmt.Sprintf("%d", qk.grep.filecounter)
	}
	var s = qk.Type.String()
	coutn := qk.view.GetItemCount()
	if coutn > 0 {
		index += 1
	}
	key := qk.searchkey.Key
	if len(scanned) > 0 {
		return fmt.Sprintf("%s %s %s %d/%d", s, key, scanned, index, coutn)
	}
	return fmt.Sprintf("%s %s %d/%d", s, key, index, coutn)
}

type DateType int

const (
	data_search DateType = iota
	data_refs
	data_callin
	data_bookmark
	data_grep_word
	data_implementation
)

func search_key_uid(key SearchKey) string {
	if len(key.File) > 0 {
		return fmt.Sprintf("%s %s:%d:%d", key.Key, key.File, key.Ranges.Start.Line+1, key.Ranges.Start.Character)
	}
	return key.Key
}
func (qk *quick_view) OnLspRefenceChanged(refs []lsp.Location, t DateType, key lspcore.SymolSearchKey) {
	// panic("unimplemented")
	qk.view.Clear()

	var Refs []ref_with_caller
	switch t {
	case data_implementation:
		for _, v := range refs {
			Refs = append(Refs, ref_with_caller{Loc: v})
		}
	case data_refs:
		Refs = get_loc_caller(qk.main, refs, key.Symbol())
	default:
		break
	}

	qk.UpdateListView(t, Refs, SearchKey{&key, nil})
	qk.save()
}
func (qk *quick_view) AddResult(end bool, t DateType, caller ref_with_caller, key SearchKey) {
	if key.Key != qk.searchkey.Key {
		qk.new_search(t, key)
	}
	if end {
		qk.save()
		if len(qk.data.Refs.Refs) < 250 {
			qk.UpdateListView(t, qk.data.Refs.Refs, key)
		}
		return
	}
	qk.data.reset_tree()
	qk.data.Refs.Refs = append(qk.data.Refs.Refs, caller)
	// _, _, width, _ := qk.view.GetRect()
	// caller.width = width
	secondline := caller.ListItem(global_prj_root, false, nil)
	if len(secondline.line) == 0 {
		return
	}
	qk.view.AddColorItem(secondline.prepend(fmt.Sprintf("%3d. ", qk.view.GetItemCount()+1), 0).line, nil, nil)
}

func (qk *quick_view) new_search(t DateType, key SearchKey) {
	qk.view.Clear()
	qk.cmd_search_key = ""
	qk.Type = t
	qk.searchkey = key
	if qk.grep != nil {
		qk.grep.close()
	}
	qk.new_quickview_data(t, nil)
	switch t {
	case data_grep_word, data_search:
		qk.view.Key = key.Key
	default:
		qk.view.Key = ""
	}
	qk.data.reset_tree()
	qk.view.SetSelectedFunc(func(index int, s1, s2 string, r rune) {
		if qk.view.GetCurrentItem() == index {
			qk.quickview.visisble = false
			vvv := qk.data.Refs.Refs[index]
			qk.cq.OpenFileHistory(vvv.Loc.URI.AsPath().String(), &vvv.Loc)
			qk.main.Tab().UpdatePageTitle()
		}
	})
	qk.view.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		vvv := qk.data.Refs.Refs[index]
		qk.cq.OpenFileHistory(vvv.Loc.URI.AsPath().String(), &vvv.Loc)
		qk.main.Tab().UpdatePageTitle()
	})
}

func (qk *quick_view) new_quickview_data(t DateType, Refs []ref_with_caller) {
	qk.data = *new_quikview_data(qk.main, t, qk.main.current_editor().Path(), &qk.searchkey, Refs, true)
}

func (qk *quick_view) UpdateListView(t DateType, Refs []ref_with_caller, key SearchKey) {
	qk.new_search(t, key)
	qk.view.SetCurrentItem(-1)
	qk.new_quickview_data(t, Refs)
	qk.data.go_build_listview_data()
	tree := qk.data.build_flextree_data(10)
	data := tree.ListColorString()
	var loaddata = func(data []colorstring, i int) {
		qk.view.Clear()
		for _, v := range data {
			qk.view.AddColorItem(v.line, nil, nil)
		}
		qk.view.SetCurrentItem(i)
		qk.main.App().ForceDraw()
	}
	qk.flex_tree = tree
	use_color := true
	lastIndex := -1
	qk.view.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		qk.quickview.visisble = false
		item, pos, more, parent := tree.GetNodeIndex(i)
		switch pos {

		case NodePostion_Root:
			{
				if lastIndex == i || tree.loadcount == 0 {
					tree.Toggle(item, use_color)
				}
				loaddata(tree.ColorstringItem, i)
			}
		case NodePostion_LastChild:
			{
				if n, err := tree.GetCaller(i); err == nil {
					qk.cq.OpenFileHistory(n.Loc.URI.AsPath().String(), &n.Loc)
					if more {
						tree.LoadMore(parent, use_color)
						loaddata(tree.ColorstringItem, i)
					}
				}
			}
		case NodePostion_Child:
			{
				if n, err := tree.GetCaller(i); err == nil {
					qk.cq.OpenFileHistory(n.Loc.URI.AsPath().String(), &n.Loc)
				}
			}
		}
		qk.main.App().ForceDraw()
	})
	qk.view.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		debug.DebugLog("quickview", "changed=", index, "currentIndex=", qk.view.GetCurrentItem())
		lastIndex = index
		if ref, err := qk.data.get_data(index); err == nil && ref != nil {
			debug.DebugLog("quickview", "open", ref.Loc)
			qk.currentstate = qk.get_current_index_string(index)
			qk.cq.OpenFileHistory(ref.Loc.URI.AsPath().String(), &ref.Loc)
			qk.main.Tab().UpdatePageTitle()
		}
		if _, pos, more, parent := tree.GetNodeIndex(index); pos == NodePostion_LastChild {
			if more {
				tree.LoadMore(parent, use_color)
				loaddata(tree.ColorstringItem, index)
			}
		}
	})
	loaddata(data, 0)
	qk.main.Tab().UpdatePageTitle()
}

func (caller *ref_with_caller) ListItem(root string, parent bool, prev *ref_with_caller) (ret *colorstring) {
	ret = &colorstring{}
	v := caller.Loc

	path := v.URI.AsPath().String()
	if len(root) > 0 {
		path = trim_project_filename(path, root)
	}
	if parent {
		// if caller.Childrens > 1 {
		// 	return fmt.Sprintf("%s %s", path, fmt_color_string(fmt.Sprint(caller.Childrens), tcell.ColorRed))
		// }
		var bg tcell.Color
		if style := global_theme.select_style(); style != nil {
			_, bg, _ = style.Decompose()
		}
		return ret.add_color_text(colortext{path, tcell.ColorYellow, bg})
		//return ret.a(path)
	}
	funcolor := global_theme.search_highlight_color_style()
	code := caller.get_code(funcolor)
	ret.a(" ")
	if caller.Caller != nil {
		var c1 colortext
		if prev != nil && (prev.Caller.ClassName == caller.Caller.ClassName && prev.Caller.Name == caller.Caller.Name) {
			prefix := strings.Repeat(" ", min(len(prev.Caller.Name+prev.Caller.ClassName), 4))
			// if prev.Caller.ClassName != "" {
			// 	c1 = strings.Repeat(" ", len(prev.Caller.ClassName)+1)
			// }
			// if prev.Caller.Name != "" {
			// 	x = strings.Repeat(" ", len(prev.Caller.Name)+2)
			// }
			return ret.a(fmt.Sprintf(":%-4d %s ", v.Range.Start.Line+1, prefix)).add_color_text_list(code.line)
		} else {
			funcolor := global_theme.search_highlight_color()
			if len(caller.Caller.ClassName) > 0 {
				if c, err := global_theme.get_lsp_color(lsp.SymbolKindClass); err == nil {
					f, _, _ := c.Decompose()
					icon := fmt.Sprintf("%c ", lspcore.IconsRunne[int(lsp.SymbolKindClass)])
					c1 = colortext{fmt.Sprint(icon, caller.Caller.ClassName+" > "), f, 0}
				}
			}
			kind := caller.Caller.Item.Kind
			caller_color := funcolor
			if c, err := global_theme.get_lsp_color(kind); err == nil {
				f, _, _ := c.Decompose()
				caller_color = f
			}
			icon := fmt.Sprintf("%c ", lspcore.IconsRunne[int(lsp.SymbolKindFunction)])
			if x, ok := lspcore.IconsRunne[int(kind)]; ok {
				icon = fmt.Sprintf("%c ", x)
			} else if s, yes := lspcore.LspIcon[int(caller.Caller.Item.Kind)]; yes {
				icon = s
			}
			callname := caller.Caller.Name
			callname = strings.TrimLeft(callname, " ")
			callname = strings.TrimRight(callname, " ")
			callname = icon + callname
			x := colortext{callname + " > ", caller_color, 0}
			liner := fmt.Sprintf(":%-4d ", v.Range.Start.Line+1)
			// ret.a(liner)
			if c1.text != "" {
				return ret.add_color_text(c1).add_color_text(x).a(" ").a(liner).add_color_text_list(code.line)
				// return fmt.Sprintf(":%-4d %s%s %s", v.Range.Start.Line+1, c1, x, line)
			} else {
				return ret.add_color_text(x).a(" ").a(liner).add_color_text_list(code.line)
			}
		}
	} else {
		return ret.a(fmt.Sprintf(":%-4d ", v.Range.Start.Line+1)).add_color_text_list(code.line)
	}
}

func (caller *ref_with_caller) get_code(funcolor tcell.Style) (ret *colorstring) {
	ret = &colorstring{}
	lines := caller.lines

	line := ""
	if caller.IsGrep {
		line = caller.Grep.Line
	}
	v := caller.Loc
	if line == "" {
		if len(lines) == 0 {
			if line == "" {
				if caller.filecache == nil {
					file := caller.LoadLines()
					if file != nil {
						line = file.lines[v.Range.Start.Line]
					}
				} else {
					line = caller.filecache.lines[v.Range.Start.Line]
				}
			}
		} else {
			line = lines[0]
		}
	}
	if v.Range.Start.Line == v.Range.End.Line {
		s := max(v.Range.Start.Character, 0)
		e := max(v.Range.End.Character, 0)
		if len(line) > s && len(line) >= e && s < e {
			a1 := line[:s]
			a := line[s:e]
			a2 := line[e:]
			if len(a1) > 0 && a1[len(a1)-1] == '*' {
				a1 += " "
			}
			if len(a2) > 0 && a2[0] == '*' {
				a2 = " " + a2
			}
			return ret.a(a1).add_string_style(a, funcolor).a(a2)
		}
	}
	return ret.a(line)
}

func (caller *ref_with_caller) LoadLines() *filecache {
	var lines []string
	v := caller.Loc
	source_file_path := v.URI.AsPath().String()
	data, err := os.ReadFile(source_file_path)
	if err != nil {
		return nil
	}
	lines = strings.Split(string(data), "\n")
	caller.filecache = &filecache{lines: lines}
	return caller.filecache
}

func (quickview *quick_view) SearchInProject(opt QueryOption) {
	var main = quickview.main
	rightmenu_select_text := opt.Query
	key := SearchKey{
		&lspcore.SymolSearchKey{
			Key: rightmenu_select_text,
		},
		&opt,
	}
	quickview.new_search(data_grep_word, key)
	add := 0
	buf := []ref_with_caller{}
	buf2 := []ref_with_caller{}
	coping := false

	quickview.grep = main.open_picker_grep(opt, func(end bool, ss ref_with_caller) bool {
		var ret = rightmenu_select_text == key.Key && quickview.Type == data_grep_word
		if ret {
			if add > 1000 {
				return false
			}
			if coping {
				buf2 = append(buf2, ss)
			} else {
				if len(buf2) > 0 {
					buf = append(buf, buf2...)
					buf2 = []ref_with_caller{}
				}
				buf = append(buf, ss)
			}
			if add < 15 || len(buf) >= 100 || end {
				main.App().QueueUpdateDraw(func() {
					coping = true
					for _, s := range buf {
						add++
						quickview.AddResult(end, data_grep_word, s, key)
					}
					main.Tab().UpdatePageTitle()
					buf = []ref_with_caller{}
					coping = false
				})
			}
		}
		return true
	})
	quickview.grep.grep_progress = func(p grep.GrepProgress) {
		quickview.currentstate = quickview.get_current_index_string(quickview.view.GetCurrentItem())
		go main.App().QueueUpdateDraw(func() {
			quickview.main.Tab().UpdatePageTitle()
		})
	}
}
