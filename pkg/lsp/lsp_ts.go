package lspcore

import (
	// "fmt"
	"github.com/tectiv3/go-lsp"
	"os/exec"
)

func (l lsp_ts) IsSource(filename string) bool {
	return false
}

// var rootFiles = []string{
// 	"pyproject.toml",
// 	"setup.py",
// 	"setup.cfg",
// 	"requirements.txt",
// 	"Pipfile",
// 	"pyrightconfig.json",
// 	".git",
// }

type lsp_ts struct {
	LanguageID string
}

// Launch_Lsp_Server implements lsplang.
func (l lsp_ts) Launch_Lsp_Server(core *lspcore, wk WorkSpace) error {
	if core.started {
		return nil
	}
	core.cmd = exec.Command("typescript-language-server", "--stdio")
	err := core.Lauch_Lsp_Server(core.cmd)
	core.started = err == nil
	return err
}

func (l lsp_ts) Resolve(sym lsp.SymbolInformation, symfile *Symbol_file) bool {
	return false
}
func (l lsp_ts) InitializeLsp(core *lspcore, wk WorkSpace) error {
	if core.inited {
		return nil
	}
	capabilities := map[string]interface{}{
		"workspace": map[string]interface{}{
			"workspaceFolders": true,
		},
	}
	core.capabilities = capabilities
	_, err := core.Initialize(wk)
	if err != nil {
		return err
	}

	// if result.ServerInfo.Name == "pylsp" {
	core.inited = true
	return nil
	// }
	// return fmt.Errorf("%s", result.ServerInfo.Name)
}

// IsSource implements lsplang.

// Resolve implements lsplang.

// IsMe implements lsplang.
func (l lsp_ts) IsMe(filename string) bool {
	if l.LanguageID == string(JAVASCRIPT) {
		return IsMe(filename, []string{"js"})
	}
	return IsMe(filename, []string{"js"})
}
