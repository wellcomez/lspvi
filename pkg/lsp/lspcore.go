package lspcore

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/tectiv3/go-lsp"
	"github.com/tectiv3/go-lsp/jsonrpc"
	"go.bug.st/json"
)

type rpchandle struct {
}

// Handle implements jsonrpc2.Handler.
func (r rpchandle) Handle(ctx context.Context, con *jsonrpc2.Conn, req *jsonrpc2.Request) {
	log.Println(con, req)
	// panic("unimplemented")
}

type lspcore struct {
	cmd                   *exec.Cmd
	conn                  *jsonrpc2.Conn
	capabilities          map[string]interface{}
	initializationOptions map[string]interface{}
	// arguments             []string
	handle     rpchandle
	rw         io.ReadWriteCloser
	LanguageID string
}
type lspclient interface {
	InitializeLsp(wk workroot) error
	Launch_Lsp_Server() error
	DidOpen(file string) error
	GetDocumentSymbol(file string) (*document_symbol, error)
	GetReferences(file string, pos lsp.Position) ([]lsp.Location, error)
	GetDeclareByLocation(loc lsp.Location) ([]lsp.Location, error)
	GetDeclare(file string, pos lsp.Position) ([]lsp.Location, error)
	TextDocumentPrepareCallHierarchy(loc lsp.Location) ([]lsp.CallHierarchyItem, error)
	CallHierarchyIncomingCalls(param lsp.CallHierarchyIncomingCallsParams) ([]lsp.CallHierarchyIncomingCall, error)
}
type lsp_base struct {
	core *lspcore
	wk   workroot
}



// DidOpen implements lspclient.
// Subtle: this method shadows the method (lsp_base).DidOpen of lsp_cpp.lsp_base.

type lsp_py struct {
	lsp_base
}

func new_lsp_base(wk workroot) lsp_base {
	return lsp_base{
		core: &lspcore{LanguageID: "cpp"},
		wk:   wk,
	}
}

// Initialize implements lspclient.
func (l lsp_base) DidOpen(file string) error {
	return l.core.DidOpen(file)
}

func (l lsp_base) TextDocumentPrepareCallHierarchy(loc lsp.Location) ([]lsp.CallHierarchyItem, error) {
	return l.core.TextDocumentPrepareCallHierarchy(loc)
}
func (l lsp_base) CallHierarchyIncomingCalls(param lsp.CallHierarchyIncomingCallsParams) ([]lsp.CallHierarchyIncomingCall, error) {
	return l.core.CallHierarchyIncomingCalls(param)
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
func (core *lspcore) Lauch_Lsp_Server(cmd *exec.Cmd) error {

	// cmd := exec.Command("clangd", "--log=verbose")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
		// log.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
		// log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		return err
		// log.Fatal(err)
	}
	core.cmd = cmd
	rwc := struct {
		io.Reader
		io.Writer
		io.Closer
	}{
		Reader: stdout,
		Writer: stdin,
		Closer: stdin,
	}
	core.rw = rwc
	conn := jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(rwc, jsonrpc2.VSCodeObjectCodec{}),
		core.handle,
	)
	core.conn = conn
	return nil
}

type workroot struct {
	path string
}

