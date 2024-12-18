package lspcore

import (
	"context"
	"errors"
	"fmt"
	"io"
	// "log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/tectiv3/go-lsp"

	// "github.com/tectiv3/go-lsp/jsonrpc"
	"zen108.com/lspvi/pkg/debug"

	// "github.com/tectiv3/go-lsp/jsonrpc"
	"go.bug.st/json"
)

type rpchandle struct {
}

// Handle implements jsonrpc2.Handler.
func (r rpchandle) Handle(ctx context.Context, con *jsonrpc2.Conn, req *jsonrpc2.Request) {
	debug.TraceLog(DebugTag, con, req)
	// panic("unimplemented")
}

type lspcore struct {
	cmd          *exec.Cmd
	conn         *jsonrpc2.Conn
	capabilities map[string]interface{}

	CapabilitiesStatus    lsp.ServerCapabilities
	initializationOptions map[string]interface{}
	// CompletionProvider              *lsp.CompletionOptions
	// SignatureHelpProvider           *lsp.SignatureHelpOptions
	// DocumentFormattingProvider      *lsp.DocumentFormattingOptions
	// DocumentRangeFormattingProvider *lsp.DocumentRangeFormattingOptions
	// arguments             []string
	handle        jsonrpc2.Handler
	rw            io.ReadWriteCloser
	LanguageID    string
	started       bool
	inited        bool
	inited_called bool

	lang lsplang
	sync *TextDocumentSyncOptions
	lock sync.Mutex

	lsp_stderr lsp_server_errorlog

	config LangConfig
}

func (core *lspcore) LspHelp() (LspUtil, error) {
	return core.lang.LspHelp(core)
}

// Handle implements jsonrpc2.Handler.
func (core *lspcore) Handle(ctx context.Context, con *jsonrpc2.Conn, req *jsonrpc2.Request) {
	debug.DebugLog(DebugTag, "lspcore.Handle", req.Method, req.Notif)
	core.handle.Handle(ctx, con, req)
}

func (core *lspcore) RunComandInConfig() bool {
	x := core.config
	if len(x.Cmd) > 0 {
		args := strings.Split(x.Cmd, " ")
		cmd := exec.Command(args[0], args[1:]...)
		core.cmd = cmd
		return true
	}
	return false
}

const DebugTag = "LSPCORE"

type lsp_server_errorlog struct {
	lsp_log LspLog
	lang    string
}

// Write implements io.Writer.
func (e lsp_server_errorlog) Write(p []byte) (n int, err error) {
	e.lsp_log.LspLogOutput(string(p), fmt.Sprintln("LspServer ", e.lang, "STDERR"))
	return len(p), nil
}

func (core *lspcore) Launch_Lsp_Server(cmd *exec.Cmd) error {

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
	cmd.Stderr = core.lsp_stderr
	if err := cmd.Start(); err != nil {
		debug.ErrorLog("failed to start", err)
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
		core,
	)
	core.conn = conn
	return nil
}

type LspLog interface {
	LspLogOutput(string, string)
}

// WorkSpace
type WorkSpace struct {
	Path         string
	Export       string
	Callback     LspLog
	NotifyHanlde lsp_notify_handle
	LspCoreList  []lspcore
	ConfigFile   string
}

func (core *lspcore) Initialized() error {
	if core.inited_called {
		return nil
	}
	core.inited_called = true
	return core.conn.Notify(context.Background(), "initialized", lsp.InitializedParams{})
	// return nil
}

type FileChangeType int
type TextChangeType int

const (
	FileChangeTypeCreated = 1
	FileChangeTypeChanged = 2
	FileChangeTypeDeleted = 3
)
const (
	TextChangeTypeInsert  = 1
	TextChangeTypeDeleted = 2
	TextChangeTypeReplace = 3
)

type TsPoint struct {
	Row    uint32
	Column uint32
}
type EditInput struct {
	StartIndex  uint32
	OldEndIndex uint32
	NewEndIndex uint32
	StartPoint  TsPoint
	OldEndPoint TsPoint
	NewEndPoint TsPoint
}
type CodeChangeEvent struct {
	Events   []TextChangeEvent
	TsEvents []EditInput
	Full     bool
	File     string
	Data     []byte
}
type TextChangeEvent struct {
	Text  string
	Type  TextChangeType
	Range lsp.Range
	Time  time.Time
}

