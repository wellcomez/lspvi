package lspcore

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tectiv3/go-lsp"
	// "zen108.com/lspvi/pkg/debug"
)

var file_extensions = []string{
    "m","mm",
	"cc", "cpp", "h", "hpp", "cxx", "hxx",
	"inl", "c", "cpp", "objc", "objcpp", "cuda", "proto"}
var root_files = []string{

	".clangd",
	".clang-tidy",
	".clang-format",
	"compile_commands.json",
	"compile_flags.txt",
	"configure.ac",
}

type lsp_lang_cpp struct {
	lsp_lang_common
}

// IsMe implements lsplang.
func (l lsp_lang_cpp) IsMe(filename string) bool {
	return IsMe(filename, file_extensions)
}
func (a lsp_lang_cpp) CompleteHelpCallback(cl lsp.CompletionList, ret *Complete, err error) {
	document := []string{}
	for index := range cl.Items {

		v := cl.Items[index]
		text := complete_cpp(v)
		document = append(document, strings.Join(text, "\n"))
	}
	ret.Result = &CompleteResult{Document: document, Complete: complete_cpp}
}

var completionItemKindMap = map[int]string{
	1:  "Text",
	2:  "Method",
	3:  "Function",
	4:  "Constructor",
	5:  "Field",
	6:  "Variable",
	7:  "Class",
	8:  "Interface",
	9:  "Module",
	10: "Property",
	11: "Unit",
	12: "Value",
	13: "Enum",
	14: "Keyword",
	15: "Snippet",
	16: "Color",
	17: "File",
	18: "Reference",
	19: "Folder",
	20: "EnumMember",
	21: "Constant",
	22: "Struct",
	23: "Event",
	24: "Operator",
	25: "TypeParameter",
}

func complete_cpp(v lsp.CompletionItem) []string {
	var text = []string{}
	var doc Document
	label := v.Label
	switch v.Kind {
	case lsp.CompletionItemKindClass:
		label = fmt.Sprintf("class %s{}", v.Label)
	case lsp.CompletionItemKindStruct:
		label = fmt.Sprintf("struct %s{}", v.Label)
	case lsp.CompletionItemKindMethod, lsp.CompletionItemKindFunction:
		label = fmt.Sprintf("%s %s{}", v.Detail, v.Label)
	case lsp.CompletionItemKindVariable:
		label = fmt.Sprintf("%s %s", v.Detail, v.Label)
	case lsp.CompletionItemKindEnum:
		label = fmt.Sprintf("enum %s{%s}", v.Detail, v.Label)
	default:
		// kindname, _ := completionItemKindMap[int(v.Kind)]
		// debug.DebugLog("complete", kindname, "detail", v.Detail, "label", v.Label)
	}
	text = append(text, label)
	if doc.Parser(v.Documentation) == nil {
		sss := strings.Split(doc.Value, "\n")
		for _, v := range sss {
			text = append(text, "//"+strings.ReplaceAll(v, "\t", "  "))
		}
	}
	return text
}

// Resolve implements lsplang.
func (l lsp_lang_cpp) Resolve(sym lsp.SymbolInformation, symfile *Symbol_file) bool {
	filename := sym.Location.URI.AsPath().String()
	yes := l.IsSource(filename)
	if !yes {
		return false
	}
	if is_memeber(sym.Kind) {
		xxx := sym
		classindex := strings.LastIndex(xxx.Name, "::")
		classname := ""
		if classindex >= 0 {
			classname = xxx.Name[0:classindex]
			member := sym
			member.Name = member.Name[classindex+2:]

			for _, v := range symfile.Class_object {
				if v.SymInfo.Name == classname {

					v.Members = append(v.Members, Symbol{
						SymInfo:   member,
						Classname: classname,
					})
					return true
				}
			}

			xxx.Name = classname
			xxx.Kind = lsp.SymbolKindClass
			classnew := Symbol{
				SymInfo: xxx,
			}

			classnew.Members = append(classnew.Members, Symbol{SymInfo: member})
			symfile.Class_object = append(symfile.Class_object, &classnew)
			return true
		}
	}

	return false
}
func (l lsp_lang_cpp) IsSource(filename string) bool {
	ext := filepath.Ext(filename)
	ext = strings.TrimPrefix(ext, ".")
	source := []string{"cpp", "cc"}
	for _, s := range source {
		if s == ext {
			return true
		}
	}
	return false
}

// InitializeLsp implements lsplang.
func (l lsp_lang_cpp) InitializeLsp(core *lspcore, wk WorkSpace) error {
	if core.inited {
		return nil
	}
	initializationOptions := map[string]interface{}{
		"ClangdFileStatus": true,
	}
	core.initializationOptions = initializationOptions
	capabilities := map[string]interface{}{
		"window": map[string]interface{}{
			"workDoneProgress": true,
		},
		"textDocument": map[string]interface{}{
			"completion": map[string]interface{}{
				"completionItem": map[string]interface{}{
					"commitCharactersSupport": true,
					"snippetSupport":          true,
				},
			},
		},
	}
	core.capabilities = capabilities
	result, err := core.Initialize(wk)
	if err != nil {
		return err
	}
	core.get_sync_option(result)
	if result.ServerInfo.Name == "clangd" {
		core.inited = true
		return nil
	}
	return fmt.Errorf("%s", result.ServerInfo.Name)
}

// IsSource implements lsplang.

// Launch_Lsp_Server implements lsplang.
func (l lsp_lang_cpp) Launch_Lsp_Server(core *lspcore, wk WorkSpace) error {
	if core.started {
		return nil
	}
	root := "--compile-commands-dir=" + wk.Path
	cmd := "clangd"
	if !core.RunComandInConfig() {
		core.cmd = exec.Command(cmd, root, "--background-index")
	}
	err := core.Launch_Lsp_Server(core.cmd)
	core.started = err == nil
	return err
}
func new_lsp_lang(wk WorkSpace, core *lspcore) lsplang {
	return lsp_lang_cpp{}
}
func (a lsp_lang_cpp) LspHelp(core *lspcore) (ret LspUtil, err error) {
	ret, _ = a.lsp_lang_common.LspHelp(core)
	ret.Complete.Document = complete_cpp
	ret.Signature.Document = func(v lsp.SignatureHelp, call SignatureHelp) (text []string) {
		for _, s := range v.Signatures {
			method := s.Label
			switch call.Kind {
			case lsp.CompletionItemKindFunction, lsp.CompletionItemKindMethod:
				// method = fmt.Sprintf("func %s", method)
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
