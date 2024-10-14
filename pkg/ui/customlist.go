package mainui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"reflect"
)

type hlItem struct {
}
type customlist struct {
	*tview.List
	hlitems       []*hlItem
	Key           string
	fuzz          bool
	default_color tcell.Color
	selected      []int
}

func (l *customlist) Clear() *customlist {
	l.List.Clear()
	l.hlitems = []*hlItem{}
	return l
}
func new_customlist(two bool) *customlist {
	ret := &customlist{default_color: global_theme.search_highlight_color()}
	ret.List = tview.NewList()
	ret.ShowSecondaryText(two)
	ret.hlitems = []*hlItem{}
	ret.fuzz = false
	return ret
}
func (l *customlist) AddItem(mainText, secondText string, selected func()) *customlist {
	l.hlitems = append(l.hlitems, &hlItem{})
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
	if s := global_theme.get_color("selection"); s != nil {
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
	for index := itemoffset; index < len(l.hlitems); index++ {
		MainText, SecondText := l.List.GetItemText(index)
		// MainText, main_postion := get_hl_postion(MainText, keys, l, keys2)
		selected := index == l.List.GetCurrentItem()
		main_text := GetColorText(MainText, []colortext{colortext{l.Key, l.default_color}})
		second_text := GetColorText(SecondText, []colortext{colortext{l.Key, l.default_color}})
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
		if len(MainText) > 0 {
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
		if l.showSecondaryText() && len(SecondText) > 0 {
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

func (list *customlist) showSecondaryText() bool {
	v := reflect.ValueOf(list.List).Elem()
	field := v.FieldByName("showSecondaryText")
	x := field.Bool()
	return x
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
