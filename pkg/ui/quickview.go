package mainui

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/reinhrst/fzf-lib"
	fzflib "github.com/reinhrst/fzf-lib"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type quick_preview struct {
	codeprev CodeEditor
	frame    *tview.Frame
	visisble bool
}

// update_preview
func (pk *quick_preview) update_preview(loc lsp.Location) {
	pk.visisble = true
	title := fmt.Sprintf("%s:%d", loc.URI.AsPath().String(), loc.Range.End.Line)
	UpdateTitleAndColor(pk.frame.Box, title)
	pk.codeprev.LoadFileNoLsp(loc.URI.AsPath().String(), loc.Range.Start.Line)
}
func new_quick_preview() *quick_preview {
	codeprev := NewCodeView(nil)
	frame := tview.NewFrame(codeprev.view)
	frame.SetBorder(true)
	return &quick_preview{
		codeprev: codeprev,
		frame:    frame,
	}
}

type logview struct {
	*view_link
	log *tview.TextView
}

// quick_view
type quick_view struct {
	// *tview.Flex
	*view_link
	quickview    *quick_preview
	view         *customlist
	Name         string
	Refs         search_reference_result
	main         MainService
	currentIndex int
	Type         DateType
	// menu         *contextmenu
	menuitem      []context_menu_item
	searchkey     lspcore.SymolSearchKey
	right_context quick_view_context

	cmd_search_key string
	grep           *greppicker
	sel            *list_multi_select

	tree *list_view_tree_extend
}
type list_view_tree_extend struct {
	tree           []list_tree_node
	tree_data_item []*list_tree_node
	filename       string
}

func (l list_view_tree_extend) NeedCreate() bool {
	return len(l.tree) == 0
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
		qf_history_data{qk.Type, qk.searchkey, qk.Refs, date, ""})
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

func new_log_view(main *mainui) *logview {
	ret := &logview{
		view_link: &view_link{id: view_log, up: view_code, right: view_callin},
		log:       tview.NewTextView(),
	}
	return ret
}

type fzf_list_item struct {
	maintext, secondText string
}
type fzf_on_listview struct {
	selected_index []int
	fzf            *fzf.Fzf
	listview       *customlist
	fuzz           bool
	list_data      []fzf_list_item
	selected       func(dataindex int, listindex int)
	query          string
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
		for _, v := range result.Matches {
			fzf.selected_index = append(fzf.selected_index, int(v.HayIndex))
		}
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
	for i := 0; i < len(fzf.list_data); i++ {
		fzf.selected_index = append(fzf.selected_index, i)
	}
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
	for i, v := range fzf.selected_index {
		list := i
		data := v
		a := fzf.list_data[data]
		fzf.listview.AddItem(a.maintext, a.secondText, func() {
			if fzf.selected != nil {
				fzf.listview.SetCurrentItem(list)
				fzf.selected(data, list)
			}
		})
	}
	fzf.listview.SetCurrentItem(0)
}

