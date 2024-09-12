package lspcore

import (
	"testing"

	treec "github.com/smacker/go-tree-sitter/c"
	treego "github.com/smacker/go-tree-sitter/golang"
)

func Test_Go(t *testing.T) {
	// filename := "/home/z/dev/lsp/goui/pkg/lsp/lsp_go.go"
	filename := "/home/z/dev/lsp/goui/main.go"
	ts := NewTreeSitter(filename)
	ts.tsname = "go"
	// /home/z/dev/lsp/goui/pkg/lsp/queries/go/highlights.scm
	ts.Loadfile(treego.GetLanguage())
}
func Test_C(t *testing.T) {
	filename := "/home/z/dev/lsp/goui/pkg/lsp/tests/cpp/d.h"
	ts := NewTreeSitter(filename)
	ts.tsname = "c"
	ts.Loadfile(treec.GetLanguage())
}
