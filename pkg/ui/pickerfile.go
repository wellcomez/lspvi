package mainui

import (
	"fmt"
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
		wk.Run(wk.filename)
	}
}

func (wk *DirWalk) traverseDir(dirPath string) {
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Error accessing", path, ":", err)
			return nil
		}
		if info.IsDir() {
			// 如果是目录，递归调用
			wk.traverseDir(path)
		} else {
			if strings.Index(strings.ToLower(path), wk.query) > 0 {
				wk.resultChannel <- path
			}
			// 如果是文件，发送到channel
		}

		return nil
	})

	if err != nil {
		fmt.Println("Error walking the path:", err)
	}
}

func (wk *DirWalk) Run(dir string) {
	// dir := "./" // 你想要遍历的目录路径

	// 开始递归遍历
	go wk.traverseDir(dir)

	// 等待所有goroutine完成
	go func() {
		wk.wg.Wait()
		close(wk.resultChannel)
	}()

	// 从channel中接收文件路径并处理
	for filePath := range wk.resultChannel {
		fmt.Println("Processed file:", filePath)
		// 这里可以添加你的文件处理逻辑
	}
}
