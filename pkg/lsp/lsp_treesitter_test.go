package lspcore

import (
	treego "github.com/smacker/go-tree-sitter/golang"
	treec "github.com/smacker/go-tree-sitter/c"
	"testing"
)

func Test_Go(t *testing.T) {
	// filename := "/home/z/dev/lsp/goui/pkg/lsp/lsp_go.go"
	filename:="/home/z/dev/lsp/goui/main.go"
	ts := NewTreeSitter(filename)
	ts.langname = "go"
	// /home/z/dev/lsp/goui/pkg/lsp/queries/go/highlights.scm
	ts.Loadfile(treego.GetLanguage())
}
func Test_C(t *testing.T) {
	filename :="/home/z/dev/lsp/goui/pkg/lsp/tests/cpp/d.h"
	ts := NewTreeSitter(filename)
	ts.langname = "c"
	ts.Loadfile(treec.GetLanguage())
}
