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
	menu         *contextmenu
	searchkey    lspcore.SymolSearchKey
}
type searchdata struct {
	Type   DateType
	Key    lspcore.SymolSearchKey
	Result search_reference_result
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

	buf, error := json.Marshal(searchdata{
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

func (h *quickfix_history) Load() ([]search_result, error) {
	var ret = []search_result{}
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
		var result search_result
		err = json.Unmarshal(buf, &result)
		if err != nil {
			continue
		}
		ret = append(ret, result)
	}
	return ret, nil
}

// new_quikview
func new_quikview(main *mainui) *quick_view {
	view := new_customlist()
	view.List.SetMainTextStyle(tcell.StyleDefault.Normal()).ShowSecondaryText(false)
	var vid view_id = view_quickview
	var items = []context_menu_item{
		{item: cmditem{cmd: cmdactor{desc: "Open"}}, handle: func() {
			qk := main.quickview
			qk.selection_handle_impl(view.GetCurrentItem(), true)
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
		menu:      new_contextmenu(main, items),
	}
	view.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		is_rightclick := (action == tview.MouseRightClick)
		menu := ret.menu
		action, event = menu.handle_mouse(action, event)
		if ret.menu.visible && is_rightclick {
			_, y, _, _ := view.GetRect()
			index := (menu.MenuPos.y - y)
			log.Println("index in list:", index)
			view.SetCurrentItem(index)
		}
		return action, event
	})
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
	loc := qk.Refs.Refs[next].Loc
	qk.quickview.update_preview(loc)
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
	var s = "Refs"
	if qk.Type == data_search {
		s = "Search"
	}
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
	data_search = iota
	data_refs
)

func search_key_uid(key lspcore.SymolSearchKey) string {
	if len(key.File) > 0 {
		return fmt.Sprintf("%s %s:%d:%d", key.Key, key.File, key.Ranges.Start.Line, key.Ranges.Start.Character)
	}
	return key.Key
}
func (qk *quick_view) OnLspRefenceChanged(refs []lsp.Location, t DateType, key lspcore.SymolSearchKey) {
	qk.Type = t
	// panic("unimplemented")
	qk.view.Clear()

	m := qk.main
	qk.Refs.Refs = get_loc_caller(refs, m.lspmgr.Current)
	qk.searchkey = key
	for _, caller := range qk.Refs.Refs {
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
		path := strings.Replace(v.URI.AsPath().String(), qk.main.root, "", -1)
		callerstr := ""
		if caller.Caller != nil {
			callerstr = caller_to_listitem(caller.Caller, qk.main.root)
		}
		code := line[begin:end]
		secondline := fmt.Sprintf("%s:%-4d%s		%s", path, v.Range.Start.Line+1, callerstr, code)
		qk.view.AddItem(secondline, "", nil)
	}
	qk.open_index(qk.view.GetCurrentItem())
}
