package lspcore

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	lsp "github.com/tectiv3/go-lsp"
)

type lsp_lang_go struct {
	LangConfig
}

func (l lsp_lang_go) IsSource(filename string) bool {
	return true
}

// Launch_Lsp_Server implements lsplang.
func (l lsp_lang_go) Launch_Lsp_Server(core *lspcore, wk WorkSpace) error {
	if core.started {
		return nil
	}
	if l.is_cmd_ok() {
		core.cmd = exec.Command(l.Cmd)
	} else {
		core.cmd = exec.Command("gopls")
	}
	err := core.Lauch_Lsp_Server(core.cmd)
	core.started = err == nil
	return err
}

func (l lsp_lang_go) Resolve(sym lsp.SymbolInformation, symfile *Symbol_file) bool {
	if sym.Kind == lsp.SymbolKindMethod {
		b := strings.Index(sym.Name, "(")
		e := strings.Index(sym.Name, ")")
		if e-b > 0 {
			sss := sym.Name[b+1 : e]
			classname := strings.TrimLeft(sss, "*")
			member := sym
			member.Name = member.Name[e+2:]
			symfile.Convert_function_method(classname, member, sym)
			return true
		}
		return true
	}
	return false
}

func (symfile *Symbol_file) Convert_function_method(classname string, member lsp.SymbolInformation, sym lsp.SymbolInformation) {
	for _, v := range symfile.Class_object {
		if v.SymInfo.Name == classname {
			v.Members = append(v.Members, Symbol{
				SymInfo:   member,
				Classname: classname,
			})
			return
		}
	}
	sym.Name = classname
	sym.Kind = lsp.SymbolKindClass
	classnew := Symbol{
		SymInfo: sym,
	}
	classnew.Members = append(classnew.Members, Symbol{SymInfo: member})
	symfile.Class_object = append(symfile.Class_object, &classnew)
}

type SaveOptions struct {
	// IncludeText is the client is supposed to include the content on save.
	IncludeText bool `json:"includeText,omitempty"`
}

// FoldingRangeProviderOptions FoldingRangeProvider options.

// TextDocumentSyncOptions TextDocumentSync options.
type TextDocumentSyncOptions struct {
	// OpenClose open and close notifications are sent to the server.
	OpenClose bool `json:"openClose,omitempty"`

	// Change notifications are sent to the server. See TextDocumentSyncKind.None, TextDocumentSyncKind.Full
	// and TextDocumentSyncKind.Incremental. If omitted it defaults to TextDocumentSyncKind.None.
	Change float64 `json:"change,omitempty"`

	// WillSave notifications are sent to the server.
	WillSave bool `json:"willSave,omitempty"`

	// WillSaveWaitUntil will save wait until requests are sent to the server.
	WillSaveWaitUntil bool `json:"willSaveWaitUntil,omitempty"`

	// Save notifications are sent to the server.
	Save *SaveOptions `json:"save,omitempty"`
}

// InitializeLsp implements lsplang.
func (l lsp_lang_go) InitializeLsp(core *lspcore, wk WorkSpace) error {
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
	result, err := core.Initialize(wk)
	if err != nil {
		return err
	}
	core.get_sync_option(result)
	if result.ServerInfo.Name == "gopls" {
		core.inited = true
		return nil
	}
	return fmt.Errorf("%s", result.ServerInfo.Name)
}

func (core *lspcore) get_sync_option(result lsp.InitializeResult) {
	var r TextDocumentSyncOptions
	if d,e:=json.MarshalIndent(result, " ", "");e==nil{
		log.Println(string(d))
	}
	if data, err := result.Capabilities.TextDocumentSync.MarshalJSON(); err == nil {
		if err = json.Unmarshal(data, &r); err == nil {
			core.sync = &r
		}
	}
}

// IsMe implements lsplang.
func (l lsp_lang_go) IsMe(filename string) bool {

	var ext = []string{"go", "gomod", "gowork", "gotmpl"}
	return IsMe(filename, ext)
}
