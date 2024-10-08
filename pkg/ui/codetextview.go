package mainui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
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
}

func (v *codetextview) Draw(screen tcell.Screen) {
	v.View.Draw(screen)
	if v.code == nil {
		return
	}
	x, y, w, _ := v.GetInnerRect()
	symbol := v.code.LspSymbol()
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
		for _, v := range " " {
			screen.SetContent(begin, y, v, nil, *textStyle)
			begin++

		}
		if len(sym.Classname) > 0 {
			s := style
			if ss, err := global_theme.get_color_style(lsp.SymbolKindClass); err == nil {
				s = &ss
			}
			f, _, _ := s.Decompose()
			if run, ok := lspcore.IconsRunne[int(lsp.SymbolKindClass)]; ok {
				screen.SetContent(begin, y, run, nil, textStyle.Foreground(f))
				begin++

				screen.SetContent(begin, y, ' ', nil, textStyle.Foreground(f))
				begin++
			}
			for _, v := range sym.Classname {
				screen.SetContent(begin, y, v, nil, textStyle.Foreground(f))
				begin++
			}
			for _, v := range " > " {
				screen.SetContent(begin, y, v, nil, *textStyle)
				begin++

			}
		}
		if len(sym.SymInfo.Name) > 0 {
			s := style
			if ss, err := global_theme.get_color_style(sym.SymInfo.Kind); err == nil {
				s = &ss
			}
			f, _, _ := s.Decompose()
			if run, ok := lspcore.IconsRunne[int(sym.SymInfo.Kind)]; ok {
				screen.SetContent(begin, y, run, nil, textStyle.Foreground(f))
				begin++
				screen.SetContent(begin, y, ' ', nil, textStyle.Foreground(f))
				begin++
			}
			for _, v := range sym.SymInfo.Name {
				screen.SetContent(begin, y, v, nil, textStyle.Foreground(f))
				begin++
			}
		}
		for {
			if begin < x+w {
				screen.SetContent(begin, y, ' ', nil, *textStyle)
				begin++
			} else {
				break
			}
		}
	}

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
	}
	root.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		style := tcell.StyleDefault
		_, topY, _, _ := root.GetInnerRect()
		bottom := root.Bottomline()
		mark := root.bookmark
		root.draw_line_mark(mark, 'B', bottom, screen, x, topY, style)
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
