package lspcore

import (
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/tectiv3/go-lsp"
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
	handle rpchandle
	rw     io.ReadWriteCloser
}
type lspclient interface {
	InitializeLsp(wk workroot) error
	Launch_Lsp_Server() error
}
type lsp_cpp struct {
	core *lspcore
	wk   workroot
}
type lsp_py struct {
	core *lspcore
	wk   workroot
}

func new_lsp_cpp(wk workroot) *lsp_cpp {
	return &lsp_cpp{
		core: &lspcore{},
		wk:   wk,
	}
}
func new_lsp_py(wk workroot) *lsp_py {
	return &lsp_py{
		core: &lspcore{},
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
