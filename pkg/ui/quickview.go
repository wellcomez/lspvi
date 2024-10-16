package mainui

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	// str "github.com/boyter/go-string"
	"github.com/gdamore/tcell/v2"
	"github.com/reinhrst/fzf-lib"
	fzflib "github.com/reinhrst/fzf-lib"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"

	// "zen108.com/lspvi/pkg/debug"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type quick_preview struct {
	codeprev CodeEditor
	frame    *tview.Frame
	visisble bool
	cq       *CodeOpenQueue
}

// update_preview
func (pk *quick_preview) update_preview(loc lsp.Location) {
	pk.visisble = true
	title := fmt.Sprintf("%s:%d", loc.URI.AsPath().String(), loc.Range.End.Line+1)
	UpdateTitleAndColor(pk.frame.Box, title)
	pk.cq.LoadFileNoLsp(loc.URI.AsPath().String(), loc.Range.Start.Line)
}
func new_quick_preview() *quick_preview {
	codeprev := NewCodeView(nil)
	frame := tview.NewFrame(codeprev.view)
	frame.SetBorder(true)
	return &quick_preview{
		codeprev: codeprev,
		frame:    frame,
		cq:       NewCodeOpenQueue(codeprev, nil),
	}
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
	menuitem      []context_menu_item
	searchkey     lspcore.SymolSearchKey
	right_context quick_view_context

	cmd_search_key string
	grep           *greppicker
	sel            *list_multi_select
	flex_tree      *FlexTreeNodeRoot
	cq             *CodeOpenQueue
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
	Key    lspcore.SymolSearchKey
	Result search_reference_result
	Date   int64
	UID    string
}

