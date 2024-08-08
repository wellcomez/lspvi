package mainui

import (
	"reflect"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type hlItem struct {
	Positions []int
}
type customlist struct {
	*tview.List
	hlitems []*hlItem
	Key     string
}

func (l *customlist) Clear() *customlist {
	l.List.Clear()
	l.hlitems = []*hlItem{}
	return l
}
func new_customlist() *customlist {
	ret := &customlist{}
	ret.List = tview.NewList()
	ret.hlitems = []*hlItem{}
	return ret
}
func (l *customlist) AddItem(mainText string, Positions []int, selected func()) *customlist {
	l.hlitems = append(l.hlitems, &hlItem{Positions: Positions})
	l.List.AddItem(mainText, "", 0, selected)
	return l
}

type keypattern struct {
	begin int
	width int
}

func find_key(s string, keys []string, offset int) []keypattern {
	for _, v := range keys {
		if len(v) == 0 {
			continue
		}
		idx := strings.Index(strings.ToLower(s), v)
		if idx >= 0 {
			pth := keypattern{begin: idx + offset, width: len(v)}
			a := []keypattern{pth}
			subret := find_key(s[idx+len(v):], keys, pth.width+idx+offset)
			return append(a, subret...)
		}
	}
	return []keypattern{}
}
func (l *customlist) Draw(screen tcell.Screen) {
	l.Box.DrawForSubclass(screen, l.Box)
	offset_x, y, _, height := l.GetInnerRect()

	bottomLimit := y + height

	selected_style := tcell.StyleDefault.Foreground(tview.Styles.PrimitiveBackgroundColor).Background(tview.Styles.PrimaryTextColor)
	selected_stylehl := tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tview.Styles.PrimaryTextColor)

	style := tcell.StyleDefault.Foreground(tview.Styles.PrimaryTextColor).Background(tview.Styles.PrimitiveBackgroundColor)
	stylehl := tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tview.Styles.PrimitiveBackgroundColor)

	itemoffset, _ := l.GetOffset()
	keys := l.get_hl_keys()
	for index := itemoffset; index < len(l.hlitems); index++ {
		MainText, SecondText := l.List.GetItemText(index)
		Positions := find_key(MainText, keys, 0)
		selected := index == l.List.GetCurrentItem()
		if y >= bottomLimit {
			break
		}
		if len(MainText) > 0 {
			if selected {
				draw_item_color(Positions, MainText, screen, offset_x, y, selected_style, selected_stylehl)
			} else {
				draw_item_color(Positions, MainText, screen, offset_x, y, style, stylehl)
			}
			y += 1
		}
		if y >= bottomLimit {
			break
		}
		if l.showSecondaryText() && len(SecondText) > 0 {
			if selected {
				draw_item_color(Positions, SecondText, screen, offset_x, y, selected_style, selected_stylehl)
			} else {
				draw_item_color(Positions, SecondText, screen, offset_x, y, style, stylehl)
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
func (l *customlist) get_hl_keys() []string {
	keys := strings.Split(l.Key, " ")
	if len(l.Key) == 0 {
		keys = []string{}
	} else {
		for i, s := range keys {
			keys[i] = strings.ToLower(s)
		}
	}
	return keys
}

func draw_item_color(Positions []keypattern, MainText string, screen tcell.Screen, offset_x int, y int, selected_style tcell.Style, selected_stylehl tcell.Style) {
	begin := 0
	for _, e := range Positions {
		normal := MainText[begin:e.begin]
		for i, r := range normal {
			screen.SetContent(offset_x+i+begin, y, r, nil, selected_style)
		}
		hltext := MainText[e.begin : e.begin+e.width]
		for i, r := range hltext {
			screen.SetContent(offset_x+i+e.begin, y, r, nil, selected_stylehl)
		}
		begin = e.begin + e.width
	}
	if begin < len(MainText) {
		normal := MainText[begin:]
		for i, r := range normal {
			screen.SetContent(offset_x+i+begin, y, r, nil, selected_style)
		}
	}
}