func (core *lspcore) Initialize(wk WorkSpace) (lsp.InitializeResult, error) {
	var ProcessID = -1
	// 发送initialize请求
	var result lsp.InitializeResult
	if err := core.conn.Call(context.Background(), "initialize", lsp.InitializeParams{
		ProcessID:             &ProcessID,
		RootURI:               lsp.NewDocumentURI(wk.Path),
		InitializationOptions: core.initializationOptions,
		Capabilities:          core.capabilities,
	}, &result); err != nil {
		return result, err
	}
	return result, nil
}

func (w WorkSpace) Handle(ctx context.Context, con *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var logerr error
	if _, err := json.MarshalIndent(req, " ", " "); err == nil {
		if a, err := json.Marshal(req.Params); err == nil {
			if ret, err := notificationDispatcher(req.Method, a, w.NotifyHanlde); err == nil {
				data := fmt.Sprintf("\nMethod: %s\n Params: %s", req.Method, ret)
				w.Callback.LspLogOutput(data, "lsp-notify")
			} else {
				logerr = err
			}
		} else {
			logerr = err
		}
	} else {
		logerr = err
	}
	w.Callback.LspLogOutput(fmt.Sprint(logerr), "lsp-notify")
}
func (core *lspcore) Progress_notify() error {
	params := &lsp.ProgressParams{}
	return core.conn.Notify(context.Background(), "$/progress", params)
}

func (core *lspcore) WorkspaceDidChangeWatchedFiles(Changes []lsp.FileEvent) error {
	param := lsp.DidChangeWatchedFilesParams{
		Changes: Changes,
	}
	return core.conn.Notify(context.Background(), "workspace/didChangeWatchedFiles", param)
}
func (core *lspcore) WorkSpaceDocumentSymbol(query string) ([]lsp.SymbolInformation, error) {
	var parameter = lsp.WorkspaceSymbolParams{
		Query: query,
	}
	var res []lsp.SymbolInformation
	err := core.conn.Call(context.Background(), "workspace/symbol", parameter, &res)
	return res, err
}

func (client *lspcore) WorkspaceSemanticTokensRefresh() error {
	ctx := context.Background()
	var result interface{}
	err := client.conn.Call(ctx, "workspace/semanticTokens/refresh", NullResult, &result)
	return err
}
func (client *lspcore) SetTrace() error {
	param := &lsp.SetTraceParams{}
	return client.conn.Notify(context.Background(), "$/setTrace", param)
}
func (client *lspcore) Exit() error {
	return client.conn.Notify(context.Background(), "exit", NullResult)
}
func (client *lspcore) DidClose(file string) error {
	param := &lsp.DidCloseTextDocumentParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: lsp.NewDocumentURI(file),
		},
	}
	return client.conn.Notify(context.Background(), "textDocument/didClose", param)
}

type SignatureHelp struct {
	Pos                 lsp.Position
	File                string
	HelpCb              func(lsp.SignatureHelp, SignatureHelp, error)
	IsVisiable          bool
	TriggerCharacter    string
	Continued           bool
	ActiveSignatureHelp *lsp.SignatureHelp
	Kind                lsp.CompletionItemKind
	Code                CompleteCodeLine
}

/*func (h SignatureHelp) CreateSignatureHelp(s string) string {*/
/*re := regexp.MustCompile(`\$\{\d+:?\}`)*/
/*return re.ReplaceAllString(h.CompleteSelected, s)*/
/*}*/

type CompleteResult struct {
	Document []string
	Complete func(v lsp.CompletionItem) []string
}
type Complete struct {
	Pos                  lsp.Position
	TriggerCharacter     string
	File                 string
	CompleteHelpCallback func(lsp.CompletionList, Complete, error)
	Continued            bool
	Sym                  *Symbol_file
	Result               *CompleteResult
}
type FormatOption struct {
	Filename string
	Range    lsp.Range
	Options  lsp.FormattingOptions
	Format   func([]lsp.TextEdit, error)
}

