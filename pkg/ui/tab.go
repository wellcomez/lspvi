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

var default_btn_style= tcell.StyleDefault.Foreground(tview.Styles.PrimaryTextColor).Background(tcell.ColorBlack)
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
