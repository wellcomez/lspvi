package mainui

import (
	// "log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TabButton struct {
	view  *tview.Button
	Name  string
	group *ButtonGroup
}
type ButtonGroup struct {
	tabs    []*TabButton
	handler func(tab *TabButton)
}

func (group ButtonGroup) Find(tab string) *TabButton {
	for _, v := range group.tabs {
		if v.Name == tab {
			return v
		}
	}
	return nil
}
func (group ButtonGroup) onselected(tab *TabButton) {
	for _, v := range group.tabs {
		if v == tab {
			v.view.Focus(nil)
		} else {
			v.view.Blur()
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
func NewTab(name string, group *ButtonGroup) *TabButton {
	var style tcell.Style
	// var style1 tcell.Style
	// style1.Foreground(tcell.ColorGreen)
	ret := &TabButton{
		Name:  name,
		view:  tview.NewButton(name).SetStyle(style),
		group: group,
	}
	ret.view.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseMove {
			return action, event
		}
		x, y := event.Position()
		if action == tview.MouseLeftDown {
			if ret.view.InRect(x, y) {
				ret.selected()
			}
		}
		return action, event
	})
	// ret.view.SetSelectedFunc(ret.selected)
	return ret
}
