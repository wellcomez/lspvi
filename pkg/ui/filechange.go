package mainui

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

// Set 是一个基于 map 的集合结构
type Set struct {
	items map[interface{}]bool
}

// NewSet 创建一个新的空集合
func NewSet() *Set {
	return &Set{items: make(map[interface{}]bool)}
}

// Add 向集合中添加元素
func (s *Set) Add(item interface{}) {
	s.items[item] = true
}

// Remove 从集合中移除元素
func (s *Set) Remove(item interface{}) {
	delete(s.items, item)
}

// Contains 判断集合中是否包含某个元素
func (s *Set) Contains(item interface{}) bool {
	_, exists := s.items[item]
	return exists
}

// Size 返回集合中元素的数量
func (s *Set) Size() int {
	return len(s.items)
}

// Union 计算两个集合的并集
func (s *Set) Union(other *Set) *Set {
	unionSet := NewSet()
	for item := range s.items {
		unionSet.Add(item)
	}
	for item := range other.items {
		unionSet.Add(item)
	}
	return unionSet
}

// Intersection 计算两个集合的交集
func (s *Set) Intersection(other *Set) *Set {
	intersectionSet := NewSet()
	for item := range s.items {
		if other.Contains(item) {
			intersectionSet.Add(item)
		}
	}
	return intersectionSet
}

// Difference 计算两个集合的差集
func (s *Set) Difference(other *Set) *Set {
	differenceSet := NewSet()
	for item := range s.items {
		if !other.Contains(item) {
			differenceSet.Add(item)
		}
	}
	return differenceSet
}

type change_reciever interface {
	OnFileChange(file string) bool
}
type FileWatch struct {
	watcher  *fsnotify.Watcher
	started  bool
	recieved []change_reciever
	files    *Set
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
	return &FileWatch{watcher: watcher,files: NewSet()}
}

func (f *FileWatch) Add(s string) error {
	if f.files.Contains(s) {
		return nil
	}
	if err := f.watcher.Add(s); err != nil {
		return err
	}
	f.files.Add(s)
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
