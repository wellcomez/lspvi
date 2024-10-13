package lspcore

import (
	"fmt"
	"os/exec"

	"github.com/tectiv3/go-lsp"
)

func (l lsp_lang_py) IsSource(filename string) bool {
	return false
}

var rootFiles = []string{
	"pyproject.toml",
	"setup.py",
	"setup.cfg",
	"requirements.txt",
	"Pipfile",
	"pyrightconfig.json",
	".git",
}

type lsp_lang_py struct {
}

// Launch_Lsp_Server implements lsplang.
func (l lsp_lang_py) Launch_Lsp_Server(core *lspcore, wk WorkSpace) error {
	if core.started {
		return nil
	}
	core.cmd = exec.Command("python3", "-m", "pylsp")
	err := core.Lauch_Lsp_Server(core.cmd)
	core.started = err == nil
	return err
}

func (l lsp_lang_py) Resolve(sym lsp.SymbolInformation, symfile *Symbol_file) bool {
	return false
}

// InitializeLsp implements lsplang.
func (l lsp_lang_py) InitializeLsp(core *lspcore, wk WorkSpace) error {
	if core.inited {
		return nil
	}
	result, err := core.Initialize(wk)
	if err != nil {
		return err
	}
	if result.ServerInfo.Name == "pylsp" {
		core.inited = true
		core.get_sync_option(result)
		return nil
	}
	return fmt.Errorf("%s", result.ServerInfo.Name)
}

// IsSource implements lsplang.

// Resolve implements lsplang.

// IsMe implements lsplang.
func (l lsp_lang_py) IsMe(filename string) bool {
	return IsMe(filename, []string{"py"})
}
