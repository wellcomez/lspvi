package mainui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
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

func caller_to_listitem(caller *lspcore.CallStackEntry, root string) string {
	if caller == nil {
		return ""
	}
	callerstr := fmt.Sprintf(" [%s %s:%d]", caller.Name,
		strings.TrimPrefix(
			caller.Item.URI.AsPath().String(), root),
		caller.Item.Range.Start.Line)
	return callerstr
}
