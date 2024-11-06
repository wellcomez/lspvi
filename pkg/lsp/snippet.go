package lspcore

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"zen108.com/lspvi/pkg/debug"
)

type snippet_arg struct {
	index   int
	name    string
	capture string
	pos     int
}
type complete_token struct {
	arg  snippet_arg
	Text string
}

func (t complete_token) is_dollar_0() bool {
	return strings.Contains(t.Text, "$0")
}
func (t complete_token) is_arg() bool {
	return len(t.arg.capture) > 0
}

type complete_code struct {
	snip      snippet
	snip_args []snippet_arg
	tokens    []complete_token
}
type snippet struct {
	raw string
}

func (r snippet) args() (args []snippet_arg) {
	newtext := r.raw
	re2 := regexp.MustCompile(`\$\{(\d+):?([^}]*)\}`)
	matches := re2.FindAllStringSubmatch(newtext, -1)
	for _, match := range matches {
		if len(match) == 3 {
			debug.DebugLog("complete", "match", "no", match[1], "default-arg", strconv.Quote(match[2]))
			if x, err := strconv.Atoi(match[1]); err == nil {
				capture := match[0]
				a := snippet_arg{
					index:   x,
					name:    match[2],
					capture: capture,
					pos:     strings.Index(r.raw, match[0]),
				}
				args = append(args, a)
			}
		}
	}
	return
}
func (code complete_code) SnipCount() (a int) {
	return len(code.snip_args)
}
func (code complete_code) Len() (a int) {
	a = len(code.tokens)
	return
}
func (code complete_code) Token(a int) (ret complete_token, err error) {
	if a < len(code.tokens) {
		ret = code.tokens[a]
	} else {
		err = fmt.Errorf("index out of range")
	}
	return
}
func (code complete_code) Text() string {
	ret := []string{}
	for _, v := range code.tokens {
		ret = append(ret, v.Text)
	}
	return strings.Join(ret, "")
}
func NewCompleteCode(raw string) (ret *complete_code) {
	ret = &complete_code{snip: snippet{raw: raw}}
	ret.snip_args = ret.snip.args()
	tokens := []complete_token{}
	s := raw
	if len(ret.snip_args) != 0 {
		for i, v := range ret.snip_args {
			ss := strings.Split(s, v.capture)
			if len(ss) > 0 {
				x1 := ss[0]
				x := ret.string_to_token(x1)
				tokens = append(tokens, x...)
				tokens = append(tokens, complete_token{Text: v.name, arg: v})
				if len(ss) > 1 {
					if len(ret.snip_args) == i+1 {
						tokens = append(tokens, ret.string_to_token(ss[1])...)
					} else {
						s = ss[1]
					}
				} else {
					break
				}
			} else {
				tokens = append(tokens, ret.string_to_token(s)...)
			}
		}
	} else {
		tokens = append(tokens, ret.string_to_token(s)...)
	}
	ret.tokens = tokens
	return
}

var snip_zero = "$0\\"

func (code *complete_code) string_to_token(sss string) (ret []complete_token) {
	x1 := sss
	for {
		if ind := strings.Index(x1, snip_zero); ind > 0 {
			arg := snippet_arg{name: "",
				capture: snip_zero,
				pos:     strings.Index(code.snip.raw, snip_zero)}
			code.snip_args = append(code.snip_args, arg)
			ret = append(ret, complete_token{Text: x1[:ind]})
			ret = append(ret, complete_token{Text: "",
				arg: arg})
			x1 = x1[ind+len(snip_zero):]
		} else {
			ret = append(ret, complete_token{Text: x1})
			return
		}
		if len(x1) == 0 {
			return
		}
	}
}