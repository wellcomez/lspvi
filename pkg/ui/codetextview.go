package mainui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
)

type codetextview struct {
	*femto.View
	bookmark   bookmarkfile
	linechange bookmarkfile
	filename   string
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

func newFunction1(mark bookmarkfile, ch rune, root *codetextview, bottom int, screen tcell.Screen, x int, topY int, style tcell.Style) {
	b := []int{}
	for _, v := range mark.LineMark {
		line := v.Line - 1
		if line >= root.Topline && line <= bottom {
			b = append(b, line)
		}
		for _, line := range b {
			by := root.GetLineNoFormDraw(line) - root.Topline
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
	}
	root.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		style := tcell.StyleDefault
		_, topY, _, _ := root.GetInnerRect()
		bottom := root.Bottomline()
		mark := root.bookmark
		newFunction1(mark, 'B', root, bottom, screen, x, topY, style)
		newFunction1(root.linechange, '*', root, bottom, screen, x, topY, style)
		return root.GetInnerRect()
	})
	// root.Buf.Settings["scrollbar"] = true
	// root.Buf.Settings["cursorline"] = true
	// root.addbookmark(1, true)
	// root.addbookmark(20, true)
	return root
}