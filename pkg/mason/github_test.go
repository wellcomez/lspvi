package mason_test

import (
	"testing"

	"zen108.com/lspvi/pkg/mason"
)

func TestClang(t *testing.T) {
	err := mason.Load("/home/z/dev/lsp/goui/pkg/mason/clangd.yaml")
	if err != nil {
		t.Fatal(err)
	}
}
func TestGo(t *testing.T) {
	err := mason.Load("/home/z/dev/lsp/goui/pkg/mason/gopls.yaml")
	if err != nil {
		t.Fatal(err)
	}
}
func TestTs(t *testing.T) {
	err := mason.Load("/home/z/dev/lsp/goui/pkg/mason/typescript-language-server/package.yaml")
	if err != nil {
		t.Fatal(err)
	}
}
func TestPy(t *testing.T) {
	err := mason.Load("/home/z/dev/lsp/goui/pkg/mason/python-lsp-server/package.yaml")
	if err != nil {
		t.Fatal(err)
	}
}
func TestRust(t *testing.T) {
	err := mason.Load("/home/z/dev/lsp/goui/pkg/mason/rust-analyzer/package.yaml")
	if err != nil {
		t.Fatal(err)
	}
}

//
