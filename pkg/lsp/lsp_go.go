package lspcore

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	// "path/filepath"
	"strings"

	lsp "github.com/tectiv3/go-lsp"
	"zen108.com/lspvi/pkg/debug"
)

type lsp_lang_go struct {
	lsp_lang_common
}

func (l lsp_lang_go) IsSource(filename string) bool {
	return true
}

var Launch_Lsp_Server = 0

// Launch_Lsp_Server implements lsplang.
func (l lsp_lang_go) Launch_Lsp_Server(core *lspcore, wk WorkSpace) error {
	if core.started {
		return nil
	}
	if Launch_Lsp_Server > 1 {
		debug.ErrorLog("Launch_Lsp_Server", "start twice ", Launch_Lsp_Server)
	}
	Launch_Lsp_Server++
	logifle := filepath.Join(
		filepath.Dir(wk.Export), "gopls.log")
	if !core.RunComandInConfig() {
		debug := false
		if !debug {
			core.cmd = exec.Command("gopls")
		} else {
			core.cmd = exec.Command("gopls", "-rpc.trace",
				"-logfile", logifle,
				"-v")
			core.cmd.Env = append(os.Environ(),
				fmt.Sprintf("GOPLS_LOGFILE=%s", logifle),
				"GOPLS_LOGLEVEL=debug",
			)
		}
	}
	err := core.Launch_Lsp_Server(core.cmd)
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
type TextDocumentSyncOptions struct {
	// Open and close notifications are sent to the server. If omitted open
	// close notification should not be sent.
	OpenClose bool `json:"openClose,omitempty"`

	// Change notifications are sent to the server. See
	// TextDocumentSyncKind.None, TextDocumentSyncKind.Full and
	// TextDocumentSyncKind.Incremental. If omitted it defaults to
	// TextDocumentSyncKind.None.
	Change lsp.TextDocumentSyncKind `json:"change,omitempty"`

	// If present will save notifications are sent to the server. If omitted
	// the notification should not be sent.
	WillSave bool `json:"willSave,omitempty"`

	// If present will save wait until requests are sent to the server. If
	// omitted the request should not be sent.
	WillSaveWaitUntil bool `json:"willSaveWaitUntil,omitempty"`

	// If present save notifications are sent to the server. If omitted the
	// notification should not be sent.
	Save      *SaveOptions `json:"save,omitempty"`
	save_bool bool
}
type TextDocumentSyncOptions2 struct {
	// OpenClose open and close notifications are sent to the server.
	OpenClose bool                     `json:"openClose,omitempty"`
	Change    lsp.TextDocumentSyncKind `json:"change,omitempty"`
	Save      bool                     `json:"save,omitempty"`
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
	/*if d, e := json.MarshalIndent(result, " ", ""); e == nil {
		log.Println(LSP_DEBUG_TAG, string(d))
	}*/
	core.CapabilitiesStatus = result.Capabilities
	if data, err := result.Capabilities.TextDocumentSync.MarshalJSON(); err == nil {
		if err = json.Unmarshal(data, &r); err == nil {
			core.sync = &r
			return

		}
		debug.ErrorLog(LSP_DEBUG_TAG, "TextDocumentSync Marsh Failed", err)
		var r2 TextDocumentSyncOptions2
		if err = json.Unmarshal(data, &r2); err == nil {
			core.sync = &TextDocumentSyncOptions{
				OpenClose: r2.OpenClose,
				save_bool: r2.Save,
				Change:    r2.Change,
			}
			return

		}
		debug.ErrorLog(LSP_DEBUG_TAG, "TextDocumentSync Marsh Failed", err)
	}
}

// IsMe implements lsplang.
func (l lsp_lang_go) IsMe(filename string) bool {
	var ext = []string{"go", "gomod", "gowork", "gotmpl"}
	return IsMe(filename, ext)
}
func (a lsp_lang_go) CompleteHelpCallback(cl lsp.CompletionList, ret *Complete, err error) {
	document := []string{}
	for index := range cl.Items {

		v := cl.Items[index]
		var text = []string{}
		text = create_complete_go(v)
		document = append(document, strings.Join(text, "\n"))
	}
	ret.Result = &CompleteResult{Document: document, Complete: create_complete_go}
}

func create_complete_go(v lsp.CompletionItem) (text []string) {
	s := ""
	switch v.Kind {
	case lsp.CompletionItemKindMethod:
		n := strings.Replace(v.Detail, "func", "", -1)
		s = fmt.Sprintf("func(...) %s %s", v.Label, n)
	case lsp.CompletionItemKindFunction:
		n := strings.Replace(v.Detail, "func", "", -1)
		s = fmt.Sprintf("func %s %s", v.Label, n)
	case lsp.CompletionItemKindVariable:
		s = fmt.Sprintf("var %s %s", v.Label, v.Detail)
	case lsp.CompletionItemKindStruct, lsp.CompletionItemKindInterface:
		s = fmt.Sprintf("type %s %s", v.Label, v.Detail)
	case lsp.CompletionItemKindClass:
		s = fmt.Sprintf("%s %s", v.Label, v.Detail)
	case lsp.CompletionItemKindConstant:
		s = fmt.Sprintf("const %s %s", v.Label, v.Detail)
	case lsp.CompletionItemKindModule:
		s = fmt.Sprintf("import (\n%s\n)//%s", v.Label, v.Detail)
	default:
		s = fmt.Sprintf("%s %s", v.Label, v.Detail)
	}
	debug.DebugLogf("complete", "go parse %d %s ", v.Kind, s)
	text = append(text, s)
	var doc Document
	if doc.Parser(v.Documentation) == nil {
		ss := strings.Split(doc.Value, "\n")
		for _, v := range ss {
			text = append(text, "//"+v)
		}
	}
	return
}

func (a lsp_lang_go) LspHelp(core *lspcore) (ret LspUtil, err error) {
	ret, _ = a.lsp_lang_common.LspHelp(core)
	ret.Complete.Document = create_complete_go
	ret.Signature.Document = func(v lsp.SignatureHelp, call SignatureHelp) (text []string) {
		for _, s := range v.Signatures {
			method := s.Label
			switch call.Kind {
			case lsp.CompletionItemKindFunction, lsp.CompletionItemKindMethod:
				method = fmt.Sprintf("func %s", method)
			}
			text = append(text, method)
			var signature_document Document
			// if len(v.Parameters) > 0 {
			// 	ret.label = v.Label
			// }
			if signature_document.Parser(s.Documentation) == nil {
				ss := strings.Split(signature_document.Value, "\n")
				for _, v := range ss {
					text = append(text, "//"+v)
				}
			}
		}
		return
	}
	return
}