func (core *lspcore) Initialize(wk workroot) (lsp.InitializeResult, error) {
	var ProcessID = -1
	// 发送initialize请求
	var result lsp.InitializeResult

	if err := core.conn.Call(context.Background(), "initialize", lsp.InitializeParams{
		ProcessID:             &ProcessID,
		RootURI:               lsp.NewDocumentURI(wk.path),
		InitializationOptions: core.initializationOptions,
		Capabilities:          core.capabilities,
	}, &result); err != nil {
		return result, err
	}
	return result, nil
}
func (core *lspcore) DidOpen(file string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	err = core.conn.Notify(context.Background(), "textDocument/didOpen", lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{
			URI:        lsp.NewDocumentURI(file),
			LanguageID: core.LanguageID,
			Text:       string(content),
			Version:    0,
		},
	})
	return err
}
func (core *lspcore) GetDeclare(file string, pos lsp.Position) ([]lsp.Location, error) {
	var referenced = lsp.DeclarationParams{}
	referenced.TextDocument.URI = lsp.NewDocumentURI(file)
	referenced.Position = pos
	var result []interface{}
	var ret []lsp.Location
	err := core.conn.Call(context.Background(), "textDocument/declaration", referenced, &result)
	if err != nil {
		return ret, err
	}
	return convert_result_to_lsp_location(result)
}
func (core lspcore) TextDocumentPrepareCallHierarchy(loc lsp.Location) ([]lsp.CallHierarchyItem, error) {
	var params = lsp.CallHierarchyPrepareParams{}
	file := LocationContent{
		location: loc,
	}.Path()
	params.TextDocument = lsp.TextDocumentIdentifier{
		URI: lsp.NewDocumentURI(file),
	}
	params.Position = loc.Range.Start
	var result []lsp.CallHierarchyItem
	err := core.conn.Call(context.Background(), "textDocument/prepareCallHierarchy", params, &result)
	if err != nil {
		return result, nil
	}
	return result, nil
}
func (core *lspcore) CallHierarchyIncomingCalls(param lsp.CallHierarchyIncomingCallsParams) ([]lsp.CallHierarchyIncomingCall, error) {
	var referenced = param
	var result []lsp.CallHierarchyIncomingCall
	err := core.conn.Call(context.Background(), "callHierarchy/incomingCalls", referenced, &result)
	if err != nil {
		return result, err
	}

	// json.Unmarshal(buf, &ret)
	return result, nil
}
func (core *lspcore) GetReferences(file string, pos lsp.Position) ([]lsp.Location, error) {
	var referenced = lsp.ReferenceParams{}
	referenced.TextDocument.URI = lsp.NewDocumentURI(file)
	referenced.Position = pos
	referenced.Context = &lsp.ReferenceContext{
		IncludeDeclaration: true,
	}
	var result []interface{}
	var ret []lsp.Location
	err := core.conn.Call(context.Background(), "textDocument/references", referenced, &result)
	if err != nil {
		return ret, err
	}
	buf, err := json.Marshal(result)
	if err != nil {
		return ret, err
	}

	json.Unmarshal(buf, &ret)
	return ret, nil
}

func convert_result_to_lsp_location(result []interface{}) ([]lsp.Location, error) {
	var ret []lsp.Location
	var result_location lsp.Location
	location_decode_config, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result: &result_location,
	})
	if err != nil {
		return ret, err
	}
	for _, v := range result {
		err := location_decode_config.Decode(v)
		if err == nil {
			ret = append(ret, result_location)
		}
	}
	return ret, nil
}

type document_symbol struct {
	DocumentSymbols   []lsp.DocumentSymbol
	SymbolInformation []lsp.SymbolInformation
}
type symbol_location struct {
	location      []lsp.Location
	LocationLinkk []lsp.LocationLink
}

type LocationContent struct {
	location lsp.Location
}

func (loc LocationContent) Path() string {
	index := strings.Index(loc.location.URI.String(), "file://")
	file := loc.location.URI.String()[index+7:]
	return file
}
func (loc LocationContent) Text() (string, error) {
	file := loc.Path()
	s, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	txt := string(s)
	lines := strings.Split(txt, "\n")
	return lines[loc.location.Range.Start.Line][loc.location.Range.Start.Character:], nil
}

