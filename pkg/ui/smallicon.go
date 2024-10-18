package mainui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func get_style_hide(hide bool) tcell.Style {
	style := global_theme.get_default_style()
	if style != nil {
		f, b, _ := style.Decompose()
		hide_stycle := style.Foreground(f).Background(b)
		x1 := style.Foreground(tcell.ColorBlue).Background(b)
		if !hide {
			hide_stycle = x1
		}
		return hide_stycle
	}
	return active_btn_style.Background(tcell.ColorBlack)
}

type icon struct {
	s          []rune
	begin, end Pos
	click      func()
	style      func() tcell.Style
}

func (s icon) Draw(screen tcell.Screen, style tcell.Style) {
	x := s.begin.X
	y := s.begin.Y
	for _, v := range s.s {
		if v == ' ' {
			screen.SetContent(x, y, v, nil, style)
		} else {
			screen.SetContent(x, y, v, nil, style)
		}
		x++
	}
}

func (s *icon) in(p Pos) bool {
	if p.GreaterEqual(s.begin) && p.LessEqual(s.end) {
		return true
	}
	return false
}
func (s *icon) relocate(x, y int) int {
	s.begin = Pos{x, y}
	s.end = Pos{x + len(s.s) - 1, y}
	return s.end.X + 1
}

var block_str = '■'
var str_back = '◀'
var str_forward = '▶'

var file_rune = '\U000f10aa'
var outline_rune = '\U000f10ab'

type minitoolbar struct {
	item  []icon
	getxy func() (int, int)
}

func (toobar *minitoolbar) get_offset_xy() (int, int) {
	if toobar.getxy != nil {
		return toobar.getxy()
	}
	return 0, 0
}
func (toolbar *minitoolbar) Relocated() {
	left, top := toolbar.get_offset_xy()
	for i := range toolbar.item {
		a := &toolbar.item[i]
		left = a.relocate(left, top)
	}
}
func (toolbar *minitoolbar) handle_mouse_event(action tview.MouseAction, event *tcell.EventMouse) (*tcell.EventMouse, tview.MouseAction) {
	if event == nil {
		return event, action
	}
	toolbar.Relocated()

	x, y := event.Position()
	loc := Pos{X: x, Y: y}
	if action == tview.MouseLeftClick {
		for i := range toolbar.item {
			a := &toolbar.item[i]
			if a.in(loc) {
				if a.click != nil {
					a.click()
				}
				return nil, tview.MouseConsumed
			}
		}
		return nil, tview.MouseConsumed
	}
	return event, action
}

func (c *minitoolbar) Width() int {
	w := 0
	for _, v := range c.item {
		w += len(v.s)
	}
	return w
}
func (c *minitoolbar) Draw(screen tcell.Screen) {
	c.Relocated()
	for i := range c.item {
		a := &c.item[i]
		a.Draw(screen, a.style())
	}
}
func new_quick_toolbar(main *mainui) *minitoolbar {
	var quick_btn icon = icon{
		s: []rune{block_str},
		click: func() {

		},
		style: func() tcell.Style {
			return get_style_hide(false)
		},
	}
	var index_bt icon = icon{
		s: []rune{block_str},
		click: func() {
			main.layout.console.resizer.toggle(view_qf_index_view.to_view_link(main))
			main.app.ForceDraw()
		},
		style: func() tcell.Style {
			return get_style_hide(view_qf_index_view.to_view_link(main).Hide)
		},
	}
	icon := []icon{quick_btn, index_bt}
	ret := &minitoolbar{
		item: icon,
	}
	ret.getxy = func() (int, int) {
		x, y, w, _ := main.page.GetRect()
		x = x + w - ret.Width()
		return x, y
	}
	return ret
}
// func new_top_toolbar(main *mainui) *minitoolbar {
// 	str_back = '\U000f0b28'
// 	str_forward = '\U000f0b2a'
// 	icons := []icon{}
// 	if false {
// 		for i := range SplitCode.index {
// 			id := SplitCode.index[i]
// 			if view, ok := SplitCode.code_collection[id]; ok {
// 				a := icon{click: func() {
// 					if id != view_code {
// 						if view, ok := SplitCode.code_collection[id]; ok {
// 							SplitClose(view).handle()
// 						}
// 					}
// 				}, style: func() tcell.Style {
// 					style := CodeIconStyle(view, main)
// 					return style
// 				}}
// 				for _, v := range FileIcon(view.FileName()) {
// 					a.s = append(a.s, v)
// 				}
// 				if id == view_code {
// 					icons = append([]icon{a}, icons...)
// 				} else {
// 					icons = append(icons, a)
// 				}
// 			}
// 		}
// 	}
// 	var back = icon{
// 		s: []rune{' ', str_back, ' '},
// 		click: func() {
// 			main.GoBack()
// 		},
// 		style: func() tcell.Style {
// 			return get_style_hide(!main.CanGoBack())
// 		},
// 	}
// 	var forward = icon{
// 		s: []rune{str_forward, ' '},
// 		click: func() {
// 			main.GoForward()
// 		},
// 		style: func() tcell.Style {
// 			return get_style_hide(!main.CanGoFoward())
// 		},
// 	}
// 	var file = icon{
// 		s: []rune{file_rune, ' '},
// 		click: func() {
// 			main.toggle_view(view_file)
// 		},
// 		style: func() tcell.Style {
// 			return get_style_hide(view_file.to_view_link(main).Hide)
// 		},
// 	}
// 	var outline = icon{
// 		s: []rune{outline_rune, ' '},
// 		click: func() {
// 			main.toggle_view(view_outline_list)
// 		},
// 		style: func() tcell.Style {
// 			return get_style_hide(view_outline_list.to_view_link(main).Hide)
// 		},
// 	}
// 	icons = append(icons, []icon{back, forward, file, outline}...)
// 	ret := &minitoolbar{
// 		item: icons,
// 	}
// 	ret.getxy = func() (int, int) {
// 		v := SplitCode.Last()
// 		if v == nil {
// 			v = main.codeview
// 		}
// 		x, y, w, _ := v.view.GetRect()
// 		x = x + w - ret.Width()
// 		return x, y
// 	}
// 	return ret
// }

func CodeIconStyle(view *CodeView, main MainService) tcell.Style {
	return view.view.IconStyle(main)
}

type IconButton struct {
	*tview.Box
	r        rune
	selected bool
	click    func(bool)
}

func NewIconButton(r rune) *IconButton {
	return &IconButton{
		Box: tview.NewBox(),
		r:   r,
	}
}
func (icon *IconButton) Draw(screen tcell.Screen) {
	icon.Box.DrawForSubclass(screen, icon)
	x, y, _, _ := icon.GetRect()
	s := get_style_hide(!icon.selected)
	screen.SetContent(x, y, icon.r, nil, s)
}
func (c *IconButton) Primitive() tview.Primitive {
	return c
}

func (c *IconButton) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if InRect(event, c) {
			if action == tview.MouseLeftClick || action == tview.MouseLeftDown {
				c.selected = !c.selected
				if c.click != nil {
					c.click(c.selected)
				}
			}
		}
		return
	}
}
