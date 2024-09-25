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
	codeprev *CodeView
	frame    *tview.Frame
	visisble bool
}

// update_preview
func (pk *quick_preview) update_preview(loc lsp.Location) {
	pk.visisble = true
	title := fmt.Sprintf("%s:%d", loc.URI.AsPath().String(), loc.Range.End.Line)
	UpdateTitleAndColor(pk.frame.Box, title)
	pk.codeprev.Load2Line(loc.URI.AsPath().String(), loc.Range.Start.Line)
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
	main         *mainui
	currentIndex int
	Type         DateType
	// menu         *contextmenu
	menuitem      []context_menu_item
	searchkey     lspcore.SymolSearchKey
	right_context quick_view_context

	cmd_search_key string
	grep           *greppicker
}
type qf_history_data struct {
	Type   DateType
	Key    lspcore.SymolSearchKey
	Result search_reference_result
	Date   int64
	UID    string
}

func save_qf_uirefresh(main *mainui, data qf_history_data) error {
	h, err := new_qf_history(main)
	if err != nil {
		return err
	}
	err = h.save_history(main.root, data, true)
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
	save_qf_uirefresh(qk.main,
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
	yes := menu.qk.main.is_tab(view_quickview.getname())
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
	var items = []context_menu_item{
		{item: cmditem{cmd: cmdactor{desc: "Open"}}, handle: func() {
			if view.GetItemCount() > 0 {
				qk := main.quickview
				qk.selection_handle_impl(view.GetCurrentItem(), true)
			}
		}},
		{item: cmditem{cmd: cmdactor{desc: "Save"}}, handle: func() {
			main.quickview.save()
		}},
	}
	ret := &quick_view{
		view_link: &view_link{id: view_quickview, up: view_code, right: view_callin},
		Name:      vid.getname(),
		view:      view,
		main:      main,
		quickview: new_quick_preview(),
		menuitem:  items,
		// menu:      new_contextmenu(main, items,view.Box),
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
	old_query := qk.cmd_search_key
	view := qk.view

	highlight_search_key(old_query, view, txt)
	qk.cmd_search_key = txt
}

func highlight_search_key(old_query string, view *customlist, new_query string) {
	sss := [][2]string{}
	ptn := ""
	if old_query != "" {
		ptn = fmt.Sprintf("**%s**", old_query)
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
			ptn = fmt.Sprintf("**%s**", new_query)
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
	vvv := qk.Refs.Refs[index]
	qk.currentIndex = index
	qk.view.SetCurrentItem(index)
	same := vvv.Loc.URI.AsPath().String() == qk.main.codeview.filename
	if open || same {
		qk.main.UpdatePageTitle()
		qk.main.gotoline(vvv.Loc)
	} else {

	}
}

type DateType int

const (
	data_search DateType = iota
	data_refs
	data_callin
	data_bookmark
	data_grep_word
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

	m := qk.main
	Refs := qk.main.get_loc_caller(refs, m.lspmgr.Current)

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
	}
	if end {
		qk.save()
		return
	}
	qk.Refs.Refs = append(qk.Refs.Refs, caller)
	_, _, width, _ := qk.view.GetRect()
	caller.width = width
	secondline := caller.ListItem(qk.main.root)
	if len(secondline) == 0 {
		return
	}
	qk.view.AddItem(fmt.Sprintf("%-3d %s", qk.view.GetItemCount()+1, secondline), "", nil)
	// qk.main.UpdatePageTitle()

	// qk.open_index(qk.view.GetCurrentItem())
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
	_, _, width, _ := qk.view.GetRect()
	m := qk.main
	for i, caller := range qk.Refs.Refs {
		caller.width = width
		v := caller.Loc
		caller.Caller = m.lspmgr.GetCallEntry(v.URI.AsPath().String(), v.Range)
		secondline := caller.ListItem(qk.main.root)
		if len(secondline) == 0 {
			continue
		}
		qk.view.AddItem(fmt.Sprintf("%-3d %s", i+1, secondline), "", nil)
	}
	qk.main.UpdatePageTitle()
	// qk.open_index(qk.view.GetCurrentItem())
}

func (caller ref_with_caller) ListItem(root string) string {
	v := caller.Loc
	source_file_path := v.URI.AsPath().String()
	data, err := os.ReadFile(source_file_path)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	line := ""
	if len(lines) > v.Range.Start.Line {
		line = lines[v.Range.Start.Line]
		if len(line) == 0 {
			return ""
		}
	}
	gap := max(40, caller.width/2)
	begin := min(len(line), max(0, v.Range.Start.Character-gap))
	end := min(len(line), v.Range.Start.Character+gap)
	path := strings.Replace(v.URI.AsPath().String(), root, "", -1)
	callerstr := ""
	if caller.Caller != nil {
		callerstr = caller_to_listitem(caller.Caller, root)
	}
	code := line[begin:end]
	secondline := fmt.Sprintf("%s -> %s:%-4d %s", callerstr, path, v.Range.Start.Line+1, code)
	return secondline
}

func new_qf_history(main *mainui) (*quickfix_history, error) {
	qf := &quickfix_history{
		Wk:   main.lspmgr.Wk,
		last: filepath.Join(main.lspmgr.Wk.Export, "quickfix_last.json"),
	}
	qfdir, err := qf.InitDir()
	qf.qfdir = qfdir
	if err != nil {
		log.Println("save ", err)
		return nil, err
	}
	return qf, nil
}
