// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"
	// lspcore "zen108.com/lspvi/pkg/lsp"
)

type codetextview struct {
	*femto.View
	bookmark             bookmarkfile
	linechange           bookmarkfile
	filename             string
	mouse_select_area    bool
	LineNumberUnderMouse int
	code                 CodeEditor
	PasteHandlerImpl     func(text string, setFocus func(tview.Primitive))
	main                 MainService
	complete             CompleteMenu
}

func (view *codetextview) IconStyle(main MainService) tcell.Style {
	focus_color := tcell.ColorYellow
	style := get_style_hide(false)
	focus := false
	focus = view.HasFocus() || (main != nil && view.code == main.current_editor())
	if focus {
		style = style.Foreground(focus_color)
	}
	return style
}

var close_icon = '\uf2d3'
var split_icon = '\ueb56'

func new_textcode_left_toolbar(code *codetextview) *minitoolbar {
	vid := code.code.vid()
	name := filepath.Base(code.code.FileName())
	FileWithIcon(name)
	var runes = []rune{}
	for _, v := range FileWithIcon(name) + " " {
		runes = append(runes, v)
	}
	sytle := code.IconStyle(code.main)
	first := SplitCode.First() == code.code
	var quick_btn icon = icon{
		s: []rune{close_icon, ' '},
		click: func() {
			if view, ok := SplitCode.code_collection[vid]; ok {
				SplitClose(view).handle()
				code.main.App().ForceDraw()
			}
		},
		style: func() tcell.Style {
			return sytle
		},
	}
	item := []icon{
		{s: runes, style: func() tcell.Style { return sytle }, click: func() {}},
	}
	if vid != view_code {
		item = append(item, quick_btn)
	}
	if first {
		item = append([]icon{FileExploreIconButton(code.main)}, item...)
	}
	ret := &minitoolbar{
		item: item,
	}
	ret.getxy = func() (int, int) {
		x, y, _, _ := code.GetRect()
		return x, y - 1
	}
	return ret
}
func FileExploreIconButton(main MainService) icon {

	return icon{
		s: []rune{left_sidebar_rune, ' '},
		click: func() {
			main.toggle_view(view_file)
		},
		style: func() tcell.Style {
			return get_style_hide(main.IsHide(view_file))
		},
	}

}
func OutlineIconButton(main MainService) icon {
	var outline = icon{
		s: []rune{right_sidebar_rune, ' '},
		click: func() {
			main.toggle_view(view_outline_list)
		},
		style: func() tcell.Style {
			return get_style_hide(main.IsHide(view_outline_list))
		},
	}
	return outline
}
func new_textcode_toolbar(code *codetextview) *minitoolbar {
	sytle := code.IconStyle(code.main)
	item := []icon{}
	vid := code.code.vid()

	is_last := false
	v := SplitCode.Last()
	if v == nil {
		is_last = true
	} else {
		is_last = v == code.code
	}
	var split_btn icon = icon{
		s: []rune{split_icon, ' '},
		click: func() {
			code.code.SplitRight()
		},
		style: func() tcell.Style {
			return sytle
		},
	}

	main := code.main
	var back = icon{
		s: []rune{' ', back_runne, ' '},
		click: func() {
			main.GoBack()
		},
		style: func() tcell.Style {
			return get_style_hide(!main.CanGoBack())
		},
	}
	var forward = icon{
		s: []rune{forward_runne, ' '},
		click: func() {
			main.GoForward()
		},
		style: func() tcell.Style {
			return get_style_hide(!main.CanGoFoward())
		},
	}

	var outline = OutlineIconButton(code.main)
	var buttom = icon{
		s: []rune{'\U000f10a9', ' '},
		click: func() {
			main.toggle_view(view_console_area)
		},
		style: func() tcell.Style {
			return get_style_hide(main.IsHide(view_console_area))
		},
	}
	if vid == view_code {
		item = append(item, split_btn)
	} else {
		item = append(item, []icon{split_btn}...)
	}
	if is_last {

		bf := []icon{back, forward}
		item = append(item, bf...)
		if code.main.RunInBrowser() {
			zoom_out := icon{
				s: []rune{'\ueb81', ' '},
				style: func() tcell.Style {
					return get_style_hide(false)
				},
				click: func() {
					main.ZoomWeb(true)
				},
			}
			zoom_in := icon{
				s: []rune{'\ueb82', ' '},
				style: func() tcell.Style {
					return get_style_hide(false)
				},
				click: func() {
					main.ZoomWeb(false)
				},
			}
			item = append(item, []icon{zoom_out, zoom_in}...)
		}
		item = append(item, []icon{buttom, outline}...)
	}

	ret := &minitoolbar{
		item: item,
	}
	ret.getxy = func() (int, int) {
		x, y, w, _ := code.GetRect()
		x = x + w - ret.Width()
		return x, y - 1
	}
	return ret
}
func (v *codetextview) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		new_textcode_left_toolbar(v).handle_mouse_event(action, event)
		new_textcode_toolbar(v).handle_mouse_event(action, event)
		if v.complete.IsShown() {
			yes, a := v.complete.MouseHandler()(action, event, setFocus)
			if yes {
				return yes, a
			}
		}
		return v.View.MouseHandler()(action, event, setFocus)
	}

}

