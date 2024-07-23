package mainui

import (
	"fmt"
	"os"
	"strings"

	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"

	"github.com/gdamore/tcell/v2"
)

func new_fzfview(main *mainui) *fzfview {
	view := tview.NewList().SetMainTextStyle(tcell.StyleDefault.Normal())
	ret := &fzfview{
		Name: "fzf",
		view: view,
		main: main,
	}
	view.SetSelectedFunc(ret.Hanlde)
	return ret

}

type fzfview struct {
	view *tview.List
	Name string
	Refs search_reference_result
	main *mainui
}

func (fzf *fzfview) Hanlde(index int, _ string, _ string, _ rune) {
	vvv := fzf.Refs.refs[index]
	fzf.main.gotoline(vvv)
}

func (fzf *fzfview) OnRefenceChanged(refs []lsp.Location) {
	// panic("unimplemented")
	fzf.view.Clear()
	fzf.Refs.refs = refs
	for _, v := range refs {
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
		secondline := fmt.Sprintf("%s:%d", path, v.Range.Start.Line+1)
		fzf.view.AddItem(line[begin:end], secondline, 0, nil)
	}
}
