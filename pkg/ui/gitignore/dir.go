package gitignore

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"zen108.com/lspvi/pkg/debug"
	// "github.com/wellsjo/SuperSearch/src/logger"
)

const (
	commentPrefix = "#"
	coreSection   = "core"
	eol           = "\n"
	excludesfile  = "excludesfile"
	gitDir        = ".git"
	gitignoreFile = ".gitignore"
	gitconfigFile = ".gitconfig"
	systemFile    = "/etc/gitconfig"
)

// ReadIgnoreFile reads a specific git ignore file.
// func ReadIgnoreFile(path []string, ignoreFile string) (ps []Pattern, err error) {
func ReadIgnoreFile(p string) (ps []Pattern, err error) {
	debug.DebugLogf("gitignore","Loading ignore file %v", p)
	parts := strings.Split(p, string(filepath.Separator))
	path := parts[1 : len(parts)-1]

	f, err := os.Open(p)
	if err == nil {
		defer f.Close()

		if data, err := ioutil.ReadAll(f); err == nil {
			x := strings.Split( string(data),eol)
			x = append([]string{gitDir, gitignoreFile}, x...)
			for _, s := range x {
				debug.TraceLogf("gitignore","gitignore processing %v", s)
				if !strings.HasPrefix(s, commentPrefix) && len(strings.TrimSpace(s)) > 0 {
					ps = append(ps, ParsePattern(s, path))
				}
			}
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	return
}
