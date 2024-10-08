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

type smallicon struct {
	back, forward icon
	main          *mainui
	x, y          int
	file, outline icon
	code          []icon
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
func (c *smallicon) Loc(loc Pos) Pos {
	loc.X += c.x
	loc.Y += c.y
	return loc
}

var block_str = '■'
var str_back = '◀'
var str_forward = '▶'

func (c *smallicon) Relocated() {
	left, top := c.get_offset_xy()
	c.code = make([]icon, len(SplitCode.code_collection))
	left = c.file.relocate(left, top)
	for i := range c.code {
		c.code[i].s = []rune{block_str}
		left = c.code[i].relocate(left, top)
	}
	left = c.outline.relocate(left, top)
	left = c.back.relocate(left, top)
	c.forward.relocate(left, top)
}

func (c *smallicon) Draw(screen tcell.Screen) {
	c.Relocated()
	main := c.main
	focus_color := tcell.ColorYellow
	x := get_style_hide(view_outline_list.to_view_link(main).Hide)
	if view_outline_list.to_box(c.main).HasFocus() {
		x = x.Foreground(focus_color)
	}
	c.outline.Draw(screen, x)

	x = get_style_hide(view_file.to_view_link(main).Hide)
	if view_file.to_box(c.main).HasFocus() {
		x = x.Foreground(focus_color)
	}
	c.file.Draw(screen, x)

	for i, v := range c.code {
		style := get_style_hide(false)
		id := SplitCode.index[i]
		focus := false
		if view, ok := SplitCode.code_collection[id]; ok {
			focus = view.view.HasFocus()||view==c.main.current_editor()
			if focus {
				style = style.Foreground(focus_color)
			}
		}
		v.Draw(screen, style)
	}

	c.back.Draw(screen, get_style_hide(!c.main.CanGoBack()))
	c.forward.Draw(screen, get_style_hide(!c.main.CanGoFoward()).Bold(true))
}

func new_small_icon(main *mainui) *smallicon {
	smallicon := &smallicon{
		file:    icon{s: []rune{block_str}},
		outline: icon{s: []rune{block_str}},
		back:    icon{s: []rune{' ', str_back, ' '}},
		forward: icon{s: []rune{str_forward}},
		main:    main,
	}
	return smallicon
}

func (icon *smallicon) handle_mouse_event(action tview.MouseAction, event *tcell.EventMouse) (*tcell.EventMouse, tview.MouseAction) {
	if event == nil {
		return event, action
	}
	icon.Relocated()

	x, y := event.Position()
	loc := Pos{X: x, Y: y}
	if action == tview.MouseLeftClick {
		// if action == tview.MouseLeftClick || action == tview.MouseLeftDown {
		for i, v := range icon.code {
			if v.in(loc) {
				id := SplitCode.index[i]
				if view, ok := SplitCode.code_collection[id]; ok {
					if i == 0 {
						view.id.setfocused(icon.main)
					} else {
						SplitClose(view).handle()
					}
				}
				return nil, tview.MouseConsumed
			}
		}
		if icon.file.in(loc) {
			icon.main.toggle_view(view_file)
		} else if icon.outline.in(loc) {
			icon.main.toggle_view(view_outline_list)
		} else if icon.back.in(loc) {
			icon.main.GoBack()
		} else if icon.forward.in(loc) {
			icon.main.GoForward()
		} else {
			return event, action
		}
		return nil, tview.MouseConsumed
	}
	return event, action
}

func (icon *smallicon) get_offset_xy() (int, int) {
	v := SplitCode.Last()
	if v == nil {
		v = icon.main.codeview
	}
	left, top, w, _ := v.view.GetRect()
	left += w
	left -= 10
	return left, top
}

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
