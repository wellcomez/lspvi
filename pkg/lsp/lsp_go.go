package lspcore

import (
	"fmt"
	"os/exec"

	lsp "github.com/tectiv3/go-lsp"
)

type lsp_lang_go struct {
}

func (l lsp_lang_go) IsSource(filename string) bool {
	return false
}

// Launch_Lsp_Server implements lsplang.
func (l lsp_lang_go) Launch_Lsp_Server(core *lspcore, wk WorkSpace) error {
	if core.started {
		return nil
	}
	core.cmd = exec.Command("gopls")
	err := core.Lauch_Lsp_Server(core.cmd)
	core.started = err == nil
	return err
}

func (l lsp_lang_go) Resolve(sym lsp.SymbolInformation, symfile *Symbol_file) bool {
	return false
}

// InitializeLsp implements lsplang.
func (l lsp_lang_go) InitializeLsp(core *lspcore, wk WorkSpace) error {
	if core.inited {
		return nil
	}
	result, err := core.Initialize(wk)
	if err != nil {
		return err
	}
	if result.ServerInfo.Name == "gopls" {
		core.inited = true
		return nil
	}
	return fmt.Errorf("%s", result.ServerInfo.Name)
}

// IsMe implements lsplang.
func (l lsp_lang_go) IsMe(filename string) bool {
	return IsMe(filename, []string{"go"})
}
