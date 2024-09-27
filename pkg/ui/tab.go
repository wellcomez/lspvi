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
	tabs []TabItem
}

func (tab *TabItem) Draw(screen tcell.Screen, x, y int, style, hl tcell.Style) int {
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

func NewTabbar() *Tabbar {
	return &Tabbar{Box: tview.NewBox()}
}
func (bar *Tabbar) Add(name string) int{
	if len(bar.tabs) > 0 {
		bar.tabs[len(bar.tabs)-1].first = false
	}
	bar.tabs = append(bar.tabs, TabItem{tview.Box{}, name, true, false})
	x := 0
	y := 0
	ret:=0
	for i := range bar.tabs {
		tab := &bar.tabs[i]
		width := len(tab.name)
		if tab.first {
			width = len(tab.name) + 1
		}
		tab.SetRect(x, y, width, 1)
		ret+=width
	}
	return ret
}
func (bar *Tabbar) Draw(screen tcell.Screen) {
	style := tcell.StyleDefault
	if s := global_theme.get_default_style(); s != nil {
		style = *s
	}
	hlstyle := style.Foreground(global_theme.search_highlight_color())
	bar.Box.DrawForSubclass(screen, bar)
	x, y, _, _ := bar.GetRect()
	posX := x
	posY := y
	for i, v := range bar.tabs {
		v.first = i == 0
		posX = v.Draw(screen, posX, posY, style, hlstyle)
	}
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
					break
				}
			}
			consumed = true
		}
		return
	})
}
