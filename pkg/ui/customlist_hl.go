package mainui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type colortext struct {
	text  string
	color tcell.Color
}

func fmt_bold_string(s string) string {
	return fmt.Sprintf("**%s**", s)
}
func fmt_color_string(s string, color tcell.Color) string {
	return fmt.Sprintf("**[%d]%s**", color, s)
}

type splitresult struct {
	b, m, a colortext
}

func parse_key_string(s string, ptn colortext) splitresult {
	key := ptn.text
	b := strings.Index(s, key)
	if b >= 0 {
		return splitresult{
			b: colortext{text: s[:b]},
			m: colortext{text: s[b : b+len(key)], color: ptn.color},
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

func GetColorText(t string, keys []colortext) []colortext {
	mark_keys := colorpaser{data: t}.Parse()
	keys_result := []colortext{}
	for _, v := range mark_keys {
		if v.color != 0 {
			keys_result = append(keys_result, v)
		} else {
			a := colorpaser{data: v.text}.ParseKey(keys)
			keys_result = append(keys_result, a...)
		}
	}
	return keys_result
}
func (p colorpaser) ParseKey(keys []colortext) []colortext {
	for _, key := range keys {
		if len(key.text) == 0 {
			continue
		}
		r3 := parse_key_string(p.data, key)
		if len(r3.m.text) == 0 {
			splitkeys := strings.Split(key.text, " ")
			if len(splitkeys) > 1 {
				split_keyword := []colortext{}
				for _, v := range splitkeys {
					split_keyword = append(split_keyword, colortext{text: v, color: key.color})
				}
				result := colorpaser{data: p.data}.ParseKey(split_keyword)
				if len(result) > 0 {
					return result
				}
			}
		}
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
func substring(s string, b, e int) (string, error) {
	if e != -1 && b > e {
		return "", errors.New("b>e")
	}
	if b < len(s) {
		if e == -1 {
			return s[b:], nil
		}
		if e < len(s) {
			return s[b:e], nil
		}
	}
	return s, errors.New("invalid range")
}
func pasrse_color_string(s string) splitresult {
	b := strings.Index(s, "**[")
	if b >= 0 {
		e := strings.Index(s, "]")
		if e >= 0 {
			var color tcell.Color
			if sub, err := substring(s, b+3, e); err == nil {
				if c, err := strconv.Atoi(sub); err == nil {
					color = tcell.Color(c)
					if sub, err := substring(s, e+1, -1); err == nil {
						if e2 := strings.Index(sub, "**"); e2 > 0 {
							x := e + 1
							if sub, err := substring(s, x+e2+2, -1); err == nil {
								return splitresult{b: colortext{text: s[:b]}, m: colortext{text: s[x : x+e2], color: color}, a: colortext{text: sub}}
							}
						}
					}
				}
			}
		}
	}
	return splitresult{b: colortext{text: s}}
}
