package lspcore

import (
	"path/filepath"
	"strings"

	"github.com/tectiv3/go-lsp"
	"zen108.com/lspvi/pkg/debug"
)

type lspclient interface {
	WorkSpaceSymbol(query string) ([]lsp.SymbolInformation, error)
	WorkspaceDidChangeWatchedFiles(Changes []lsp.FileEvent) error
	Semantictokens_full(file string) (*lsp.SemanticTokens, error)
	InitializeLsp(wk WorkSpace) error
	IsReady() bool
	Format(opt FormatOption) ([]lsp.TextEdit, error)
	Launch_Lsp_Server() error
	DidOpen(file SourceCode, version int) error
	DidComplete(param Complete) (lsp.CompletionList, error)
	CompletionItemResolve(param *lsp.CompletionItem) (*lsp.CompletionItem, error)
	IsTrigger(param string) (TriggerChar, error)
	SignatureHelp(arg SignatureHelp) (lsp.SignatureHelp, error)
	DidClose(file string) error
	DidSave(file string, text string) error
	DidChange(file string, verion int, ContentChanges []lsp.TextDocumentContentChangeEvent) error
	GetDocumentSymbol(file string) (*document_symbol, error)
	GetReferences(file string, pos lsp.Position) ([]lsp.Location, error)
	GetImplement(file string, pos lsp.Position) (ImplementationResult, error)
	GetDeclareByLocation(loc lsp.Location) ([]lsp.Location, error)
	GetDeclare(file string, pos lsp.Position) ([]lsp.Location, error)
	GetDefine(file string, pos lsp.Position) ([]lsp.Location, error)
	PrepareCallHierarchy(loc lsp.Location) ([]lsp.CallHierarchyItem, error)
	CallHierarchyOutcomingCalls(param lsp.CallHierarchyItem) ([]lsp.CallHierarchyOutgoingCall, error)
	CallHierarchyIncomingCalls(param lsp.CallHierarchyItem) ([]lsp.CallHierarchyIncomingCall, error)
	IsMe(filename string) bool
	IsSource(filename string) bool
	Resolve(sym lsp.SymbolInformation, symbolfile *Symbol_file) bool
	Close()
	syncOption() *TextDocumentSyncOptions

	HelpTrigger() LspSignatureHelp
	CompleteTrigger() LspCompleteUtil
}
type lsp_base struct {
	core *lspcore
	wk   *WorkSpace
}

func (l lsp_base) HelpTrigger() LspSignatureHelp {
	return LspSignatureHelp{}
}
func (l lsp_base) CompleteTrigger() LspCompleteUtil {
	return LspCompleteUtil{}
}
func (l lsp_base) Format(opt FormatOption) ([]lsp.TextEdit, error) {
	return l.core.TextDocumentFormatting(opt)
}
func (l lsp_base) syncOption() *TextDocumentSyncOptions {
	return l.core.sync
}
func (l lsp_base) Semantictokens_full(file string) (*lsp.SemanticTokens, error) {
	return l.core.document_semantictokens_full(file)
}
func (l lsp_base) IsReady() bool {
	return l.core.inited
}
func (l lsp_base) InitializeLsp(wk WorkSpace) (err error) {
	l.core.lock.Lock()
	defer l.core.lock.Unlock()
	if !l.core.inited {
		err = l.core.lang.InitializeLsp(l.core, wk)
		if err == nil {
			if err = l.core.Initialized(); err == nil {
				l.core.Progress_notify()
				debug.InfoLog(DebugTag, "InitializeLsp OK: ", l.core.lang)
				return
			}
		}
		debug.ErrorLog(DebugTag, "InitializeLsp failed: ", err)
	}
	return
}
func (l lsp_base) Launch_Lsp_Server() error {
	l.core.lock.Lock()
	defer l.core.lock.Unlock()
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
	if l.core.cmd == nil {
		return
	}
	if l.core.cmd.Process != nil {
		l.core.cmd.Process.Kill()
	}
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
func (l lsp_base) DidClose(file string) error {
	return l.core.DidClose(file)
}

type SourceCode struct {
	Path   string
	Cotent string
}

func (l lsp_base) SignatureHelp(arg SignatureHelp) (lsp.SignatureHelp, error) {
	return l.core.SignatureHelp(arg)
}
func (l lsp_base) CompletionItemResolve(param *lsp.CompletionItem) (*lsp.CompletionItem, error) {
	return l.core.CompletionItemResolve(param)
}
func (l lsp_base) IsTrigger(param string) (TriggerChar, error) {
	return l.core.IsTrigger(param)
}
func (l lsp_base) DidComplete(param Complete) (lsp.CompletionList, error) {
	var cb = param.CompleteHelpCallback
	if cb != nil {
		param.CompleteHelpCallback = func(cl lsp.CompletionList, c Complete, err error) {
			l.core.lang.CompleteHelpCallback(cl, &c, err)
			cb(cl, c, err)
		}
	}
	return l.core.DidComplete(param)
}

// Initialize implements lspclient.
func (l lsp_base) DidOpen(file SourceCode, version int) error {
	return l.core.DidOpen(file, version)
}
func (l lsp_base) DidSave(file, text string) error {
	return l.core.DidSave(file, text)
}
func (l lsp_base) DidChange(file string, verion int, ContentChanges []lsp.TextDocumentContentChangeEvent) error {
	return l.core.DidChange(file, verion, ContentChanges)
}

func (l lsp_base) PrepareCallHierarchy(loc lsp.Location) ([]lsp.CallHierarchyItem, error) {
	return l.core.TextDocumentPrepareCallHierarchy(loc)
}
func (l lsp_base) GetDefine(file string, pos lsp.Position) ([]lsp.Location, error) {
	ret, err := l.core.GetDefine(file, pos)
	if err != nil {
		debug.ErrorLog(DebugTag, file, err)
	}
	return ret, err
}
func (l lsp_base) CallHierarchyOutcomingCalls(param lsp.CallHierarchyItem) ([]lsp.CallHierarchyOutgoingCall, error) {
	return l.core.CallHierarchyOutgoingCalls(param)
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

	ret, err := l.core.GetDeclare(file, pos)
	if err != nil {
		debug.ErrorLog(DebugTag, file, err)
	}
	return ret, err
}
func (l lsp_base) GetImplement(file string, pos lsp.Position) (ImplementationResult, error) {
	ret, err := l.core.GetImplement(file, pos)
	if err != nil {
		debug.ErrorLog(DebugTag, file, err)
	}
	return ret, err
}
func (l lsp_base) GetReferences(file string, pos lsp.Position) ([]lsp.Location, error) {
	ret, err := l.core.GetReferences(file, pos)
	if err != nil {
		debug.ErrorLog(DebugTag, file, err)
	}
	return ret, err
}
func (l lsp_base) GetDocumentSymbol(file string) (*document_symbol, error) {
	ret, err := l.core.GetDocumentSymbol(file)
	if err != nil {
		debug.ErrorLog(DebugTag, file, err)
	}
	return ret, err
}

func (l lsp_base) WorkSpaceSymbol(query string) ([]lsp.SymbolInformation, error) {
	return l.core.WorkSpaceDocumentSymbol(query)
}

func (l lsp_base) WorkspaceDidChangeWatchedFiles(Changes []lsp.FileEvent) error {
	return l.core.WorkspaceDidChangeWatchedFiles(Changes)
}
