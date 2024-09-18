package mainui

import (
	"errors"

	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	// lspcore "zen108.com/lspvi/pkg/lsp"
)

var console_board_color = tcell.ColorGreen
var focused_border_color = tcell.ColorGreenYellow

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
		fg, bg, _ := n.Decompose()
		if mgr.main != nil {
			mgr.main.set_widget_theme(fg, bg)
		}
		code.bgcolor = bg
	}
	code.view.Buf.SetTreesitter(code.tree_sitter)
	code.view.SetColorscheme(mgr.colorscheme)
	go func() {
		if GlobalApp != nil {
			GlobalApp.QueueUpdateDraw(func() {
			})
		}
	}()
	return false
}

func (main *mainui) set_widget_theme(fg, bg tcell.Color) {
	tview.Styles.PrimitiveBackgroundColor = bg
	tview.Styles.PrimaryTextColor = fg
	tview.Styles.BorderColor = fg
	tview.Styles.TitleColor = fg
	tview.Styles.InverseTextColor = bg

	for _, v := range all_view_list {
		v.to_box(main).SetBackgroundColor(bg)
	}

	tview.Styles.PrimitiveBackgroundColor = bg
	main.layout.console.SetBackgroundColor(bg)
	main.page.SetBackgroundColor(bg)

	main.symboltree.view.SetGraphicsColor(fg)
	main.fileexplorer.view.SetGraphicsColor(fg)
	main.page.SetBorderColor(fg)

	main.uml.file.view.SetGraphicsColor(fg)
	main.uml.file.view.SetBackgroundColor(bg)

	main.console_index_list.SetBackgroundColor(bg)
	main.layout.editor_area.SetBackgroundColor(bg)
	main.layout.tab_area.SetBackgroundColor(bg)
	main.statusbar.SetBackgroundColor(bg)
	main.console_index_list.SetBackgroundColor(bg)
	main.layout.dialog.Frame.SetBackgroundColor(bg)
	main.symboltree.update(main.lspmgr.Current)

	// default_primarytext_color = fg
	// default_backgroudColor = bg
	input := main.cmdline.input
	input.SetBackgroundColor(bg)
	input.SetFieldTextColor(fg)
	main.fileexplorer.ChangeDir(main.fileexplorer.rootdir)
}

// func (code *CodeView)update_with_ts_inited(ts *lspcore.TreeSitter) {
// 	if code.main == nil {
// 		return
// 	} else if len(ts.Outline) > 0 {
// 		code.ts = ts
// 		if ts.DefaultOutline() {
// 			lsp := code.main.symboltree.upate_with_ts(ts)
// 			code.main.lspmgr.Current = lsp
// 		} else {
// 			code.main.OnSymbolistChanged(nil, nil)
// 		}
// 	}
// 	code.main.app.Draw()
// }