func (core *lspcore) TextDocumentDeclaration(file string, pos lsp.Position) (*symbol_location, error) {
	var result []interface{}
	var parameter = lsp.DeclarationParams{}
	parameter.TextDocument = lsp.TextDocumentIdentifier{URI: lsp.NewDocumentURI(file)}
	parameter.Position = pos
	err := core.conn.Call(context.Background(), "textDocument/declaration", parameter, &result)
	if err != nil {
		return nil, err
	}
	ret := symbol_location{}
	var result_link lsp.LocationLink
	link_decode, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result: &result_link,
	})
	if err != nil {
		return nil, err
	}
	var result_location lsp.Location
	location_decode_config, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result: &result_location,
	})
	if err != nil {
		return nil, err
	}

	for _, v := range result {
		err := location_decode_config.Decode(v)
		if err == nil {
			ret.location = append(ret.location, result_location)
			continue
		}
		err = link_decode.Decode(v)
		if err == nil {
			ret.LocationLinkk = append(ret.LocationLinkk, result_link)
		}
	}
	return &ret, nil
}
func (core *lspcore) GetDocumentSymbol(file string) (*document_symbol, error) {
	uri := lsp.NewDocumentURI(file)
	var parameter = lsp.DocumentSymbolParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: uri,
		},
	}
	var result []interface{}
	err := core.conn.Call(context.Background(), "textDocument/documentSymbol", parameter, &result)
	if err != nil {
		return nil, err
	}
	// var result_symbol lsp.SymbolInformation
	// symbol_decode, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
	// 	Result: &result_symbol,
	// })
	// if err != nil {
	// 	return nil, err
	// }
	// var result_document lsp.DocumentSymbol
	// doc_decode, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
	// 	Result: &result_document,
	// })

	// ret := &document_symbol{}

	// for _, v := range result {

	// 	err := symbol_decode.Decode(v)
	// 	if err == nil {
	// 		ret.SymbolInformation = append(ret.SymbolInformation, result_symbol)
	// 	}
	// 	err = doc_decode.Decode(v)
	// 	if err == nil {
	// 		ret.DocumentSymbols = append(ret.DocumentSymbols, result_document)
	// 	}
	// }
	// return ret, err
	resp, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	var documentSymbols []lsp.DocumentSymbol
	if err := json.Unmarshal(resp, &documentSymbols); err == nil {
		return &document_symbol{
			DocumentSymbols: documentSymbols,
		}, nil
	}
	var sym []lsp.SymbolInformation
	if err := json.Unmarshal(resp, &sym); err == nil {
		return &document_symbol{
			SymbolInformation: sym,
		}, nil
	}
	return nil, fmt.Errorf("not found")
}
func mainxx2() {
	// 启动clangd进程
	cmd := exec.Command("clangd", "--log=verbose")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	rwc := struct {
		io.Reader
		io.Writer
		io.Closer
	}{
		Reader: stdout,
		Writer: stdin,
		Closer: stdin,
	}
	var handle = rpchandle{}
	// 使用jsonrpc2库创建与clangd的连接
	conn := jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(rwc, jsonrpc2.VSCodeObjectCodec{}),
		handle,
	)
	var path = "/home/z/dev/lsp/pylspclient/tests"
	initializationOptions := map[string]interface{}{
		"clangdFileStatus": true,
	}

	capabilities := map[string]interface{}{
		"window": map[string]bool{
			"workDoneProgress": true,
		},
		"textDocument": map[string]interface{}{
			"completion": map[string]interface{}{
				"completionItem": map[string]bool{
					"commitCharactersSupport": true,
					"snippetSupport":          true,
				},
			},
		},
	}
	var ProcessID = -1
	// 发送initialize请求
	var result lsp.InitializeResult
	if err := conn.Call(context.Background(), "initialize", lsp.InitializeParams{
		ProcessID:             &ProcessID,
		RootURI:               lsp.NewDocumentURI(path),
		InitializationOptions: initializationOptions,
		Capabilities:          capabilities,
	}, &result); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("clangd initialized: %+v %+v\n", result.ServerInfo.Name, result.ServerInfo.Version)

}

type servhandle struct {
}

// ClientRegisterCapability implements lsp.ServerMessagesHandler.
func (s servhandle) ClientRegisterCapability(context.Context, jsonrpc.FunctionLogger, *lsp.RegistrationParams) *jsonrpc.ResponseError {
	panic("unimplemented")
}

// ClientUnregisterCapability implements lsp.ServerMessagesHandler.
func (s servhandle) ClientUnregisterCapability(context.Context, jsonrpc.FunctionLogger, *lsp.UnregistrationParams) *jsonrpc.ResponseError {
	panic("unimplemented")
}

