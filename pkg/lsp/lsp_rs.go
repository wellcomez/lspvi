package lspcore

import (
	"context"
	"fmt"
	"os/exec"

	lsp "github.com/tectiv3/go-lsp"
)

type lsp_lang_rs struct {
	lsp_lang_common
}

func (l lsp_lang_rs) IsSource(filename string) bool {
	return true
}

// Launch_Lsp_Server implements lsplang.
func (l lsp_lang_rs) Launch_Lsp_Server(core *lspcore, wk WorkSpace) error {
	if core.started {
		return nil
	}
	if !core.RunComandInConfig() {
		core.cmd = exec.Command("rust-analyzer")
	}
	err := core.Launch_Lsp_Server(core.cmd)
	core.started = err == nil
	return err
}

func (l lsp_lang_rs) Resolve(sym lsp.SymbolInformation, symfile *Symbol_file) bool {
	return false
}
func (lang *lsp_lang_rs) Initialize(wk WorkSpace, core *lspcore) (lsp.InitializeResult, error) {
	var ProcessID = core.cmd.Process.Pid
	// 发送initialize请求
	var result lsp.InitializeResult
	if err := core.conn.Call(context.Background(), "initialize", lsp.InitializeParams{
		ProcessID:             &ProcessID,
		RootURI:               lsp.NewDocumentURI(wk.Path),
		InitializationOptions: core.initializationOptions,
		Capabilities:          core.capabilities,
	}, &result); err != nil {
		return result, err
	}
	return result, nil
}

// InitializeLsp implements lsplang.
func (l lsp_lang_rs) InitializeLsp(core *lspcore, wk WorkSpace) error {
	if core.inited {
		return nil
	}
	defaultCapabilities := map[string]interface{}{
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
	core.capabilities = defaultCapabilities
	core.initializationOptions = map[string]interface{}{}
	result, err := l.Initialize(wk, core)
	if err != nil {
		return err
	}
	if result.ServerInfo.Name != "rust-analyzer" {
		return fmt.Errorf("worng rust lsp %s", result.ServerInfo.Name)
	}
	core.get_sync_option(result)
	core.inited = true
	return nil
	// }
	// return fmt.Errorf("%s", result.ServerInfo.Name)
}

func (l lsp_lang_rs) IsMe(filename string) bool {
	var ext = []string{"rs"}
	return IsMe(filename, ext)
}
