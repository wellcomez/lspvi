package mainui

import (
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/tectiv3/go-lsp"
)

type quick_preview struct {
	codeprev *CodeView
}

func (preview quick_preview) open_file(file string) {
	preview.codeprev.Load(file)
}
func new_quick_preview(main *mainui) *quick_preview {
	return &quick_preview{
		codeprev: NewCodeView(main),
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
}

func new_quikview(main *mainui) *quick_view {
	view := new_customlist()
	view.List.SetMainTextStyle(tcell.StyleDefault.Normal())
	var vid view_id = view_quickview
	ret := &quick_view{
		view_link: &view_link{up: view_code, right: view_callin},
		Name:      vid.getname(),
		view:      view,
		main:      main,
		quickview: new_quick_preview(main),
	}
	view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		ch := event.Rune()
		if ch == 'j' {
			ret.go_next()
		} else if ch == 'k' {
			ret.go_prev()
		} else {
			return event
		}
		return nil
	})
	view.SetSelectedFunc(ret.Hanlde)
	return ret

}
func (fzf *quick_view) DrawPreview(screen tcell.Screen) bool {
	return false
}
func (fzf *quick_view) go_prev() {
	next := (fzf.view.GetCurrentItem() - 1 + fzf.view.GetItemCount()) % fzf.view.GetItemCount()
	fzf.view.SetCurrentItem(next)
	if fzf.Type == data_refs {
		fzf.Hanlde(next, "", "", 1)
	}
}
func (fzf *quick_view) go_next() {
	next := (fzf.view.GetCurrentItem() + 1) % fzf.view.GetItemCount()
	fzf.view.SetCurrentItem(next)
	if fzf.Type == data_refs {
		fzf.Hanlde(next, "", "", 1)
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

// Hanlde
func (fzf *quick_view) Hanlde(index int, _ string, _ string, _ rune) {
	vvv := fzf.Refs.refs[index]
	fzf.currentIndex = index
	fzf.main.UpdatePageTitle()
	fzf.main.gotoline(vvv.loc)

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
		fzf.view.AddItem(secondline, "", nil)
	}
}
