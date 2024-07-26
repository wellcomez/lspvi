package lspcore

import (
	"path/filepath"
	"strings"

	"github.com/tectiv3/go-lsp"
)

type lspclient interface {
	Semantictokens_full(file string) (*lsp.SemanticTokens, error)
	InitializeLsp(wk WorkSpace) error
	Launch_Lsp_Server() error
	DidOpen(file string) error
	GetDocumentSymbol(file string) (*document_symbol, error)
	GetReferences(file string, pos lsp.Position) ([]lsp.Location, error)
	GetDeclareByLocation(loc lsp.Location) ([]lsp.Location, error)
	GetDeclare(file string, pos lsp.Position) ([]lsp.Location, error)
	GetDefine(file string, pos lsp.Position) ([]lsp.Location, error)
	PrepareCallHierarchy(loc lsp.Location) ([]lsp.CallHierarchyItem, error)
	CallHierarchyIncomingCalls(param lsp.CallHierarchyItem) ([]lsp.CallHierarchyIncomingCall, error)
	IsMe(filename string) bool
	IsSource(filename string) bool
	Resolve(sym lsp.SymbolInformation, symbolfile *Symbol_file) bool
	Close()
}
type lsp_base struct {
	core *lspcore
	wk   *WorkSpace
}

func (l lsp_base) Semantictokens_full(file string) (*lsp.SemanticTokens, error) {
	return l.core.document_semantictokens_full(file)
}
func (l lsp_base) InitializeLsp(wk WorkSpace) error {
	err:=l.core.lang.InitializeLsp(l.core, wk)
	if err!=nil{
		return err
	}
	l.core.Initialized()
	l.core.Progress_notify()
	return nil
}
func (l lsp_base) Launch_Lsp_Server() error {
	return l.core.lang.Launch_Lsp_Server(l.core, *l.wk)
}

func (l lsp_base) IsSource(filename string) bool {
	return l.core.lang.IsSource(filename)
}
func (l lsp_base) Resolve(sym lsp.SymbolInformation, symbolfile *Symbol_file) bool {
	return l.core.lang.Resolve(sym, symbolfile)
}

// DidOpen implements lspclient.
// Subtle: this method shadows the method (lsp_base).DidOpen of lsp_py.lsp_base.
func (l lsp_base) Close() {
	l.core.cmd.Process.Kill()
}
func IsMe(filename string, file_extensions []string) bool {
	ext := filepath.Ext(filename)
	ext = strings.TrimPrefix(ext, ".")
	for _, v := range file_extensions {
		if v == ext {
			return true
		}
	}
	return false
}
func (l lsp_base) IsMe(filename string) bool {
	return l.core.lang.IsMe(filename)
}

// Initialize implements lspclient.
func (l lsp_base) DidOpen(file string) error {
	return l.core.DidOpen(file)
}

func (l lsp_base) PrepareCallHierarchy(loc lsp.Location) ([]lsp.CallHierarchyItem, error) {
	return l.core.TextDocumentPrepareCallHierarchy(loc)
}
func (l lsp_base) GetDefine(file string, pos lsp.Position) ([]lsp.Location, error) {
	return l.core.GetDefine(file, pos)
}
func (l lsp_base) CallHierarchyIncomingCalls(param lsp.CallHierarchyItem) ([]lsp.CallHierarchyIncomingCall, error) {
	return l.core.CallHierarchyIncomingCalls(lsp.CallHierarchyIncomingCallsParams{
		Item: param,
	})
}
func (l lsp_base) GetDeclareByLocation(loc lsp.Location) ([]lsp.Location, error) {
	path := LocationContent{
		location: loc,
	}.Path()
	return l.core.GetDeclare(path, lsp.Position{
		Line:      loc.Range.Start.Line,
		Character: loc.Range.Start.Character,
	})
}
func (l lsp_base) GetDeclare(file string, pos lsp.Position) ([]lsp.Location, error) {

	return l.core.GetDeclare(file, pos)
}
func (l lsp_base) GetReferences(file string, pos lsp.Position) ([]lsp.Location, error) {
	return l.core.GetReferences(file, pos)
}
func (l lsp_base) GetDocumentSymbol(file string) (*document_symbol, error) {
	return l.core.GetDocumentSymbol(file)
}
