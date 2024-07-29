package mainui

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type dir_walk struct {
	wg            sync.WaitGroup
	resultChannel chan string
}

func NewDirWalk() *dir_walk {
	return &dir_walk{
		resultChannel: make(chan string),
	}
}

// 定义一个channel来接收遍历结果
// var resultChannel = make(chan string)

// 用于等待所有goroutine完成
// var wg sync.WaitGroup

// 递归遍历目录的函数
func (wk *dir_walk) traverseDir(dirPath string) {
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Error accessing", path, ":", err)
			return nil
		}

		if info.IsDir() {
			// 如果是目录，递归调用
			wk.wg.Add(1)
			go func(subDir string) {
				defer wk.wg.Done()
				wk.traverseDir(subDir)
			}(path)
		} else {
			// 如果是文件，发送到channel
			wk.resultChannel <- path
		}

		return nil
	})

	if err != nil {
		fmt.Println("Error walking the path:", err)
	}
}

func (wk *dir_walk) Run(dir string) {
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
