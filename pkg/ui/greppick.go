package mainui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func new_grep_picker(v *fzfmain, code CodeEditor) *greppicker {
	grep := &greppicker{
		livewgreppicker: new_live_grep_picker(v, code),
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

// UpdateQuery implements picker.
// Subtle: this method shadows the method (*livewgreppicker).UpdateQuery of greppicker.livewgreppicker.
func (g *greppicker) UpdateQuery(query string) {
	// g.query = query
	if g.impl.fzf_on_result != nil {
		g.impl.fzf_on_result.OnSearch(query, true)
		g.grep_list_view.Key = query
	}
}

// handle implements picker.
// Subtle: this method shadows the method (*livewgreppicker).handle of greppicker.livewgreppicker.
func (g *greppicker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		var key = event.Key()
		if key == tcell.KeyEnter {
			if g.parent.input.HasFocus() {
				if g.impl.last != g.impl.query_option {
					RunQuery(g)
				}
			} else {
				g.grep_list_view.InputHandler()(event, nil)
			}
		} else if key == tcell.KeyCtrlS {
			g.Save()
		} else {
			switch key {
			case tcell.KeyDown, tcell.KeyUp:
				g.grep_list_view.List.Focus(nil)
			}
		}
	}
}

func RunQuery(g *greppicker) {
	g.impl.fzf_on_result = nil
	g.parent.input.SetLabel(">")
	g.impl.query_option.query = g.parent.input.GetText()
	g.livewgreppicker.__updatequery(g.impl.query_option)
}

// name implements picker.
func (g *greppicker) name() string {
	return "grep word"
}
