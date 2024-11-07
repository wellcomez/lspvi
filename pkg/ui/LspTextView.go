package mainui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	"zen108.com/lspvi/pkg/debug"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

const n40 = 60

type LspTextView struct {
	*tview.Box
	lines  []string
	HlLine lspcore.TreesiterSymbolLine
	main   MainService
}
type HelpBox struct {
	*LspTextView
	begin femto.Loc
	// end   femto.Loc
	prev *lsp.SignatureHelp

	doc       []*help_signature_docs
	loaded    bool
	Complete  *lspcore.CompleteCodeLine
	hasborder bool
	current   int
	//Complete            *lspcore.complete_code
}

func (b *HelpBox) SetBorder(show bool) {
	b.Box.SetBorder(show)
	b.hasborder = show
	if show {
		b.SetBorderStyle(global_theme.select_style().Foreground(tview.Styles.BorderColor))
	}
}
func (helpview *HelpBox) UpdateLayout(complete *completemenu) {
	var ret = []string{}
	var v = helpview.doc[helpview.current]
	filename := complete.filename()
	if len(helpview.doc) > 1 {
		ret = append(ret, fmt.Sprintf("%d/%d", helpview.current+1, len(helpview.doc)))
	}
	ret = append(ret, v.lines...)
	txt := strings.Join(ret, "\n")
	height := len(helpview.lines)
	if !helpview.loaded {
		height = helpview.Load(txt, filename)
		helpview.loaded = true
	}
	width := n40
	if helpview.hasborder {
		height += 2
		width = width + 2
	}
	loc := complete.editor.Cursor.Loc
	edit_x, edit_y, _, _ := complete.editor.GetRect()
	Y := edit_y + loc.Y - complete.editor.Topline - (height - 1)
	helpview.SetRect(helpview.begin.X+edit_x, Y, width, height)
}
func (v HelpBox) IsShown(view *codetextview) bool {
	loc := view.Cursor.Loc
	if v.begin.Y == loc.Y {
		begin := v.begin
		line := view.Buf.Line(begin.Y)
		if begin.X > len(line) {
			return false
		}
		ss := line[begin.X:]
		end := begin
		if index := strings.Index(ss, ")"); index >= 0 {
			end.X = begin.X + index
		} else if v.begin.LessThan(loc) {
			return true
		}
		if v.begin.GreaterThan(loc) || end.LessThan(loc) {
			return false
		}
		return true
	}
	return false
}
func NewHelpBox() *HelpBox {
	ret := &HelpBox{
		LspTextView: &LspTextView{
			Box: tview.NewBox(),
		},
	}
	// x := global_theme.get_color("selection")
	ret.SetBorder(true)
	return ret
}
func (help *HelpBox) handle_key(event *tcell.EventKey) {
	switch event.Key() {
	case tcell.KeyDown:
		help.current++
		if help.current >= len(help.doc) {
			help.current = 0
		}
		help.loaded = false
	case tcell.KeyUp:
		help.current--
		if help.current < 0 {
			help.current = len(help.doc) - 1
		}
		help.loaded = false
	}
}

func (helpview *LspTextView) Load(txt string, filename string) int {
	if helpview.Box == nil {
		helpview.Box = tview.NewBox()
	}
	helpview.lines = strings.Split(txt, "\n")
	ts := lspcore.NewTreeSitterParse(filename, txt)
	ts.Init(func(ts *lspcore.TreeSitter) {
		debug.DebugLog("init")
		helpview.main.App().QueueUpdateDraw(func() {
			helpview.HlLine = ts.HlLine
		})
	})
	return len(helpview.lines)
}

func (l *LspTextView) Draw(screen tcell.Screen) {
	l.Box.DrawForSubclass(screen, l)
	begingX, y, w, _ := l.GetInnerRect()
	default_style := *global_theme.select_style()
	_, bg, _ := default_style.Decompose()
	menu_width := w
	for i := range l.lines {
		v := l.lines[i]
		line := []rune(v)
		menu_width = max(len(line), menu_width)
	}
	for i, v := range l.lines {
		PosY := y + i
		var symline *[]lspcore.TreeSitterSymbol
		if sym, ok := l.HlLine[i]; ok {
			symline = &sym
		}
		line := []rune(v)
		for col, v := range line {
			style := default_style
			if symline != nil {
				if s, e := GetColumnStyle(symline, uint32(col), bg); e == nil {
					style = s
				}
			}
			Posx := begingX + col
			screen.SetContent(Posx, PosY, v, nil, style)
		}
		for col := len(line); col < menu_width; col++ {
			Posx := begingX + col
			screen.SetContent(Posx, PosY, ' ', nil, default_style)
		}
	}
}

func GetColumnStyle(symline *[]lspcore.TreeSitterSymbol, col uint32, bg tcell.Color) (style tcell.Style, err error) {
	for _, pos := range *symline {
		if col >= pos.Begin.Column && col < pos.End.Column {
			if s, e := get_position_style(pos); e == nil {
				style = s.Background(bg)
				return
			}
		}
	}
	return style, fmt.Errorf("not found")
}

func get_position_style(pos lspcore.TreeSitterSymbol) (tcell.Style, error) {
	style := global_theme.get_color(pos.CaptureName)
	if style == nil {
		style = global_theme.get_color("@" + pos.CaptureName)
	}
	if style == nil {
		return tcell.Style{}, fmt.Errorf("not found")
	}
	return *style, nil
}
