package hlresult

import lspcore "zen108.com/lspvi/pkg/lsp"

type MatchPosition struct {
	Begin, End, Y int
}
type SearchLine struct {
	Lines map[int][]MatchPosition
}

func NewSearchLine() SearchLine {
	return SearchLine{Lines: make(map[int][]MatchPosition)}
}
func (l *SearchLine) Add(pos MatchPosition) {
	if _, yes := l.Lines[pos.Y]; !yes {
		l.Lines[pos.Y] = []MatchPosition{}
	}
	l.Lines[pos.Y] = append(l.Lines[pos.Y], pos)
}

type HLResult struct {
	Tree         lspcore.TreesiterSymbolLine
	Current      MatchPosition
	SearchResult SearchLine
	Diagnos      SearchLine
}

func (h *HLResult) Update() {
	for line := range h.SearchResult.Lines {
		h.update_line(line)
	}
}
func (h *HLResult) UpdateErrorPosition(l SearchLine) {
	h.Diagnos = l
}
func (h *HLResult) GetErrorPosition(lineno int) []MatchPosition {
	return h.Diagnos.Lines[lineno]
}
func (h *HLResult) GetMatchPosition(lineno int) []MatchPosition {
	return h.SearchResult.Lines[lineno]
}
func (h *HLResult) update_line(lineno int) {
	sym_in_line := h.Tree[lineno]
	search_result := h.SearchResult.Lines[lineno]
	if len(search_result) > 0 {
		var new_outline = []lspcore.TreeSitterSymbol{}
		word_index := 0
		for _, outline := range sym_in_line {
			split := false
			for i := word_index; i < len(search_result); i++ {
				word := search_result[i]
				if word.End < int(outline.Begin.Column) {
					word_index++
					continue
				}
				if int(outline.Begin.Column) <= word.Begin && int(outline.End.Column) >= word.End {
					// word_index++
					col1 := outline
					col1.End.Column = uint32(word.Begin)
					col2 := outline
					col2.Begin.Column = uint32(word.End)
					split = true
					if col1.Begin.Column < col1.End.Column {
						new_outline = append(new_outline, col1)
					}
					// new_outline = append(new_outline, lspcore.TreeSitterSymbol{
					// 	SymbolName: "search",
					// 	Begin:      lspcore.Point{Row: uint32(lineno), Column: uint32(word.Begin)},
					// 	End:        lspcore.Point{Row: uint32(lineno), Column: uint32(word.End)},
					// })
					if col2.End.Column > col2.Begin.Column {
						// new_outline = append(new_outline, col2)
					}
				} else if word.End < int(outline.Begin.Column) {
					// word_index++
					// new_outline = append(new_outline, lspcore.TreeSitterSymbol{
					// 	SymbolName: "search",
					// 	Begin:      lspcore.Point{Row: uint32(lineno), Column: uint32(word.Begin)},
					// 	End:        lspcore.Point{Row: uint32(lineno), Column: uint32(word.End)},
					// })
				} else {

					break
				}
			}
			if !split {
				new_outline = append(new_outline, outline)
			}
		}
		h.Tree[lineno] = new_outline
	}
}