func (client *lspcore) TextDocumentFormatting(param FormatOption) (ret []lsp.TextEdit, err error) {
	if param.Range.End != param.Range.Start && client.CapabilitiesStatus.DocumentRangeFormattingProvider != nil {
		return client.TextDocumentRangeFormatting(param)
	}
	var ret2 []lsp.TextEdit
	if err = client.conn.Call(context.Background(), "textDocument/formatting", lsp.DocumentFormattingParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: lsp.NewDocumentURI(param.Filename)},
		Options:      param.Options,
	}, &ret2); err == nil {
		ret = ret2
	}
	if param.Format != nil {
		param.Format(ret, err)
	}
	return
}
func (client *lspcore) TextDocumentRangeFormatting(opt FormatOption) (ret []lsp.TextEdit, err error) {
	var param = lsp.DocumentRangeFormattingParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: lsp.NewDocumentURI(opt.Filename),
		},
		Range:   opt.Range,
		Options: opt.Options,
	}
	var ret2 []lsp.TextEdit
	err = client.conn.Call(context.Background(), "textDocument/rangeFormatting", param, &ret2)
	if err == nil {
		ret = ret2
	}
	if opt.Format != nil {
		opt.Format(ret, err)
	}
	return
}
func (client *lspcore) CompletionItemResolve(param *lsp.CompletionItem) (*lsp.CompletionItem, error) {
	ctx := context.Background()
	var res lsp.CompletionItem
	err := client.conn.Call(ctx, "completionItem/resolve", param, &res)
	return &res, err
}

func (client *lspcore) TextDocumentHover(file string, Pos lsp.Position) (*lsp.Hover, error) {
	var res lsp.Hover
	var param = lsp.HoverParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: lsp.NewDocumentURI(file)},
			Position:     Pos,
		},
	}
	err := client.conn.Call(context.Background(), "textDocument/hover", param, &res)
	return &res, err
}

func (client *lspcore) SignatureHelp(arg SignatureHelp) (lsp.SignatureHelp, error) {
	var param = lsp.SignatureHelpParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{
				URI: lsp.NewDocumentURI(arg.File),
			},
			Position: arg.Pos,
		},
	}
	var res lsp.SignatureHelp
	TriggerKind := lsp.SignatureHelpTriggerKindInvoked
	if client.CapabilitiesStatus.SignatureHelpProvider != nil {
		cc := client.CapabilitiesStatus.SignatureHelpProvider.TriggerCharacters
		for _, v := range cc {
			if arg.TriggerCharacter == v {
				TriggerKind = lsp.SignatureHelpTriggerKindTriggerCharacter
			}
		}
		param.Context = &lsp.SignatureHelpContext{
			TriggerKind:         TriggerKind,
			IsRetrigger:         arg.IsVisiable,
			TriggerCharacter:    arg.TriggerCharacter,
			ActiveSignatureHelp: arg.ActiveSignatureHelp,
		}
	}
	err := client.conn.Call(context.Background(), "textDocument/signatureHelp", param, &res)
	if arg.HelpCb != nil {
		arg.HelpCb(res, arg, err)
	}
	return res, err
}

type TriggerCharType int

const (
	TriggerCharHelp TriggerCharType = iota
	TriggerCharComplete
)

type TriggerChar struct {
	Type TriggerCharType
	Ch   string
}

