package lspcore

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

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
}
type lsp_base struct {
	core *lspcore
	wk   workroot
}

type lsp_cpp struct {
	lsp_base
}
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
func (l lsp_cpp) InitializeLsp(wk workroot) error {
	result, err := l.core.Initialize(wk)
	if err != nil {
		return err
	}
	if result.ServerInfo.Name == "clangd" {
		return nil
	}
	return fmt.Errorf("%s", result.ServerInfo.Name)
}

// Launch_Lsp_Server implements lspclient.
func (l lsp_cpp) Launch_Lsp_Server() error {
	l.core.cmd = exec.Command("clangd")
	return l.core.Lauch_Lsp_Server(l.core.cmd)
}
func (l lsp_base) DidOpen(file string) error {
	return l.core.DidOpen(file)
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
func (core *lspcore) GetDocumentSymbol(file string) error {
	return nil
}
func main2() {
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
