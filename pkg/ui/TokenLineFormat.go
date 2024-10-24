package mainui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pgavlin/femto"
	"github.com/tectiv3/go-lsp"
	"zen108.com/lspvi/pkg/debug"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

func (format *TokenLineFormat) Run() (code_change lspcore.CodeChangeEvent) {

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
	debug.DebugLog("format", strings.Repeat("=", 20), len(lines))
	for _, tl := range lines {
		tl.print()
	}
	debug.DebugLog("format", strings.Repeat("=", 20), len(lines))
	for _, tl := range lines {
		i := tl.lineno
		if tl.removed {
			debug.DebugLog("format", "x:", i)
		} else {
			debug.DebugLog("format", "*:", i, tl.FormatOutput(true))
		}
	}
	deleted := 0

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
	debug.DebugLogf("format", "line:%d inline-editor:%d  line-editor:%s", currentLine.lineno, currentLine.editorcount, e)
	debug.DebugLog("format", "old--", codeprint(currentLine.line))
	debug.DebugLog("format", "new--", codeprint(currentLine.FormatOutput(false)))
}
func (currentLine *TokenLine) Run(format *TokenLineFormat) {
	if currentLine.formated {
		return
	}
	currentLine.formated = true
	if edit := currentLine.line_edit; edit != nil {
		start, end := GetEditLoc(*edit)
		newLines := strings.Split(edit.NewText, "\n")
		if len(newLines) == 1 {
			lastline := format.lines[end.Y]
			lastline.Run(format)
			replace, left := lastline.split(end)
			currentLine.replaced = append(currentLine.replaced, replace...)
			lastline.removed = true
			to_left := append([]Token{{
				edit.NewText,
				nil,
				-1, -1,
			}}, left...)
			currentLine.appends = append(currentLine.appends, to_left...)
			currentLine.print()
		} else {
			for i := range newLines {
				v := newLines[i]
				if i == 0 {
					currentLine.appends = append(currentLine.appends, Token{data: v})
				} else {
					lineNr := i + start.Y
					line := format.lines[lineNr]
					line.newline = []Token{{data: v}}
				}
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
	add = lastline.SubFrom(end.X)
	replace = lastline.SubTo(end.X)
	return
}

func create_line(Buf *femto.Buffer, edits []lsp.TextEdit) (tokenline *TokenLine, next_index int) {
	next_index = -1
	if len(edits) > 0 {
		lineNr := edits[0].Range.Start.Line
		tokenline = create_from_bufline(Buf, lineNr)
		line := tokenline.line

		begin_x := 0
		for next_index = 0; next_index < len(edits); next_index++ {
			_, end := GetEditLoc(edits[next_index])
			if end.Y != lineNr {
				break
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

		if next_index < len(edits) {
			break_line_edit := edits[next_index]
			start, stop := GetEditLoc(break_line_edit)
			if start.Y != stop.Y {
				tokenline.line_edit = &break_line_edit
				tokenline.replaced = []Token{{line[start.X:], &break_line_edit, start.X, len(line)}}
			}
			if lineNr == start.Y {
				next_index++
			}
		}
		tokenline.Tokens = line_tokens

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
			ret = ret + v.data
		}
		return
	}
	for _, v := range t.Tokens {
		ret = ret + v.replace()
	}
	for _, v := range t.appends {
		ret = ret + v.data
	}
	return
}

type TokenLineFormat struct {
	lines map[int]*TokenLine
	Buf   *femto.Buffer
}

func NewTokenLineFormat(Buf *femto.Buffer, edits []lsp.TextEdit) (f *TokenLineFormat) {
	f = &TokenLineFormat{Buf: Buf}
	f.lines = make(map[int]*TokenLine)
	for {
		if line, next_index := create_line(Buf, edits); line != nil {
			f.lines[line.lineno] = line
			edits = edits[next_index:]
			if next_edit := line.line_edit; next_edit != nil {
				start, stop := GetEditLoc(*next_edit)
				for i := start.Y + 1; i < stop.Y+1; i++ {
					x1 := create_from_bufline(Buf, i)
					f.lines[i] = x1
				}
			}
		} else {
			return
		}
	}
}

func create_from_bufline(Buf *femto.Buffer, i int) *TokenLine {
	x := Buf.Line(i)
	token := Token{x, nil, 0, len(x)}
	x1 := TokenLine{[]Token{token}, 0, x, nil, i, nil, nil, nil, false, false}
	return &x1
}
