// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	// "zen108.com/lspvi/pkg/debug"
)

type colortext struct {
	text  string
	color tcell.Color
	bg    tcell.Color
}

func fmt_bold_string(s string) string {
	return fmt.Sprintf("**%s**", s)
}

type colorstring struct {
	line []colortext
	text string
}

func (line *colorstring) len() (ret int) {
	for _, v := range line.line {
		ret = ret + len([]rune(v.text))
	}
	return
}
func (line *colorstring) setfg(bg tcell.Color) {
	for i := range line.line {
		line.line[i].color = bg
	}
}
func (line *colorstring) setbg(bg tcell.Color) {
	for i := range line.line {
		line.line[i].bg = bg
	}
}

// func (line *colorstring) Sprintf(format string, v ...any) {
// 	var param []any = []any{}
// 	param = append(param, v...)
// 	p := format
// 	idx := 0
// 	for {
// 		v_index := strings.Index(p, "%v")
// 		if v_index > 0 {
// 			line.add_string_color(fmt.Sprintf(p[:v_index], param[v_index]), 0)
// 			idx++
// 			x := param[v_index]
// 			switch v := x.(type) {
// 			case colortext:
// 				line.add_color_text(v)
// 			case colorstring:
// 				line.add_color_text_list(v.line)
// 			default:
// 				line.add_string_color(fmt.Sprintf("%v", x), 0)
// 				debug.ErrorLog("fmt_color_string %v", v)
// 			}
// 			idx++
// 			p = p[v_index+2:]
// 		} else {
// 			line.add_string_color(fmt.Sprintf(p, param[idx:]...), 0)
// 			break
// 		}
// 	}

// }
func (line *colorstring) ColorText() (ret string) {
	for _, v := range line.line {
		ret = ret + fmt_color_string(v.text, v.color)
	}
	return
}
func (line *colorstring) plaintext() (ret string) {
	if len(line.text) == 0 {
		for _, v := range line.line {
			ret = ret + v.text
		}
		line.text = ret
	}
	return line.text
}
func (line *colorstring) add_color_text(v colortext) *colorstring {
	return line.add_string_color_bg(v.text, v.color, v.bg)
}
func (line *colorstring) add_color_text_list(s []colortext) *colorstring {
	for _, v := range s {
		line.add_string_color_bg(v.text, v.color, v.bg)
	}
	return line
}
func (line *colorstring) Replace(old string, new string, n int) {
	for i := range line.line {
		v := &line.line[i]
		if count := strings.Count(v.text, old); count > 0 && n > 0 {
			replace_count := min(count, n)
			v.text = strings.Replace(v.text, old, new, replace_count)
			n = n - replace_count
		}
	}
}
func (line *colorstring) prepend(s string, color tcell.Color) *colorstring {
	if len(s) > 0 {
		line.line = append([]colortext{{s, color, 0}}, line.line...)
	}
	return line
}
func (line *colorstring) a(s string) *colorstring {
	return line.add_string_color_bg(s, 0, 0)
}
func (line *colorstring) add_string_color_bg(s string, color, bg tcell.Color) *colorstring {
	if len(s) > 0 {
		line.line = append(line.line, colortext{s, color, bg})
	}
	return line
}
func (line *colorstring) add_string_style(s string, color tcell.Style) *colorstring {
	if len(s) > 0 {
		f,g,_:=color.Decompose()
		line.line = append(line.line, colortext{s, f, g})
	}
	return line
}
func (line *colorstring) add_string_color(s string, color tcell.Color) *colorstring {
	if len(s) > 0 {
		line.line = append(line.line, colortext{s, color, 0})
	}
	return line
}
func fmt_color_string(s string, color tcell.Color) string {
	if color == 0 {
		return s
	}
	return fmt.Sprintf("**[%d]%s**", color, s)
}

type splitresult struct {
	b, m, a colortext
}

func parse_key_string(s string, ptn colortext) splitresult {
	key := ptn.text
	b := strings.Index(s, key)
	if b >= 0 {
		x := b + len(key)
		c := ""
		if x < len(s) {
			c = s[x:]
		}
		return splitresult{
			b: colortext{text: s[:b]},
			m: colortext{text: s[b : b+len(key)], color: ptn.color},
			a: colortext{text: c},
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
			c := ""
			b1 := b + 2
			b2 := b1 + e
			c1 := b2 + 2
			if c1 < len(s) {
				c = s[c1:]
			}
			return splitresult{
				b: colortext{text: s[:b]},
				m: colortext{text: s[b1:b2], color: tcell.ColorYellow},
				a: colortext{text: c},
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
		if e <= len(s) {
			return s[b:e], nil
		}
	}
	return s, errors.New("invalid range")
}
func pasrse_color_string(s string) splitresult {
	b := strings.Index(s, "**[")
	if b >= 0 {
		e := strings.Index(s[b:], "]")
		if e >= 0 {
			e += b
			var color tcell.Color
			if sub, err := substring(s, b+3, e); err == nil {
				if c, err := strconv.Atoi(sub); err == nil {
					color = tcell.Color(c)
					if sub, err := substring(s, e+1, -1); err == nil {
						if e2 := strings.Index(sub, "**"); e2 > 0 {
							x := e + 1
							if sub, err := substring(s, x+e2+2, -1); err == nil {
								return splitresult{b: colortext{text: s[:b]}, m: colortext{text: s[x : x+e2], color: color}, a: colortext{text: sub}}
							} else {
								return splitresult{b: colortext{text: s[:b]}, m: colortext{text: s[x : x+e2], color: color}, a: colortext{text: ""}}
							}
						}
					}
				}
			}
		}
	}
	return splitresult{b: colortext{text: s}}
}
