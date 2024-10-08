package mainui

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
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
type colortext struct {
	text  string
	color tcell.Color
}

func fmt_color_string(s string, color tcell.Color) string {
	return fmt.Sprintf("**[%d]%s**", color, s)
}

type splitresult struct {
	b, m, a colortext
}

func pasrse_bold_color_string(s string) splitresult {
	b := strings.Index(s, "**")
	if b >= 0 {
		e := strings.Index(s[b+2:], "**")
		if e >= 0 {
			return splitresult{
				b: colortext{text: s[:b]},
				m: colortext{text: s[b+2 : b+2+e],color: tcell.ColorYellow},
				a: colortext{text: s[b+2+e+2:]},
			}
		}
	}
	return splitresult{b: colortext{text: s}}
}

type colorpaser struct {
	data string
}

func (p *colorpaser) Parse() []colortext {
	r3 := pasrse_color_string(p.data)
	if r3.m.text == "" {
		r3 = pasrse_bold_color_string(p.data)
	}
	if len(r3.m.text) > 0 {
	var before_part []colortext
		if r3.b.text != "" {
			aa := colorpaser{data: r3.b.text}
			result := aa.Parse()
			for _, v := range result {
				if len(v.text) > 0 {
					before_part = append(before_part, v)
				}
			}
		}
		var after_part []colortext
		if r3.a.text != "" {
			aa := colorpaser{data: r3.a.text}
			result := aa.Parse()
			for _, v := range result {
				if len(v.text) > 0 {
					after_part = append(after_part, v)
				}
			}
		}
		before_part = append(before_part, r3.m)
		before_part = append(before_part, after_part...)
		return before_part 
	}
	return []colortext{r3.a}
}
func pasrse_color_string(s string) splitresult {
	b := strings.Index(s, "**[")
	if b >= 0 {
		e := strings.Index(s, "]")
		if e >= 0 {
			var color tcell.Color
			if c, err := strconv.Atoi(s[b+3 : e]); err == nil {
				color = tcell.Color(c)
				if e2 := strings.Index(s[e+1:], "**"); e2 > 0 {
					x := e + 1
					return splitresult{b: colortext{text: s[:b]}, m: colortext{text: s[x : x+e2], color: color}, a: colortext{text: s[x+e2+2:]}}
				}
			}
		}
	}
	return splitresult{b: colortext{text: s}}
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

type keypattern struct {
	begin int
	width int
	color tcell.Color
}

func find_key_fuzzy2(s string, colore_keys []colorkey, offset int) []keypattern {
	for i := 0; i < len(colore_keys); i++ {
		v := ""
		for _, k := range colore_keys[:len(colore_keys)-i] {
			v = v + k.str
		}
		idx := strings.Index(strings.ToLower(s), v)
		if idx >= 0 {
			pth := keypattern{begin: idx + offset, width: len(v), color: global_theme.search_highlight_color()}
			a := []keypattern{pth}
			subret := find_key_fuzzy2(s[idx+len(v):], colore_keys[len(v):], pth.width+idx+offset)
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
func (l *customlist) find_hl_key(ss string) ([]colorkey, string) {
	key := "**"
	return l.remove_hl_flag(ss, key)
}

func (l *customlist) remove_hl_flag(ss string, key string) ([]colorkey, string) {
	keys := []colorkey{}
	s := ss
	for len(s) > 0 {
		b := strings.Index(s, key)
		if b >= 0 {
			e := strings.Index(s[b+1:], key)
			if e > 0 {
				key := l.NewDefaultColorKey(s[b+2 : b+e+1])
				key.color = global_theme.search_highlight_color()
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
		ss = strings.ReplaceAll(ss, fmt.Sprintf("%s%s%s", key, v.str, key), v.str)
	}
	return keys, ss
}
func find_key(s string, keys []colorkey, offset int) []keypattern {
	idx := len(s)
	var k = colorkey{str: ""}
	for _, v := range keys {
		if len(v.str) == 0 {
			continue
		}
		n := strings.Index(strings.ToLower(s), v.str)
		if n >= 0 && n < idx {
			k = v
			idx = n
		}
	}
	v := k
	if len(k.str) > 0 {
		pth := keypattern{begin: idx + offset, width: len(v.str), color: v.color}
		a := []keypattern{pth}
		subret := find_key(s[idx+len(v.str):], keys, pth.width+idx+offset)
		return append(a, subret...)
	}
	return []keypattern{}
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
	offset_x, y, _, height := l.GetInnerRect()

	bottomLimit := y + height
	select_color := global_theme.search_highlight_color()
	selected_style := tcell.StyleDefault.Foreground(tview.Styles.PrimitiveBackgroundColor).Background(tview.Styles.PrimaryTextColor)
	selected_stylehl := tcell.StyleDefault.Foreground(select_color).Background(tview.Styles.PrimaryTextColor)

	style := tcell.StyleDefault.Foreground(tview.Styles.PrimaryTextColor).Background(tview.Styles.PrimitiveBackgroundColor)
	stylehl := tcell.StyleDefault.Foreground(select_color).Background(tview.Styles.PrimitiveBackgroundColor)
	theme_style := stylehl
	if s := global_theme.get_color("selection"); s != nil {
		theme_style = *s
	}

	itemoffset, _ := l.GetOffset()
	keys := []colorkey{}
	for _, v := range l.get_hl_keys() {
		keys = append(keys, l.NewDefaultColorKey(v))
	}
	keys2 := []colorkey{}
	for _, v := range l.Key {
		keys2 = append(keys2, l.NewDefaultColorKey(string(v)))
	}
	for index := itemoffset; index < len(l.hlitems); index++ {
		MainText, SecondText := l.List.GetItemText(index)
		MainText, main_postion := get_hl_postion(MainText, keys, l, keys2)
		selected := index == l.List.GetCurrentItem()
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
				l.draw_item_color([]keypattern{}, MainText, screen, offset_x, y, theme_style, theme_style)
			} else if selected {
				l.draw_item_color(main_postion, MainText, screen, offset_x, y, selected_style, selected_stylehl)
			} else {
				l.draw_item_color(main_postion, MainText, screen, offset_x, y, style, stylehl)
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
			l.draw_item_color(main_postion, SecondText, screen, offset_x, y, style, stylehl)
			// }
			y += 1
			if y >= bottomLimit {
				break
			}
		}
	}

}

func get_hl_postion(MainText string, keys []colorkey, l *customlist, keys2 []colorkey) (string, []keypattern) {
	k, MainText := l.find_hl_key(MainText)
	for i := range k {
		k[i].str = strings.ToLower(k[i].str)
	}
	hlkey := k
	for _, v := range k {
		k1 := v.str
		for _, v2 := range keys {
			k2 := strings.ToLower(v2.str)
			if k1 == k2 {
				continue
			} else {
				hlkey = append(hlkey, v2)
			}
		}
	}
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

func (l *customlist) draw_item_color(Positions []keypattern, MainText string, screen tcell.Screen, offset_x int, y int, normal_style tcell.Style, hlStyle tcell.Style) {
	begin := 0
	for _, e := range Positions {
		normal := MainText[begin:e.begin]
		for i, r := range normal {
			x := offset_x + i + begin
			if !l.InInnerRect(x, y) {
				return
			}
			screen.SetContent(x, y, r, nil, normal_style)
		}
		hltext := MainText[e.begin : e.begin+e.width]
		for i, r := range hltext {
			s := hlStyle
			x := offset_x + i + e.begin
			if !l.InInnerRect(x, y) {
				return
			}
			screen.SetContent(x, y, r, nil, s.Foreground(e.color))
		}
		begin = e.begin + e.width
	}
	if begin < len(MainText) {
		normal := MainText[begin:]
		for i, r := range normal {
			x := offset_x + i + begin
			if !l.InInnerRect(x, y) {
				return
			}
			screen.SetContent(x, y, r, nil, normal_style)
		}
	}
}
