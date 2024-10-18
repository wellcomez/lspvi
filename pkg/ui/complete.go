package mainui

import (
	"time"

	"github.com/pgavlin/femto"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type CompleteMenu struct {
	*customlist
	show          bool
	loc           femto.Loc
	width, height int
	editor        *codetextview
}

func NewCompleteMenu(main MainService, txt *codetextview) *CompleteMenu {
	ret := &CompleteMenu{
		new_customlist(false), false, femto.Loc{X: 0, Y: 0}, 0, 0, txt}

	// ret.customlist.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
	// 	idx := ret.GetCurrentItem()
	// 	count := ret.GetItemCount()
	// 	if event.Key() == tcell.KeyUp {
	// 		idx--
	// 		idx = max(idx, 0)
	// 		ret.SetCurrentItem(idx)
	// 		return nil
	// 	}
	// 	if event.Key() == tcell.KeyDown {
	// 		idx++
	// 		idx = min(idx, count-1)
	// 		ret.SetCurrentItem(idx)
	// 		return nil
	// 	}
	// 	if event.Key() == tcell.KeyEnter {
	// 		ret.customlist.List.InputHandler()(event, nil)
	// 	}
	// 	return event
	// })
	return ret
}

func (code *CodeView) handle_complete_key(after []lspcore.CodeChangeEvent) {
	codetext := code.view
	lsp := code.LspSymbol()
	if lsp == nil {
		return
	}
	if !lsp.HasLsp() {
		return
	}
	if complete := codetext.complete; complete != nil {
		for _, v := range after {
			if codetext.run_complete(v, lsp, complete, codetext) {
				break
			}
		}
	}
}

func (view *codetextview) run_complete(v lspcore.CodeChangeEvent, sym *lspcore.Symbol_file, complete *CompleteMenu, codetext *codetextview) bool {
	for _, e := range v.Events {
		if e.Type == lspcore.TextChangeTypeInsert && len(e.Text) == 1 {
			if e.Text == "\n" {
				continue
			}
			req := complete.newFunction1(e)
			go sym.DidComplete(req)
			return true
		}
	}
	return false
}

func (complete *CompleteMenu) newFunction1(e lspcore.TextChangeEvent) lspcore.Complete {
	var codetext = complete.editor
	var cb = func(cl lsp.CompletionList, err error) {
		if err != nil {
			return
		}
		if !cl.IsIncomplete {
			return
		}
		complete.Clear()
		width := 0
		for i := range cl.Items {
			v := cl.Items[i]
			width = max(len(v.Label)+2, width)
			complete.AddItem(v.Label, "", func() {
				complete.show = false
				codetext.Buf.Insert(complete.loc, v.Label)
			})
		}
		complete.height = len(cl.Items)
		complete.width = width
		complete.loc = codetext.Cursor.Loc
		complete.show = true
		go func() {
			<-time.After(10 * time.Second)
			complete.show = false
		}()
	}
	req := lspcore.Complete{
		Pos:              e.Range.End,
		TriggerCharacter: e.Text,
		Cb:               cb}
	return req
}