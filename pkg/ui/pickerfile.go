package mainui

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type DirWalk struct {
	wg            sync.WaitGroup
	resultChannel chan string
	filename      string
	query         string
	ret           []string
}

func NewDirWalk(root string) *DirWalk {
	return &DirWalk{
		resultChannel: make(chan string),
		filename:      root,
	}
}

func (wk *DirWalk) UpdateQuery(query string) {
	var run = len(wk.query) == 0
	wk.query = query
	if run {
		go Run(wk)
	}
}

func asyncWalk(wk *DirWalk, root string, wg *sync.WaitGroup, fileChan chan string, dirChan chan string) {
	defer wg.Done()
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error walking path %s: %v\n", path, err)
			return err
		}
		if info.IsDir() {
			asyncWalk(wk, path, wg, fileChan, dirChan)
		} else {
			if strings.Contains(strings.ToLower(path), wk.query) {
				fileChan <- path // 发送文件路径到通道
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path %s: %v\n", root, err)
	}
}

func Run(walk *DirWalk) {
	go findfile(walk)
	println("Run")
}
func findfile(walk *DirWalk) {

	var wg sync.WaitGroup
	fileChan := make(chan string)
	dirChan := make(chan string)

	// 启动异步遍历
	wg.Add(1)
	go asyncWalk(walk, walk.filename, &wg, fileChan, dirChan)

	// 处理结果
	go func() {
		wg.Wait()
		close(fileChan)
		close(dirChan)
	}()

	// 从通道中读取并处理结果
	for file := range fileChan {
		fmt.Println("File:", file)
		walk.ret = append(walk.ret, file)
	}

	for dir := range dirChan {
		fmt.Println("Directory:", dir)
	}
}
