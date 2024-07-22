package lspcore

import (
	// "context"
	// "fmt"
	"log"
	"os/exec"
	"testing"

	"github.com/tectiv3/go-lsp"
	// "github.com/tectiv3/go-lsp"
)

var path = "/home/z/dev/lsp/pylspclient/tests/cpp/"

// var d_cpp = "/home/z/dev/lsp/pylspclient/tests/cpp/d.cpp"
var d_cpp = "/home/z/dev/lsp/pylspclient/tests/cpp/test_main.cpp"
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
	symbol, err := client.GetDocumentSymbol(d_cpp)
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range symbol.SymbolInformation {
		t.Log(v.Name)
	}

	data, err := client.GetReferences(d_cpp, lsp.Position{
		Line:      12,
		Character: 7,
	})
	print(data, err)
	if len(data) == 0 {
		t.Fatalf("fail to get reference")
	}
	var call_in_range = lsp.Range{}
	call_in_range.Start = lsp.Position{

		Line:      12,
		Character: 7,
	}
	call_in_range.End = call_in_range.Start
	call_preare_item, err := client.TextDocumentPrepareCallHierarchy(lsp.Location{
		URI:   lsp.NewDocumentURI(d_cpp),
		Range: call_in_range,
	})
	if len(call_preare_item) == 0 || err != nil {
		t.Fatalf("fail to call_prepare")
	}
	callin, err := client.CallHierarchyIncomingCalls(call_preare_item[0])
	if len(callin) == 0 || err != nil {
		t.Fatalf("fail to call in")
	}

	for _, v := range data {
		var a = LocationContent{
			location: v,
		}
		code, _ := a.Text()
		t.Logf("!!! REFERENCE >%s<\n", code)
		client.GetDeclare(a.Path(), lsp.Position{Line: v.Range.Start.Line, Character: v.Range.Start.Character})
	}
	// for _, v := range symbol.SymbolInformation{
	// 	uri := v.Location.URI.String()
	// 	index := strings.Index(uri,"file://")
	// 	var path = uri[index+7:]
	// 	var loc =LocationContent{location: v.Location}
	// 	ttt,err := loc.Text()
	// 	print(ttt)
	// 	data, err := client.GetReferences(path, lsp.Position{
	// 		Line:      v.Location.Range.Start.Line,
	// 		Character: v.Location.Range.Start.Character,
	// 	})
	// 	if err != nil {
	// 		print(err)
	// 	}
	// 	print(data)
	// 	decl, err := client.GetDeclare(d_cpp, lsp.Position{
	// 		Line:      v.Location.Range.Start.Line,
	// 		Character: v.Location.Range.Start.Character,
	// 	})
	// 	if err != nil {
	// 		print(err)
	// 	}
	// 	print(decl)
	// }
	t.Log(symbol)
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