var BoxDrawingsDoubleHorizontal rune = '\u2550' // ═
var BoxDrawingsLightHorizontal rune = '\u2500'  // ─
func (b *codetextview) SetRect(x, y, width, height int) {
	if b.code.vid().is_editor() {
		b.View.SetRect(x, y+1, width, height-1)
	} else {
		b.View.SetRect(x, y, width, height)
	}
}
func (v *codetextview) Draw(screen tcell.Screen) {
	v.View.Draw(screen)
	if v.code == nil {
		return
	}
	vid := v.code.vid()
	if vid == view_code_below {
		x, y, w, _ := v.GetRect()
		v.DrawNavigationBar(x-1, y, w+2, screen)
	} else if v.code.vid().is_editor() {
		x, y, w, _ := v.GetRect()
		v.DrawNavigationBar(x, y+1, w, screen)
		y--
		ch := BoxDrawingsLightHorizontal
		if v.HasFocus() {
			ch = BoxDrawingsDoubleHorizontal
		}
		_, b := new_textcode_left_toolbar(v).Draw(screen)
		e, _ := new_textcode_toolbar(v).Draw(screen)
		for i := b; i < e; i++ {
			code_navbar_draw_runne(screen, i, y, ch, tcell.StyleDefault.Foreground(tview.Styles.BorderColor).Background(tview.Styles.PrimitiveBackgroundColor))
		}
		v.complete.Draw(screen)
	}
	// newFunction1(v, x, y, w,screen)
}

var BoxDrawingsHeavyVertical rune = '\u2503'  // ┃
var BoxDrawingsLightVertical rune = '\u2502'  // │
var BoxDrawingsDoubleVertical rune = '\u2551' // ║

func (code *codetextview) PasteHandler() func(text string, setFocus func(tview.Primitive)) {
	return code.PasteHandlerImpl
}
func (v *codetextview) DrawNavigationBar(x int, y int, w int, screen tcell.Screen) {
	y = y - 1
	var code = v.code
	var symbol = code.LspSymbol()
	border_style := tcell.StyleDefault.Foreground(tview.Styles.BorderColor).Background(v.GetBackgroundColor())
	code.LspSymbol()
	if symbol == nil {
		return
	}
	loc := v.Cursor.Loc
	r := lsp.Range{
		Start: lsp.Position{Line: loc.Y, Character: loc.X},
		End:   lsp.Position{Line: loc.Y, Character: loc.X},
	}
	sym := GetClosestSymbol(symbol, r)
	begin := x
	style := global_theme.get_default_style()
	textStyle := global_theme.select_style()

	b1 := BoxDrawingsLightVertical
	if v.HasFocus() {
		b1 = BoxDrawingsDoubleVertical
	}
	begin = code_navbar_draw_runne(screen, begin, y, b1,
		border_style)
	x1 := code.FileName()
	x1 = strings.ReplaceAll(x1, "/", " > ") + " > "
	for _, v := range x1 {
		begin = code_navbar_draw_runne(
			screen, begin, y, v, *textStyle)
	}
	if sym != nil {
		begin = code_navbar_draw_runne(screen, begin, y, ' ', *textStyle)

		if len(sym.Classname) > 0 {
			s := style
			if ss, err := global_theme.get_lsp_color(lsp.SymbolKindClass); err == nil {
				s = &ss
			}
			f, _, _ := s.Decompose()
			if run, ok := lspcore.IconsRunne[int(lsp.SymbolKindClass)]; ok {
				keys := []rune{run, ' '}
				for _, run := range keys {
					begin = code_navbar_draw_runne(screen, begin, y, run, textStyle.Foreground(f))
				}
			}

			for _, v := range sym.Classname {
				begin = code_navbar_draw_runne(
					screen, begin, y, v, textStyle.Foreground(f))
			}
			for _, v := range " >" {
				begin = code_navbar_draw_runne(screen,
					begin, y, v, *textStyle)
			}
		}
		if len(sym.SymInfo.Name) > 0 {
			s := style
			if ss, err := global_theme.get_lsp_color(sym.SymInfo.Kind); err == nil {
				s = &ss
			}
			f, _, _ := s.Decompose()
			if run, ok := lspcore.IconsRunne[int(sym.SymInfo.Kind)]; ok {
				keys := []rune{run, ' '}
				for _, run := range keys {
					begin = code_navbar_draw_runne(screen,
						begin, y, run, textStyle.Foreground(f))
				}
			} else {
				x1 := sym.Icon() + " "
				if len(x1) > 0 {
					for _, v := range x1 {
						begin = code_navbar_draw_runne(screen,
							begin, y, v, textStyle.Foreground(f))
					}
				}
			}
			for _, v := range sym.SymInfo.Name {
				begin = code_navbar_draw_runne(screen,
					begin, y, v, textStyle.Foreground(f))
			}
		}
	}
	for {
		if begin < x+w-1 {
			begin = code_navbar_draw_runne(screen,
				begin, y, ' ', *textStyle)
		} else {
			break
		}
	}
	code_navbar_draw_runne(screen, begin, y, b1, border_style)
}

