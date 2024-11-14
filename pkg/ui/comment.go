package mainui

import (
	"path/filepath"

	"github.com/pgavlin/femto"
)

func (code CodeView) GetLineCommentChar() (commentChar string) {
	commentChar = "//"
	ext := filepath.Ext(code.FileName())
	switch ext {
	case ".go", ".c", ".cpp", ".h", ".hpp", ".cc", ".hxx", ".cxx", ".java", ".js", ".ts", ".tsx", ".rs", ".sh", ".bash", ".zsh", ".fish", ".php", ".html", ".css", ".xml", ".json", ".yaml", ".yml", ".toml", ".ini", ".cfg", ".conf", ".md", ".rst", ".tex", ".bib", ".bibtex", ".biblate":
		commentChar = "// "
	case ".py":
		commentChar = "#"
	}
	return
}
func (code *CodeView) CommentLine() {
	sel := code.view.Cursor.CurSelection
	line := []int{}
	for i := min(sel[0].Y, sel[1].Y); i <= max(sel[0].Y, sel[1].Y); i++ {
		line = append(line, i)
	}
	for _, v := range line {
		code.view.Buf.Insert(femto.Loc{X: 0, Y: v}, code.GetLineCommentChar())
	}
}
func (code *CodeView) CommentLineNo(line []int) {
	for _, v := range line {
		code.view.Buf.Insert(femto.Loc{X: 0, Y: v}, code.GetLineCommentChar())
	}
}
