package filewalk

import (
	"bufio"
	"os"
	"sync"

	"github.com/charlievieth/fastwalk"
	"zen108.com/lspvi/pkg/debug"
)


type Filewalk struct {
	loadcb   func(t []string)
	filelist []string
	event    int32
	killed   bool
	mutex    sync.Mutex
	root     string
	ignores  []string
	filelist_config string 
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

func new_filewalk(root string) *Filewalk {
	ret := &Filewalk{
		filelist: []string{},
		root:     root,
	}
	// filegit := filepath.Join(root, ".gitignore")
	// fp, err := os.OpenFile(filegit, os.O_RDONLY, 0666)
	// if err == nil {
	// 	defer fp.Close()
	// 	scanner := bufio.NewScanner(fp)
	// 	for scanner.Scan() {
	// 		ret.ignores = append(ret.ignores, scanner.Text())
	// 	}
	// }
	ret.load()
	loader := &Filewalk{
		filelist: []string{},
		root:     root,
	}
	loader_cb := func(t []string) {
	}
	loader.loadcb = loader_cb
	go loader.readFiles(root)
	return ret
}
type walkerOpts struct {
	file   bool
	dir    bool
	hidden bool
	follow bool
}
func (r *Filewalk) readFiles(root string) bool {
	opts := walkerOpts{
		file:   true,
		dir:    true,
		hidden: false,
		follow: false,
	}
	r.filelist = []string{}
	// r.killed = false
	conf := fastwalk.Config{
		Follow: opts.follow,
		// Use forward slashes when running a Windows binary under WSL or MSYS
		ToSlash: fastwalk.DefaultToSlash(),
		Sort:    fastwalk.SortFilesFirst,
	}
	fn := func(path string, de os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		return nil
	}
	yes := fastwalk.Walk(&conf, root, fn) == nil
	debug.InfoLog("file count %d", len(r.filelist))
	r.loadcb(r.filelist)
	r.save()
	return yes
}

