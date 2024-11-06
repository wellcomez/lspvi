package lspcore

import "github.com/tectiv3/go-lsp"

type lsp_dummy struct {
	lsp_lang_base
}

// InitializeLsp implements lsplang.
func (l lsp_dummy) InitializeLsp(core *lspcore, wk WorkSpace) error {
	return nil
}

// IsMe implements lsplang.
func (l lsp_dummy) IsMe(filename string) bool {
	return false
}

// IsSource implements lsplang.
func (l lsp_dummy) IsSource(filename string) bool {
	return false
}

// Launch_Lsp_Server implements lsplang.
func (l lsp_dummy) Launch_Lsp_Server(core *lspcore, wk WorkSpace) error {
	return nil
}

// Resolve implements lsplang.
func (l lsp_dummy) Resolve(sym lsp.SymbolInformation, symfile *Symbol_file) bool {
	return false
}
