package lspcore

import "github.com/tectiv3/go-lsp"

type lsp_md struct {
	lsp_lang_common
}

// InitializeLsp implements lsplang.
func (l lsp_md) InitializeLsp(core *lspcore, wk WorkSpace) error {
	panic("unimplemented")
}

// IsMe implements lsplang.
func (l lsp_md) IsMe(filename string) bool {
	return IsMe(filename, []string{"md"})
}

// IsSource implements lsplang.
func (l lsp_md) IsSource(filename string) bool {
	panic("unimplemented")
}

// Launch_Lsp_Server implements lsplang.
func (l lsp_md) Launch_Lsp_Server(core *lspcore, wk WorkSpace) error {
	panic("unimplemented")
}

// Resolve implements lsplang.
func (l lsp_md) Resolve(sym lsp.SymbolInformation, symfile *Symbol_file) bool {
	panic("unimplemented")
}
