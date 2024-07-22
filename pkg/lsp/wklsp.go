package lspcore

type lspwk struct {
	cpp lsp_cpp
	py  lsp_py
}

func NewLspWk(wk workroot) *lspwk {
	return &lspwk{
		cpp: lsp_cpp{new_lsp_base(wk)},
		py:  lsp_py{new_lsp_base(wk)},
	}
}
