// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package gitignore

import (
	// "github.com/steakknife/bloomfilter"
	"hash/fnv"
	"path/filepath"
	"strings"
)

// BloomFilter structure
type BloomFilter struct {
	bitArray  []bool
	size      uint64
	hashCount uint64
}

// NewBloomFilter creates a new Bloom filter
func NewBloomFilter(size uint64, hashCount uint64) *BloomFilter {
	return &BloomFilter{
		bitArray:  make([]bool, size),
		size:      size,
		hashCount: hashCount,
	}
}

// hash function to generate a hash for a given string
func (b *BloomFilter) hash(data string, seed int) uint64 {
	h := fnv.New64a()
	h.Write([]byte(data))
	h.Write([]byte{byte(seed)})
	return h.Sum64() % b.size
}

// Add adds an item to the Bloom filter
func (b *BloomFilter) Add(item string) {
	for i := 0; i < int(b.hashCount); i++ {
		hash := b.hash(item, i)
		b.bitArray[hash] = true
	}
}

// Contains checks if an item is in the Bloom filter
func (b *BloomFilter) Contains(item string) bool {
	for i := 0; i < int(b.hashCount); i++ {
		hash := b.hash(item, i)
		if !b.bitArray[hash] {
			return false
		}
	}
	return true
}

// Matcher defines a global multi-pattern matcher for gitignore patterns
// type Matcher interface {
// 	// Match matches patterns in the order of priorities. As soon as an inclusion or
// 	// exclusion is found, not further matching is performed.
// 	Match(path []string, isDir bool) bool

// 	AddPatterns(ps []Pattern)
// 	Patterns() []Pattern
// 	MatchFile(string, bool) bool
// 	Enter(dir string)
// }

// NewMatcher constructs a new global matcher. Patterns must be given in the order of
// increasing priority. That is most generic settings files first, then the content of
// the repo .gitignore, then content of .gitignore down the path or the repo and then
// the content command line arguments.
func NewMatcher(ps []Pattern, bffilter bool) Matcher {
	// const probCollide = 0.0000001
	// bf, _ := bloomfilter.NewOptimal(1000, probCollide)
	var bf *BloomFilter
	if bffilter {
		bf = NewBloomFilter(100000, 3)
	}
	return Matcher{ps, bf}
}

type Matcher struct {
	patterns []Pattern
	// bf       *bloomfilter.Filter
	bf *BloomFilter
}

func (m *Matcher) Enter(dir string) {
	if d, err := EnterDir(dir); len(d) > 0 && err == nil {
		m.AddPatterns(d)
	}
}

func EnterDir(dir string) ([]Pattern, error) {
	return ReadIgnoreFile(filepath.Join(dir, ".gitignore"))
}

func (m Matcher) MatchFile(file string, isdir bool) bool {
	if m.bf != nil {
		if !m.bf.Contains(file) {
			return false
		}
	}
	ss := strings.Split(file, "/")
	return m.Match(ss[1:], isdir)
}
func (m Matcher) Match(path []string, isDir bool) bool {
	n := len(m.patterns)
	for i := n - 1; i >= 0; i-- {
		if match := m.patterns[i].Match(path, isDir); match > NoMatch {
			return match == Exclude
		}
	}
	return false
}

func (m *Matcher) AddPatterns(ps []Pattern) {
	if m.bf != nil {
		for _, v := range ps {
			for _, b := range v.BfPth() {
				m.bf.Add(b)
			}
		}
	}
	m.patterns = append(m.patterns, ps...)
}

func (m *Matcher) Patterns() []Pattern {
	return m.patterns
}
