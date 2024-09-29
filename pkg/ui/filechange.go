package mainui

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

type change_reciever interface {
	OnFileChange(file string) bool
}
type FileWatch struct {
	watcher  *fsnotify.Watcher
	started  bool
	recieved []change_reciever
	files    []string
}

func (f *FileWatch) AddReciever(reciever change_reciever) {
	f.recieved = append(f.recieved, reciever)
}
func (f *FileWatch) Remove(reciever change_reciever) {
	recieved := []change_reciever{}
	for _, v := range f.recieved {
		if v == reciever {
			continue
		}
		recieved = append(recieved, v)
	}
	f.recieved = recieved
}

func NewFileWatch() *FileWatch {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil
	}
	return &FileWatch{watcher: watcher}
}

func (f *FileWatch) Add(s string) error {
	if err := f.watcher.Add(s); err != nil {
		return err
	}
	f.files = append(f.files, s)
	return nil
}

func (f *FileWatch) Run(root string) error {
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
	if err := f.Add(root); err == nil {
		f.started = true
	} else {
		return err
	}
	<-make(chan struct{})
	return nil
}
