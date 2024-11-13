// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	rgb "image/color"

	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
	"github.com/pgavlin/femto/runtime"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	"zen108.com/lspvi/pkg/debug"
	"zen108.com/lspvi/pkg/treesittertheme"
	// lspcore "zen108.com/lspvi/pkg/lsp"
)

func ColorToCellColor(c rgb.Color) tcell.Color {
	r, g, b, _ := c.RGBA()
	// 将每个通道的值从16位转换为8位
	r8 := int32(r >> 8)
	g8 := int32(g >> 8)
	b8 := int32(b >> 8)
	return tcell.NewRGBColor(r8, g8, b8)
}
func lightenColor(c rgb.Color, factor float64) rgb.Color {
	r, g, b, a := c.RGBA()
	r = uint32(float64(r) * (1 + factor))
	g = uint32(float64(g) * (1 + factor))
	b = uint32(float64(b) * (1 + factor))
	if r > 65535 {
		r = 65535
	}
	if g > 65535 {
		g = 65535
	}
	if b > 65535 {
		b = 65535
	}
	return rgb.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}

// var console_board_color = tcell.ColorGreen

// var focused_border_color = tcell.ColorGreenYellow

type symbol_colortheme struct {
	colorscheme femto.Colorscheme
	main        *mainui
	name        string
}

