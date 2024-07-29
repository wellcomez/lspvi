package mainui

import (
	"github.com/sabhiram/go-gitignore"
	"path/filepath"
)

type gitignore struct {
	check *ignore.GitIgnore
	root  string
}

func NewGitIgnore(root string) *gitignore {
	a := &gitignore{root: root}
	a.init()
	return a
}
func (t *gitignore) init() {
	f := filepath.Join(t.root, ".gitignore")
	check, err := ignore.CompileIgnoreFileAndLines(f,".git")
	if err == nil {
		t.check = check
	}
}
func (t gitignore) Ignore(filename string) bool {
	if t.check != nil {
		return t.check.MatchesPath(filename)
	}
	return false
}
