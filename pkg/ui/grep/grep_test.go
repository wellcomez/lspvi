package grep

import (
	"os"
	"path/filepath"
	"testing"

	str "github.com/boyter/go-string"
	gi "zen108.com/lspvi/pkg/ui/gitignore"
)

func TestGoString(t *testing.T) {
	if _, err := os.ReadFile("/home/z/dev/lsp/goui/pkg/ui/gogrepimpl.go"); err == nil {
		ret := str.IndexAll("err123456\nerr", "err", -1)
		for _, v := range ret {
			t.Log(v)
		}
	}
}
func TestGitIgnore(t *testing.T) {
	home, _ := os.UserHomeDir()
	root := "/home/z/dev/lsp/goui"
	// .gitignore"
	ps, _ := gi.ReadIgnoreFile(filepath.Join(home, ".gitignore_global"))
	m := gi.NewMatcher(ps)
	ignorepath := filepath.Join(root, ".gitignore")
	ps2, _ := gi.ReadIgnoreFile(ignorepath)
	if len(ps2) > 0 {
		m.AddPatterns(ps2)
	}
	ret := m.MatchFile(filepath.Join(root, "__debug_bin3307010684"))
	if ret == false {
		t.Errorf("not match")
	}
	if data,err:= os.ReadDir(root);err==nil{
		for _, v := range data {
			path:=filepath.Join(root, v.Name())		
			t.Log(path, m.MatchFile(path))
		}
	}
}