func (c symbol_colortheme) select_style() *tcell.Style {
	if style := c.get_color("selection"); style != nil {
		return style
	}
	s := c.get_default_style()
	return s
}
func (mgr symbol_colortheme) get_styles(names []string) (tcell.Style, error) {
	for _, v := range names {
		if s, ok := mgr.colorscheme[v]; ok {
			return s, nil
		}
	}
	return tcell.Style{}, fmt.Errorf("not found")
}
func (mgr symbol_colortheme) get_lsp_complete_color(kind lsp.CompletionItemKind) (ret tcell.Style, err error) {
	var get_styles = func(names []string) (tcell.Style, error) {
		return mgr.get_styles(names)
	}
	var get_style = func(names string) (tcell.Style, error) {
		return get_styles([]string{names})
	}
	switch kind {
	case lsp.CompletionItemKindClass:
		return mgr.get_lsp_color(lsp.SymbolKindClass)
	case lsp.CompletionItemKindText:
		return get_style("@text")
	case lsp.CompletionItemKindMethod:
		return mgr.get_lsp_color(lsp.SymbolKindMethod)
	case lsp.CompletionItemKindFunction:
		return mgr.get_lsp_color(lsp.SymbolKindFunction)
	case lsp.CompletionItemKindConstructor:
		return mgr.get_lsp_color(lsp.SymbolKindConstructor)
	case lsp.CompletionItemKindField:
		return mgr.get_lsp_color(lsp.SymbolKindField)
	case lsp.CompletionItemKindVariable:
		return mgr.get_lsp_color(lsp.SymbolKindVariable)
	case lsp.CompletionItemKindInterface:
		return mgr.get_lsp_color(lsp.SymbolKindInterface)
	case lsp.CompletionItemKindModule:
		return mgr.get_lsp_color(lsp.SymbolKindModule)
	case lsp.CompletionItemKindProperty:
		return mgr.get_lsp_color(lsp.SymbolKindProperty)
	case lsp.CompletionItemKindEnum:
		return mgr.get_lsp_color(lsp.SymbolKindEnum)
	case lsp.CompletionItemKindKeyword:
		return get_styles([]string{"@keyword", "keyword"})
	case lsp.CompletionItemKindFile:
		return mgr.get_lsp_color(lsp.SymbolKindFile)
	case lsp.CompletionItemKindReference:
	case lsp.CompletionItemKindFolder:
		return mgr.get_lsp_color(lsp.SymbolKindFile)
	case lsp.CompletionItemKindEnumMember:
		return mgr.get_lsp_color(lsp.SymbolKindEnumMember)
	case lsp.CompletionItemKindConstant:
		return mgr.get_lsp_color(lsp.SymbolKindConstant)
	case lsp.CompletionItemKindStruct:
		return mgr.get_lsp_color(lsp.SymbolKindStruct)
	case lsp.CompletionItemKindEvent:
		return mgr.get_lsp_color(lsp.SymbolKindEvent)
	case lsp.CompletionItemKindOperator:
		return mgr.get_lsp_color(lsp.SymbolKindOperator)
	case lsp.CompletionItemKindTypeParameter:
		return mgr.get_lsp_color(lsp.SymbolKindTypeParameter)
	case lsp.CompletionItemKindUnit:
	case lsp.CompletionItemKindValue:
	case lsp.CompletionItemKindSnippet:
	case lsp.CompletionItemKindColor:
	}

	return tcell.StyleDefault, errors.New("not found")
}
func (mgr symbol_colortheme) get_lsp_color(kind lsp.SymbolKind) (tcell.Style, error) {
	var styles = []string{}
	switch kind {
	case lsp.SymbolKindFile:

	case lsp.SymbolKindModule:
		styles = ([]string{"@module"})
	case lsp.SymbolKindNamespace:
		styles = ([]string{"@namespace", "@lsp.type.namespace"})
	case lsp.SymbolKindPackage:
		// styles = ([]string{"@namespace", "@lsp.type.namespace"})
	case lsp.SymbolKindClass:
		styles = ([]string{"@class", "@type.class", "@lsp.type.class"})
	case lsp.SymbolKindMethod:
		styles = []string{"@method", "@function.method", "lsp.type.method"}
	case lsp.SymbolKindProperty:
		styles = ([]string{"@property", "lsp.type.property"})
	case lsp.SymbolKindField:
		styles = ([]string{"@field"})
	case lsp.SymbolKindConstructor:
		styles = ([]string{"@construct"})
	case lsp.SymbolKindEnum:
		styles = ([]string{"@enum", "@lsp.type.enum"})
	case lsp.SymbolKindInterface:
		styles = ([]string{"@interface", "@type.class", "lsp.type.interface"})
	case lsp.SymbolKindFunction:
		styles = []string{"@function", "@lsp.type.function", "function"}
	case lsp.SymbolKindVariable:
		styles = ([]string{"@variable", "@lsp.type.variable"})
	case lsp.SymbolKindConstant:
		styles = []string{"@constant", "constant"}
	case lsp.SymbolKindString:
		styles = []string{"@string", "string"}
	case lsp.SymbolKindNumber:
		styles = []string{"@number", "number"}
	case lsp.SymbolKindBoolean:
		styles = []string{"@boolean", "boolean"}
	case lsp.SymbolKindArray:
	case lsp.SymbolKindObject:
	case lsp.SymbolKindKey:
	case lsp.SymbolKindNull:
	case lsp.SymbolKindEnumMember:
		styles = ([]string{"@enum", "@lsp.type.enummember", "@enum.member"})
	case lsp.SymbolKindStruct:
		styles = ([]string{"@structure", "@lsp.type.structure", "structure"})
	case lsp.SymbolKindEvent:
		styles = ([]string{"@event", "@lsp.type.event"})
	case lsp.SymbolKindOperator:
		styles = ([]string{"@operator", "operator"})
	case lsp.SymbolKindTypeParameter:
		styles = []string{"@type.parameter", "@lsp.type.typeparameter", "parameter"}
	}
	if len(styles) > 0 {
		return mgr.get_styles(styles)
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

func (mgr symbol_colortheme) get_default_style() *tcell.Style {
	if n, ok := mgr.colorscheme["normal"]; ok {
		return &n
	}
	ss := tcell.StyleDefault.Foreground(tview.Styles.PrimaryTextColor).Background(tview.Styles.PrimitiveBackgroundColor)
	return &ss
}
func IntToRGB(colorInt tcell.Color) rgb.RGBA {
	r, g, b := colorInt.RGB()
	return rgb.RGBA{uint8(r), uint8(g), uint8(b), 255} // 默认Alpha通道为255（完全不透明）
}
func HexToRGB(hexString string) (int, int, int, error) {
	hexString = strings.TrimPrefix(hexString, "#")

	value, err := strconv.ParseUint(hexString, 16, 32)
	if err != nil {
		return 0, 0, 0, err
	}

	blue := int(value & 255)
	green := int((value >> 8) & 255)
	red := int((value >> 16) & 255)

	return red, green, blue, nil
}
func (mgr *symbol_colortheme) select_line_style() tcell.Style {
	return mgr.colorscheme["cursor-line"]
}
func (mgr *symbol_colortheme) set_currsor_line() {
	if ret := mgr.get_color("cursorline"); ret != nil {
		_, bg, _ := ret.Decompose()
		// ss := bg.Hex()
		// debug.DebugLogf("color", "#%x %s", ss, mgr.name)
		// x := lightenColor(IntToRGB(bg), 0.2)
		// v := ColorToCellColor(x)
		// s := ret.Background(v)
		// _, bg, _ = ret.Decompose()
		debug.DebugLogf("color", "#%x %s", bg.Hex(), mgr.name)
		mgr.colorscheme["cursor-line"] = ret.Foreground(bg)
	}
	if bg := global_config.Color.Cursorline; bg != nil {
		if color, err := hexToCellColor(*bg); err == nil {
			s := tcell.StyleDefault.Background(color)
			mgr.colorscheme["cursor-line"] = s
		}
	}
	if ret := mgr.get_color("visual"); ret != nil {
		mgr.colorscheme["selection"] = *ret

	}
	if ret := mgr.get_color("linenr"); ret != nil {
		mgr.colorscheme["line-number"] = *ret
		if line := mgr.get_color("cursorline"); line != nil {
			f, _, _ := mgr.get_color("keyword").Decompose()
			_, b, _ := line.Decompose()
			mgr.colorscheme["current-line-number"] = ret.Background(b).Foreground(f)
		}
	}
	if n, ok := mgr.colorscheme["normal"]; ok {
		mgr.colorscheme["default"] = n
	}
}

func hexToCellColor(hex string) (ret tcell.Color, err error) {
	if r, g, b, e := hexToRGB(hex); e != nil {
		err = e
		return
	} else {
		ret = tcell.NewRGBColor(r, g, b)
		return
	}
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
func (mgr *symbol_colortheme) search_highlight_color_style() (ret tcell.Style) {
	if global_config.Color.Highlight != nil {
		if rgb := global_config.Color.Highlight.Search; rgb != "" {
			if color, err := hexToCellColor(rgb); err == nil {
				return mgr.get_default_style().Foreground(color)
			}
			// r,g,b := femto.ParseHexColor(global_config.Color.Highlight.Search)
		}
	}
	for _, key := range []string{"search", "keyword"} {
		if color := mgr.get_color(key); color != nil {
			ret = *color
			return
		}
	}
	return mgr.get_default_style().Foreground(tcell.ColorYellow)
}
func (mgr *symbol_colortheme) search_highlight_color() tcell.Color {
	if global_config.Color.Highlight != nil {

		if rgb := global_config.Color.Highlight.Search; rgb != "" {
			if r, err := hexToCellColor(rgb); err == nil {
				return r
			}
			// r,g,b := femto.ParseHexColor(global_config.Color.Highlight.Search)
		}
	}
	for _, key := range []string{"keyword"} {
		if color := mgr.get_color(key); color != nil {
			a, _, _ := color.Decompose()
			return a
		}
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

	trees := []*Tree{
		main.symboltree.view,
		main.fileexplorer.view,
		main.callinview.view,
		main.uml.file.view}
	for _, v := range trees {
		colorscheme.update_tree_color(v)
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
	textview := []*tview.TextView{main.log.log, main.symboltree.waiter, main.statusbar}
	for _, v := range textview {
		colorscheme.update_textview(v)
	}
	for _, x1 := range []*flex_area{main.layout.console, main.layout.editor_area} {
		colorscheme.update_flex_area(x1)
	}

	main.layout.tab_area.SetBackgroundColor(bg)

	colorscheme.update_listbox_color(main.console_index_list.List)

	x1 := main.Dialog()
	colorscheme.update_dialog_color(x1)

	x := main.current_editor()
	main.symboltree.update_with_ts(x.TreeSitter(), x.LspSymbol())

	main.page.SetTitleColor(fg)
	main.page.SetBackgroundColor(bg)
	main.page.SetBorderColor(fg)

	main.layout.mainlayout.spacemenu.table.SetBackgroundColor(bg)
	main.layout.mainlayout.spacemenu.load_spacemenu()

	main.right_context_menu.table.SetBackgroundColor(bg)

	// sp= fg
	// default_backgroudColor = bg
	inputs := []*tview.InputField{main.cmdline.input, main.Dialog().input}
	for _, input := range inputs {
		colorscheme.update_input_color(input)
	}

	for _, v := range SplitCode.code_collection {
		v.set_codeview_colortheme(colorscheme)
	}
	main.codeview2.set_codeview_colortheme(colorscheme)
	main.fileexplorer.ChangeDir(main.fileexplorer.rootdir)
	main.uml.file.view.SetBackgroundColor(bg)
	main.uml.file.ChangeDir(filepath.Join(lspviroot.Export, "uml"))
}

func (c symbol_colortheme) update_dialog_color(x1 *fzfmain) {
	fg, bg, _ := c.get_default_style().Decompose()
	x1.Frame.SetBackgroundColor(bg)
	x1.Frame.SetTitleColor(fg)
	x1.Frame.SetBorderColor(tview.Styles.BorderColor)
	c.update_input_color(x1.input)
}

func (c symbol_colortheme) update_tree_color(input *Tree) {
	fg, bg, _ := c.select_style().Decompose()
	input.SetBackgroundColor(bg)
	input.SetBorderColor(tview.Styles.BorderColor)
	input.SetGraphicsColor(fg)
}
func (c symbol_colortheme) update_listbox_color(input *List) {
	_, bg, _ := c.get_default_style().Decompose()
	input.SetBackgroundColor(bg)
	input.SetBorderColor(tview.Styles.BorderColor)
}
func (c symbol_colortheme) update_input_color(input *tview.InputField) {
	fg, bg, _ := c.get_default_style().Decompose()
	input.SetFieldBackgroundColor(bg)
	input.SetFieldTextColor(fg)
	input.SetBackgroundColor(bg)
	input.SetFieldTextColor(fg)
	input.SetLabelColor(fg)
}
func (color *symbol_colortheme) update_flex_area(x1 *flex_area) {
	_, bg, _ := color.get_default_style().Decompose()
	x1.SetBackgroundColor(bg)
}
func (color *symbol_colortheme) update_textview(x1 *tview.TextView) {
	fg, bg, _ := color.get_default_style().Decompose()
	x1.SetTextColor(fg)
	x1.SetBackgroundColor(bg)
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

func new_ui_theme(theme string, main *mainui) (uicolorscheme *symbol_colortheme) {
	var colorscheme femto.Colorscheme
	micro_buffer := []byte{}
	if monokai := runtime.Files.FindFile(femto.RTColorscheme, theme); monokai != nil {
		if data, err := monokai.Data(); err == nil {
			micro_buffer = data

		}
	}
	uicolorscheme = &symbol_colortheme{
		main: main,
		name: theme,
	}
	buf, err := treesittertheme.LoadTreesitterTheme(theme)
	if err == nil {
		micro_buffer = append(micro_buffer, buf...)
		if len(micro_buffer) > 0 {
			colorscheme = femto.ParseColorscheme(string(micro_buffer))
			uicolorscheme.colorscheme = colorscheme
		}
	} else {
		if len(micro_buffer) > 0 {
			colorscheme = femto.ParseColorscheme(string(micro_buffer))
			uicolorscheme.colorscheme = colorscheme
		}
	}
	return
}
