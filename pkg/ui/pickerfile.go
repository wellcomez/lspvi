package mainui

import (
	"github.com/reinhrst/fzf-lib"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

type DirWalk struct {
	query    string
	cur      *querytask
	root     string
	cb       func(t querytask)
	hayStack []string
}
type querytask struct {
	// filename string
	query string
	ret   []file_picker_item
	done  bool
}
type file_picker_item struct {
	name string
	path string
}

func NewDirWalk(root string, cb func(t querytask)) *DirWalk {

	var hayStack = walk(root)
	return &DirWalk{root: root, cb: cb, hayStack: hayStack}
}
func (wk *DirWalk) UpdateQueryOld(query string) {
	cur := wk.cur
	r := cur == nil || !strings.Contains(query, cur.query)
	if cur != nil && cur.done {
		r = true
		cur.ret = []file_picker_item{}
	}
	wk.query = query
	if r {
		wk.cur = &querytask{
			query: query,
			ret:   []file_picker_item{},
		}
	}
	if r {
		go wk.Run()
	}
}

func (wk *DirWalk) UpdateQuery(query string) {
	wk.query = query
	wk.cur = &querytask{
		query: query,
		ret:   []file_picker_item{},
	}
	go wk.Run()
}

func (wk *DirWalk) asyncWalk(task *querytask, root string) {
	var options = fzf.DefaultOptions()
	// update any options here
	// var hayStack = walk(root)
	var myFzf = fzf.New(wk.hayStack, options)
	var result fzf.SearchResult
	myFzf.Search(task.query)
	result = <-myFzf.GetResultChannel()
	for _, v := range result.Matches {
		task.ret = append(task.ret, file_picker_item{
			name: strings.ReplaceAll(v.Key, root, ""),
			path: v.Key,
		})
	}
	var t querytask
	var check = NewGitIgnore(root)
	for _, v := range wk.cur.ret {
		if !check.Ignore(v.path) {
			t.ret = append(t.ret, v)
		}
	}
	wk.cb(t)
}
func walk(root string) []string {
	ret := []string{}
	files, err := ioutil.ReadDir(root)
	if err == nil {
		for _, v := range files {
			path := filepath.Join(root, v.Name())
			if v.IsDir() {
				ret = append(ret, walk(path)...)
			} else {
				ret = append(ret, path)
			}

		}
	}
	return ret
}

func (wk *DirWalk) Run() {
	root := wk.root
	wk.asyncWalk(wk.cur, root)

	log.Printf("Run")
}