func (log *logview) update_log_view(s string) {
	t := log.log.GetText(true)
	log.log.SetText(t + s)
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
			if ret.tree == nil {
				data = ret.BuildListString("", main.lspmgr)
			} else {
				x := ret.tree.tree_data_item
				for _, v := range x {
					data = append(data, v.text)
				}
			}
			if len(ss) > 0 {

				sss := data[ss[0]:ss[1]]
				data := strings.Join(sss, "\n")
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
func (qk *quick_view) DrawPreview(screen tcell.Screen, top, left, width, height int) bool {
	qk.quickview.draw(width, height, screen)
	return false
}

func (qk *quick_preview) draw(width int, height int, screen tcell.Screen) {
	if !qk.visisble {
		return
	}
	width, height = screen.Size()
	w := width
	h := height * 1 / 4
	frame := qk.frame
	frame.SetRect(0, height/3, w, h)
	frame.Draw(screen)
}
func (qk *quick_view) go_prev() {
	if qk.view.GetItemCount() == 0 {
		return
	}
	next := (qk.view.GetCurrentItem() - 1 + qk.view.GetItemCount()) % qk.view.GetItemCount()
	qk.view.SetCurrentItem(next)
	qk.open_index(next)
	if qk.Type == data_refs {
		qk.selection_handle_impl(next, false)
	}
}

func (qk *quick_view) open_index(next int) {
	if len(qk.Refs.Refs) > 0 {
		loc := qk.Refs.Refs[next].Loc
		qk.quickview.update_preview(loc)
	}
}
func (qk *quick_view) go_next() {
	if qk.view.GetItemCount() == 0 {
		return
	}
	next := (qk.view.GetCurrentItem() + 1) % qk.view.GetItemCount()
	loc := qk.Refs.Refs[next].Loc
	qk.quickview.update_preview(loc)
	qk.view.SetCurrentItem(next)
	if qk.Type == data_refs {
		qk.selection_handle_impl(next, false)
	}
}
func (qk *quick_view) OnSearch(txt string) {
	// old_query := qk.cmd_search_key
	// view := qk.view
	// highlight_search_key(old_query, view, txt)
	qk.cmd_search_key = txt
}

func highlight_search_key(old_query string, view *customlist, new_query string) {
	sss := [][2]string{}
	ptn := ""
	if old_query != "" {
		ptn = fmt_bold_string(old_query)
	}
	for i := 0; i < view.GetItemCount(); i++ {
		m, s := view.GetItemText(i)
		if len(ptn) > 0 {
			m = strings.ReplaceAll(m, ptn, old_query)
			s = strings.ReplaceAll(s, ptn, old_query)
		}
		sss = append(sss, [2]string{m, s})
	}
	if len(new_query) > 0 {
		if new_query != "" {
			ptn = fmt_bold_string(new_query)
		}
		for i := range sss {
			v := &sss[i]
			m := v[0]
			s := v[1]
			m = strings.ReplaceAll(m, new_query, ptn)
			s = strings.ReplaceAll(s, new_query, ptn)
			v[0] = m
			v[1] = s
		}
	}
	view.Clear()
	for _, v := range sss {
		view.AddItem(v[0], v[1], nil)
	}
}

// String
func (qk *quick_view) String() string {
	var s = qk.Type.String()
	index := qk.currentIndex
	if len(qk.Refs.Refs) > 0 {
		index++
	}
	key := qk.searchkey.Key
	return fmt.Sprintf("%s %s %d/%d", s, key, index, len(qk.Refs.Refs))
}

// selection_handle
func (qk *quick_view) selection_handle(index int, _ string, _ string, _ rune) {
	qk.selection_handle_impl(index, true)
	qk.quickview.visisble = false
}

func (qk *quick_view) selection_handle_impl(index int, open bool) {
	if qk.tree != nil {
		qk.view.SetCurrentItem(index)
		node := qk.tree.tree_data_item[index]
		need_draw := false
		if node.parent {
			node.expand = !node.expand
			data := qk.tree.BuildListStringGroup(qk, global_prj_root, qk.main.Lspmgr())
			qk.view.Clear()
			for _, v := range data {
				qk.view.AddItem(v.text, "", func() {

				})
			}
			need_draw = true
		}
		refindex := node.ref_index
		vvv := qk.Refs.Refs[refindex]
		qk.main.Tab().UpdatePageTitle()
		qk.main.OpenFileHistory(vvv.Loc.URI.AsPath().String(), &vvv.Loc)
		if need_draw {
			GlobalApp.ForceDraw()
		}
	} else {
		vvv := qk.Refs.Refs[index]
		qk.currentIndex = index
		qk.view.SetCurrentItem(index)
		same := vvv.Loc.URI.AsPath().String() == qk.main.current_editor().Path()
		if open || same {
			qk.main.Tab().UpdatePageTitle()
			qk.main.OpenFileHistory(vvv.Loc.URI.AsPath().String(), &vvv.Loc)
		} else {

		}
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
		qk.reset_tree()
	}
	if end {
		qk.save()
		return
	}
	qk.reset_tree()
	qk.Refs.Refs = append(qk.Refs.Refs, caller)
	// _, _, width, _ := qk.view.GetRect()
	// caller.width = width
	secondline := caller.ListItem(global_prj_root, false)
	if len(secondline) == 0 {
		return
	}
	qk.view.AddItem(fmt.Sprintf("%3d. %s", qk.view.GetItemCount()+1, secondline), "", nil)
	// qk.main.UpdatePageTitle()

	// qk.open_index(qk.view.GetCurrentItem())
}

func (qk *quick_view) reset_tree() {
	qk.tree = nil
}
func (qk *quick_view) UpdateListView(t DateType, Refs []ref_with_caller, key lspcore.SymolSearchKey) {
	if qk.grep != nil {
		qk.grep.close()
	}
	qk.Type = t
	qk.Refs.Refs = Refs
	qk.searchkey = key
	qk.view.Key = key.Key
	qk.view.Clear()
	qk.view.SetCurrentItem(-1)
	qk.currentIndex = 0
	qk.cmd_search_key = ""
	qk.reset_tree()
	// _, _, width, _ := qk.view.GetRect()
	m := qk.main
	lspmgr := m.Lspmgr()
	qk.tree = &list_view_tree_extend{filename: qk.main.current_editor().Path()}
	qk.tree.build_tree(Refs)
	data := qk.tree.BuildListStringGroup(qk, global_prj_root, lspmgr)
	for _, v := range data {
		qk.view.AddItem(v.text, "", func() {

		})
	}
	qk.main.Tab().UpdatePageTitle()
}

func (qk *list_view_tree_extend) build_tree(Refs []ref_with_caller) {
	group := make(map[string]list_tree_node)
	for i := range Refs {
		caller := Refs[i]
		v := caller.Loc
		x := v.URI.AsPath().String()
		if s, ok := group[x]; ok {
			s.children = append(s.children, list_tree_node{ref_index: i})
			group[x] = s
		} else {
			s := list_tree_node{ref_index: i, parent: true, expand: true}
			s.children = append(s.children, list_tree_node{ref_index: i})
			group[x] = s
		}
	}
	trees := []list_tree_node{}
	for k, v := range group {
		if k == qk.filename {
			aaa := []list_tree_node{v}
			trees = append(aaa, trees...)
			continue
		}
		trees = append(trees, v)
	}
	qk.tree = trees
}
func (qk *list_view_tree_extend) BuildListStringGroup(view *quick_view, root string, lspmgr *lspcore.LspWorkspace) []*list_tree_node {
	var data = []*list_tree_node{}
	lineno := 1
	for i := range qk.tree {
		a := &qk.tree[i]
		a.quickfix_listitem_string(view, lspmgr, lineno)
		data = append(data, a)
		if a.expand {
			for i := range a.children {
				c := &a.children[i]
				c.quickfix_listitem_string(view, lspmgr, lineno)
				data = append(data, c)
			}
		}
		lineno++
	}
	qk.tree_data_item = data
	return data
}

func (tree *list_tree_node) quickfix_listitem_string(qk *quick_view, lspmgr *lspcore.LspWorkspace, lineno int) {
	caller := &qk.Refs.Refs[tree.ref_index]
	parent := tree.parent
	root := lspmgr.Wk.Path
	switch qk.Type {
	case data_refs, data_search, data_grep_word:
		v := caller.Loc
		if caller.Caller == nil || len(caller.Caller.Name) == 0 {
			caller.Caller = lspmgr.GetCallEntry(v.URI.AsPath().String(), v.Range)
		}
	}
	color := tview.Styles.BorderColor
	list_text := caller.ListItem(root, parent)
	if parent {
		tree.text = fmt.Sprintf("%3d. %s", lineno, list_text)
		if len(tree.children) > 0 {
			if !tree.expand {
				tree.text = fmt_color_string(fmt.Sprintf("%c", IconCollapse), color) + tree.text
			} else {
				tree.text = fmt_color_string(fmt.Sprintf("%c", IconExpaned), color) + tree.text
			}
		} else {
			tree.text = " " + tree.text
		}
	} else {
		tree.text = fmt.Sprintf(" %s", list_text)
	}
}
func (qk *quick_view) BuildListString(root string, lspmgr *lspcore.LspWorkspace) []string {
	var data = []string{}
	for i, caller := range qk.Refs.Refs {
		// caller.width = width
		switch qk.Type {
		case data_refs:
			v := caller.Loc
			if caller.Caller == nil || len(caller.Caller.Name) == 0 {
				caller.Caller = lspmgr.GetCallEntry(v.URI.AsPath().String(), v.Range)
			}
		}
		secondline := caller.ListItem(root, true)
		if len(secondline) == 0 {
			continue
		}
		x := fmt.Sprintf("%3d. %s", i+1, secondline)
		data = append(data, x)
	}
	return data
}

type list_tree_node struct {
	ref_index int
	expand    bool
	parent    bool
	children  []list_tree_node
	text      string
}

func (caller ref_with_caller) ListItem(root string, parent bool) string {
	v := caller.Loc

	line := ""
	path := v.URI.AsPath().String()
	if len(root) > 0 {
		path = trim_project_filename(path, root)
	}

	source_file_path := v.URI.AsPath().String()
	data, err := os.ReadFile(source_file_path)
	funcolor := global_theme.search_highlight_color()
	caller_color := funcolor
	if c, err := global_theme.get_lsp_color(lsp.SymbolKindFunction); err == nil {
		f, _, _ := c.Decompose()
		caller_color = f
	}
	icon := fmt.Sprintf("%c ", lspcore.IconsRunne[int(lsp.CompletionItemKindFunction)])
	if err != nil {
		line = err.Error()
	} else {
		lines := strings.Split(string(data), "\n")
		if len(lines) > v.Range.Start.Line {
			line = lines[v.Range.Start.Line]
			s := v.Range.Start.Character
			e := v.Range.End.Character
			if v.Range.Start.Line == v.Range.End.Line {
				if len(line) > s && len(line) > e && s < e {
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
		} else {
			line = "File changed lines not exsited"
		}
	}

	if caller.Caller != nil {
		callname := caller.Caller.Name
		callname = strings.TrimLeft(callname, " ")
		callname = strings.TrimRight(callname, " ")
		callname = icon + callname
		if parent {
			return path
		} else {
			return fmt.Sprintf(":%-4d %s %s", v.Range.Start.Line+1, fmt_color_string(callname, caller_color), line)
		}
	} else {
		if parent {
			return path
			// return fmt.Sprintf("%s:%-4d %s", path, v.Range.Start.Line+1, line)
		} else {
			return fmt.Sprintf(":%-4d %s", v.Range.Start.Line+1, strings.TrimLeft(line, "\t "))
		}
	}
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
