package mainui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/pgavlin/femto"
	"github.com/tectiv3/go-lsp"
	"zen108.com/lspvi/pkg/debug"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

func (format *TokenLineFormat) Run(replace bool) (code_change lspcore.CodeChangeEvent, err error) {
	var lines []*TokenLine
	for linenr := range format.lines {
		tl := format.lines[linenr]
		lines = append(lines, tl)
	}
	sort.SliceStable(lines, func(i, j int) bool {
		return lines[i].lineno < lines[j].lineno
	})
	for _, tl := range lines {
		tl.Run(format)
	}
	debug.DebugLog("format", "print", strings.Repeat("=", 40), len(lines))
	for _, tl := range lines {
		tl.print()
	}
	debug.DebugLog("format", strings.Repeat("=", 40), len(lines))
	for _, tl := range lines {
		i := tl.lineno
		if tl.removed {
			debug.DebugLog("format", "x:", i)
		} else {
			debug.DebugLog("format", "*:", i, tl.FormatOutput(true))
		}
	}
	deleted := 0
	if !replace {
		err = fmt.Errorf("not replacing")
		return
	}
	var Events []lspcore.TextChangeEvent
	for i := 0; i < len(lines); i++ {
		lineno := len(lines) - i - 1
		line := lines[lineno]
		if !line.removed {
			var s, e femto.Loc
			newline := line.FormatOutput(true)
			if deleted != 0 {
				s, _ = line.Range()
				_, e = lines[lineno+deleted].Range()
			} else {
				s, e = line.Range()
			}
			s, e = ord_loc(s, e)
			format.Buf.Replace(s, e, newline)
			a := lspcore.TextChangeEvent{
				Text: newline,
				Type: lspcore.TextChangeTypeReplace,
				Range: lsp.Range{
					Start: lsp.Position{
						Line:      s.Y,
						Character: s.X,
					},
					End: lsp.Position{
						Line:      e.Y,
						Character: e.X,
					},
				},
			}
			Events = append(Events, a)
			deleted = 0
		} else {
			deleted++
		}
	}
	code_change.Events = Events
	return
}

var codeprint = func(s string) (ret string) {
	ret = fmt.Sprintf("[%s]", s)
	return
}

func (currentLine TokenLine) print() {
	e := ""
	if currentLine.line_edit != nil {
		e = string_editor(*currentLine.line_edit)
	}
	debug.DebugLogf("format", "%d inline-editor:%d  line-editor:%s", currentLine.lineno, currentLine.editorcount, e)
	x1 := codeprint(currentLine.line)
	debug.DebugLog("format", "old--", len(x1), x1)
	x := codeprint(currentLine.FormatOutput(false))
	debug.DebugLog("format", "new--", len(x), x)
}
func (cur *TokenLine) Run(format *TokenLineFormat) {
	if cur.formated {
		return
	}
	debug.DebugLog("format", cur.lineno, "**********Run-start@", cur.editorcount, "EditText", cur.line_edit)
	defer debug.DebugLog("format", cur.lineno, "**********Run-end@")
	cur.formated = true
	if edit := cur.line_edit; edit != nil {
		debug.DebugLog("format", cur.lineno, "edit start@", cur.editorcount, string_editor(*edit))
		defer debug.DebugLog("format", cur.lineno, "edit end@")
		_, end := GetEditLoc(*edit)
		// newLines := strings.Split(edit.NewText, "\n")
		enter_count := strings.Count(edit.NewText, "\n")
		switch enter_count {

		case 0:
			{
				lastline := format.lines[end.Y]
				lastline.Run(format)
				replace, left := lastline.split(end)
				cur.replaced = append(cur.replaced, replace...)
				lastline.removed = true
				to_left := append([]Token{{
					edit.NewText,
					nil,
					-1, -1,
				}}, left...)
				cur.appends = append(cur.appends, to_left...)
				cur.print()
			}
		default:
			{
				ss := strings.Split(edit.NewText, "\n")
				s1 := ss[0]
				s2 := ss[len(ss)-1]
				newline := ss[1 : len(ss)-1]
				if len(newline) > 0 {
					for i := 0; i < len(newline); i++ {
						debug.DebugLog("format", cur.lineno, i+cur.lineno+1, "editmulti-insert", strconv.Quote(newline[i]))
						// v := newline[i]
						// n := []Token{{data: v}}
						// l := TokenLine{line: v, lineno: i + cur.lineno,
						// 	newline: n}
						// cur.insertline = append(cur.insertline, l)
					}
				}
				debug.DebugLog("format", cur.lineno, "s1-", strconv.Quote(s1), "s2-", strconv.Quote(s2))
				cur.appends = append(cur.appends, Token{data: s1})
				cur.print()

				nextline := format.lines[end.Y]
				nextline.Run(format)
				_, right := Split(end.X, nextline.Tokens)
				r := []Token{{data: s2, b: 0, e: end.X}}
				r = append(r, right...)
				r = append(r, nextline.appends...)
				nextline.newline = r
				nextline.print()
			}
		}
	}
}
func (line TokenLine) Range() (start, end femto.Loc) {
	start = femto.Loc{Y: line.lineno, X: 0}
	end = femto.Loc{Y: line.lineno, X: len(line.line)}
	return
}
func (lastline *TokenLine) split(end femto.Loc) (replace []Token, add []Token) {
	replace, add = Split(end.X, lastline.Tokens)
	add = append(add, lastline.appends...)
	return
}

