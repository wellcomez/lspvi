package filewalk

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/charlievieth/fastwalk"
	"zen108.com/lspvi/pkg/debug"
	gi "zen108.com/lspvi/pkg/ui/gitignore"
)

type Filewalk struct {
	filelist        []string
	filelist_config string
	root            string
	filereciver     chan string
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
	}
	var end = make(chan bool)
	go func() {
		for {
			select {
			case ss := <-ret.filereciver:
				ret.filelist = append(ret.filelist, ss)
			}
			end <- true
			break
		}
	}()
	<-end
	return ret
}

func (r Filewalk) Walk(root string, m gi.Matcher) {
	if dirs, err := os.ReadDir(root); err == nil {
		m.Enter(root)
		for _, v := range dirs {
			if !m.MatchFile(filepath.Join(root, v.Name())) {
				continue
			}
			if v.IsDir() {
				go r.Walk(filepath.Join(root, v.Name()), m)
			} else {
				r.filereciver <- strings.TrimLeft( filepath.Join(root, v.Name()),r.root)
			}
		}
	}
}
