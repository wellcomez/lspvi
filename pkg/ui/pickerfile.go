package mainui

import (
	"bufio"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/charlievieth/fastwalk"
	fzflib "github.com/reinhrst/fzf-lib"
)

var WalkerSkip = []string{".git", "node_modules"}

type DirWalk struct {
	query    string
	cur      *querytask
	root     string
	cb       func(t querytask)
	hayStack []string
	fzf      *fzflib.Fzf
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
	cb      func(t querytask)
	ret     []string
	event   int32
	killed  bool
	mutex   sync.Mutex
	root    string
	ignores []string
}

func (f *filewalk) load() error {
	fp, err := os.OpenFile(f.root+".files", os.O_RDONLY, 0666)
	if err == nil {
		defer fp.Close()
		scanner := bufio.NewScanner(fp)
		for scanner.Scan() {
			f.ret = append(f.ret, scanner.Text())
		}
		return nil
	}
	return err
}
func (f *filewalk) save() error {
	fp, err := os.OpenFile(filepath.Join(f.root, ".files"), os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer fp.Close()
	for _, v := range f.ret {
		fp.Write([]byte(v + "\n"))
	}
	return nil
}

var global_walk *filewalk = nil

func new_filewalk(root string, cb func(t querytask)) *filewalk {
	if global_walk != nil {
		return global_walk
	}
	ret := &filewalk{
		ret:     []string{},
		root:    root,
		ignores: WalkerSkip,
		cb:      cb,
	}
	global_walk = ret
	// filegit := filepath.Join(root, ".gitignore")
	// fp, err := os.OpenFile(filegit, os.O_RDONLY, 0666)
	// if err == nil {
	// 	defer fp.Close()
	// 	scanner := bufio.NewScanner(fp)
	// 	for scanner.Scan() {
	// 		ret.ignores = append(ret.ignores, scanner.Text())
	// 	}
	// }
	global_walk.load()
	loader := &filewalk{
		ret:     []string{},
		root:    root,
		ignores: WalkerSkip,
	}
	loader_cb := func(t querytask) {
		global_walk = loader
		cb(t)
	}
	loader.cb = loader_cb
	go loader.readFiles(root)
	return ret
}
func (r *filewalk) pusher(s string) bool {
	if len(r.ret) > 1000 {
		return false
	}
	r.ret = append(r.ret, s)
	return true
}

func (r *filewalk) readFiles(root string) bool {
	opts := walkerOpts{
		file:   true,
		dir:    true,
		hidden: false,
		follow: false,
	}
	r.ret = []string{}
	// r.killed = false
	conf := fastwalk.Config{
		Follow: opts.follow,
		// Use forward slashes when running a Windows binary under WSL or MSYS
		ToSlash: fastwalk.DefaultToSlash(),
		Sort:    fastwalk.SortFilesFirst,
	}
	var ignoremap map[string]*gitignore = map[string]*gitignore{}
	ignoremap[root] = NewGitIgnore(root)
	fn := func(path string, de os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		path = trimPath(path)
		if path != "." {
			dir := path
			if !de.IsDir() {
				dir = filepath.Dir(path)
			}
			r.mutex.Lock()
			ignore := ignoremap[dir]
			if ignore == nil {
				ignore = NewGitIgnore(dir)
				if ignore.check == nil {
					d := filepath.Dir(dir)
					ignore = ignoremap[d]
				}
				ignoremap[dir] = ignore
			}

			defer r.mutex.Unlock()
			isDir := de.IsDir()
			notadd := ignore.Ignore(path)
			if isDir || opts.follow && isSymlinkToDir(path, de) {
				base := filepath.Base(path)
				if !opts.hidden && base[0] == '.' {
					return filepath.SkipDir
				}
				for _, ignore := range r.ignores {
					if ignore == base {
						return filepath.SkipDir
					}
				}
				if notadd {
					return filepath.SkipDir
				}

			}
			if notadd {
				return nil
			}
			if ((opts.file && !isDir) || (opts.dir && isDir)) && r.pusher(path) {
				atomic.StoreInt32(&r.event, int32(EvtReadNew))
				// global_walk.cb(querytask{
				// 	count:        len(r.ret),
				// 	update_count: true,
				// })
			}
		}
		// r.mutex.Lock()
		// defer r.mutex.Unlock()
		// if r.killed {
		// 	return nil
		// }
		return nil
	}
	yes := fastwalk.Walk(&conf, root, fn) == nil
	r.save()
	return yes
}

type querytask struct {
	count       int
	match_count int
	// filename string
	query        string
	ret          []file_picker_item
	done         bool
	update_count bool
}
type file_picker_item struct {
	name string
	path string
}

func NewDirWalk(root string, cb func(t querytask)) *DirWalk {

	var hayStack = walk(root, cb)
	ret := &DirWalk{root: root, cb: cb, hayStack: hayStack}
	var options = fzflib.DefaultOptions()
	// options.Fuzzy = true
	options.Sort = []fzflib.Criterion{}
	// update any options here
	// var hayStack = walk(root)
	ret.fzf = fzflib.New(ret.hayStack, options)
	return ret
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

	var t querytask
	t.count = len(wk.hayStack)
	wk.cb(t)
	var result fzflib.SearchResult
	wk.fzf.Search(task.query)
	result = <-wk.fzf.GetResultChannel()
	cout := 0
	t.match_count = len(result.Matches)
	log.Println(t.match_count, len(wk.hayStack))
	for _, v := range result.Matches {
		task.ret = append(task.ret, file_picker_item{
			name: strings.ReplaceAll(v.Key, root, ""),
			path: v.Key,
		})
		cout++
		if cout > 1000 {
			break
		}
	}
	t.ret = append(t.ret, task.ret...)
	wk.cb(t)
}
func walk(root string, cb func(t querytask)) []string {
	walk := new_filewalk(root, cb)

	return walk.ret
}

func (wk *DirWalk) Run() {
	root := wk.root
	wk.asyncWalk(wk.cur, root)

	log.Printf("Run")
}
