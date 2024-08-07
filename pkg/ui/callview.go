package mainui

import (
	"fmt"
	"os"
	"strings"

	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"

	"github.com/gdamore/tcell/v2"
)

func new_fzfview(main *mainui) *fzfview {
	view := tview.NewList().SetMainTextStyle(tcell.StyleDefault.Normal())
	ret := &fzfview{
		view_link: &view_link{up: view_code, right: view_callin},
		Name:      "fzf",
		view:      view,
		main:      main,
	}
	view.SetSelectedFunc(ret.Hanlde)
	return ret

}

type DateType int

const (
	data_search = iota
	data_refs
)

// fzfview
type fzfview struct {
	*view_link
	view         *tview.List
	Name         string
	Refs         search_reference_result
	main         *mainui
	currentIndex int
	Type         DateType
}

func (fzf *fzfview) go_prev() {
	next := (fzf.view.GetCurrentItem() - 1 + fzf.view.GetItemCount()) % fzf.view.GetItemCount()
	fzf.view.SetCurrentItem(next)
	if fzf.Type == data_refs {
		fzf.Hanlde(next, "", "", 1)
	}
}
func (fzf *fzfview) go_next() {
	next := (fzf.view.GetCurrentItem() + 1) % fzf.view.GetItemCount()
	fzf.view.SetCurrentItem(next)
	if fzf.Type == data_refs {
		fzf.Hanlde(next, "", "", 1)
	}
}
func (main *fzfview) OnSearch(txt string) {
}

// String
func (fzf fzfview) String() string {
	var s = "Refs"
	if fzf.Type == data_search {
		s = "Search"
	}
	return fmt.Sprintf("%s %d/%d", s, fzf.currentIndex, len(fzf.Refs.refs))
}

// Hanlde
func (fzf *fzfview) Hanlde(index int, _ string, _ string, _ rune) {
	vvv := fzf.Refs.refs[index]
	fzf.currentIndex = index
	fzf.main.UpdatePageTitle()
	fzf.main.gotoline(vvv.loc)

}

func (fzf *fzfview) OnRefenceChanged(refs []lsp.Location, t DateType) {
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
			callerstr = newFunction(caller.caller, fzf.main.root)
		}
		secondline := fmt.Sprintf("%s:%d%s", path, v.Range.Start.Line+1, callerstr)
		fzf.view.AddItem(secondline, line[begin:end], 0, nil)
	}
}

func newFunction(items *lspcore.CallStackEntry, root string) string {
	callerstr := fmt.Sprintf(" [%s %s:%d]", items.Name,
		strings.TrimPrefix(
			items.Item.URI.AsPath().String(), root),
		items.Item.Range.Start.Line)
	return callerstr
}
