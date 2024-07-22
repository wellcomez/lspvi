package lspcore

type lspwk struct {
	cpp *lsp_cpp
	py  *lsp_py
}

func NewLspWk(wk workroot) *lspwk {
	return &lspwk{
		cpp: new_lsp_cpp(wk),
		py:  new_lsp_py(wk),
	}
}
