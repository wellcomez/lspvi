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

	"github.com/gdamore/tcell/v2"
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
	pk.frame.SetTitle(title)
	pk.codeprev.Load(loc.URI.AsPath().String())
	pk.codeprev.gotoline(loc.Range.Start.Line)
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
}
type qf_history_data struct {
	Type   DateType
	Key    lspcore.SymolSearchKey
	Result search_reference_result
}

func (h *qf_history_data) ListItem() string {

	return ""
}

func (qk quick_view) save() error {
	h := quickfix_history{Wk: qk.main.lspmgr.Wk}
	dir, err := h.InitDir()
	if err != nil {
		log.Println("save ", err)
		return err
	}
	uid := search_key_uid(qk.searchkey)
	uid = strings.ReplaceAll(uid, qk.main.root, "")
	hexbuf := md5.Sum([]byte(uid))
	uid = hex.EncodeToString(hexbuf[:])
	filename := filepath.Join(dir, uid+".json")

	buf, error := json.Marshal(qf_history_data{
		Key:    qk.searchkey,
		Type:   qk.Type,
		Result: qk.Refs,
	})
	if error != nil {
		return error
	}
	return os.WriteFile(filename, buf, 0666)
}

type quickfix_history struct {
	Wk lspcore.WorkSpace
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
	return menu.qk.view.Box
}

// menuitem implements context_menu_handle.
func (menu quick_view_context) menuitem() []context_menu_item {
	return menu.qk.menuitem
}

func new_log_view(main *mainui) *logview {
	ret := &logview{
		view_link: &view_link{up: view_code, right: view_callin},
		log:       tview.NewTextView(),
	}
	return ret
}
func (log *logview) update_log_view(s string) {
	t := log.log.GetText(true)
	log.log.SetText(t + s)
}

// new_quikview
func new_quikview(main *mainui) *quick_view {
	view := new_customlist()
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
		view_link: &view_link{up: view_code, right: view_callin},
		Name:      vid.getname(),
		view:      view,
		main:      main,
		quickview: new_quick_preview(),
		menuitem:  items,
		// menu:      new_contextmenu(main, items,view.Box),
	}
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
func (main *quick_view) OnSearch(txt string) {
}

// String
func (qk quick_view) String() string {
	var s = qk.Type.String()
	return fmt.Sprintf("%s %d/%d", s, qk.currentIndex+1, len(qk.Refs.Refs))
}

// selection_handle
func (qk *quick_view) selection_handle(index int, _ string, _ string, _ rune) {
	qk.selection_handle_impl(index, true)
	qk.quickview.visisble = false
}

func (qk *quick_view) selection_handle_impl(index int, open bool) {
	vvv := qk.Refs.Refs[index]
	qk.currentIndex = index
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
	Refs := get_loc_caller(refs, m.lspmgr.Current)

	qk.UpdateListView(t, Refs, key)
}

func (qk *quick_view) UpdateListView(t DateType, Refs []ref_with_caller, key lspcore.SymolSearchKey) {
	qk.view.Clear()
	qk.Type = t
	qk.Refs.Refs = Refs
	qk.searchkey = key
	for _, caller := range qk.Refs.Refs {
		secondline := caller.ListItem(qk.main.root)
		if len(secondline) == 0 {
			continue
		}
		qk.view.AddItem(secondline, "", nil)
	}
	qk.open_index(qk.view.GetCurrentItem())
}

func (caller ref_with_caller) ListItem(root string) string {
	v := caller.Loc
	source_file_path := v.URI.AsPath().String()
	data, err := os.ReadFile(source_file_path)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	line := lines[v.Range.Start.Line]
	if len(line) == 0 {
		return ""
	}
	gap := 40
	begin := max(0, v.Range.Start.Character-gap)
	end := min(len(line), v.Range.Start.Character+gap)
	path := strings.Replace(v.URI.AsPath().String(), root, "", -1)
	callerstr := ""
	if caller.Caller != nil {
		callerstr = caller_to_listitem(caller.Caller, root)
	}
	code := line[begin:end]
	secondline := fmt.Sprintf("%s:%-4d%s		%s", path, v.Range.Start.Line+1, callerstr, code)
	return secondline
}
