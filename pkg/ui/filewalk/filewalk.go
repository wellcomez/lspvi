package filewalk

import (
	// "bufio"
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/charlievieth/fastwalk"
	"zen108.com/lspvi/pkg/debug"
	gi "zen108.com/lspvi/pkg/ui/gitignore"
)

type Filewalk struct {
	waitReports    sync.WaitGroup
	Filelist       []string
	Root           string
	filereciver    chan string
	end            chan bool
	filecount      chan int
	resultfile     string
	use_git_ignore bool
	skipcount      int
}

func (f *Filewalk) Load() error {
	fp, err := os.OpenFile(f.resultfile, os.O_RDONLY, 0666)
	if err == nil {
		defer fp.Close()
		scanner := bufio.NewScanner(fp)
		for scanner.Scan() {
			f.Filelist = append(f.Filelist, scanner.Text())
		}
		return nil
	}
	return err
}
func (f *Filewalk) Save() error {
	fp, err := os.OpenFile(f.resultfile, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer fp.Close()
	data := strings.Join(f.Filelist, "\n")
	_, err = fp.WriteString(data)
	return err
}

func NewFilewalk(root string) *Filewalk {
	ret := &Filewalk{
		Filelist:       []string{},
		Root:           root,
		filereciver:    make(chan string, 10),
		waitReports:    sync.WaitGroup{},
		end:            make(chan bool),
		filecount:      make(chan int),
		use_git_ignore: true,
	}
	return ret
}

func (r *Filewalk) Walk() {
	var exit = make(chan bool)
	var total = 0
	go func() {
		for {
			select {
			case c := <-r.filecount:
				total += c
			case s := <-r.filereciver:
				// println(s)
				r.Filelist = append(r.Filelist, s)
			case <-r.end:
				debug.InfoLog("Filewalk", "report end")
				exit <- true
				return
			}
		}
	}()
	r.waitReports.Add(1)
	go r.walk(r.Root)
	r.waitReports.Wait()
	r.end <- true
	<-exit
	debug.InfoLog("Filewalk", "save", total, "Filtered", r.skipcount)
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
	count := 0
	debug.InfoLog("Filewalk", "START", root)
	defer func() {
		debug.InfoLog("Filewalk", "END", root, count)
	}()
	home, _ := os.UserHomeDir()
	ps, _ := gi.ReadIgnoreFile(filepath.Join(home, ".gitignore_global"))
	matcher := gi.NewMatcher(ps,true)
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
			debug.ErrorLog("Filewalk", "Error ", err, path)
			return err
		}
		skip := false
		if r.use_git_ignore {
			skip = matcher.MatchFile(path, de.IsDir())
		}
		if de.IsDir() {
			if filepath.Base(path)[0] == '.' {
				skip = true
			}
		}
		if skip {
			r.skipcount++
			if de.IsDir() {
				return fastwalk.SkipDir
			} else {
				return nil
			}
		}
		if is_git_root(path) {
			println(path)
			r.waitReports.Add(1)
			go r.walk(path)
			return fastwalk.SkipDir
		}
		r.filereciver <- path
		count++
		return nil
	})
	r.filecount <- count
	r.waitReports.Done()
}
