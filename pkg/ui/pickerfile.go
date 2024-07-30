package mainui

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/charlievieth/fastwalk"
	"github.com/reinhrst/fzf-lib"
)

type DirWalk struct {
	query    string
	cur      *querytask
	root     string
	cb       func(t querytask)
	hayStack []string
}

type walkerOpts struct {
	file   bool
	dir    bool
	hidden bool
	follow bool
}

func stringBytes(data string) []byte {
	return unsafe.Slice(unsafe.StringData(data), len(data))
}

func byteString(data []byte) string {
	return unsafe.String(unsafe.SliceData(data), len(data))
}

func trimPath(path string) string {
	bytes := stringBytes(path)

	for len(bytes) > 1 && bytes[0] == '.' && (bytes[1] == '/' || bytes[1] == '\\') {
		bytes = bytes[2:]
	}

	if len(bytes) == 0 {
		return "."
	}

	return byteString(bytes)
}
func isSymlinkToDir(path string, de os.DirEntry) bool {
	if de.Type()&fs.ModeSymlink == 0 {
		return false
	}
	if s, err := os.Stat(path); err == nil {
		return s.IsDir()
	}
	return false
}

type EventType int

// fzf events
const (
	EvtReadNew EventType = iota
	EvtReadFin
	EvtSearchNew
	EvtSearchProgress
	EvtSearchFin
	EvtHeader
	EvtReady
	EvtQuit
)

type filewalk struct {
	ret    []string
	event  int32
	killed bool
	mutex  sync.Mutex
}

func new_filewalk(root string) *filewalk {
	ret := &filewalk{
		ret: []string{},
	}
	return ret
}
func (r *filewalk) pusher(s string) bool {
	r.ret = append(r.ret, s)
	return true
}

var WalkerSkip = []string{".git", "node_modules"}

func (r *filewalk) readFiles(root string, ignores []string) bool {
	opts := walkerOpts{
		file:   true,
		dir:    true,
		hidden: false,
		follow: false,
	}
	// r.killed = false
	conf := fastwalk.Config{
		Follow: opts.follow,
		// Use forward slashes when running a Windows binary under WSL or MSYS
		ToSlash: fastwalk.DefaultToSlash(),
		Sort:    fastwalk.SortFilesFirst,
	}
	fn := func(path string, de os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		path = trimPath(path)
		if path != "." {
			isDir := de.IsDir()
			if isDir || opts.follow && isSymlinkToDir(path, de) {
				base := filepath.Base(path)
				if !opts.hidden && base[0] == '.' {
					return filepath.SkipDir
				}
				for _, ignore := range ignores {
					if ignore == base {
						return filepath.SkipDir
					}
				}
			}
			if ((opts.file && !isDir) || (opts.dir && isDir)) && r.pusher(path) {
				atomic.StoreInt32(&r.event, int32(EvtReadNew))
			}
		}
		// r.mutex.Lock()
		// defer r.mutex.Unlock()
		// if r.killed {
		// 	return nil
		// }
		return nil
	}
	return fastwalk.Walk(&conf, root, fn) == nil
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
	walk := new_filewalk(root)
	walk.readFiles(root, WalkerSkip)
	return walk.ret
}

func (wk *DirWalk) Run() {
	root := wk.root
	wk.asyncWalk(wk.cur, root)

	log.Printf("Run")
}
