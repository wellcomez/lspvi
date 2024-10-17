package mainui

import (
	"fmt"
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
}

func (v *codetextview) Draw(screen tcell.Screen) {
	v.View.Draw(screen)
	if v.code == nil {
		return
	}
	x, y, w, _ := v.GetInnerRect()
	if v.code.vid() == view_code_below {
		v.code.DrawNavigationBar(x, y, w, screen)
	}
	// newFunction1(v, x, y, w,screen)
}
func (code *codetextview) PasteHandler() func(text string, setFocus func(tview.Primitive)) {
	return code.PasteHandlerImpl
}
func (code *CodeView) DrawNavigationBar(x int, y int, w int, screen tcell.Screen) {
	var v = code.view
	var symbol = code.LspSymbol()
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
	textStyle := global_theme.get_color("selection")

	if sym != nil {
		begin = code_navbar_draw_runne(screen, begin, y, ' ', *textStyle)

		x1 := code.FileName()
		x1 = strings.ReplaceAll(x1, "/", " > ") + " > "
		for _, v := range x1 {
			begin = code_navbar_draw_runne(
				screen, begin, y, v, *textStyle)
		}

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
		for {
			if begin < x+w {
				begin = code_navbar_draw_runne(screen,
					begin, y, ' ', *textStyle)
			} else {
				break
			}
		}
	}
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
func new_codetext_view(buffer *femto.Buffer) *codetextview {

	root := &codetextview{
		femto.NewView(buffer),
		bookmarkfile{},
		bookmarkfile{},
		"",
		false, 0,
		nil,
		nil,
	}
	root.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		style := tcell.StyleDefault
		_, topY, _, _ := root.GetInnerRect()
		bottom := root.Bottomline()
		mark := root.bookmark
		bookmark_icon:='\uf02e'
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
