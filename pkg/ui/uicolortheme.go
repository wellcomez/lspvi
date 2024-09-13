package mainui

import (
	"errors"

	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type symbol_colortheme struct {
	colorscheme femto.Colorscheme
	main        *mainui
}

func (mgr *symbol_colortheme) get_color_style(kind lsp.SymbolKind) (tcell.Style, error) {
	switch kind {
	case lsp.SymbolKindClass, lsp.SymbolKindInterface:
		return mgr.colorscheme.GetColor("@type.class"), nil
	case lsp.SymbolKindFunction:
		return mgr.colorscheme.GetColor("@function"), nil
	case lsp.SymbolKindMethod:
		return mgr.colorscheme.GetColor("@function.method"), nil
	case lsp.SymbolKindStruct:
		return mgr.colorscheme.GetColor("structure"), nil
	case lsp.SymbolKindConstant:
		return mgr.colorscheme.GetColor("@constant"), nil
	}
	return tcell.StyleDefault, errors.New("not found")
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
			main.symboltree.update(main.lspmgr.Current)
		}
		code.bgcolor = bg
	}

	code.view.SetColorscheme(mgr.colorscheme, func(ts *lspcore.TreeSitter) {
		if code.main == nil {
			return
		} else if len(ts.Outline) > 0 {
			code.ts =ts
			code.main.OnSymbolistChanged(nil,nil)
		}
		code.main.app.Draw()
	})
	return false
}

