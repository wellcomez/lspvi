package lspcore

import (
	// "fmt"
	"os/exec"

	"github.com/tectiv3/go-lsp"
	"zen108.com/lspvi/pkg/mason"
)

func find_git_ancestor(filename string) (dir string, err error) {
	return
}

var rootpattern = []string{"buildServer.json", "*.xcodeproj", "*.xcworkspace", "Package.swift"}

func (l lsp_swift) IsSource(filename string) bool {
	return false
}

type lsp_swift struct {
	lsp_lang_common
	LanguageID string
}

// Launch_Lsp_Server implements lsplang.
func (l lsp_swift) Launch_Lsp_Server(core *lspcore, wk WorkSpace) error {
	if core.started {
		return nil
	}
	cmd, err := wk.GetLspBin("sourcekit-lsp", mason.ToolLsp_swift)
	if err != nil {
		return err
	}

	if args := core.config.Args; len(args) > 0 {
		core.cmd = exec.Command(cmd, args...)
	} else if !core.RunComandInConfig() {
		core.cmd = exec.Command(cmd)
	}
	err = core.Launch_Lsp_Server(core.cmd)
	core.started = err == nil
	return err
}

func (l lsp_swift) Resolve(sym lsp.SymbolInformation, symfile *Symbol_file) bool {
	return false
}
func (l lsp_swift) InitializeLsp(core *lspcore, wk WorkSpace) (err error) {
	if core.inited {
		return
	}
	capabilities := map[string]interface{}{
		// "workspace": map[string]interface{}{
		// 	"workspaceFolders": true,
		// },
		"textDocument": map[string]interface{}{
			"completion": map[string]interface{}{
				"completionItem": map[string]interface{}{
					"commitCharactersSupport": true,
					"documentationFormat":     []interface{}{"markdown", "plaintext"},
					"snippetSupport":          true,
				},
			},
		},
		"textDocumentSync": map[string]interface{}{
			"openClose": true,                                // Notify server when documents are opened/closed
			"change":    lsp.TextDocumentSyncKindIncremental, // Send incremental updates
			"willSave":  true,                                // Notify before saving
			"save": map[string]interface{}{
				"includeText": true, // Send full document content when saving
			},
		},
	}
	core.capabilities = capabilities
	if result, e := core.Initialize(wk); e == nil {
		core.inited = true
		core.get_sync_option(result)
	} else {
		err = e
	}
	return
}

// IsSource implements lsplang.

// Resolve implements lsplang.

// IsMe implements lsplang.
func (l lsp_swift) IsMe(filename string) bool {
	return IsMe(filename, []string{"swift"})
}
