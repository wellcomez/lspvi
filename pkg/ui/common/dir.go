package common

import (
	"os"
	"path/filepath"
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
func NewWorkdir(root string) Workdir {
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
	wk := Workdir{
		Root:               root,
		Configfile:         filepath.Join(globalroot, "config.yaml"),
		Logfile:            filepath.Join(root, "lspvi.log"),
		History:            filepath.Join(root, "history.log"),
		Bookmark:           filepath.Join(root, "bookmark.json"),
		Cmdhistory:         filepath.Join(root, "cmdhistory.log"),
		Search_cmd_history: filepath.Join(root, "search_cmd_history.log"),
		Export:             export,
		Temp:               filepath.Join(root, "temp"),
		UML:                filepath.Join(export, "uml"),
		Filelist:           filepath.Join(root, ".file"),
	}
	ensure_dir(root)
	ensure_dir(export)
	ensure_dir(wk.Temp)
	ensure_dir(wk.UML)
	return wk
}
func ensure_dir(root string) {
	if _, err := os.Stat(root); err != nil {
		if err := os.MkdirAll(root, 0755); err != nil {
			panic(err)
		}
	}
}