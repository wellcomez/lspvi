package mainui

import (
	"strings"

	"github.com/pgavlin/femto"
	"github.com/tectiv3/go-lsp"
)


func (format *TokenLineFormat) FormatOnTokenline() {
	for linenr := range format.lines {
		tl := format.lines[linenr]
		tl.Replace()
		tl.format_on_breakline(format)
	}
}

func (currentLine *TokenLine) format_on_breakline(format *TokenLineFormat) {
	if edit := currentLine.line_edit; edit != nil {
		start, end := GetEditLoc(*edit)
		newLines := strings.Split(edit.NewText, "\n")
		if len(newLines) == 1 {
			lastline := format.lines[end.Y]
			lastline.format_on_breakline(format)
			lastline.SubFrom(end.X)
			tokens := lastline.SubTo(end.X)
			currentLine.replaced = append(currentLine.replaced, tokens...)
			tokens = []Token{{
				edit.NewText,
				nil,
				-1, -1,
			}}
			tokens = append(tokens, lastline.Tokens...)
			lastline.appends = tokens
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

func create_line(Buf *femto.Buffer, edits []lsp.TextEdit) (tokenline TokenLine, next_index int) {
	next_index = -1
	if len(edits) > 0 {
		lineNr := edits[0].Range.Start.Line
		tokenline.lineno = lineNr
		line := Buf.Line(lineNr)

		begin_x := 0
		for next_index = 1; next_index < len(edits); next_index++ {
			_, end := GetEditLoc(edits[next_index])
			if end.Y != lineNr {
				break
			}
		}
		var s = []Token{}
		for i := 0; i < next_index; i++ {
			start, end := GetEditLoc(edits[i])
			t := Token{line[begin_x:start.X], nil, begin_x, start.X}
			s = append(s, t)
			begin_x = start.X
			t = Token{line[begin_x:end.X], &edits[i], begin_x, end.X}
			s = append(s, t)
		}
		if next_index < len(edits) {
			break_line_edit := edits[next_index]
			start, stop := GetEditLoc(break_line_edit)
			if start.Y != stop.Y {
				tokenline.line_edit = &break_line_edit
				tokenline.replaced = []Token{{line[start.X:], &break_line_edit, start.X, len(line)}}
			}
		}
		tokenline.Tokens = s
		tokenline.line = line
	}
	return
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
	Tokens    []Token
	line      string
	line_edit *lsp.TextEdit
	lineno    int
	replaced  []Token
	appends   []Token
	newline   []Token
}

func (line *TokenLine) SubTo(index int) (t []Token) {
	for _, v := range line.Tokens {
		if v.e < index {
			t = append(t, v)
		}
	}
	return
}
func (line *TokenLine) SubFrom(index int) {
	t := []Token{}
	for _, v := range line.Tokens {
		if v.b >= index {
			if v.e >= index {
				t = append(t, Token{line.line[index:v.e], v.edit, index, v.e})
			} else {
				t = append(t, v)
			}
		}
	}
	line.Tokens = t
}
func (t *TokenLine) Replace() (ret string) {
	for _, v := range t.Tokens {
		ret = ret + v.replace()
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
		if line, next_index := create_line(Buf, edits); next_index > 0 {
			f.lines[line.lineno] = &line
			edits = edits[next_index:]
			if next_edit := line.line_edit; next_edit != nil {
				start, stop := GetEditLoc(*next_edit)
				for i := start.Y + 1; i < stop.Y+1; i++ {
					x := Buf.Line(i)
					token := Token{x, nil, 0, len(x)}
					f.lines[i] = &TokenLine{[]Token{token}, x, nil, i, nil, nil, nil}
				}
			}
		} else {
			return
		}
	}
}
