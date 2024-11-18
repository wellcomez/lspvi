package lspcore

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/tectiv3/go-lsp"
	"zen108.com/lspvi/pkg/mason"
	"zen108.com/lspvi/pkg/ui/common"
)

func (l lsp_lang_jedi) IsSource(filename string) bool {
	return false
}

type lsp_lang_jedi struct {
	lsp_lang_common
}

// Launch_Lsp_Server implements lsplang.
func (l lsp_lang_jedi) Launch_Lsp_Server(core *lspcore, wk WorkSpace) (err error) {
	if core.started {
		return nil
	}
	var path string
	path, err = exec.LookPath("jedi-language-server")
	if err != nil {
		var w common.Workdir
		w, err = common.NewMkWorkdir(wk.Path)
		if err != nil {
			return
		}
		if soft, e := mason.NewSoftManager(w).FindLsp(mason.ToolLsp_java_jedi); e != nil {
			err = e
			return
		} else {
			path = soft.Executable()
			if path == "" {
				err = fmt.Errorf("not found jedi-language-server")
				return
			}
		}
	}
	core.cmd = exec.Command(path)
	err = core.Launch_Lsp_Server(core.cmd)
	core.started = err == nil
	return err
}

func (l lsp_lang_jedi) Resolve(sym lsp.SymbolInformation, symfile *Symbol_file) bool {
	return false
}

// InitializeLsp implements lsplang.
func (l lsp_lang_jedi) InitializeLsp(core *lspcore, wk WorkSpace) error {
	if core.inited {
		return nil
	}
	var capabilities = map[string]interface{}{
		"textDocument": map[string]interface{}{
			"hover": map[string]interface{}{
				"dynamicRegistration": true,
				"contentFormat":       []interface{}{"plaintext", "markdown"},
			},
			"synchronization": map[string]interface{}{
				"dynamicRegistration": true,
				"willSave":            false,
				"didSave":             false,
				"willSaveWaitUntil":   false,
			},
			"completion": map[string]interface{}{
				"dynamicRegistration": true,
				"completionItem": map[string]interface{}{
					"snippetSupport":          false,
					"commitCharactersSupport": true,
					"documentationFormat":     []interface{}{"plaintext", "markdown"},
					"deprecatedSupport":       false,
					"preselectSupport":        false,
				},
				"contextSupport": false,
			},
			"signatureHelp": map[string]interface{}{
				"dynamicRegistration": true,
				"signatureInformation": map[string]interface{}{
					"documentationFormat": []interface{}{"plaintext", "markdown"},
				},
			},
			"declaration": map[string]interface{}{
				"dynamicRegistration": true,
				"linkSupport":         true,
			},
			"definition": map[string]interface{}{
				"dynamicRegistration": true,
				"linkSupport":         true,
			},
			"typeDefinition": map[string]interface{}{
				"dynamicRegistration": true,
				"linkSupport":         true,
			},
			"implementation": map[string]interface{}{
				"dynamicRegistration": true,
				"linkSupport":         true,
			},
		},
		"workspace": map[string]interface{}{
			"didChangeConfiguration": map[string]interface{}{
				"dynamicRegistration": true,
			},
		},
	}
	core.capabilities = capabilities
	core.initializationOptions = map[string]interface{}{}
	var result lsp.InitializeResult
	if err := core.conn.Call(context.Background(), "initialize", lsp.InitializeParams{
		RootURI:               lsp.NewDocumentURI(wk.Path),
		InitializationOptions: core.initializationOptions,
		Capabilities:          core.capabilities,
	}, &result); err != nil {
		return err
	} else {
		if result.ServerInfo.Name == "jedi-language-server" {
			core.inited = true
			core.get_sync_option(result)
			return nil
		}
	}
	return fmt.Errorf("%s", result.ServerInfo.Name)
}

func (l lsp_lang_jedi) IsMe(filename string) bool {
	return IsMe(filename, []string{"java"})
}