func (main *mainui) save_qf_uirefresh(data qf_history_data) error {
	h, err := new_qf_history(main)
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

func (qf *quickfix_history) save_history(
	root string,
	data qf_history_data, add bool,
) error {
	dir := qf.qfdir
	uid := ""
	if add {
		uid = search_key_uid(data.Key)
		uid = strings.ReplaceAll(uid, root, "")
		hexbuf := md5.Sum([]byte(uid))
		uid = hex.EncodeToString(hexbuf[:])
		data.UID = uid
	} else {
		uid = data.UID
	}
	filename := filepath.Join(dir, uid+".json")
	if !add {
		if len(data.UID) != 0 {
			return os.Remove(filename)
		} else if data.Type == data_callin {
			return os.RemoveAll(data.Key.File)
		} else {
			return fmt.Errorf("uid is empty")
		}
	}
	buf, error := json.Marshal(data)
	if error != nil {
		return error
	}
	os.WriteFile(qf.last, buf, 0666)
	return os.WriteFile(filename, buf, 0666)
}

type quickfix_history struct {
	Wk    lspcore.WorkSpace
	last  string
	qfdir string
}

func (qk *quick_view) RestoreLast() {
	data, _ := qk.ReadLast()
	if data != nil {
		qk.UpdateListView(data.Type, data.Result.Refs, data.Key)
	}
}
func (qk *quick_view) ReadLast() (*qf_history_data, error) {
	h, err := new_qf_history(qk.main)
	if err != nil {
		return nil, err
	}
	buf, err := os.ReadFile(h.last)
	if err != nil {
		return nil, err
	}
	var ret qf_history_data
	err = json.Unmarshal(buf, &ret)
	if err == nil {
		return &ret, nil

	}
	return nil, err
}
func (h *quickfix_history) Load() ([]qf_history_data, error) {
	var ret = []qf_history_data{}
	dir, err := h.InitDir()
	if err != nil {
		return ret, err
	}
	dirs, err := os.ReadDir(dir)
	if err != nil {
		return ret, err
	}
	for _, v := range dirs {
		if v.IsDir() {
			continue
		}
		filename := filepath.Join(dir, v.Name())
		buf, err := os.ReadFile(filename)
		if err != nil {
			continue
		}
		var result qf_history_data
		err = json.Unmarshal(buf, &result)
		if err != nil {
			continue
		}
		ret = append(ret, result)
	}
	umlDir := filepath.Join(h.Wk.Export, "uml")
	dirs, err = os.ReadDir(umlDir)
	if err != nil {
		return ret, err
	}
	for _, dir := range dirs {
		var result = qf_history_data{
			Type: data_callin,
			Key: lspcore.SymolSearchKey{
				Key:  dir.Name(),
				File: filepath.Join(umlDir, dir.Name()),
			},
		}
		ret = append(ret, result)

	}
	return ret, nil
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
	selected_index   []int
	selected_text    []string
	selected_postion [][]int
	fzf              *fzf.Fzf
	listview         *customlist
	fuzz             bool
	list_data        []fzf_list_item
	selected         func(dataindex int, listindex int)
	query            string
}

func new_fzf_on_list_data(list *customlist, data []string, fuzz bool) *fzf_on_listview {
	ret := &fzf_on_listview{
		listview:       list,
		fuzz:           fuzz,
		selected_index: []int{},
	}
	opt := fzf.DefaultOptions()
	opt.Fuzzy = fuzz
	key := []string{}
	for i := 0; i < len(data); i++ {
		a := data[i]
		ret.list_data = append(ret.list_data, fzf_list_item{a, ""})
		key = append(key, a)
		ret.selected_index = append(ret.selected_index, i)
	}
	if len(key) > 0 {
		ret.fzf = fzf.New(key, opt)
	}
	return ret
}
func new_fzf_on_list(list *customlist, fuzz bool) *fzf_on_listview {
	ret := &fzf_on_listview{
		listview:       list,
		fuzz:           fuzz,
		selected_index: []int{},
	}
	opt := fzf.DefaultOptions()
	opt.Fuzzy = fuzz
	key := []string{}
	for i := 0; i < list.GetItemCount(); i++ {
		a, b := list.GetItemText(i)
		ret.list_data = append(ret.list_data, fzf_list_item{a, b})
		key = append(key, a+b)
		ret.selected_index = append(ret.selected_index, i)
	}
	if len(key) > 0 {
		ret.fzf = fzf.New(key, opt)
	}
	return ret
}
func (fzf *fzf_on_listview) OnSearch(txt string, update bool) string {
	var result fzflib.SearchResult
	old := fzf.query
	if len(txt) > 0 && fzf.fzf != nil {
		fzf.fzf.Search(txt)
		fzf.query = txt
		result = <-fzf.fzf.GetResultChannel()
		fzf.selected_index = []int{}
		fzf.selected_postion = [][]int{}
		ss := []string{}
		for _, v := range result.Matches {
			fzf.selected_index = append(fzf.selected_index, int(v.HayIndex))
			fzf.selected_postion = append(fzf.selected_postion, v.Positions)
			ss = append(ss, v.Key)
		}
		fzf.selected_text = ss
	} else {
		fzf.reset_selection_index()

	}
	if update {
		fzf.refresh_list()
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
	fzf.selected_text = make([]string, len(fzf.selected_index))
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
func fzf_color_pos(colors []int, s string) []Pos {
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
			} else {
				colors2 = append(colors2, Pos{v, v + 1})
			}
		}
	}
	return colors2
}
func fzf_color_with_color(colors []int, s string, normal tcell.Color, hl tcell.Color) string {
	if hl == 0 {
		hl = tcell.ColorYellow
	}
	if len(colors) < len(s) {
		ss := []string{}
		var colors2 = fzf_color_pos(colors, s)
		begin := 0
		for _, v := range colors2 {
			normal_text := s[begin:v.X]
			if normal != 0 {
				normal_text = fmt_color_string(normal_text, normal)
			}
			ss = append(ss, normal_text)
			x := s[v.X:v.Y]
			x = fmt_color_string(x, hl)
			ss = append(ss, x)
			begin = v.Y
		}
		if begin < len(s) {
			ss = append(ss, s[begin:])
		}
		if len(ss) > 0 {
			s = strings.Join(ss, "")
		}
	}
	return s
}
func fzf_color(colors []int, s string) string {
	ss := []string{}
	var colors2 = fzf_color_pos(colors, s)
	begin := 0
	for _, v := range colors2 {
		ss = append(ss, s[begin:v.X])
		x := s[v.X:v.Y]
		x = fmt_color_string(x, tcell.ColorYellow)
		ss = append(ss, x)
		begin = v.Y
	}
	if begin < len(s) {
		ss = append(ss, s[begin:])
	}
	if len(ss) > 0 {
		s = strings.Join(ss, "")
	}
	return s
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
		quickview: new_quick_preview(),
		cq:        NewCodeOpenQueue(nil, main),
	}

	// ret.Flex = layout
	ret.right_context.qk = ret
	// view.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	// 	is_rightclick := (action == tview.MouseRightClick)
	// 	menu := ret.menu
	// 	action, event = menu.handle_mouse(action, event)
	// 	if ret.menu.visible && is_rightclick {
	// 		_, y, _, _ := view.GetRect()
	// 		index := (menu.MenuPos.y - y)
	// 		log.Println("index in list:", index)
	// 		view.SetCurrentItem(index)
	// 	}
	// 	return action, event
	// })
	view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		ch := event.Rune()
		if ch == 'j' || event.Key() == tcell.KeyDown {
			ret.go_next()
		} else if ch == 'k' || event.Key() == tcell.KeyUp {
			ret.go_prev()
		} else {
			return event
		}
		return nil
	})
	view.SetSelectedFunc(ret.selection_handle)

	ret.menuitem = []context_menu_item{
		{item: cmditem{cmd: cmdactor{desc: "Open "}}, handle: func() {
			if view.GetItemCount() > 0 {
				ret.selection_handle_impl(view.GetCurrentItem(), true)
			}
		}},
		{item: cmditem{cmd: cmdactor{desc: "Save "}}, handle: func() {
			ret.save()
		}},
		{item: cmditem{cmd: cmdactor{desc: "Copy"}}, handle: func() {
			ss := ret.sel.list.selected
			var data []string
			if ret.data.tree == nil {
				data = ret.data.BuildListString("")
			} else {
				x := ret.data.tree.tree_data_item
				for _, v := range x {
					data = append(data, v.text)
				}
			}
			if len(ss) > 0 {
				sss := data[ss[0]:ss[1]]
				aa := remove_color(sss)
				data := strings.Join(aa, "\n")
				main.CopyToClipboard(data)
				ret.sel.clear()
				main.app.ForceDraw()
			} else {
				main.CopyToClipboard(data[view.GetCurrentItem()])
			}
		}},
	}
	ret.sel = &list_multi_select{list: view}
	main.sel.Add(ret.sel)
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
func (qf *quickfix_history) InitDir() (string, error) {
	Dir := filepath.Join(qf.Wk.Export, "qf")
	if checkDirExists(Dir) {
		return Dir, nil
	}
	err := os.Mkdir(Dir, 0755)
	return Dir, err
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
func (qk *quick_view) DrawPreview(screen tcell.Screen, left, top, width, height int) bool {
	qk.quickview.draw(left, top, width, height, screen)
	return false
}

func (qk *quick_preview) draw(left, top, width, height int, screen tcell.Screen) {
	if !qk.visisble {
		return
	}
	frame := qk.frame
	frame.SetRect(left, top, width, height)
	frame.Draw(screen)
}
func (qk *quick_view) go_prev() {
	if qk.view.GetItemCount() == 0 {
		return
	}
	next := (qk.view.GetCurrentItem() - 1 + qk.view.GetItemCount()) % qk.view.GetItemCount()
	qk.view.SetCurrentItem(next)
	qk.open_index(next)
	qk.selection_handle_impl(next, false)
}

func (qk *quick_view) open_index(next int) {
	if len(qk.data.Refs.Refs) > 0 {
		if a, e := qk.data.get_data(next); e == nil {
			qk.quickview.update_preview(a.Loc)
		}
	}
}
func (qk *quick_view) go_next() {
	if qk.view.GetItemCount() == 0 {
		return
	}
	next := (qk.view.GetCurrentItem() + 1) % qk.view.GetItemCount()
	if loc, err := qk.data.get_data(next); err == nil {
		qk.quickview.update_preview(loc.Loc)
		qk.view.SetCurrentItem(next)
		qk.selection_handle_impl(next, false)
	}
}
func (qk *quick_view) OnSearch(txt string) {
	// old_query := qk.cmd_search_key
	// view := qk.view
	// highlight_search_key(old_query, view, txt)
	qk.cmd_search_key = txt
}

// func highlight_search_key(old_query string, view *customlist, new_query string) {
// 	sss := [][2]string{}
// 	ptn := ""
// 	if old_query != "" {
// 		ptn = fmt_bold_string(old_query)
// 	}
// 	for i := 0; i < view.GetItemCount(); i++ {
// 		m, s := view.GetItemText(i)
// 		if len(ptn) > 0 {
// 			m = strings.ReplaceAll(m, ptn, old_query)
// 			s = strings.ReplaceAll(s, ptn, old_query)
// 		}
// 		sss = append(sss, [2]string{m, s})
// 	}
// 	if len(new_query) > 0 {
// 		if new_query != "" {
// 			ptn = fmt_bold_string(new_query)
// 		}
// 		for i := range sss {
// 			v := &sss[i]
// 			m := v[0]
// 			s := v[1]
// 			m = strings.ReplaceAll(m, new_query, ptn)
// 			s = strings.ReplaceAll(s, new_query, ptn)
// 			v[0] = m
// 			v[1] = s
// 		}
// 	}
// 	view.Clear()
// 	for _, v := range sss {
// 		view.AddItem(v[0], v[1], nil)
// 	}
// }

// String
func (qk *quick_view) String() string {
	var s = qk.Type.String()
	coutn := qk.view.GetItemCount()
	index := qk.view.GetCurrentItem()
	if coutn > 0 {
		index += 1
	}
	key := qk.searchkey.Key
	return fmt.Sprintf("%s %s %d/%d", s, key, index, coutn)
}

// selection_handle
func (qk *quick_view) selection_handle(index int, _ string, _ string, _ rune) {
	qk.selection_handle_impl(index, true)
	if qk.quickview != nil {
		qk.quickview.visisble = false
	}
}

func (qk *quick_view) selection_handle_impl(index int, click bool) {
	if qk.data.tree != nil {
		qk.view.SetCurrentItem(index)
		node := qk.data.tree.tree_data_item[index]
		need_draw := false
		if click {
			if node.parent {
				node.expand = !node.expand
				data := qk.data.tree_to_listemitem()
				qk.view.Clear()
				for _, v := range data {
					qk.view.AddItem(v.text, "", func() {

					})
				}
				need_draw = true
			}
		}

		if vvv, err := qk.data.get_data(index); err == nil {
			qk.main.Tab().UpdatePageTitle()
			qk.cq.OpenFileHistory(vvv.Loc.URI.AsPath().String(), &vvv.Loc)
			if need_draw {
				GlobalApp.ForceDraw()
			}
		}
	} else {
		vvv := qk.data.Refs.Refs[index]
		qk.view.SetCurrentItem(index)
		qk.main.Tab().UpdatePageTitle()
		qk.cq.OpenFileHistory(vvv.Loc.URI.AsPath().String(), &vvv.Loc)
	}
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

func search_key_uid(key lspcore.SymolSearchKey) string {
	if len(key.File) > 0 {
		return fmt.Sprintf("%s %s:%d:%d", key.Key, key.File, key.Ranges.Start.Line, key.Ranges.Start.Character)
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

	qk.UpdateListView(t, Refs, key)
	qk.save()
}
func (qk *quick_view) AddResult(end bool, t DateType, caller ref_with_caller, key lspcore.SymolSearchKey) {
	if key.Key != qk.searchkey.Key {
		qk.view.Clear()
		qk.cmd_search_key = ""
		qk.Type = t
		qk.searchkey = key
		qk.grep.close()
		qk.data.reset_tree()
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
	if len(secondline) == 0 {
		return
	}
	qk.view.AddItem(fmt.Sprintf("%3d. %s", qk.view.GetItemCount()+1, secondline), "", nil)
	// qk.main.UpdatePageTitle()

	// qk.open_index(qk.view.GetCurrentItem())
}

func (qk *quick_view) UpdateListView(t DateType, Refs []ref_with_caller, key lspcore.SymolSearchKey) {
	if qk.grep != nil {
		qk.grep.close()
	}
	qk.Type = t
	qk.searchkey = key
	switch t {
	case data_grep_word, data_search:
		qk.view.Key = key.Key
	default:
		qk.view.Key = ""
	}
	qk.view.Clear()
	qk.view.SetCurrentItem(-1)
	qk.cmd_search_key = ""
	qk.data = *new_quikview_data(qk.main, t, qk.main.current_editor().Path(), Refs)
	qk.data.tree_to_listemitem()
	tree := qk.data.build_flextree_data(10)
	data := tree.ListString()
	var loaddata = func(data []string) {
		qk.view.Clear()
		for _, v := range data {
			qk.view.AddItem(v, "", nil)
		}
		qk.main.App().ForceDraw()
	}
	qk.flex_tree = tree
	qk.view.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		item := tree.GetNodeIndex(i)
		if item.IsParent() {
			if item.HasMore() {
				tree.LoadMore(item)
			} else {
				tree.Toggle(item)
			}
			loaddata(tree.ListItem)
			if r, e := item.GetRange(tree); e == nil {
				qk.view.SetCurrentItem(r[0])
			}
		} else {
			n := tree.GetCaller(i)
			qk.cq.OpenFileHistory(n.Loc.URI.AsPath().String(), &n.Loc)
		}
	})
	loaddata(data)
	qk.main.Tab().UpdatePageTitle()
}



func (caller *ref_with_caller) ListItem(root string, parent bool, prev *ref_with_caller) string {
	v := caller.Loc

	path := v.URI.AsPath().String()
	if len(root) > 0 {
		path = trim_project_filename(path, root)
	}
	if parent {
		// if caller.Childrens > 1 {
		// 	return fmt.Sprintf("%s %s", path, fmt_color_string(fmt.Sprint(caller.Childrens), tcell.ColorRed))
		// }
		return path
	}
	funcolor := global_theme.search_highlight_color()
	line := caller.get_code(funcolor)
	if caller.Caller != nil {
		c1 := ""
		x := ""
		if prev != nil && (prev.Caller.ClassName == caller.Caller.ClassName && prev.Caller.Name == caller.Caller.Name) {
			prefix := strings.Repeat(" ", min(len(prev.Caller.Name+prev.Caller.ClassName), 4))
			// if prev.Caller.ClassName != "" {
			// 	c1 = strings.Repeat(" ", len(prev.Caller.ClassName)+1)
			// }
			// if prev.Caller.Name != "" {
			// 	x = strings.Repeat(" ", len(prev.Caller.Name)+2)
			// }
			return fmt.Sprintf(":%-4d %s %s", v.Range.Start.Line+1, prefix, line)
		} else {
			funcolor := global_theme.search_highlight_color()
			if len(caller.Caller.ClassName) > 0 {
				if c, err := global_theme.get_lsp_color(lsp.SymbolKindClass); err == nil {
					f, _, _ := c.Decompose()
					icon := fmt.Sprintf("%c ", lspcore.IconsRunne[int(lsp.SymbolKindClass)])
					c1 = fmt_color_string(fmt.Sprint(icon, caller.Caller.ClassName+" > "), f)
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
			x = fmt_color_string(callname+" > ", caller_color)
			if c1 != "" {
				return fmt.Sprintf(":%-4d %s%s %s", v.Range.Start.Line+1, c1, x, line)
			} else {
				return fmt.Sprintf(":%-4d %s %s", v.Range.Start.Line+1, x, line)
			}
		}
	} else {
		return fmt.Sprintf(":%-4d %s", v.Range.Start.Line+1, strings.TrimLeft(line, "\t "))
	}
}

func (caller *ref_with_caller) get_code(funcolor tcell.Color) string {
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
		s := v.Range.Start.Character
		e := v.Range.End.Character
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
			line = strings.Join([]string{a1, fmt_color_string(a, funcolor), a2}, "")
		}
	}
	return line
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

func new_qf_history(main MainService) (*quickfix_history, error) {
	qf := &quickfix_history{
		Wk:   main.Lspmgr().Wk,
		last: filepath.Join(main.Lspmgr().Wk.Export, "quickfix_last.json"),
	}
	qfdir, err := qf.InitDir()
	qf.qfdir = qfdir
	if err != nil {
		log.Println("save ", err)
		return nil, err
	}
	return qf, nil
}
