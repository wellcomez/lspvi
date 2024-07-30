package mainui

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sabhiram/go-gitignore"
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
func (t *gitignore) init() error {
	f, err := has_gitignore_file(t.root)
	if err != nil {
		return err
	}
	check, err := ignore.CompileIgnoreFileAndLines(f, ".git", "*.o")
	if err == nil {
		t.check = check
		return nil
	}
	return err
}

func has_gitignore_file(root string) (string, error) {
	f := filepath.Join(root, ".gitignore")
	_, err := os.Stat(f)
	return f, err
}
func (t gitignore) Ignore(filename string) bool {
	if t.check != nil {
		f:=strings.ReplaceAll(filename, t.root, "")
		if t.check.MatchesPath(f){
			log.Println(f,filename)
			return true
		}
		return false 
	}
	return false
}
