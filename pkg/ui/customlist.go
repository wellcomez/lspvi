// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type customlist struct {
	*List
	Key               string
	fuzz              bool
	default_color     tcell.Color
	selected          []int
	main_color_text   [][]colortext
	second_color_text [][]colortext
}

func (l *customlist) Clear() *customlist {
	l.List.Clear()
	l.main_color_text = [][]colortext{}
	l.second_color_text = [][]colortext{}
	return l
}
func new_customlist(two bool) *customlist {
	ret := &customlist{default_color: global_theme.search_highlight_color()}
	ret.List = NewList()
	ret.ShowSecondaryText(two)
	ret.main_color_text = [][]colortext{}
	ret.second_color_text = [][]colortext{}
	ret.fuzz = false
	return ret
}
func (l *customlist) AddColorItem(main, second []colortext, selected func()) *customlist {
	l.main_color_text = append(l.main_color_text, main)
	maintext := ""
	for _, v := range main {
		maintext += v.text
	}
	second_text := ""
	for _, v := range second {
		second_text += v.text
	}
	l.List.AddItem(maintext, second_text, 0, selected)
	return l
}
func (l *customlist) AddItem(mainText, secondText string, selected func()) *customlist {
	// l.hlitems = append(l.hlitems, &hlItem{})
	l.List.AddItem(mainText, secondText, 0, selected)
	return l
}

type colorkey struct {
	str   string
	color tcell.Color
}

func (l *customlist) NewDefaultColorKey(key string) colorkey {
	return colorkey{str: key, color: l.default_color}
}
func (l *customlist) Draw(screen tcell.Screen) {
	l.Box.DrawForSubclass(screen, l.Box)
	offset_x, y, width, height := l.GetInnerRect()

	bottomLimit := y + height
	select_color := global_theme.search_highlight_color()
	selected_style := tcell.StyleDefault.Foreground(tview.Styles.PrimitiveBackgroundColor).Background(tview.Styles.PrimaryTextColor)

	style := tcell.StyleDefault.Foreground(tview.Styles.PrimaryTextColor).Background(tview.Styles.PrimitiveBackgroundColor)
	stylehl := tcell.StyleDefault.Foreground(select_color).Background(tview.Styles.PrimitiveBackgroundColor)
	theme_style := stylehl
	if s := global_theme.select_style(); s != nil {
		theme_style = *s
	}

	itemoffset, _ := l.GetOffset()
	// keys := []colorkey{}
	// for _, v := range l.get_hl_keys() {
	// 	keys = append(keys, l.NewDefaultColorKey(v))
	// }
	// keys2 := []colorkey{}
	// for _, v := range l.Key {
	// 	keys2 = append(keys2, l.NewDefaultColorKey(string(v)))
	// }
	for index := itemoffset; index < l.GetItemCount(); index++ {
		var main_text, second_text []colortext
		selected := index == l.List.GetCurrentItem()
		var has_main, has_second bool
		if len(l.main_color_text) > 0 {
			main_text = append([]colortext{}, l.main_color_text[index]...)
			if len(l.second_color_text) > 0 {
				second_text = append([]colortext{}, l.second_color_text[index]...)
			}
		} else {
			MainText, SecondText := l.List.GetItemText(index)
			main_text = GetColorText(MainText, []colortext{{l.Key, l.default_color, 0}})
			second_text = GetColorText(SecondText, []colortext{{l.Key, l.default_color, 0}})
		}
		has_main = len(main_text) > 0
		has_second = len(second_text) > 0
		if selected {
			for i := range main_text {
				main_text[i].color = 0
			}
			for i := range second_text {
				second_text[i].color = 0
			}
		}

		multiselected := false
		if len(l.selected) > 0 {
			if index >= l.selected[0] && index <= l.selected[1] {
				multiselected = true
			}
		}
		if y >= bottomLimit {
			break
		}
		if has_main {
			if multiselected {
				l.draw_item_color_new(main_text, screen, offset_x, y, width, theme_style)
			} else if selected {
				l.draw_item_color_new(main_text, screen, offset_x, y, width, selected_style)
			} else {
				l.draw_item_color_new(main_text, screen, offset_x, y, width, style)
			}
			y += 1
		}
		if y >= bottomLimit {
			break
		}
		if l.showSecondaryText && has_second {
			if selected {
				l.draw_item_color_new(second_text, screen, offset_x, y, width, selected_style)
			} else {
				l.draw_item_color_new(second_text, screen, offset_x, y, width, style)
			}
			y += 1
			if y >= bottomLimit {
				break
			}
		}
	}

}

func (l *customlist) draw_item_color_new(segment []colortext, screen tcell.Screen, offset_x int, y int, width int, normal_style tcell.Style) {
	x := offset_x
	max := x + width
	for _, e := range segment {
		for _, r := range e.text {
			if x < max {
				if e.color == 0 {
					screen.SetContent(x, y, r, nil, normal_style)
				} else {
					screen.SetContent(x, y, r, nil, normal_style.Foreground(e.color))
				}
				x++
			} else {
				break
			}
		}
	}
}
