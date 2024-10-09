package mainui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

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

func parse_key_string(s string, key string) splitresult {
	b := strings.Index(s, key)
	if b >= 0 {
		return splitresult{
			b: colortext{text: s[:b]},
			m: colortext{text: s[b : b+len(key)], color: tcell.ColorYellow},
			a: colortext{text: s[b+len(key):]},
		}
	}
	return splitresult{b: colortext{text: s}}
}
func color_maintext(sss []colortext) string {
	ss := ""
	for _, s := range sss {
		ss += s.text
	}
	return ss
}
func pasrse_bold_color_string(s string) splitresult {
	b := strings.Index(s, "**")
	if b >= 0 {
		e := strings.Index(s[b+2:], "**")
		if e >= 0 {
			return splitresult{
				b: colortext{text: s[:b]},
				m: colortext{text: s[b+2 : b+2+e], color: tcell.ColorYellow},
				a: colortext{text: s[b+2+e+2:]},
			}
		}
	}
	return splitresult{b: colortext{text: s}}
}

type colorpaser struct {
	data string
}

func (p colorpaser) ParseKey(keys []string) []colortext {
	for _, v := range keys {
		if len(v) == 0 {
			continue
		}
		r3 := parse_key_string(p.data, v)
		if len(r3.m.text) > 0 {
			var before_part []colortext
			if r3.b.text != "" {
				aa := colorpaser{data: r3.b.text}
				result := aa.ParseKey(keys)
				for _, v := range result {
					if len(v.text) > 0 {
						before_part = append(before_part, v)
					}
				}
			}
			var after_part []colortext
			if r3.a.text != "" {
				aa := colorpaser{data: r3.a.text}
				result := aa.ParseKey(keys)
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
	}
	b := colortext{text: p.data}
	return []colortext{b}
}
func (p colorpaser) Parse() []colortext {
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
	return []colortext{r3.b}
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