func code_navbar_draw_runne(screen tcell.Screen, begin int, y int, v rune, textStyle tcell.Style) int {
	screen.SetContent(begin, y, v, nil, textStyle)
	begin++
	return begin
}
func (view *codetextview) has_bookmark() bool {
	var line = view.Cursor.Loc.Y + 1
	for _, v := range view.bookmark.LineMark {
		if v.Line == line {
			return true
		}
	}
	return false
}
func (view *codetextview) addbookmark(add bool, comment string) {
	var line = view.Cursor.Loc.Y + 1
	view.bookmark.Add(line, comment, view.Buf.Line(line-1), add)
}
func (root *codetextview) change_line_color(screen tcell.Screen, x int, topY int, style tcell.Style) {
	line := root.Cursor.Loc.Y
	x1 := root.GetLineNoFormDraw(line)
	by := x1 - root.Topline
	sss := fmt.Sprintf("%d", line)
	for i, ch := range sss {
		screen.SetContent(x+i, by+topY, ch, nil,
			style.Foreground(tcell.ColorDarkGreen).Background(root.GetBackgroundColor()))
	}
}
func (root *codetextview) draw_line_mark(mark bookmarkfile, ch rune, bottom int, screen tcell.Screen, x int, topY int, style tcell.Style) {
	b := []int{}
	for _, v := range mark.LineMark {
		line := v.Line - 1
		if line >= root.Topline && line <= bottom {
			b = append(b, line)
		}
		for _, line := range b {
			x1 := root.GetLineNoFormDraw(line)
			by := x1 - root.Topline
			screen.SetContent(x, by+topY, ch, nil,
				style.Foreground(global_theme.search_highlight_color()).Background(root.GetBackgroundColor()))
		}
	}
}
func new_codetext_view(buffer *femto.Buffer, main MainService) *codetextview {

	root := &codetextview{
		femto.NewView(buffer),
		bookmarkfile{},
		bookmarkfile{},
		"",
		false, 0,
		nil,
		nil,
		main,
		nil,
	}
	root.complete = Newcompletemenu(main, root)
	root.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		style := tcell.StyleDefault
		_, topY, _, _ := root.GetInnerRect()
		bottom := root.Bottomline()
		mark := root.bookmark
		bookmark_icon := '\uf02e'
		root.draw_line_mark(mark, bookmark_icon, bottom, screen, x, topY, style)
		root.draw_line_mark(root.linechange, '*', bottom, screen, x, topY, style)
		// root.change_line_color(screen, x, topY, style)
		return root.GetInnerRect()
	})
	// root.Buf.Settings["scrollbar"] = true
	// root.Buf.Settings["cursorline"] = true
	// root.addbookmark(1, true)
	// root.addbookmark(20, true)
	return root
}