func (core *lspcore) CompleteTrigger() (ret []string) {
	if c := core.CapabilitiesStatus.CompletionProvider; c != nil {
		ret = c.TriggerCharacters
	}
	return
}
func (core *lspcore) HelpTrigger() (ret []string) {
	if c := core.CapabilitiesStatus.SignatureHelpProvider; c != nil {
		ret = c.TriggerCharacters
	}
	return
}
func (core *lspcore) IsTrigger(param string) (TriggerChar, error) {
	if c := core.CapabilitiesStatus.CompletionProvider; c != nil {
		for _, a := range c.TriggerCharacters {
			if param == a {
				return TriggerChar{TriggerCharComplete, param}, nil
			}
		}
	}
	if c := core.CapabilitiesStatus.SignatureHelpProvider; c != nil {
		for _, a := range c.TriggerCharacters {
			if param == a {
				return TriggerChar{TriggerCharHelp, param}, nil
			}
		}
	}
	return TriggerChar{}, errors.New("not found")
}
func (core *lspcore) DidComplete(param Complete) (result lsp.CompletionList, err error) {
	complete := lsp.CompletionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{
				URI: lsp.NewDocumentURI(param.File),
			},
			Position: param.Pos,
		},
	}
	if CompletionProvider := core.CapabilitiesStatus.CompletionProvider; CompletionProvider != nil {
		var context = lsp.CompletionContext{
			TriggerKind: lsp.CompletionTriggerKindInvoked,
		}
		cc := CompletionProvider.TriggerCharacters
		for _, v := range cc {
			if v == param.TriggerCharacter {
				context.TriggerKind = lsp.CompletionTriggerKindTriggerCharacter
				context.TriggerCharacter = v
				break
			}
		}
		complete.Context = &context
	}
	var sss interface{}
	err = core.conn.Call(context.Background(), "textDocument/completion", complete, &sss)
	if param.CompleteHelpCallback != nil {
		b, _ := json.Marshal(sss)
		err := json.Unmarshal(b, &result)
		param.CompleteHelpCallback(result, param, err)
	}
	return
}
func (core *lspcore) DidOpen(file SourceCode, version int) error {
	x, err := core.newTextDocument(file.Path, version, file.Cotent)
	if err != nil {
		return err
	}
	err = core.conn.Notify(context.Background(), "textDocument/didOpen", lsp.DidOpenTextDocumentParams{
		TextDocument: x,
	})
	return err
}
func (core *lspcore) DidChange(file string, verion int, ContentChanges []lsp.TextDocumentContentChangeEvent) error {
	Method := "textDocument/didChange"
	data := lsp.DidChangeTextDocumentParams{
		TextDocument: lsp.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: lsp.TextDocumentIdentifier{URI: lsp.NewDocumentURI(file)},
			Version:                verion,
		},
		ContentChanges: ContentChanges,
	}
	err := core.conn.Notify(context.Background(), Method, data)
	debug.DebugLog(DebugTag, data.TextDocument)
	return err
}
func (core *lspcore) DidSave(file string, text string) error {
	err := core.conn.Notify(context.Background(), "textDocument/didSave", lsp.DidSaveTextDocumentParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: lsp.NewDocumentURI(file)},
		Text:         text,
	})
	return err
}

func (core *lspcore) newTextDocument(file string, version int, content string) (lsp.TextDocumentItem, error) {
	if content == "" {
		if c, err := os.ReadFile(file); err != nil {
			return lsp.TextDocumentItem{}, err
		} else {
			content = string(c)
		}
	}
	x := lsp.TextDocumentItem{
		URI:        lsp.NewDocumentURI(file),
		LanguageID: core.LanguageID,
		Text:       content,
		Version:    version,
	}
	return x, nil
}
func (core *lspcore) document_semantictokens_full(file string) (result *lsp.SemanticTokens, err error) {
	params := lsp.SemanticTokensParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: lsp.NewDocumentURI(file),
		},
	}
	var r lsp.SemanticTokens
	if err = core.conn.Call(context.Background(), "textDocument/semanticTokens/full", params, &r); err == nil {
		result = &r
	}
	return
}

func (core *lspcore) GetDeclare(file string, pos lsp.Position) ([]lsp.Location, error) {
	var referenced = lsp.DeclarationParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{
				URI: lsp.NewDocumentURI(file),
			},
		},
	}
	referenced.Position = pos
	var result []interface{}
	err := core.conn.Call(context.Background(), "textDocument/declaration", referenced, &result)
	if err != nil {
		var ret []lsp.Location
		return ret, err
	}
	return convert_result_to_lsp_location(result)
}
func (core *lspcore) TextDocumentPrepareCallHierarchy(loc lsp.Location) ([]lsp.CallHierarchyItem, error) {
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
func (client *lspcore) CallHierarchyOutgoingCalls(Item lsp.CallHierarchyItem) ([]lsp.CallHierarchyOutgoingCall, error) {
	param := lsp.CallHierarchyOutgoingCallsParams{
		Item: Item,
	}
	var result []lsp.CallHierarchyOutgoingCall
	if err := client.conn.Call(context.Background(), "callHierarchy/outgoingCalls", param, &result); err == nil {
		return result, nil
	} else {
		return nil, err
	}
}

var NullResult = []byte("null")

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
func (core *lspcore) GetDefine(file string, pos lsp.Position) ([]lsp.Location, error) {
	var ret []lsp.Location
	param := lsp.DefinitionParams{}
	param.TextDocument = lsp.TextDocumentIdentifier{
		URI: lsp.NewDocumentURI(file),
	}
	param.Position = pos
	var result []interface{}
	if err := core.conn.Call(context.Background(), "textDocument/definition", param, &result); err != nil {
		return ret, err
	}
	return convert_result_to_lsp_location(result)
}
func (core *lspcore) GetReferences(file string, pos lsp.Position) ([]lsp.Location, error) {
	var referenced = lsp.ReferenceParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			Position: pos,
			TextDocument: lsp.TextDocumentIdentifier{
				URI: lsp.NewDocumentURI(file),
			},
		},
		Context: &lsp.ReferenceContext{
			IncludeDeclaration: true,
		},
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

type ImplementationResult struct {
	Loc     []lsp.Location
	LocLink []lsp.LocationLink
}

func (core *lspcore) GetImplement(file string, pos lsp.Position) (ImplementationResult, error) {
	var referenced = lsp.ImplementationParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{
				URI: lsp.NewDocumentURI(file),
			},
			Position: pos,
		},
	}
	var ret ImplementationResult
	var result []interface{}
	var method = "textDocument/implementation"
	err := core.conn.Call(context.Background(), method, referenced, &result)
	if err != nil {
		return ret, err
	}
	buf, err := json.Marshal(result)
	if err != nil {
		return ret, err
	}

	if err := json.Unmarshal(buf, &ret.Loc); err == nil {
		return ret, nil
	}
	var loc lsp.Location
	if err := json.Unmarshal(buf, &loc); err == nil {
		ret.Loc = append(ret.Loc, loc)
		return ret, nil
	}
	if err := json.Unmarshal(buf, &ret.LocLink); err == nil {
		return ret, nil
	}
	return ret, fmt.Errorf("not implemented found")
}

