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

// CallHierarchyIncomingCalls implements lspclient.
// Subtle: this method shadows the method (lsp_base).CallHierarchyIncomingCalls of lsp_cpp.lsp_base.
func (l lsp_cpp) CallHierarchyIncomingCalls(param lsp.CallHierarchyItem) ([]lsp.CallHierarchyIncomingCall, error) {
	return lsp_base.CallHierarchyIncomingCalls(l.lsp_base, param)
}

// DidOpen implements lspclient.
// Subtle: this method shadows the method (lsp_base).DidOpen of lsp_cpp.lsp_base.
func (l lsp_cpp) DidOpen(file string) error {
	return lsp_base.DidOpen(l.lsp_base, file)
}

// GetDeclare implements lspclient.
// Subtle: this method shadows the method (lsp_base).GetDeclare of lsp_cpp.lsp_base.
func (l lsp_cpp) GetDeclare(file string, pos lsp.Position) ([]lsp.Location, error) {
	return lsp_base.GetDeclare(l.lsp_base, file, pos)
}

// GetDeclareByLocation implements lspclient.
// Subtle: this method shadows the method (lsp_base).GetDeclareByLocation of lsp_cpp.lsp_base.
func (l lsp_cpp) GetDeclareByLocation(loc lsp.Location) ([]lsp.Location, error) {
	return lsp_base.GetDeclareByLocation(l.lsp_base, loc)
}

// GetDocumentSymbol implements lspclient.
// Subtle: this method shadows the method (lsp_base).GetDocumentSymbol of lsp_cpp.lsp_base.
func (l lsp_cpp) GetDocumentSymbol(file string) (*document_symbol, error) {
	return lsp_base.GetDocumentSymbol(l.lsp_base, file)
}

// GetReferences implements lspclient.
// Subtle: this method shadows the method (lsp_base).GetReferences of lsp_cpp.lsp_base.
func (l lsp_cpp) GetReferences(file string, pos lsp.Position) ([]lsp.Location, error) {
	return lsp_base.GetReferences(l.lsp_base, file, pos)
}

// InitializeLsp implements lspclient.

// IsMe implements lspclient.
// Subtle: this method shadows the method (lsp_base).IsMe of lsp_cpp.lsp_base.
func (l lsp_cpp) IsMe(filename string) bool {
	return lsp_base.IsMe(l.lsp_base, filename)
}

// PrepareCallHierarchy implements lspclient.
// Subtle: this method shadows the method (lsp_base).PrepareCallHierarchy of lsp_cpp.lsp_base.
func (l lsp_cpp) PrepareCallHierarchy(loc lsp.Location) ([]lsp.CallHierarchyItem, error) {
	return lsp_base.PrepareCallHierarchy(l.lsp_base, loc)
}

func new_lsp_cpp(wk WorkSpace,core *lspcore ) lsp_cpp {
	ret := lsp_cpp{
		new_lsp_base(wk,core),
	}
	ret.file_extensions = file_extensions
	ret.root_files = root_files
	return ret
}
func (l lsp_cpp) Resolve(sym lsp.SymbolInformation, symfile *Symbol_file) bool {
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
func (l lsp_cpp) InitializeLsp(wk WorkSpace) error {
	if l.core.inited {
		return nil
	}
	result, err := l.core.Initialize(wk)
	if err != nil {
		return err
	}
	if result.ServerInfo.Name == "clangd" {
		l.core.inited = true
		return nil
	}
	return fmt.Errorf("%s", result.ServerInfo.Name)
}

func (l lsp_cpp) Close() {
	l.core.cmd.Process.Kill()
}

// Launch_Lsp_Server implements lspclient.
func (l lsp_cpp) Launch_Lsp_Server() error {
	if l.core.started {
		return nil
	}
	root := "--compile-commands-dir=" + l.wk.Path
	l.core.cmd = exec.Command("clangd", root)
	err := l.core.Lauch_Lsp_Server(l.core.cmd)
	l.core.started = err == nil
	return err
}
