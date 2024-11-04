package hlresult

import lspcore "zen108.com/lspvi/pkg/lsp"

type MatchPosition struct {
	X, Y int
}
type HLResult struct {
	Tree    lspcore.TreesiterSymbolLine
	Matches []MatchPosition
}