func convert_result_to_lsp_location(result []interface{}) ([]lsp.Location, error) {
	var ret []lsp.Location
	data, err := json.Marshal(result)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)

	if err != nil {
		var loc lsp.Location
		err = json.Unmarshal(data, &loc)
		if err != nil {
			return ret, err
		}
		return []lsp.Location{loc}, nil
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
		debug.ErrorLog(DebugTag, err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		debug.ErrorLog(DebugTag, err)
	}
	if err := cmd.Start(); err != nil {
		debug.ErrorLog(DebugTag, err)
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
		debug.ErrorLog(DebugTag, err)
		return
	}

	debug.DebugLogf("clangd initialized: %+v %+v\n", result.ServerInfo.Name, result.ServerInfo.Version)

}

type lsp_notify_handle interface {
	PublishDiagnostics(lsp.PublishDiagnosticsParams)
}

func notificationDispatcher(method string, req json.RawMessage, cb lsp_notify_handle) (ret string, err error) {
	switch method {
	case "$/progress":
		var param lsp.ProgressParams
		if err = json.Unmarshal(req, &param); err != nil {
			return
		}
		ret = fmt.Sprintf("value=%s", string(param.Value))
		// client.handler.Progress(logger, &param)
	case "$/cancelRequest":
		// should not reach here
	case "$/logTrace":
		var param lsp.LogTraceParams
		if err = json.Unmarshal(req, &param); err != nil {
			// client.errorHandler(err)
			return
		}
		ret = param.Message
		// client.handler.LogTrace(logger, &param)
	case "window/showMessage":
		var param lsp.ShowMessageParams
		if err = json.Unmarshal(req, &param); err != nil {
			// client.errorHandler(err)
			return
		}
		ret = param.Message
		// client.handler.WindowShowMessage(logger, &param)
	case "LogMessage":
		fallthrough
	case "window/logMessage":
		var param lsp.LogMessageParams
		if err = json.Unmarshal(req, &param); err != nil {
			// client.errorHandler(err)
			return
		}
		ret = param.Message
		// client.handler.WindowLogMessage(logger, &param)
	case "featureFlagsNotification":
		// params: FeatureFlags
	case "telemetry/event":
		// params: ‘object’ | ‘number’ | ‘boolean’ | ‘string’;
		// client.handler.TelemetryEvent(logger, req) // passthrough
	case "textDocument/publishDiagnostics":
		var param lsp.PublishDiagnosticsParams
		if err = json.Unmarshal(req, &param); err != nil {
			// client.errorHandler(err)
			return
		}
		s, _ := json.Marshal(param)
		ret = string(s)
		cb.PublishDiagnostics(param)
		// client.handler.TextDocumentPublishDiagnostics(logger, &param)
	default:
		// if handler, ok := client.customNotification[method]; ok {
		// 	handler(logger, req)
		// } else {
		// 	panic("unimplemented notification: " + method)
		// }
	}
	return
}
