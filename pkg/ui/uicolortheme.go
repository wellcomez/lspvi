// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
	"github.com/pgavlin/femto/runtime"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	"zen108.com/lspvi/pkg/treesittertheme"
	// lspcore "zen108.com/lspvi/pkg/lsp"
)

// var console_board_color = tcell.ColorGreen

// var focused_border_color = tcell.ColorGreenYellow

type symbol_colortheme struct {
	colorscheme femto.Colorscheme
	main        *mainui
	name        string
}

func (mgr symbol_colortheme) get_lsp_color(kind lsp.SymbolKind) (tcell.Style, error) {
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
func (mgr *symbol_colortheme) update_controller_theme() bool {
	mgr.update_default_color()
	if style := mgr.get_default_style(); style != nil {
		fg, bg, _ := style.Decompose()
		if mgr.main != nil {
			mgr.set_widget_theme(fg, bg, mgr.main)
		}

	}
	return false
}

func (mgr *symbol_colortheme) get_default_style() *tcell.Style {
	if n, ok := mgr.colorscheme["normal"]; ok {
		return &n
	}
	return nil
}
func (mgr *symbol_colortheme) set_currsor_line() *tcell.Style {
	ret := mgr.get_color("cursorline")
	if ret != nil {
		mgr.colorscheme["cursor-line"] = *ret
		// if _, ok := mgr.colorscheme["selection"]; !ok {
		// 	mgr.colorscheme["selection"] = *ret
		// }
	}
	if ret := mgr.get_color("visual"); ret != nil {
		mgr.colorscheme["selection"] = *ret

	}
	if ret := mgr.get_color("linenr"); ret != nil {
		mgr.colorscheme["line-number"] = *ret
	}
	if ret := mgr.get_color("@function"); ret != nil {
		f, _, _ := ret.Decompose()
		mgr.colorscheme["current-line-number"] = ret.Foreground(f)
	}
	if n, ok := mgr.colorscheme["normal"]; ok {
		mgr.colorscheme["default"] = n
		return &n
	}
	return ret

}

func hexToRGB(hex string) (int32, int32, int32, error) {
	if len(hex) != 7 || hex[0] != '#' {
		return 0, 0, 0, fmt.Errorf("invalid hex color code")
	}

	r, err := strconv.ParseInt(hex[1:3], 16, 0)
	if err != nil {
		return 0, 0, 0, err
	}

	g, err := strconv.ParseInt(hex[3:5], 16, 0)
	if err != nil {
		return 0, 0, 0, err
	}

	b, err := strconv.ParseInt(hex[5:7], 16, 0)
	if err != nil {
		return 0, 0, 0, err
	}

	return int32(r), int32(g), int32(b), nil
}

func (mgr *symbol_colortheme) search_highlight_color() tcell.Color {
	if rgb := global_config.Color.Highlight.Search; rgb != "" {
		if r, g, b, err := hexToRGB(rgb); err == nil {
			return tcell.NewRGBColor(r, g, b)
		}
		// r,g,b := femto.ParseHexColor(global_config.Color.Highlight.Search)
	}
	if color := mgr.get_color("keyword"); color != nil {
		a, _, _ := color.Decompose()
		return a
	}
	return tcell.ColorYellow
}

func (mgr *symbol_colortheme) get_color(name string) *tcell.Style {
	if n, ok := mgr.colorscheme[name]; ok {
		return &n
	}
	return nil
}
func (mgr *symbol_colortheme) StatusLine() *tcell.Style {
	return mgr.get_color("statusline")
}

func (colorscheme *symbol_colortheme) set_widget_theme(fg, bg tcell.Color, main *mainui) {
	colorscheme.set_currsor_line()
	if color := colorscheme.StatusLine(); color != nil {
		f, b, _ := color.Decompose()
		main.statusbar.SetBackgroundColor(b)
		main.statusbar.SetTextColor(f)
	}
	for _, v := range all_view_list {
		b := v.to_box(main)
		if b != nil {
			b.SetBackgroundColor(bg)
		}
	}

	main.layout.console.SetBackgroundColor(bg)

	main.page.SetBackgroundColor(bg)
	main.page.SetBorderColor(fg)

	trees := []*tview.TreeView{
		main.symboltree.view,
		main.fileexplorer.view,
		main.callinview.view,
		main.uml.file.view}
	for _, v := range trees {
		v.SetGraphicsColor(fg)
	}

	for _, v := range all_view_list {
		if v == view_code {
			continue
		}
		view := main.get_view_from_id(v)
		if view != nil {
			view.SetBackgroundColor(bg)
		}
	}
	main.uml.file.view.SetBackgroundColor(bg)

	main.log.log.SetTextColor(fg)
	main.log.log.SetBackgroundColor(bg)

	main.layout.editor_area.SetBackgroundColor(bg)
	main.layout.tab_area.SetBackgroundColor(bg)
	main.statusbar.SetBackgroundColor(bg)
	main.console_index_list.SetBackgroundColor(bg)
	main.layout.dialog.Frame.SetBackgroundColor(bg)
	x := main.current_editor()
	main.symboltree.update_with_ts(x.TreeSitter(), x.LspSymbol())
	main.symboltree.waiter.SetBackgroundColor(bg)
	main.symboltree.waiter.SetTextColor(fg)

	main.page.SetTitleColor(fg)

	main.layout.spacemenu.table.SetBackgroundColor(bg)
	main.layout.spacemenu.load_spacemenu()

	main.right_context_menu.table.SetBackgroundColor(bg)

	// sp= fg
	// default_backgroudColor = bg
	inputs := []*tview.InputField{main.cmdline.input, main.layout.dialog.input}
	for _, input := range inputs {
		input.SetFieldBackgroundColor(bg)
		input.SetFieldTextColor(fg)
		input.SetBackgroundColor(bg)
		input.SetFieldTextColor(fg)
		input.SetLabelColor(fg)
	}

	for _, v := range SplitCode.code_collection {
		v.set_codeview_colortheme(colorscheme)
	}
	main.codeview2.set_codeview_colortheme(colorscheme)
	main.fileexplorer.ChangeDir(main.fileexplorer.rootdir)
}

func (coloretheme *symbol_colortheme) update_default_color() {
	if style := coloretheme.get_default_style(); style != nil {
		fg, bg, _ := style.Decompose()
		coloretheme.__update_default_color(bg, fg)
		if ret := coloretheme.get_color("@function"); ret != nil {
			f, _, _ := ret.Decompose()
			tview.Styles.BorderColor = f
		}
	}
}
func (coloretheme *symbol_colortheme) __update_default_color(bg tcell.Color, fg tcell.Color) {
	tview.Styles.PrimitiveBackgroundColor = bg
	tview.Styles.PrimaryTextColor = fg
	tview.Styles.BorderColor = fg
	tview.Styles.TitleColor = fg
	tview.Styles.InverseTextColor = bg
}

//	func (code *CodeView)update_with_ts_inited(ts *lspcore.TreeSitter) {
//		if code.main == nil {
//			return
//		} else if len(ts.Outline) > 0 {
//			code.ts = ts
//			if ts.DefaultOutline() {
//				lsp := code.main.symboltree.upate_with_ts(ts)
//				code.main.lspmgr.Current = lsp
//			} else {
//				code.main.OnSymbolistChanged(nil, nil)
//			}
//		}
//		code.main.app.Draw()
//	}
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
			colorscheme: colorscheme,
			main:        main,
			name:        theme,
		}
	}
	return uicolorscheme
}
