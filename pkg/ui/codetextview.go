// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	"zen108.com/lspvi/pkg/devicon"
	lspcore "zen108.com/lspvi/pkg/lsp"
	nerd "zen108.com/lspvi/pkg/ui/icon"
	// lspcore "zen108.com/lspvi/pkg/lsp"
)

type hover_dector struct {
	Pos   femto.Loc
	move  bool
	Abort bool
}

func new_hover(pos femto.Loc, move bool, cb func()) (ret *hover_dector) {
	ret = &hover_dector{Pos: pos, move: move}
	ret.start(cb)
	return
}
func (h *hover_dector) start(cb func()) {
	go func() {
		s := time.Second * 1
		if h.move {
			s = time.Millisecond * 5
		}
		timer := time.NewTimer(s)
		<-timer.C
		if h.Abort {
			return
		}
		cb()
	}()
}

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
	hover                *hover_dector
	error                *LspTextView
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

var close_icon = nerd.Nf_fa_window_close
var split_icon = nerd.Nf_cod_split_horizontal

func new_textcode_left_toolbar(code *codetextview) *minitoolbar {
	vid := code.code.vid()
	changemark := ""
	if len(code.linechange.LineMark) > 0 {
		changemark = "*"
	}
	name := filepath.Base(code.code.FileName())
	FileWithIcon(name)
	var runes = []rune{}
	for _, v := range FileWithIcon(name) + changemark + " " {
		runes = append(runes, v)
	}

	style := code.IconStyle(code.main)
	codestyle := code.IconStyle(code.main)
	if icon, err := devicon.FindIconPath(name); err == nil {
		x := icon.Color
		if c, e := hexToCellColor(x); e == nil {
			codestyle = style.Foreground(c)
		}
	}
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
			return style
		},
	}
	item := []icon{
		{s: runes, style: func() tcell.Style { return codestyle }, click: func() {}},
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
		s: []rune{nerd.Nf_md_dock_bottom, ' '},
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
				s: []rune{nerd.Nf_cod_zoom_in, ' '},
				style: func() tcell.Style {
					return get_style_hide(false)
				},
				click: func() {
					main.ZoomWeb(true)
				},
			}
			zoom_in := icon{
				s: []rune{nerd.Nf_cod_zoom_out, ' '},
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
		if !v.InRect(event.Position()) {
			return false, nil
		}
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
		if v.error != nil {
			v.error.Draw(screen)
		}
	}
	// newFunction1(v, x, y, w,screen)
}

var BoxDrawingsHeavyVertical rune = '\u2503'  // ┃
var BoxDrawingsLightVertical rune = '\u2502'  // │
var BoxDrawingsDoubleVertical rune = '\u2551' // ║

func (code *codetextview) PasteHandler() func(text string, setFocus func(tview.Primitive)) {
	return code.PasteHandlerImpl
}

type colorchar struct {
	begin, y  int
	v         rune
	textStyle tcell.Style
}
type navtext struct {
	data []colorchar
}