func create_line(Buf *femto.Buffer, edits []lsp.TextEdit) (tokenline *TokenLine, next_index int, err error) {
	next_index = -1
	if len(edits) > 0 {
		lineNr := edits[0].Range.Start.Line
		tokenline = create_from_bufline(Buf, lineNr)
		line := tokenline.line

		begin_x := 0
		for next_index = 0; next_index < len(edits); next_index++ {
			start, end := GetEditLoc(edits[next_index])
			if end.Y != lineNr {
				break
			} else if start.X > len(line) || end.X > len(line) {
				err = fmt.Errorf("%v %v %s", start, end, line)
				tokenline = nil
				return
			}
		}
		var line_tokens = []Token{}
		for i := 0; i < next_index; i++ {
			tokenline.editorcount++
			start, end := GetEditLoc(edits[i])
			t := Token{line[begin_x:start.X], nil, begin_x, start.X}
			line_tokens = append(line_tokens, t)
			begin_x = start.X
			t = Token{line[begin_x:end.X], &edits[i], begin_x, end.X}
			line_tokens = append(line_tokens, t)
			debug.DebugLog("format", "createline", string_editor(edits[i]), "find edit")
			begin_x = end.X
		}
		if begin_x < len(line) {
			lasttoken := Token{line[begin_x:], nil, begin_x, len(line)}
			line_tokens = append(line_tokens, lasttoken)
		}
		if next_index == 0 {
			e := edits[0]
			debug.DebugLog("format", "createline", string_editor(e), "just multi-line")
		}

		tokenline.Tokens = line_tokens

		if next_index < len(edits) {
			break_line_edit := edits[next_index]
			start, stop := GetEditLoc(break_line_edit)
			if start.Y != stop.Y && lineNr == start.Y {
				if start.X < len(tokenline.line) {
					debug.DebugLog("format", "createline-line-edit", start.X, ":", len(tokenline.line))
				}
				tokenline.line_edit = &break_line_edit
				left, right := Split(start.X, tokenline.Tokens)
				tokenline.Tokens = left
				for i := range right {
					right[i].edit = &break_line_edit
				}
				tokenline.replaced = right
			}
			if lineNr == start.Y {
				next_index++
			}
		}
	}
	return
}
func string_editor(e lsp.TextEdit) string {
	start, end := GetEditLoc(e)
	return fmt.Sprintf("%d:%d ->  %d:%d %s %v", start.Y, start.X, end.Y, end.X, string_editor_text(e), []rune(e.NewText))
}
func string_editor_text(e lsp.TextEdit) string {
	x := codeprint(strings.ReplaceAll(e.NewText, "\n", "\\n"))
	return x
}

