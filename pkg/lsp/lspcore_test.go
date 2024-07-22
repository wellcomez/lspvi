package lspcore

import (
	"log"
	"os/exec"
	"testing"
)

var path = "/home/z/dev/lsp/pylspclient/tests"
var wk = workroot{path: path}

func Test_lspcore_init(t *testing.T) {
	lspcore := &lspcore{}
	cmd := exec.Command("clangd")
	err := lspcore.Lauch_Lsp_Server(cmd)
	if err != nil {
		log.Fatal(err)
	}
	resutl,err := lspcore.Initialize(wk)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(resutl)
}
func Test_lspcpp_init(t *testing.T) {
	cpp := lsp_cpp{core: &lspcore{}}
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
