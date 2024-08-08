package mainui

import (
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
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
	quickview *quick_preview
	//view         *customlist
	view         *tview.List
	Name         string
	Refs         search_reference_result
	main         *mainui
	currentIndex int
	Type         DateType
}

// new_quikview
func new_quikview(main *mainui) *quick_view {
	//view := new_customlist()
	//view.List.SetMainTextStyle(tcell.StyleDefault.Normal())
	view := tview.NewList()
	view.SetMainTextStyle(tcell.StyleDefault.Normal()).ShowSecondaryText(false)
	var vid view_id = view_quickview
	ret := &quick_view{
		view_link: &view_link{up: view_code, right: view_callin},
		Name:      vid.getname(),
		view:      view,
		main:      main,
		quickview: new_quick_preview(),
	}
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
func (fzf *quick_view) DrawPreview(screen tcell.Screen, top, left, width, height int) bool {
	fzf.quickview.draw(width, height, screen)
	return false
}

func (fzf *quick_preview) draw(width int, height int, screen tcell.Screen) {
	if !fzf.visisble {
		return
	}
	width, height = screen.Size()
	w := width
	h := height * 1 / 4
	frame := fzf.frame
	frame.SetRect(0, height/3, w, h)
	frame.Draw(screen)
}
func (fzf *quick_view) go_prev() {
	next := (fzf.view.GetCurrentItem() - 1 + fzf.view.GetItemCount()) % fzf.view.GetItemCount()
	fzf.view.SetCurrentItem(next)
	fzf.open_index(next)
	if fzf.Type == data_refs {
		fzf.selection_handle_impl(next, false)
	}
}

func (fzf *quick_view) open_index(next int) {
	loc := fzf.Refs.refs[next].loc
	fzf.quickview.update_preview(loc)
}
func (fzf *quick_view) go_next() {
	next := (fzf.view.GetCurrentItem() + 1) % fzf.view.GetItemCount()
	loc := fzf.Refs.refs[next].loc
	fzf.quickview.update_preview(loc)
	fzf.view.SetCurrentItem(next)
	if fzf.Type == data_refs {
		fzf.selection_handle_impl(next, false)
	}
}
func (main *quick_view) OnSearch(txt string) {
}

// String
func (fzf quick_view) String() string {
	var s = "Refs"
	if fzf.Type == data_search {
		s = "Search"
	}
	return fmt.Sprintf("%s %d/%d", s, fzf.currentIndex, len(fzf.Refs.refs))
}

// selection_handle
func (fzf *quick_view) selection_handle(index int, _ string, _ string, _ rune) {
	fzf.selection_handle_impl(index, true)
	fzf.quickview.visisble = false
	//fzf.main.set_viewid_focus(view_code)
}

func (fzf *quick_view) selection_handle_impl(index int, open bool) {
	vvv := fzf.Refs.refs[index]
	fzf.currentIndex = index
	same := vvv.loc.URI.AsPath().String() == fzf.main.codeview.filename
	if open || same {
		fzf.main.UpdatePageTitle()
		fzf.main.gotoline(vvv.loc)
	} else {

	}
}

type DateType int

const (
	data_search = iota
	data_refs
)

func (fzf *quick_view) OnRefenceChanged(refs []lsp.Location, t DateType) {
	fzf.Type = t
	// panic("unimplemented")
	fzf.view.Clear()

	m := fzf.main
	fzf.Refs.refs = get_loc_caller(refs, m.lspmgr.Current)

	for _, caller := range fzf.Refs.refs {
		v := caller.loc
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
		path := strings.Replace(v.URI.AsPath().String(), fzf.main.root, "", -1)
		callerstr := ""
		if caller.caller != nil {
			callerstr = caller_to_listitem(caller.caller, fzf.main.root)
		}
		code := line[begin:end]
		secondline := fmt.Sprintf("%s:%-4d%s		%s", path, v.Range.Start.Line+1, callerstr, code)
		//fzf.view.AddItem(secondline, "", nil)
		fzf.view.AddItem(secondline, "", 0, nil)
	}
	fzf.open_index(fzf.view.GetCurrentItem())
}