// GetDiagnosticChannel implements lsp.ServerMessagesHandler.
func (s servhandle) GetDiagnosticChannel() chan *lsp.PublishDiagnosticsParams {
	panic("unimplemented")
}

// LogTrace implements lsp.ServerMessagesHandler.
func (s servhandle) LogTrace(jsonrpc.FunctionLogger, *lsp.LogTraceParams) {
	panic("unimplemented")
}

// Progress implements lsp.ServerMessagesHandler.
func (s servhandle) Progress(jsonrpc.FunctionLogger, *lsp.ProgressParams) {
	panic("unimplemented")
}

// TelemetryEvent implements lsp.ServerMessagesHandler.
func (s servhandle) TelemetryEvent(jsonrpc.FunctionLogger, json.RawMessage) {
	panic("unimplemented")
}

// TextDocumentPublishDiagnostics implements lsp.ServerMessagesHandler.
func (s servhandle) TextDocumentPublishDiagnostics(jsonrpc.FunctionLogger, *lsp.PublishDiagnosticsParams) {
	panic("unimplemented")
}

// WindowLogMessage implements lsp.ServerMessagesHandler.
func (s servhandle) WindowLogMessage(jsonrpc.FunctionLogger, *lsp.LogMessageParams) {
	panic("unimplemented")
}

// WindowShowDocument implements lsp.ServerMessagesHandler.
func (s servhandle) WindowShowDocument(context.Context, jsonrpc.FunctionLogger, *lsp.ShowDocumentParams) (*lsp.ShowDocumentResult, *jsonrpc.ResponseError) {
	panic("unimplemented")
}

// WindowShowMessage implements lsp.ServerMessagesHandler.
func (s servhandle) WindowShowMessage(jsonrpc.FunctionLogger, *lsp.ShowMessageParams) {
	panic("unimplemented")
}

// WindowShowMessageRequest implements lsp.ServerMessagesHandler.
func (s servhandle) WindowShowMessageRequest(context.Context, jsonrpc.FunctionLogger, *lsp.ShowMessageRequestParams) (*lsp.MessageActionItem, *jsonrpc.ResponseError) {
	panic("unimplemented")
}

// WindowWorkDoneProgressCreate implements lsp.ServerMessagesHandler.
func (s servhandle) WindowWorkDoneProgressCreate(context.Context, jsonrpc.FunctionLogger, *lsp.WorkDoneProgressCreateParams) *jsonrpc.ResponseError {
	panic("unimplemented")
}

// WorkspaceApplyEdit implements lsp.ServerMessagesHandler.
func (s servhandle) WorkspaceApplyEdit(context.Context, jsonrpc.FunctionLogger, *lsp.ApplyWorkspaceEditParams) (*lsp.ApplyWorkspaceEditResult, *jsonrpc.ResponseError) {
	panic("unimplemented")
}

// WorkspaceCodeLensRefresh implements lsp.ServerMessagesHandler.
func (s servhandle) WorkspaceCodeLensRefresh(context.Context, jsonrpc.FunctionLogger) *jsonrpc.ResponseError {
	panic("unimplemented")
}

// WorkspaceConfiguration implements lsp.ServerMessagesHandler.
func (s servhandle) WorkspaceConfiguration(context.Context, jsonrpc.FunctionLogger, *lsp.ConfigurationParams) ([]json.RawMessage, *jsonrpc.ResponseError) {
	panic("unimplemented")
}

// WorkspaceWorkspaceFolders implements lsp.ServerMessagesHandler.
func (s servhandle) WorkspaceWorkspaceFolders(context.Context, jsonrpc.FunctionLogger) ([]lsp.WorkspaceFolder, *jsonrpc.ResponseError) {
	panic("unimplemented")
}

type readwriter struct {
	w io.WriteCloser
	r io.ReadCloser
}

// Read implements io.Reader.
func (r readwriter) Read(p []byte) (n int, err error) {
	return r.r.Read(p)
}

// Write implements io.Writer.
func (r readwriter) Write(p []byte) (n int, err error) {
	return r.w.Write(p)
}
