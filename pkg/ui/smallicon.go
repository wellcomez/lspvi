package mainui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func get_style_hide(hide bool) tcell.Style {
	style := *global_theme.get_default_style()
	hl := global_theme.search_highlight_color()
	f, b, _ := style.Decompose()
	hide_stycle := style.Foreground(f).Background(b)
	x1 := style.Foreground(hl).Background(b)
	if !hide {
		hide_stycle = x1
	}
	return hide_stycle
}

type smallicon struct {
	file, code, outline Pos
	back, forward       Pos
	main                *mainui
	x, y                int
}

func (c *smallicon) Loc(loc Pos) Pos {
	loc.X += c.x
	loc.Y += c.y
	return loc
}
func (c *smallicon) Draw(screen tcell.Screen) {
	main := c.main
	ch := '█'
	ch = '■'

	left, top := c.get_offset_xy()
	forward := '→'
	back := '←'

	back = '◀'
	forward = '▶'
	style := *global_theme.get_default_style()
	screen.SetContent(c.file.X+left, c.file.Y+top, ch, nil, get_style_hide(view_file.to_view_link(main).Hide))
	screen.SetContent(c.code.X+left, c.code.Y+top, ch, nil, get_style_hide(view_code.to_view_link(main).Hide))
	screen.SetContent(c.outline.X+left, c.outline.Y+top, ch, nil, get_style_hide(view_outline_list.to_view_link(main).Hide))

	screen.SetContent(c.back.X-1+left, top+c.back.Y, ' ', nil, style.Foreground(tcell.ColorWhite).Bold(true))
	screen.SetContent(c.back.X+left, top+c.back.Y, back, nil, get_style_hide(!c.main.CanGoBack()))
	screen.SetContent(c.back.X+1+left, top+c.back.Y, ' ', nil, style.Foreground(tcell.ColorWhite))
	screen.SetContent(c.forward.X+left, top+c.forward.Y, forward, nil, get_style_hide(!c.main.CanGoFoward()).Bold(true))
	screen.SetContent(c.forward.X+1+left, top+c.forward.Y, ' ', nil, style.Foreground(tcell.ColorWhite))
}
func new_small_icon(main *mainui) *smallicon {
	smallicon := &smallicon{
		file:    Pos{0, 0},
		code:    Pos{1, 0},
		outline: Pos{2, 0},
		back:    Pos{4, 0},
		forward: Pos{6, 0},
		main:    main,
	}

	return smallicon
}
func (icon *smallicon) handle_mouse_event(action tview.MouseAction, event *tcell.EventMouse) (*tcell.EventMouse, tview.MouseAction) {
	if event == nil {
		return event, action
	}
	x, y := event.Position()
	left, top := icon.get_offset_xy()
	loc := Pos{X: x - left, Y: y - top}
	if action == tview.MouseLeftClick {
		// if action == tview.MouseLeftClick || action == tview.MouseLeftDown {
		switch loc {
		case icon.code:
			{

			}
		case icon.file:
			{
				icon.main.toggle_view(view_file)
			}
		case icon.outline:
			{
				icon.main.toggle_view(view_outline_list)
			}
		case icon.back:
			{
				icon.main.GoBack()
			}
		case icon.forward:
			{
				icon.main.GoForward()
			}
		default:
			return event, action
		}
		return nil, tview.MouseConsumed
	}
	return event, action
}

func (icon *smallicon) get_offset_xy() (int, int) {
	left, top, w, _ := icon.main.codeview.view.GetRect()
	left += w
	left -= 10
	return left, top
}
