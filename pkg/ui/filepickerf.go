package mainui

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/charlievieth/fastwalk"
	"github.com/gdamore/tcell/v2"
	"github.com/reinhrst/fzf-lib"
	fzflib "github.com/reinhrst/fzf-lib"
	"github.com/rivo/tview"
)

// new_fzf_file
func (pk *DirWalk) grid(input *tview.InputField) *tview.Grid {
	return pk.fzflist_impl.grid(input)
}

var WalkerSkip = []string{".git", "node_modules"}

type filepicker struct {
	impl *DirWalk
}

// close implements picker.
func (f filepicker) close() {
}

// name implements picker.
func (f filepicker) name() string {
	return "files"
}

// UpdateQuery implements picker.
func (f filepicker) UpdateQuery(query string) {
	f.impl.UpdateQuery(query)
}

// handle implements picker.
func (f filepicker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return f.impl.list.InputHandler()
}

type DirWalk struct {
	*fzflist_impl
	query    string
	cur      *querytask
	root     string
	cb       func(t querytask)
	hayStack []string
	fzf      *fzf.Fzf
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
	loadcb   func(t []string)
	filelist []string
	event    int32
	killed   bool
	mutex    sync.Mutex
	root     string
	ignores  []string
	list     *customlist
}

func (f *filewalk) load() error {
	fp, err := os.OpenFile(lspviroot.filelist, os.O_RDONLY, 0666)
	if err == nil {
		defer fp.Close()
		scanner := bufio.NewScanner(fp)
		for scanner.Scan() {
			f.filelist = append(f.filelist, scanner.Text())
		}
		return nil
	}
	return err
}
func (f *filewalk) save() error {
	fp, err := os.OpenFile(lspviroot.filelist, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer fp.Close()
	for _, v := range f.filelist {
		fp.Write([]byte(v + "\n"))
	}
	return nil
}

var global_walk *filewalk = nil

func new_filewalk(root string) *filewalk {
	if global_walk != nil {
		if global_walk.root == root {
			return global_walk
		}
	}
	ret := &filewalk{
		filelist: []string{},
		root:     root,
		ignores:  WalkerSkip,
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
		filelist: []string{},
		root:     root,
		ignores:  WalkerSkip,
	}
	loader_cb := func(t []string) {
		global_walk = loader
		global_walk.filelist = t
	}
	loader.loadcb = loader_cb
	go loader.readFiles(root)
	return ret
}
func (r *filewalk) pusher(s string) bool {
	if len(r.filelist) > 1000 {
		return false
	}
	r.filelist = append(r.filelist, s)
	return true
}

var wrongfile = "/chrome/buildcef/chromium/src/out/Debug/obj/chrome/gpu/gpu/chrome_content_gpu_client.o"

func (r *filewalk) readFiles(root string) bool {
	opts := walkerOpts{
		file:   true,
		dir:    true,
		hidden: false,
		follow: false,
	}
	r.filelist = []string{}
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
			if isDir || opts.follow && isSymlinkToDir(path, de) {
				notadd := ignore.Ignore(path)
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
			if !isDir {
				if path == wrongfile {
					log.Print(path)
				}
			}
			// if notadd {
			// 	return nil
			// }
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
	log.Printf("file count %d", len(r.filelist))
	r.loadcb(r.filelist)
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
	name      string
	path      string
	Positions []int
}

func NewDirWalk(root string, v *fzfmain) *DirWalk {
	impl := new_fzflist_impl(nil, v)
	var hayStack = walk(root)
	ret := &DirWalk{
		fzflist_impl: impl,
		root:         root,
		cb: func(t querytask) {
			v.app.QueueUpdate(func() {
				list := impl.list
				update_list_view(list, t, v, v.main)
				v.app.ForceDraw()
			})
		}, hayStack: hayStack,
	}
	var options = fzflib.DefaultOptions()
	options.Fuzzy = false
	options.CaseMode = fzflib.CaseIgnore
	options.Sort = []fzflib.Criterion{}
	// update any options here
	// var hayStack = walk(root)
	ret.fzf = fzflib.New(ret.hayStack, options)
	if global_walk != nil && len(global_walk.filelist) > 0 {
		task := querytask{match_count: 0, count: len(global_walk.filelist)}
		for i := 0; i < min(len(global_walk.filelist), 100); i++ {
			v := global_walk.filelist[i]
			task.ret = append(task.ret, file_picker_item{
				name: strings.TrimLeft(strings.ReplaceAll(v, root, ""), "/"),
				path: v,
			})
		}
		update_list_view(ret.list, task, v, v.main)
	}
	return ret
}

func update_list_view(list *customlist, t querytask, v *fzfmain, main MainService) {
	UpdateTitleAndColor(list.Box, fmt.Sprintf("Files %d/%d", t.match_count, t.count))
	if t.update_count {
		return
	}
	list.Clear()
	list.Key = t.query
	for i := 0; i < min(len(t.ret), 1000); i++ {
		a := t.ret[i]
		list.AddItem(a.name, "", func() {
			idx := list.GetCurrentItem()
			f := t.ret[idx]
			v.hide()
			main.OpenFileHistory(f.path, nil)
		})
	}
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

	var t = querytask{
		query: task.query,
	}
	t.count = len(wk.hayStack)
	wk.cb(t)
	if t.count == 0 {
		return
	}
	var result fzflib.SearchResult
	wk.fzf.Search(task.query)
	result = <-wk.fzf.GetResultChannel()
	cout := 0
	t.match_count = len(result.Matches)
	log.Println(t.match_count, len(wk.hayStack))
	matched := result.Matches
	if len(matched) < 1000 {
		sort.Slice(matched, func(i, j int) bool {
			return matched[i].Score > matched[j].Score
		})
	} else {
		matched = matched[:1000]
	}
	for _, v := range matched {
		// ss:=""
		// sort.Slice(v.Positions,func(i, j int) bool {
		// 	return i < j
		// })
		// for _, i := range v.Positions {
		// 	ss=ss+fmt.Sprintf("%c", v.Key[i])
		// }
		task.ret = append(task.ret, file_picker_item{
			name:      strings.TrimLeft(strings.ReplaceAll(v.Key, root, ""), "/"),
			path:      v.Key,
			Positions: v.Positions,
		})
		cout++
		if cout > 1000 {
			break
		}
	}
	t.ret = append(t.ret, task.ret...)
	wk.cb(t)
}
func walk(root string) []string {
	walk := new_filewalk(root)

	return walk.filelist
}

func (wk *DirWalk) Run() {
	root := wk.root
	wk.asyncWalk(wk.cur, root)

	log.Printf("Run")
}
