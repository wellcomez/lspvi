package mainui

import (
	"github.com/pgavlin/femto"
	"github.com/rivo/tview"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type symbol_colortheme struct {
	colorscheme femto.Colorscheme
	main        *mainui
}

func (mgr *symbol_colortheme) update_controller_theme(code *CodeView) bool {

	if n, ok := mgr.colorscheme["normal"]; ok {
		mgr.colorscheme["default"] = n
		_, bg, _ := n.Decompose()
		main := mgr.main
		if main != nil {
			for _, v := range all_view_list {
				v.to_box(main).SetBackgroundColor(bg)
			}
			tview.Styles.PrimitiveBackgroundColor = bg
			main.layout.console.SetBackgroundColor(bg)
			main.page.SetBackgroundColor(bg)
			main.layout.editor_area.SetBackgroundColor(bg)
			main.layout.tab_area.SetBackgroundColor(bg)
			main.statusbar.SetBackgroundColor(bg)
			main.console_index_list.SetBackgroundColor(bg)
			main.layout.dialog.Frame.SetBackgroundColor(bg)
		}
		code.bgcolor = bg
	}

	code.view.SetColorscheme(mgr.colorscheme, func(ts *lspcore.TreeSitter) {
		if code.main == nil {
			return
		} else if len(ts.Outline) > 0 {
			lspmgr := code.main.lspmgr
			if lspmgr.Current == nil || !lspmgr.Current.HasLsp() {
				lspmgr.Current = &lspcore.Symbol_file{
					Class_object: ts.Outline,
				}
				code.main.symboltree.update(lspmgr.Current)
			}
		}
		code.main.app.Draw()
	})
	return false
}
