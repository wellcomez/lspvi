// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	lspcore "zen108.com/lspvi/pkg/lsp"
	"zen108.com/lspvi/pkg/ui/common"
	web "zen108.com/lspvi/pkg/ui/xterm"
)

type Project struct {
	Name string `json:"name"`
	Root string `json:"root"`
}
type workspace_list struct {
	Projects []Project `json:"projects"`
}

var gload_workspace_list workspace_list
var global_prj_root string
var global_file_watch = NewFileWatch()

func (prj *Project) Load(arg *common.Arguments, main *mainui) {
	root := prj.Root
	lspviroot = common.NewWorkdir(root)
	global_config = NewLspviconfig()
	global_config.Load()
	// go servmain(lspviroot.uml, 18080, func(port int) {
	// 	httport = port
	// })

	// handle := LspHandle{}
	// var main = &mainui{
	main.bf = NewBackForward(NewHistory(lspviroot.History))
	main.bookmark = &proj_bookmark{path: lspviroot.Bookmark, Bookmark: []bookmarkfile{}, root: root}
	main.tty = arg.Tty
	main.ws = arg.Ws
	// }
	main.bookmark.load()
	if main.bookmark_view != nil {
		main.bookmark_view.update_redraw()
	}
	// handle.main = main
	if !filepath.IsAbs(root) {
		root, _ = filepath.Abs(root)
	}
	ConfigFile := lspviroot.Configfile
	lspmgr := lspcore.NewLspWk(lspcore.WorkSpace{
		Path:         root,
		Export:       lspviroot.Export,
		Callback:     main,
		NotifyHanlde: main,
		ConfigFile:   ConfigFile})
	main.lspmgr = lspmgr
	main.lspmgr.Handle = main
	global_prj_root = root
	go web.OpenInPrj(root)
	if !global_file_watch.started {
		go global_file_watch.Run(global_prj_root)
	} else {
		global_file_watch.Add(global_prj_root)
	}
	theme := global_config.Colorscheme
	if global_theme == nil {
		global_theme = new_ui_theme(theme, main)
	} else {
		main.on_change_color(theme)
	}
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
	root, err := common.CreateLspviRoot()
	if err != nil {
		return "", err
	}
	config := filepath.Join(root, "workspace.json")
	return config, nil
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
	c.fzf.OnSearch(query, false)
	UpdateColorFzfList(c.fzf).SetCurrentItem(0)
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
		new_fzflist_impl(v),
	}
	gload_workspace_list.Load()
	ret := &workspace_picker{impl: impl}
	fzfdata := []string{}
	for i := range gload_workspace_list.Projects {
		a := gload_workspace_list.Projects[i]
		x := fmt.Sprintf("%-10s %-30s", a.Name, a.Root)
		fzfdata = append(fzfdata, x)
		impl.list.AddItem(x, "", nil)
	}

	ret.fzf = new_fzf_on_list_data(ret.impl.list, fzfdata, true)
	lastindex := -1
	impl.list.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		if i == lastindex {
			c := &gload_workspace_list.Projects[ret.fzf.get_data_index(i)]
			impl.parent.main.on_select_project(c)
			impl.parent.main.current_editor().Acitve()
			impl.parent.hide()
		}
	})
	impl.list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		lastindex = index
	})
	return ret
}





var lspviroot common.Workdir
var global_config *LspviConfig
