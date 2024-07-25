package lspcore

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tectiv3/go-lsp"
)

var file_extensions = []string{"cc", "cpp", "h", "hpp", "cxx", "hxx",
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
}

// IsMe implements lsplang.
func (l lsp_lang_cpp) IsMe(filename string) bool {
	return IsMe(filename, file_extensions)
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
						classname: classname,
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
	core.cmd = exec.Command("clangd", root)
	err := core.Lauch_Lsp_Server(core.cmd)
	core.started = err == nil
	return err
}
func new_lsp_lang(wk WorkSpace, core *lspcore) lsplang {
	return lsp_lang_cpp{}
}
