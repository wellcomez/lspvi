package lspcore

import (
	"fmt"
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
	}
	core.capabilities = defaultCapabilities
	core.initializationOptions = map[string]interface{}{}
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

	var ext = []string{"go", "gomod", "gowork", "gotmpl"}
	return IsMe(filename, ext)
}
