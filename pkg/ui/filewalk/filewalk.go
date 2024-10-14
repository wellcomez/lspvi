package filewalk

import (
	"bufio"
	"os"
	"path/filepath"
	"sync"

	"github.com/charlievieth/fastwalk"
	"zen108.com/lspvi/pkg/debug"
	gi "zen108.com/lspvi/pkg/ui/gitignore"
)

type Filewalk struct {
	waitReports     sync.WaitGroup
	filelist        []string
	filelist_config string
	root            string
	filereciver     chan string
	end             chan bool
}

func (f *Filewalk) load() error {
	fp, err := os.OpenFile(f.filelist_config, os.O_RDONLY, 0666)
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
func (f *Filewalk) save() error {
	fp, err := os.OpenFile(f.filelist_config, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer fp.Close()
	for _, v := range f.filelist {
		fp.Write([]byte(v + "\n"))
	}
	return nil
}

func NewFilewalk(root string) *Filewalk {
	ret := &Filewalk{
		filelist:    []string{},
		root:        root,
		filereciver: make(chan string, 10),
		waitReports: sync.WaitGroup{},
		end:         make(chan bool),
	}
	return ret
}

func (r *Filewalk) Walk() {
	r.waitReports.Add(1)
	var exit = make(chan bool)
	go func() {
		for {
			select {
			case s := <-r.filereciver:
				println(s)
				r.filelist = append(r.filelist, s)
			case <-r.end:
				debug.InfoLog("Filewalk", "report end")
				exit <- true
				return
			}
		}
	}()
	go r.walk(r.root)
	r.waitReports.Wait()
	r.end <- true
	<-exit
	debug.InfoLog("Filewalk", "save")
}
func is_git_root(path string) bool {
	fi, err := os.Stat(filepath.Join(path, ".git"))
	if err == nil {
		if fi.IsDir() {
			return true
		}
	}
	return false
}
func (r *Filewalk) walk(root string) {
	debug.InfoLog("Filewalk", "walk", root)
	home, _ := os.UserHomeDir()
	ps, _ := gi.ReadIgnoreFile(filepath.Join(home, ".gitignore_global"))
	matcher := gi.NewMatcher(ps)
	matcher.Enter(root)
	conf := fastwalk.Config{
		Follow:  true,
		ToSlash: fastwalk.DefaultToSlash(),
		Sort:    fastwalk.SortFilesFirst,
	}
	fastwalk.Walk(&conf, root, func(path string, de os.DirEntry, err error) error {
		if root == path {
			return nil
		}
		if err != nil {
			return err
		}
		if matcher.MatchFile(path) {
			debug.InfoLogf("Filewalk", "Skip %s", path)
			return fastwalk.ErrSkipFiles
		}
		if is_git_root(path) {
			r.waitReports.Add(1)
			go r.walk(path)
			return fastwalk.SkipDir
		}
		r.filereciver <- path
		return nil
	})
	r.waitReports.Done()
}
