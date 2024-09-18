package mainui

import (
	"errors"

	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
	"github.com/pgavlin/femto/runtime"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	"zen108.com/lspvi/pkg/treesittertheme"
	// lspcore "zen108.com/lspvi/pkg/lsp"
)

var console_board_color = tcell.ColorGreen
// var focused_border_color = tcell.ColorGreenYellow

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
	mgr.update_default_color()
	// fg, bg, _ := n.Decompose()
	if style := mgr.get_default_style(); style != nil {
		fg, bg, _ := style.Decompose()
		if mgr.main != nil {
			mgr.main.set_widget_theme(fg, bg)
		}
		if code != nil {
			code.bgcolor = bg
		}
	}
	if code == nil {
		return false
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

func (mgr *symbol_colortheme) get_default_style() *tcell.Style {
	if n, ok := mgr.colorscheme["normal"]; ok {
		mgr.colorscheme["default"] = n
		return &n
	}
	return nil
}
func (mgr *symbol_colortheme) CursorLine() *tcell.Style {
	return mgr.newMethod("cursorline")
}
func (mgr *symbol_colortheme) search_highlight_color() tcell.Color {
	if color := mgr.newMethod("function"); color != nil {
		a, _, _ := color.Decompose()
		return a
	}
	return tcell.ColorYellow
}

func (mgr *symbol_colortheme) newMethod(name string) *tcell.Style {
	if n, ok := mgr.colorscheme[name]; ok {
		return &n
	}
	return nil
}
func (mgr *symbol_colortheme) StatusLine() *tcell.Style {
	return mgr.newMethod("statusline")
}

func (main *mainui) set_widget_theme(fg, bg tcell.Color) {
	var colorscheme = main.codeview.colorscheme
	if color := colorscheme.StatusLine(); color != nil {
		f, b, _ := color.Decompose()
		main.statusbar.SetBackgroundColor(b)
		main.statusbar.SetTextColor(f)
	}
	for _, v := range all_view_list {
		v.to_box(main).SetBackgroundColor(bg)
	}

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

	main.page.SetTitleColor(fg)

	// default_primarytext_color = fg
	// default_backgroudColor = bg
	input := main.cmdline.input
	input.SetBackgroundColor(bg)
	input.SetFieldTextColor(fg)
	main.fileexplorer.ChangeDir(main.fileexplorer.rootdir)
}

func (coloretheme *symbol_colortheme) update_default_color() {
	if style := coloretheme.get_default_style(); style != nil {
		fg, bg, _ := style.Decompose()
		coloretheme.__update_default_color(bg, fg)
	}
}
func (coloretheme *symbol_colortheme) __update_default_color(bg tcell.Color, fg tcell.Color) {

	tview.Styles.PrimitiveBackgroundColor = bg
	tview.Styles.PrimaryTextColor = fg
	tview.Styles.BorderColor = fg
	tview.Styles.TitleColor = fg
	tview.Styles.InverseTextColor = bg
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
var global_theme *symbol_colortheme
func new_ui_theme(theme string, main *mainui) *symbol_colortheme {
	var uicolorscheme *symbol_colortheme
	var colorscheme femto.Colorscheme
	micro_buffer := []byte{}
	if monokai := runtime.Files.FindFile(femto.RTColorscheme, theme); monokai != nil {
		if data, err := monokai.Data(); err == nil {
			micro_buffer = data

		}
	}
	buf, err := treesittertheme.LoadTreesitterTheme(theme)

	if err == nil {
		micro_buffer = append(micro_buffer, buf...)
	}
	if len(micro_buffer) > 0 {
		colorscheme = femto.ParseColorscheme(string(micro_buffer))
		uicolorscheme = &symbol_colortheme{
			colorscheme,
			main,
		}
	}
	return uicolorscheme
}
