package gitignore

import (
	"path/filepath"
	"strings"
)

// Matcher defines a global multi-pattern matcher for gitignore patterns
type Matcher interface {
	// Match matches patterns in the order of priorities. As soon as an inclusion or
	// exclusion is found, not further matching is performed.
	Match(path []string, isDir bool) bool

	AddPatterns(ps []Pattern)
	Patterns() []Pattern
	MatchFile(string,bool) bool
	Enter(dir string)
}

// NewMatcher constructs a new global matcher. Patterns must be given in the order of
// increasing priority. That is most generic settings files first, then the content of
// the repo .gitignore, then content of .gitignore down the path or the repo and then
// the content command line arguments.
func NewMatcher(ps []Pattern) Matcher {
	return &matcher{ps}
}

type matcher struct {
	patterns []Pattern
}

func (m *matcher) Enter(dir string) {
	ps, _ := ReadIgnoreFile(filepath.Join(dir, ".gitignore"))
	if len(ps) > 0 {
		m.AddPatterns(ps)
	}
}
func (m *matcher) MatchFile(filepath string,isdir bool) bool {
	ss := strings.Split(filepath, "/")
	return m.Match(ss[1:], isdir)
}
func (m *matcher) Match(path []string, isDir bool) bool {
	n := len(m.patterns)
	for i := n - 1; i >= 0; i-- {
		if match := m.patterns[i].Match(path, isDir); match > NoMatch {
			return match == Exclude
		}
	}
	return false
}

func (m *matcher) AddPatterns(ps []Pattern) {
	m.patterns = append(m.patterns, ps...)
}

func (m *matcher) Patterns() []Pattern {
	return m.patterns
}
