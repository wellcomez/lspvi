package mainui

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type hlItem struct {
}
type customlist struct {
	*tview.List
	hlitems []*hlItem
	Key     string
	fuzz    bool
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
	ret.fuzz = false
	return ret
}
func (l *customlist) AddItem(mainText, secondText string, selected func()) *customlist {
	l.hlitems = append(l.hlitems, &hlItem{})
	l.List.AddItem(mainText, secondText, 0, selected)
	return l
}

type keypattern struct {
	begin int
	width int
}

func find_key_fuzzy2(s string, keys []string, offset int) []keypattern {
	for i := 0; i < len(keys); i++ {
		v := strings.Join(keys[:len(keys)-i], "")
		idx := strings.Index(strings.ToLower(s), v)
		if idx >= 0 {
			pth := keypattern{begin: idx + offset, width: len(v)}
			a := []keypattern{pth}
			subret := find_key_fuzzy2(s[idx+len(v):], keys[len(v):], pth.width+idx+offset)
			return append(a, subret...)
		}
	}
	return []keypattern{}
}

//	func find_key_fuzzy(s string, keys []string, offset int) []keypattern {
//		for i, v := range keys {
//			if len(v) == 0 {
//				continue
//			}
//			idx := strings.Index(strings.ToLower(s), v)
//			if idx >= 0 {
//				pth := keypattern{begin: idx + offset, width: len(v)}
//				a := []keypattern{pth}
//				subret := find_key_fuzzy(s[idx+len(v):], keys[i+1:], pth.width+idx+offset)
//				return append(a, subret...)
//			}
//		}
//		return []keypattern{}
//	}
func find_hl_key(ss string) ([]string, string) {
	key := "**"
	return remove_hl_flag(ss, key)
}

func remove_hl_flag(ss string, key string) ([]string, string) {
	keys := []string{}
	s := ss
	for len(s) > 0 {
		b := strings.Index(s, key)
		if b >= 0 {
			e := strings.Index(s[b+1:], key)
			if e > 0 {
				key := s[b+2 : b+e+1]
				keys = append(keys, key)
				s = s[b+e+3:]
			} else {
				break
			}
		} else {
			break
		}
	}
	for _, v := range keys {
		ss = strings.ReplaceAll(ss, fmt.Sprintf("%s%s%s", key, v, key), v)
	}
	return keys, ss
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
	keys2 := []string{}
	for _, v := range l.Key {
		keys2 = append(keys2, string(v))
	}
	for index := itemoffset; index < len(l.hlitems); index++ {
		MainText, SecondText := l.List.GetItemText(index)
		MainText, main_postion := get_hl_postion(MainText, keys, l, keys2)
		selected := index == l.List.GetCurrentItem()
		if y >= bottomLimit {
			break
		}
		if len(MainText) > 0 {
			if selected {
				draw_item_color(main_postion, MainText, screen, offset_x, y, selected_style, selected_stylehl)
			} else {
				draw_item_color(main_postion, MainText, screen, offset_x, y, style, stylehl)
			}
			y += 1
		}
		if y >= bottomLimit {
			break
		}
		if l.showSecondaryText() && len(SecondText) > 0 {
			SecondText, main_postion := get_hl_postion(SecondText, keys, l, keys2)
			// if selected {
			// 	draw_item_color(main_postion, SecondText, screen, offset_x, y, selected_style, selected_stylehl)
			// } else {
			draw_item_color(main_postion, SecondText, screen, offset_x, y, style, stylehl)
			// }
			y += 1
			if y >= bottomLimit {
				break
			}
		}
	}

}

func get_hl_postion(MainText string, keys []string, l *customlist, keys2 []string) (string, []keypattern) {
	hlkey, MainText := find_hl_key(MainText)
	hlkey = append(hlkey, keys...)
	main_postion := find_key(MainText, hlkey, 0)
	if l.fuzz && len(main_postion) == 0 && len(keys2) > 0 {
		main_postion = find_key_fuzzy2(MainText, keys2, 0)
	}
	return MainText, main_postion
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
