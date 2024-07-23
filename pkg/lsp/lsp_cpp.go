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

type lsp_cpp struct {
	lsp_base
}

func new_lsp_cpp(wk WorkSpace) lsp_cpp {
	ret := lsp_cpp{
		new_lsp_base(wk),
	}
	ret.file_extensions = file_extensions
	ret.root_files = root_files
	return ret
}
func (l lsp_cpp) Resolve(sym lsp.SymbolInformation) (*lsp.SymbolInformation, bool) {
	filename := sym.Location.URI.AsPath().String()
	yes := l.IsSource(filename)
	if yes == false {
		return nil, false
	}
	return nil, false
}
func (l lsp_cpp) IsSource(filename string) bool {
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
func (l *lsp_cpp) InitializeLsp(wk WorkSpace) error {
	if l.inited {
		return nil
	}
	result, err := l.core.Initialize(wk)
	if err != nil {
		return err
	}
	if result.ServerInfo.Name == "clangd" {
		l.inited = true
		return nil
	}
	return fmt.Errorf("%s", result.ServerInfo.Name)
}

// Launch_Lsp_Server implements lspclient.
func (l lsp_cpp) Launch_Lsp_Server() error {
	if l.started {
		return nil
	}
	root := "--compile-commands-dir=" + l.wk.Path
	l.core.cmd = exec.Command("clangd", root)
	err := l.core.Lauch_Lsp_Server(l.core.cmd)
	l.started = err == nil
	return err
}
