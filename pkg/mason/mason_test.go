package mason_test

import (
	"testing"

	"zen108.com/lspvi/pkg/mason"
)

func TestClang(t *testing.T) {
	_, err := mason.Load(nil, "/home/z/dev/lsp/goui/pkg/mason/config/clangd/package.yaml")
	if err != nil {
		t.Fatal(err)
	}
}
func TestGo(t *testing.T) {
	_, err := mason.Load(nil, "/home/z/dev/lsp/goui/pkg/mason/config/go/package.yaml")
	if err != nil {
		t.Fatal(err)
	}
}
func TestTs(t *testing.T) {
	_, err := mason.Load(nil, "/home/z/dev/lsp/goui/pkg/mason/config/typescript-language-server/package.yaml")
	if err != nil {
		t.Fatal(err)
	}
}
func TestPy(t *testing.T) {
	_, err := mason.Load(nil, "/home/z/dev/lsp/goui/pkg/mason/config/python-lsp-server/package.yaml")
	if err != nil {
		t.Fatal(err)
	}
}
func TestRust(t *testing.T) {
	_, err := mason.Load(nil, "/home/z/dev/lsp/goui/pkg/mason/config/rust-analyzer/package.yaml")
	if err != nil {
		t.Fatal(err)
	}
}
func TestJava(t *testing.T) {
	_, err := mason.Load(nil, "/home/z/dev/lsp/goui/pkg/mason/config/java-language-server/package.yaml")
	if err != nil {
		t.Fatal(err)
	}
}
//
