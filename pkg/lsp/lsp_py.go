package lspcore

import "github.com/tectiv3/go-lsp"

type lsp_py struct {
	lsp_base
}

// GetDefine implements lspclient.
func (l lsp_py) GetDefine(file string, pos lsp.Position) ([]lsp.Location, error) {
	return lsp_base.GetDefine(l.lsp_base, file, pos)
}

// IsSource implements lspclient.
// Subtle: this method shadows the method (lsp_base).IsSource of lsp_py.lsp_base.
func (l lsp_py) IsSource(filename string) bool {
	return lsp_base.IsSource(l.lsp_base, filename)
}

// Launch_Lsp_Server implements lspclient.
func (l lsp_py) Launch_Lsp_Server() error {
	panic("unimplemented")
}

// Resolve implements lspclient.
// Subtle: this method shadows the method (lsp_base).Resolve of lsp_py.lsp_base.
func (l lsp_py) Resolve(sym lsp.SymbolInformation, symfile *Symbol_file) bool {
	return false
}

// InitializeLsp implements lspclient.
func (l lsp_py) InitializeLsp(wk WorkSpace) error {
	return nil
}

func (l lsp_py) CallHierarchyIncomingCalls(param lsp.CallHierarchyItem) ([]lsp.CallHierarchyIncomingCall, error) {
	return lsp_base.CallHierarchyIncomingCalls(l.lsp_base, param)
}

// DidOpen implements lspclient.
// Subtle: this method shadows the method (lsp_base).DidOpen of lsp_py.lsp_base.
func (l lsp_py) DidOpen(file string) error {
	return lsp_base.DidOpen(l.lsp_base, file)
}

// Close
func (l lsp_py) Close() {}

// GetDeclare implements lspclient.
// Subtle: this method shadows the method (lsp_base).GetDeclare of lsp_py.lsp_base.
func (l lsp_py) GetDeclare(file string, pos lsp.Position) ([]lsp.Location, error) {
	return lsp_base.GetDeclare(l.lsp_base, file, pos)
}

// GetDeclareByLocation implements lspclient.
// Subtle: this method shadows the method (lsp_base).GetDeclareByLocation of lsp_py.lsp_base.
func (l lsp_py) GetDeclareByLocation(loc lsp.Location) ([]lsp.Location, error) {
	return lsp_base.GetDeclareByLocation(l.lsp_base, loc)
}

// GetDocumentSymbol implements lspclient.
// Subtle: this method shadows the method (lsp_base).GetDocumentSymbol of lsp_py.lsp_base.
func (l lsp_py) GetDocumentSymbol(file string) (*document_symbol, error) {
	return lsp_base.GetDocumentSymbol(l.lsp_base, file)
}

// GetReferences implements lspclient.
// Subtle: this method shadows the method (lsp_base).GetReferences of lsp_py.lsp_base.
func (l lsp_py) GetReferences(file string, pos lsp.Position) ([]lsp.Location, error) {
	return lsp_base.GetReferences(l.lsp_base, file, pos)
}

// InitializeLsp implements lspclient.

// IsMe implements lspclient.
// Subtle: this method shadows the method (lsp_base).IsMe of lsp_py.lsp_base.
func (l lsp_py) IsMe(filename string) bool {
	return lsp_base.IsMe(l.lsp_base, filename)
}

// PrepareCallHierarchy implements lspclient.
// Subtle: this method shadows the method (lsp_base).PrepareCallHierarchy of lsp_py.lsp_base.
func (l lsp_py) PrepareCallHierarchy(loc lsp.Location) ([]lsp.CallHierarchyItem, error) {
	return lsp_base.PrepareCallHierarchy(l.lsp_base, loc)
}
