package lspcore

import (
	"testing"

	treec "github.com/smacker/go-tree-sitter/c"
	treego "github.com/smacker/go-tree-sitter/golang"
	ts_yaml "github.com/smacker/go-tree-sitter/yaml"
)

func Test_Go(t *testing.T) {
	// filename := "/home/z/dev/lsp/goui/pkg/lsp/lsp_go.go"
	filename := "/home/z/dev/lsp/goui/main.go"
	ts := NewTreeSitter(filename,nil)
	// ts.tsname = "go"
	// /home/z/dev/lsp/goui/pkg/lsp/queries/go/highlights.scm
	ts.Loadfile(treego.GetLanguage(), func(ts *TreeSitter) {

	})
}
func Test_C(t *testing.T) {
	filename := "/home/z/dev/lsp/goui/pkg/lsp/tests/cpp/d.h"
	ts := NewTreeSitter(filename,nil)
	// ts.tsname = "c"
	ts.Loadfile(treec.GetLanguage(), func(ts *TreeSitter) {})
}
func Test_cpp_outline(t *testing.T) {
	filename := "/home/z/dev/lsp/goui/pkg/lsp/tests/cpp/d.h"
	ts := GetNewTreeSitter(filename,CodeChangeEvent{})
	ts.Init(func(ts *TreeSitter) {

	})
}

func Test_rs_outline(t *testing.T) {
	filename := "/home/z/dev/gnvim/ui/src/render.rs"
	ts := GetNewTreeSitter(filename,CodeChangeEvent{})
	ts.Init(func(ts *TreeSitter) {

	})
}
func Test_yml_outline(t *testing.T) {
	filename := "/home/z/dev/lsp/goui/pkg/ui/colorscheme/dracula.yml"
	// ts := GetNewTreeSitter(filename)
	def:=new_tsdef("yaml", lsp_dummy{}, ts_yaml.GetLanguage()).set_ext([]string{"yaml", "yml"}).setparser(rs_outline)
	def.load_scm()
	ts := def.create_treesitter(filename)
	ts.Loadfile(def.tslang, func(ts *TreeSitter){

	})
	// ts.Init(func(ts *TreeSitter) {

	// })
}