type Token struct {
	data string
	edit *lsp.TextEdit
	b, e int
}

func NewToken(data string) *Token {
	return &Token{
		data: data,
	}
}

func (t *Token) replace() string {
	if t.edit == nil {
		return t.data
	}
	return t.edit.NewText
}

type TokenLine struct {
	Tokens      []Token
	editorcount int
	line        string
	line_edit   *lsp.TextEdit
	lineno      int
	replaced    []Token
	appends     []Token
	newline     []Token
	removed     bool
	formated    bool
	insertline  []TokenLine
}

func Split(index int, tokes []Token) (left, rigtht []Token) {
	for i, v := range tokes {
		if index < v.e {
			left = append(left, tokes[:i]...)
			if index < v.e {
				left = append(left, Token{v.data[:index-v.b], v.edit, v.b, index})
				rigtht = append(rigtht, Token{v.data[index-v.b:], v.edit, index, v.e})
			}
			rigtht = append(rigtht, tokes[i+1:]...)
			return
		}
	}
	left = append(left, tokes...)
	return
}
func (line *TokenLine) SubTo(index int) (t []Token) {
	for _, v := range line.Tokens {
		if index <= v.e {
			if index > v.b {
				vv := v
				vv.data = line.line[v.b:index]
				vv.b = v.b
				vv.e = index
				t = append(t, vv)
			} else {
				t = append(t, v)
			}
		}
	}
	return
}
func (line *TokenLine) SubFrom(index int) (t []Token) {
	for _, v := range line.Tokens {
		if index <= v.b {
			t = append(t, v)
		} else {
			if index < v.e {
				t = append(t, Token{line.line[index:v.e], v.edit, index, v.e})
			}
		}
	}
	t = append(t, line.appends...)
	return
}
func (t *TokenLine) FormatOutput(no_print bool) (ret string) {
	if t.removed {
		if !no_print {
			ret = "deleted?????????????????????????????????"
		}
		return
	}
	if len(t.newline) > 0 {
		for _, v := range t.newline {
			ret = ret + v.replace()
		}
		return
	}
	for _, v := range t.Tokens {
		ret = ret + v.replace()
	}
	for _, v := range t.appends {
		ret = ret + v.replace()
	}
	return
}

type TokenLineFormat struct {
	lines map[int]*TokenLine
	Buf   *femto.Buffer
}

func NewTokenLineFormat(Buf *femto.Buffer, edits []lsp.TextEdit) (f *TokenLineFormat, err error) {
	if len(edits) == 0 {
		err = fmt.Errorf("no edits")
		return
	}
	f = &TokenLineFormat{Buf: Buf}
	f.lines = make(map[int]*TokenLine)
	for {
		if line, next_index, e := create_line(Buf, edits); line != nil {
			f.lines[line.lineno] = line
			edits = edits[next_index:]
			if next_edit := line.line_edit; next_edit != nil {
				start, stop := GetEditLoc(*next_edit)
				for i := start.Y + 1; i < stop.Y+1; i++ {
					x1 := create_from_bufline(Buf, i)
					f.lines[i] = x1
				}
			}
		} else if e != nil {
			err = e
			return
		} else {
			return
		}
	}
}

func create_from_bufline(Buf *femto.Buffer, i int) *TokenLine {
	x := Buf.Line(i)
	token := Token{x, nil, 0, len(x)}
	x1 := TokenLine{[]Token{token}, 0, x, nil, i, nil, nil, nil, false, false, nil}
	return &x1
}
