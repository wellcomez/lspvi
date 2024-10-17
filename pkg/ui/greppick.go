package mainui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func new_grep_picker(v *fzfmain) *greppicker {
	grep := &greppicker{
		livewgreppicker: new_live_grep_picker(v),
	}
	grep.not_live = true
	return grep
}

// greppicker
type greppicker struct {
	*livewgreppicker
}

// close implements picker.
func (g *greppicker) close() {
	g.livewgreppicker.close()
}

func (g *greppicker) UpdateQuery(query string) {
	// g.query = query
	if g.impl.fzf_on_result != nil {
		g.impl.fzf_on_result.OnSearch(query, true)
		g.grep_list_view.Key = query
	}
}

func (g *greppicker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		var key = event.Key()
		if g.parent.input.HasFocus() {
			if key == tcell.KeyEnter {
				RunQuery(g)
				return
			} 
		}

		switch key {
		case tcell.KeyDown, tcell.KeyUp:
			g.grep_list_view.List.Focus(nil)
		case tcell.KeyCtrlS:
			{
				g.Save()

			}
		}
	}
}

func RunQuery(g *greppicker) {
	g.impl.fzf_on_result = nil
	g.parent.input.SetLabel(">")
	g.impl.query_option.Query= g.parent.input.GetText()
	g.livewgreppicker.__updatequery(g.impl.query_option)
}

// name implements picker.
func (g *greppicker) name() string {
	return "grep word"
}
