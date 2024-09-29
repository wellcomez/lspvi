package mainui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TabButton struct {
	*tview.Button
	Name  string
	group *ButtonGroup
}
type ButtonGroup struct {
	tabs      []*TabButton
	handler   func(tab *TabButton)
	activetab *TabButton
}

func (group ButtonGroup) Find(tab string) *TabButton {
	for _, v := range group.tabs {
		if v.Name == tab {
			return v
		}
	}
	return nil
}
func (group *ButtonGroup) onselected(tab *TabButton) {
	for _, v := range group.tabs {
		if v == tab {
			group.activetab = v
			v.Focus(nil)
		} else {
			v.Blur()
		}
	}
	group.handler(tab)
}
func NewButtonGroup(tabs []string, handler func(tab *TabButton)) *ButtonGroup {
	ret := &ButtonGroup{
		handler: handler,
	}
	var i = 0
	for i = 0; i < len(tabs); i++ {
		ret.tabs = append(ret.tabs, NewTab(tabs[i], ret))
	}
	return ret
}
func (btn *TabButton) selected() {
	btn.group.onselected(btn)
}

var default_btn_style = tcell.StyleDefault.Foreground(tview.Styles.PrimaryTextColor).Background(tcell.ColorBlack)
var active_btn_style = tcell.StyleDefault.Background(tview.Styles.InverseTextColor).Foreground(tview.Styles.PrimaryTextColor)

func NewTab(name string, group *ButtonGroup) *TabButton {
	Button := tview.NewButton(name).SetStyle(default_btn_style).SetActivatedStyle(active_btn_style)
	// var style1 tcell.Style
	// style1.Foreground(tcell.ColorGreen)
	ret := &TabButton{
		Button: Button,
		Name:   name,
		group:  group,
	}
	ret.SetBlurFunc(func() {
		if group.activetab == ret {
			ret.SetStyle(active_btn_style)
		} else {
			ret.SetStyle(default_btn_style)
		}
	})
	ret.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseMove {
			return action, event
		}
		x, y := event.Position()
		if action == tview.MouseLeftDown {
			if ret.InRect(x, y) {
				ret.selected()
			}
		}
		return action, event
	})
	// ret.view.SetSelectedFunc(ret.selected)
	return ret
}

type TabItem struct {
	tview.Box
	name   string
	first  bool
	active bool
}
type Tabbar struct {
	*tview.Box
	tabs   []TabItem
	active func(string)
	mode   tab_style
}

var space = true

func (tab *TabItem) Draw(screen tcell.Screen, x, y int, style, hl tcell.Style, prevactive bool) int {
	x = tab.draw_btn_mode(screen, x, y, style, hl, prevactive)
	return x
}
func (tab *TabItem) draw_btn_mode(screen tcell.Screen, x, y int, style, hl tcell.Style, prevactive bool) int {
	if !tab.first {
		screen.SetContent(x, y, ' ', nil, style)
		x++
	}
	s := style
	if tab.active {
		s = hl
	}
	for _, ch := range tab.name {
		screen.SetContent(x, y, ch, nil, s)
		x++
	}
	return x
}
func (tab *TabItem) draw_tab_mode(screen tcell.Screen, x, y int, style, hl tcell.Style, prevactive bool) int {
	s := style
	if tab.active {
		s = hl
	}
	if space {
		if !tab.first && !prevactive {
			screen.SetContent(x, y, ' ', nil, s)
			x++
		}
	}

	for _, ch := range tab.name {
		screen.SetContent(x, y, ch, nil, s)
		x++
	}
	if space {
		if tab.active {
			screen.SetContent(x, y, ' ', nil, s)
			x++
		}
	}
	return x
}

func NewTabbar(active func(string)) *Tabbar {
	return &Tabbar{Box: tview.NewBox(), active: active}
}
func (bar *Tabbar) Add(name string) int {
	if len(bar.tabs) > 0 {
		bar.tabs[len(bar.tabs)-1].first = false
	}
	bar.tabs = append(bar.tabs, TabItem{tview.Box{}, name, true, false})
	x := 0
	y := 0
	ret := 0
	for i := range bar.tabs {
		tab := &bar.tabs[i]
		width := len(tab.name)
		if space {
			if !tab.first {
				width = len(tab.name) + 1
			}
		}
		tab.SetRect(x, y, width, 1)
		ret += width
		x += width
	}
	return ret
}
func (bar *Tabbar) Active(s string) {
	for i := range bar.tabs {
		v := &bar.tabs[i]
		if v.name == s {
			v.active = true
		} else {
			v.active = false
		}
	}
}
func (bar *Tabbar) Draw(screen tcell.Screen) {

	// .Underline(true)
	hlstyle, style := style_mode(bar.mode)
	bar.Box.DrawForSubclass(screen, bar)
	x, y, _, _ := bar.GetRect()
	posX := x
	posY := y
	for i, v := range bar.tabs {
		v.first = i == 0
		pre_active := false
		if i > 0 {
			pre_active = bar.tabs[i-1].active
		}
		switch bar.mode {
		case tab_style_btn:
			posX = v.draw_btn_mode(screen, posX, posY, style, hlstyle, pre_active)
		case tab_style_tab:
			posX = v.draw_tab_mode(screen, posX, posY, style, hlstyle, pre_active)
		}
	}
}

type tab_style int

const (
	tab_style_btn tab_style = iota
	tab_style_tab
)

func style_mode(mode tab_style) (tcell.Style, tcell.Style) {
	style := tcell.StyleDefault
	if s := global_theme.get_default_style(); s != nil {
		style = *s
	}
	_, b, _ := active_btn_style.Decompose()
	hlstyle := style.Foreground(b)
	seleted := tcell.ColorBlack
	if s := global_theme.get_color("selection"); s != nil {
		_, b, _ := s.Decompose()
		seleted = b
	}
	switch mode {
	case tab_style_btn:
		style = style.Background(seleted)
		return style, hlstyle
	case tab_style_tab:
		return hlstyle, style
	}
	return hlstyle, style
}
func (l *Tabbar) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return l.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	})
}

// MouseHandler returns the mouse handler for this primitive.
func (l *Tabbar) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return l.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if !l.InRect(event.Position()) {
			return false, nil
		}
		x, y, _, _ := l.GetRect()

		// Process mouse event.
		switch action {
		case tview.MouseLeftClick:
			for i := range l.tabs {
				b := l.tabs[i].Box
				bx, by, bw, bh := b.GetRect()
				b.SetRect(bx+x, by+y, bw, bh)
				if b.InRect(event.Position()) {
					l.tabs[i].active = true
					if l.active != nil {
						l.active(l.tabs[i].name)
					}
				} else {
					l.tabs[i].active = false
				}
			}
			consumed = true
		}
		return
	})
}
