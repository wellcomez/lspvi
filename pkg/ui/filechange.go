package mainui

import (
	"github.com/fsnotify/fsnotify"
	"log"
)

type change_reciever interface {
	OnFileChange(file string) bool
}
type FileWatch struct {
	watcher  *fsnotify.Watcher
	root     string
	started  bool
	recieved []change_reciever
}

func NewFileWatch() *FileWatch {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil
	}
	return &FileWatch{watcher: watcher}
}

func (f *FileWatch) Change(s string) error {
	if err := f.watcher.Add(s); err != nil {
		return err
	}
	if len(f.root) > 0 {
		if err := f.watcher.Remove(f.root); err != nil {
			return err
		}
	}
	f.root = s
	return nil
}
func (f *FileWatch) Run() error {
	if f.started {
		return nil
	}
	// 创建一个 watcher。
	watcher := f.watcher
	defer watcher.Close()

	// 处理事件。
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Printf("Watched file %s", event.Name)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Printf("modified file: %s", event.Name)
					for _, v := range f.recieved {
						if v.OnFileChange(event.Name) {
							break
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()
	err := watcher.Add(f.root)
	if err != nil {
		return err
	}
	f.started = true
	<-make(chan struct{})
	return nil
}
