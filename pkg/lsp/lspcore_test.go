package lspcore

import (
	// "context"
	// "fmt"
	"log"
	"os/exec"
	"testing"
	// "github.com/tectiv3/go-lsp"
)

var path = "/home/z/dev/lsp/pylspclient/tests/cpp/"
var d_cpp = "/home/z/dev/lsp/pylspclient/tests/cpp/d.cpp"
var wk = workroot{path: path}

func Test_lspcore_init(t *testing.T) {
	lspcore := &lspcore{}
	cmd := exec.Command("clangd")
	err := lspcore.Lauch_Lsp_Server(cmd)
	if err != nil {
		log.Fatal(err)
	}
	resutl, err := lspcore.Initialize(wk)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(resutl)
}
func Test_lspcpp_init(t *testing.T) {
	cpp := lsp_cpp{new_lsp_base(wk)}
	var client lspclient = cpp
	err := client.Launch_Lsp_Server()
	if err != nil {
		log.Fatal(err)
	}
	err = client.InitializeLsp(wk)
	if err != nil {
		log.Fatal(err)
	}
}
func Test_lspcpp_open(t *testing.T) {
	cpp := lsp_cpp{new_lsp_base(wk)}
	var client lspclient = cpp
	err := client.Launch_Lsp_Server()
	if err != nil {
		log.Fatal(err)
	}
	err = client.InitializeLsp(wk)
	if err != nil {
		log.Fatal(err)
	}

	err = client.DidOpen(d_cpp)
	if err != nil {
		log.Fatal(err)
	}
	err = client.GetDocumentSymbol(d_cpp)
	if err != nil {
		log.Fatal(err)
	}
}

// func Test_new_client(t *testing.T) {
// 	// 启动clangd进程
// 	cmd := exec.Command("clangd", "--log=verbose")
// 	stdin, err := cmd.StdinPipe()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	stdout, err := cmd.StdoutPipe()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	if err := cmd.Start(); err != nil {
// 		log.Fatal(err)
// 	}
// 	var rw = readwriter{w: stdin, r: stdout}
// 	var handle = servhandle{}
// 	client := lsp.NewClient(rw, rw, handle, func(e error) {})
// 	var path = "/home/z/dev/lsp/pylspclient/tests"
// 	initializationOptions := map[string]interface{}{
// 		"clangdFileStatus": true,
// 	}

// 	capabilities := map[string]interface{}{
// 		"window": map[string]bool{
// 			"workDoneProgress": true,
// 		},
// 		"textDocument": map[string]interface{}{
// 			"completion": map[string]interface{}{
// 				"completionItem": map[string]bool{
// 					"commitCharactersSupport": true,
// 					"snippetSupport":          true,
// 				},
// 			},
// 		},
// 	}
// 	var ProcessID = -1
// 	client.Run()
// 	// 发送initialize请求
// 	result, err_resp, err := client.Initialize(context.Background(),
// 		&lsp.InitializeParams{
// 			ProcessID:             &ProcessID,
// 			RootURI:               lsp.NewDocumentURI(path),
// 			InitializationOptions: initializationOptions,
// 			Capabilities:          capabilities,
// 		},
// 	)

// 	fmt.Printf("%v%v", err_resp, err)
// 	fmt.Printf("clangd initialized: %+v %+v\n", result.ServerInfo.Name, result.ServerInfo.Version)

// }
