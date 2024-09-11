package mainui

import (
	"encoding/json"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"log"
	"os"
	"path/filepath"
)

type Project struct {
	Name string `json:"name"`
	Root string `json:"root"`
}
type workspace_list struct {
	Projects []Project `json:"projects"`
}

var gload_workspace_list workspace_list

func (wk *workspace_list) Add(root string) error {
	if !checkDirExists(root) {
		return os.ErrNotExist
	}
	for _, v := range wk.Projects {
		if v.Root == root {
			return nil
		}
	}
	wk.Projects = append(wk.Projects, Project{
		Name: filepath.Base(root),
		Root: root,
	})
	return save_workspace_list(wk)
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
	if err == nil {
		return nil
	}
	return json.Unmarshal(buf, wk)
}

func (*workspace_list) get_config_file() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	root := filepath.Join(home, ".lspvi")
	os.Mkdir(root, 0755)
	if _, err := os.Stat(root); err != nil {
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
