package mainui

import (
	"encoding/json"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"log"
	"os"
	"path/filepath"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type Project struct {
	Name string `json:"name"`
	Root string `json:"root"`
}
type workspace_list struct {
	Projects []Project `json:"projects"`
}

var gload_workspace_list workspace_list

func (prj *Project) Load(arg *Arguments, main *mainui) {
	root := prj.Root
	lspviroot = new_workdir(root)
	global_config, _ = LspviConfig{}.Load()
	// go servmain(lspviroot.uml, 18080, func(port int) {
	// 	httport = port
	// })

	handle := LspHandle{}
	// var main = &mainui{
	main.bf = NewBackForward(NewHistory(lspviroot.history))
	main.bookmark = &proj_bookmark{path: lspviroot.bookmark, Bookmark: []bookmarkfile{},root: root}
	main.tty = arg.Tty
	main.ws = arg.Ws
	// }
	main.bookmark.load()
	if main.bookmark_view != nil {
		main.bookmark_view.update_redraw()
	}
	handle.main = main
	if !filepath.IsAbs(root) {
		root, _ = filepath.Abs(root)
	}
	ConfigFile := lspviroot.configfile
	lspmgr := lspcore.NewLspWk(lspcore.WorkSpace{Path: root, Export: lspviroot.export, Callback: handle, ConfigFile: ConfigFile})
	main.lspmgr = lspmgr
	main.lspmgr.Handle = main
	main.root = root
}
func (wk *workspace_list) Add(root string) (*Project, error) {
	if !checkDirExists(root) {
		return nil, os.ErrNotExist
	}
	for i := range wk.Projects {
		v := wk.Projects[i]
		if v.Root == root {
			return &v, nil
		}
	}
	x := new_prj(root)
	wk.Projects = append(wk.Projects, x)
	return &x, save_workspace_list(wk)
}

func new_prj(root string) Project {
	x := Project{
		Name: filepath.Base(root),
		Root: root,
	}
	return x
}

func save_workspace_list(wk *workspace_list) error {
	buf, err := json.Marshal(wk)
	if err != nil {
		return err
	}
	if file, err := wk.get_config_file(); err == nil {
		return os.WriteFile(file, buf, 0666)
	} else {
		return err
	}
}

func (wk *workspace_list) Load() error {
	config, err := wk.get_config_file()
	if err != nil {
		return err
	}
	buf, err := os.ReadFile(config)
	if err != nil {
		return nil
	}
	return json.Unmarshal(buf, wk)
}

func (*workspace_list) get_config_file() (string, error) {
	root, err := CreateLspviRoot()
	if err != nil {
		return "", err
	}
	config := filepath.Join(root, "workspace.json")
	return config, nil
}

func CreateLspviRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	root := filepath.Join(home, ".lspvi")
	os.Mkdir(root, 0755)
	if _, err := os.Stat(root); err != nil {
		return "", err
	}
	return root, nil
}

type wk_picker_impl struct {
	*fzflist_impl
}
type workspace_picker struct {
	impl *wk_picker_impl
	fzf  *fzf_on_listview
}

// close implements picker.
func (pk *workspace_picker) close() {
}

func (pk *workspace_picker) grid(input *tview.InputField) *tview.Grid {
	return pk.impl.grid(input)
}

// UpdateQuery implements picker.
func (c *workspace_picker) UpdateQuery(query string) {
	c.fzf.OnSearch(query, true)
}
func (pk workspace_picker) handle_key_override(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := pk.impl.list.InputHandler()
	handle(event, setFocus)
}

func (pk workspace_picker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return pk.handle_key_override
}

func (c *workspace_picker) name() string {
	return "workspace_picker"
}

func new_workspace_picker(v *fzfmain) *workspace_picker {
	impl := &wk_picker_impl{
		new_fzflist_impl(nil, v),
	}
	gload_workspace_list.Load()
	ret := &workspace_picker{impl: impl}
	for i := range gload_workspace_list.Projects {
		a := gload_workspace_list.Projects[i]
		impl.list.AddItem(fmt.Sprintf("%-10s %-30s", a.Name, a.Root), "", func() {
			ret.on_select(&a)
		})
	}

	ret.fzf = new_fzf_on_list(ret.impl.list, true)
	ret.fzf.selected = func(dataindex, listindex int) {
		a := gload_workspace_list.Projects[dataindex]
		log.Println(a)
		ret.on_select(&a)
	}
	return ret
}

func (pk *workspace_picker) on_select(c *Project) {
	pk.impl.parent.main.on_select_project(c)
	pk.impl.parent.hide()
}

type workdir struct {
	root               string
	logfile            string
	configfile         string
	uml                string
	history            string
	cmdhistory         string
	search_cmd_history string
	export             string
	temp               string
	filelist           string
	bookmark           string
}

func new_workdir(root string) workdir {
	config_root := false
	globalroot, err := CreateLspviRoot()
	if err == nil {
		full, err := filepath.Abs(root)
		if err == nil {
			root = filepath.Join(globalroot, filepath.Base(full))
			config_root = true
		}
	}
	if !config_root {
		root = filepath.Join(root, ".lspvi")
	}
	export := filepath.Join(root, "export")
	wk := workdir{
		root:               root,
		configfile:         filepath.Join(globalroot, "config.yaml"),
		logfile:            filepath.Join(root, "lspvi.log"),
		history:            filepath.Join(root, "history.log"),
		bookmark:           filepath.Join(root, "bookmark.json"),
		cmdhistory:         filepath.Join(root, "cmdhistory.log"),
		search_cmd_history: filepath.Join(root, "search_cmd_history.log"),
		export:             export,
		temp:               filepath.Join(root, "temp"),
		uml:                filepath.Join(export, "uml"),
		filelist:           filepath.Join(root, ".file"),
	}
	ensure_dir(root)
	ensure_dir(export)
	ensure_dir(wk.temp)
	ensure_dir(wk.uml)
	return wk
}

func ensure_dir(root string) {
	if _, err := os.Stat(root); err != nil {
		if err := os.MkdirAll(root, 0755); err != nil {
			panic(err)
		}
	}
}

var lspviroot workdir
var global_config *LspviConfig
