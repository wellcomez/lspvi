package mason_test

import (
	"os"
	"path/filepath"
	"testing"

	"zen108.com/lspvi/pkg/mason"
)

func TestClang(t *testing.T) {
	_, err := mason.Load(nil, root("clangd"), ".")
	if err != nil {
		t.Fatal(err)
	}
}
func TestGo(t *testing.T) {
	_, err := mason.Load(nil, root("go"), ".")
	if err != nil {
		t.Fatal(err)
	}
}
func TestTs(t *testing.T) {
	_, err := mason.Load(nil, root("typescript-language-server"), ".")
	if err != nil {
		t.Fatal(err)
	}
}
func TestPy(t *testing.T) {
	_, err := mason.Load(nil, root("python-lsp-server"), ".")
	if err != nil {
		t.Fatal(err)
	}
}
func TestRust(t *testing.T) {
	_, err := mason.Load(nil, root("rust-analyzer"), ".")
	if err != nil {
		t.Fatal(err)
	}
}
func TestJava(t *testing.T) {
	_, err := mason.Load(nil, root("java-language-server"), ".")
	if err != nil {
		t.Fatal(err)
	}
}
func TestSwift(t *testing.T) {
	file := root("sourcekit-lsp")
	app, err := mason.Load(nil, file, ".")
	if err != nil {
		t.Fatal(err)
	}
	app.Run()
}

func root(s string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "dev", "lspvi", "pkg/mason/config", s, "package.yaml")
}
