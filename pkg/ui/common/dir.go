package common

import (
	"os"
	"path/filepath"
	"strings"
)

type Arguments struct {
	File string
	Root string
	Ws   string
	Tty  bool
	Cert string
	Grep bool
	Help bool
}

type Workdir struct {
	Root               string
	Logfile            string
	Configfile         string
	UML                string
	History            string
	Cmdhistory         string
	Search_cmd_history string
	Export             string
	Temp               string
	Filelist           string
	Bookmark           string
}

func GetLspviRoot() (root string, err error) {
	var home string
	home, err = os.UserHomeDir()
	if err != nil {
		return "", err
	}
	root = filepath.Join(home, ".lspvi")
	return
}
func CreateLspviRoot() (root string, err error) {
	root, err = GetLspviRoot()
	if err != nil {
		return
	}
	os.Mkdir(root, 0755)
	if _, err := os.Stat(root); err != nil {
		return "", err
	}
	return root, nil
}
func NewMkWorkdir(root string) (wk Workdir, err error) {
	var globalroot string
	globalroot, err = CreateLspviRoot()
	if err != nil {
		return
	}
	wk = NewWorkDir(globalroot, root)
	ensure_dir(wk.Export)
	ensure_dir(wk.Temp)
	ensure_dir(wk.UML)
	return
}

func NewWorkDir(globalroot string, root string) (wk Workdir) {
	root_under_config := filepath.Join(globalroot, filepath.Base(root))
	export := filepath.Join(root_under_config, "export")
	wk = Workdir{
		Root:               root_under_config,
		Configfile:         filepath.Join(globalroot, "config.yaml"),
		Logfile:            filepath.Join(root_under_config, "lspvi.log"),
		History:            filepath.Join(root_under_config, "history.log"),
		Bookmark:           filepath.Join(root_under_config, "bookmark.json"),
		Cmdhistory:         filepath.Join(root_under_config, "cmdhistory.log"),
		Search_cmd_history: filepath.Join(root_under_config, "search_cmd_history.log"),
		Export:             export,
		Temp:               filepath.Join(root_under_config, "temp"),
		UML:                filepath.Join(export, "uml"),
		Filelist:           filepath.Join(root_under_config, ".file"),
	}
	return wk
}
func ensure_dir(root string) {
	if _, err := os.Stat(root); err != nil {
		if err := os.MkdirAll(root, 0755); err != nil {
			panic(err)
		}
	}
}

func Trim_project_filename(x, y string) string {
	if strings.HasPrefix(x, y) {
		x = strings.TrimPrefix(x, y)
		x = strings.TrimPrefix(x, "/")
	}
	return x
}
func Is_open_as_md(files string) bool {
	var sss = []string{".md", ".puml"}
	return newFunction(files, sss)
}

func newFunction(files string, sss []string) bool {
	var ext = filepath.Ext(files)
	for _, v := range sss {
		if v == ext {
			return true
		}
	}
	return false
}
func Is_image(ext string) bool {
	var sss = []string{".jpg", ".png", ".gif", ".jpeg", ".bmp"}
	return newFunction(ext, sss)
}
