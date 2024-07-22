package lspcore

type lspwk struct {
	cpp lsp_cpp
	py  lsp_py
	wk  workroot
}

func (wk lspwk) GetClient(filename string) *lsp_base {
	if wk.cpp.IsMe(filename) {
		return &wk.cpp.lsp_base
	}
	if wk.py.IsMe(filename) {
		return &wk.py.lsp_base
	}
	return nil
}
func NewLspWk(wk workroot) *lspwk {
	return &lspwk{
		cpp: new_lsp_cpp(wk),
		py:  lsp_py{new_lsp_base(wk)},
		wk:  wk,
	}
}
