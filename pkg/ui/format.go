package mainui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pgavlin/femto"
	"github.com/tectiv3/go-lsp"
	"zen108.com/lspvi/pkg/debug"
)

func format2(ret []lsp.TextEdit, code *CodeView) {
	for i := range ret {
		y := len(ret) - i - 1
		v := ret[y]
		if v.NewText == "\n" {

		}
		start, end := GetEditLoc(v)
		startline := code.view.Buf.Line(start.Y)
		endline := code.view.Buf.Line(end.Y)
		debug.DebugLog("format", y, ">", []rune(v.NewText), "<", start.Y, ":", start.X, "->", end.Y, ":", end.X)
		debug.DebugLog("format", y, "before-start", startline, len(startline))
		yes := end.Y > start.Y
		if yes {
			continue
		}
		code.view.Buf.Replace(start, end, v.NewText)
		x := code.view.Buf.Line(start.Y)
		debug.DebugLog("format", y, "after       ", x, len(x))
		if yes {
			debug.DebugLog("format", y, "end       ", endline, len(endline))
			x := code.view.Buf.Line(end.Y)
			debug.DebugLog("format", y, "end-after ", x, len(x))
		}
	}
}

func GetEditLoc(v lsp.TextEdit) (femto.Loc, femto.Loc) {
	start := femto.Loc{
		X: v.Range.Start.Character,
		Y: v.Range.Start.Line,
	}
	end := femto.Loc{
		X: v.Range.End.Character,
		Y: v.Range.End.Line,
	}
	return start, end
}

func format3(edits []lsp.TextEdit, code *CodeView) error {
	f := Format{Lines: code.view.Buf}
	return f.run(edits)
}

type Format struct {
	Lines *femto.Buffer
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

func (d *Format) run(edits []lsp.TextEdit) error {
	sort.SliceStable(edits, func(i, j int) bool {
		// Compare lines first
		if edits[i].Range.Start.Line != edits[j].Range.Start.Line {
			return edits[i].Range.Start.Line > edits[j].Range.Start.Line
		}
		// If same line, compare characters
		return edits[i].Range.Start.Character > edits[j].Range.Start.Character
	})
	for _, edit := range edits {
		if err := d.applyEdit(edit); err != nil {
			return err
		}
	}
	return nil
}

// applyEdit applies a single TextEdit
func (d *Format) applyEdit(edit lsp.TextEdit) error {
	start := edit.Range.Start
	end := edit.Range.End

	// Handle single line edit
	if start.Line == end.Line {
		return d.applySingleLineEdit(edit)
	}

	// Handle multi-line edit
	return d.applyMultiLineEdit(edit)
	// return nil
}

// applySingleLineEdit handles edits within a single line
func (d *Format) applySingleLineEdit(edit lsp.TextEdit) error {
	// line := d.Lines.Line(edit.Range.Start.Line)

	// // Create new line content
	// newLine := line[:edit.Range.Start.Character] +
	// 	edit.NewText +
	// 	line[edit.Range.End.Character:]

	// // Replace the line
	// d.Lines[edit.Range.Start.Line] = newLine
	start, end := GetEditLoc(edit)
	d.Lines.Replace(start, end, edit.NewText)
	return nil
}

// applyMultiLineEdit handles edits that span multiple lines
func (d *Format) applyMultiLineEdit(edit lsp.TextEdit) error {
	// Get the prefix of the first line
	start, _ := GetEditLoc(edit)

	firstLine := d.Lines.Line(edit.Range.Start.Line)
	prefix := firstLine[:edit.Range.Start.Character]
	// Get the suffix of the last line
	lastLine := d.Lines.Line(edit.Range.End.Line)
	suffix := lastLine[edit.Range.End.Character:]

	// Split the new text into lines
	newLines := strings.Split(edit.NewText, "\n")

	// Combine prefix with first new line
	if len(newLines) > 0 {
		newLines[0] = prefix + newLines[0]
	} else {
		newLines = []string{prefix}
	}

	// Combine suffix with last new line
	lastIndex := len(newLines) - 1
	newLines[lastIndex] = newLines[lastIndex] + suffix
	newFunction2(edit, d)
	for i, v := range newLines {
		debug.DebugLog("format", "++", i+start.Y, ":", len(v), fmt.Sprintf("[%s]", v))
	}
	if len(newLines) == 1 {
		start, end := GetEditLoc(edit)
		end.X = max(end.X, len(lastLine))
		d.Lines.Replace(start, end, edit.NewText+lastLine)
	} else {
		for i := range newLines {
			v := newLines[i]
			lineNr := i + start.Y
			x1 := len(d.Lines.Line(lineNr))
			d.Lines.Replace(femto.Loc{Y: lineNr, X: 0}, femto.Loc{Y: i + start.Y, X: x1}, v)
			newline := d.Lines.Line(lineNr)
			debug.DebugLog("format", "ReplaceAfter-", v, "-", lineNr, ":", len(newline), fmt.Sprintf("[%s]", newline))
		}
	}
	// d.Lines.Replace(femto.Loc{Y: edit.Range.Start.Line, X: 0}, end, strings.Join(newLines, "\n"))
	// Replace the old lines with new ones
	// d.Lines = append(
	// 	d.Lines[:edit.Range.Start.Line],
	// 	append(
	// 		newLines,
	// 		d.Lines[edit.Range.End.Line+1:]...,
	// 	)...,
	// )
	return nil
}

func newFunction2(edit lsp.TextEdit, d *Format) {
	start, end := GetEditLoc(edit)
	var lastLine = d.Lines.Line(end.Y)
	var firstLine = d.Lines.Line(start.Y)

	if edit.Range.End.Line+1 < d.Lines.LinesNum() {
		end = femto.Loc{Y: edit.Range.End.Line + 1, X: 0}
	}
	x := '?'
	if len(lastLine) > end.X {
		x = rune(lastLine[end.X])
	}
	x1 := '?'
	if start.X < len(firstLine) {

		x1 = rune(firstLine[start.X])
	}
	debug.DebugLog("format", start.Y, []rune(edit.NewText),
		fmt.Sprintf("end %d:%d [%c] '%s' %d", end.Y, end.X, x, lastLine, len(lastLine)),
		fmt.Sprintf("start %d:%d [%c] '%s' %d", start.Y, start.X, x1, firstLine, len(firstLine)))
}
