package mainui

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"sync"
)

type DirWalk struct {
	query string
	cur   *querytask
	root  string
	cb    func(t querytask)
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
	return &DirWalk{root: root, cb: cb}
}

func (wk *DirWalk) UpdateQuery(query string) {
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

func asyncWalk(wk *querytask, root string, wg *sync.WaitGroup, fileChan chan string, dirChan chan string) {
	// 发送文件路径到通道
	defer wg.Done()
	newFunction1(root, wk, wg, fileChan, dirChan)
}

func newFunction1(root string, wk *querytask, wg *sync.WaitGroup, fileChan chan string, dirChan chan string) {
	files, err := ioutil.ReadDir(root)
	if err == nil {
		for _, v := range files {
			path := filepath.Join(root, v.Name())
			if v.IsDir() {
				newFunction1(path, wk, wg, fileChan, dirChan)
			} else {
				if strings.Contains(strings.ToLower(path), wk.query) {
					fileChan <- path
				}

			}

		}
	}
}

func (wk *DirWalk) Run() {
	root := wk.root
	walk := wk.cur
	findfile(root, walk)
	var ret querytask
	var check =NewGitIgnore(root)
	for _, v := range wk.cur.ret {
		if !check.Ignore(v.path) {
			ret.ret = append(ret.ret, v)
		}
	}
	wk.cb(ret)
	log.Printf("Run")
}
func findfile(root string, task *querytask) {
	// task
	var wg sync.WaitGroup
	fileChan := make(chan string)
	dirChan := make(chan string)

	// 启动异步遍历
	wg.Add(1)
	go asyncWalk(task, root, &wg, fileChan, dirChan)

	// 处理结果
	go func() {
		wg.Wait()
		close(fileChan)
		close(dirChan)
	}()

	// 从通道中读取并处理结果
	for file := range fileChan {
		log.Println("File:", file)
		task.ret = append(task.ret, file_picker_item{path: file, name: strings.Replace(file, root, "", -1)})
	}

	// for dir := range dirChan {
	// 	log.Println("Directory:", dir)
	// }
	task.done = true
}