func (n *navtext) Add(s colorchar) int {
	n.data = append(n.data, s)
	return s.begin + 1
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
	fg, _, _ := global_theme.get_default_style().Decompose()
	textStyle := global_theme.select_line_style().Foreground(fg)

	b1 := BoxDrawingsLightVertical
	if v.HasFocus() {
		b1 = BoxDrawingsDoubleVertical
	}
	var txt navtext
	begin = txt.Add(colorchar{begin, y, b1,
		border_style})
	x1 := code.FileName()
	x1 = strings.ReplaceAll(x1, "/", " > ") + " > "
	for _, v := range x1 {
		begin = txt.Add(colorchar{
			begin, y, v, textStyle})
	}
	if sym != nil {
		begin = txt.Add(colorchar{begin, y, ' ', textStyle})

		if len(sym.Classname) > 0 {
			s := style
			if ss, err := global_theme.get_lsp_color(lsp.SymbolKindClass); err == nil {
				s = &ss
			}
			f, _, _ := s.Decompose()
			if run, ok := lspcore.IconsRunne[int(lsp.SymbolKindClass)]; ok {
				keys := []rune{run, ' '}
				for _, run := range keys {
					begin = txt.Add(colorchar{begin, y, run, textStyle.Foreground(f)})
				}
			}

			for _, v := range sym.Classname {
				begin = txt.Add(colorchar{
					begin, y, v, textStyle.Foreground(f)})
			}
			for _, v := range " >" {
				begin = txt.Add(colorchar{
					begin, y, v, textStyle})
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
					begin = txt.Add(colorchar{
						begin, y, run, textStyle.Foreground(f)})
				}
			} else {
				x1 := sym.Icon() + " "
				if len(x1) > 0 {
					for _, v := range x1 {
						begin = txt.Add(colorchar{
							begin, y, v, textStyle.Foreground(f)})
					}
				}
			}
			for _, v := range sym.SymInfo.Name {
				begin = txt.Add(colorchar{
					begin, y, v, textStyle.Foreground(f)})
			}
		}
	}
	for {
		if begin < x+w-1 {
			begin = txt.Add(colorchar{
				begin, y, ' ', textStyle})
		} else {
			break
		}
	}
	for _, v := range txt.data {
		if v.begin < x+w-1 {
			code_navbar_draw_runne(screen, v.begin, v.y, v.v, v.textStyle)
		}
	}
	code_navbar_draw_runne(screen, x+w-1, y, b1, border_style)
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
	Y := root.GetLineNoForDraw(line)
	by := Y - root.Topline
	LineR := by + topY
	sss := fmt.Sprintf("%d", line)
	for i, ch := range sss {
		screen.SetContent(x+i, LineR, ch, nil,
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
			x1 := root.GetLineNoForDraw(line)
			by := x1 - root.Topline
			screen.SetContent(x, by+topY, ch, nil, style)
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
		nil, nil, nil,
	}
	root.complete = Newcompletemenu(main, root)
	root.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		root.draw_line_number(screen, x)
		return root.GetInnerRect()
	})
	// root.Buf.Settings["scrollbar"] = true
	// root.Buf.Settings["cursorline"] = true
	// root.addbookmark(1, true)
	// root.addbookmark(20, true)
	return root
}

func (root *codetextview) draw_line_number(screen tcell.Screen, x int) {
	style := tcell.StyleDefault
	style = style.Foreground(global_theme.search_highlight_color()).Background(root.GetBackgroundColor())
	_, topY, _, _ := root.GetInnerRect()
	bottom := root.Bottomline()
	mark := root.bookmark
	bookmark_icon := nerd.Nf_fa_bookmark
	root.draw_line_mark(mark, bookmark_icon, bottom, screen, x, topY, style)
	root.draw_line_mark(root.linechange, '*', bottom, screen, x, topY, style)
	dialogsize := root.code.Dianostic()
	if !dialogsize.data.IsClear {
		mark := get_dialog_line(dialogsize, lsp.DiagnosticSeverityError)
		if len(mark.LineMark) > 0 {
			error_style := style
			if s, e := global_theme.get_styles([]string{"@error", "error"}); e == nil {
				error_style = s
			}
			root.draw_line_mark(mark, 'E', bottom, screen, x, topY, error_style)
		}
		mark = get_dialog_line(dialogsize, lsp.DiagnosticSeverityWarning)
		if len(mark.LineMark) > 0 {
			error_style := style
			if s, e := global_theme.get_styles([]string{"@text.warning", "warningmsg"}); e == nil {
				error_style = s
			}
			root.draw_line_mark(mark, 'W', bottom, screen, x, topY, error_style)
		}
	}
}

func get_dialog_line(dialogsize editor_diagnostic, level lsp.DiagnosticSeverity) (errline bookmarkfile) {
	for _, v := range dialogsize.data.Diagnostics {
		if v.Severity == level {
			errline.LineMark = append(errline.LineMark, LineMark{Line: v.Range.Start.Line + 1})
		}
	}
	return
}
