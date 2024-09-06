package lspcore

import (
	"github.com/tectiv3/go-lsp"
)

type lsplang interface {
	Launch_Lsp_Server(core *lspcore, wk WorkSpace) error
	InitializeLsp(core *lspcore, wk WorkSpace) error
	IsSource(filename string) bool
	Resolve(sym lsp.SymbolInformation, symfile *Symbol_file) bool
	IsMe(filename string) bool
	TreeSymbolParser(*TreeSitter) []lsp.SymbolInformation
}
