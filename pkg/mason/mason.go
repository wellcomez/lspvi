package mason

import (
	"embed"
	"time"

	grab "github.com/cavaliergopher/grab/v3"

	// "errors"
	"fmt"

	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v2"
	"zen108.com/lspvi/pkg/debug"
	"zen108.com/lspvi/pkg/ui/common"
)

type PackSpec struct {
	Name   string `yaml:"name"`
	Source string `yaml:"source"`
}

// Source represents the source section of the YAML configuration.
type Source struct {
	ID    string  `yaml:"id"`
	Asset []Asset `yaml:"asset"`
}
type pkgtype int

const (
	pkg_github pkgtype = iota
	pkg_go
	pkg_npm
	pkg_pypi
)

type InstallResult int

const (
	InstallSuccess InstallResult = iota
	InstallFail
	InstallBreak
)

type SoftInstallResult func(SoftwareTask, InstallResult, error)
type SoftwareTask struct {
	Type pkgtype
	data string

	//for asset field
	asset_bin  string
	asset_file string

	excute   bool
	onend    SoftInstallResult
	onupdate func(string)
	Config   Config
	Id       ToolType
	zipdir   string
}

func isExecutableInPath(executable string) bool {
	_, err := exec.LookPath(executable)
	return err == nil
}

type Executable struct {
	Path       string
	Download   string
	DownloadOk bool
}

func (s SoftwareTask) GetBin() (bin Executable, err error) {
	if s.Config.Bin.Value == "{{source.asset.bin}}" {
		if len(s.asset_bin) > 0 {
			path := filepath.Join(s.zipdir, s.asset_bin)
			bin.Path, _ = exec.LookPath(filepath.Base(s.asset_bin))
			bin.Download = path
			bin.DownloadOk = false
			if fi, e := os.Stat(path); e == nil && !fi.IsDir() {
				bin.DownloadOk = true
				return
			}
			if len(bin.Path) > 0 {
				return
			}
		}
	}
	bin.Path, err = exec.LookPath(s.Config.Bin.Key)
	if err == nil {
		return
	}
	return
}

// func (s software_task) Installed() (ret bool, err error) {
// 	if s.Config.Bin.Value == "{{source.asset.bin}}" {
// 		if len(s.bin) > 0 {
// 			if ret = isExecutableInPath(filepath.Base(s.bin)); ret {
// 				return
// 			}
// 		}
// 	}
// 	if ret = isExecutableInPath(s.Config.Bin.Key); ret {
// 		ss, _ := exec.LookPath(s.Config.Bin.Key)
// 		debug.DebugLog("mason", "isExecutableInPath", ss)
// 		return
// 	}
// 	return false, errors.New("not support")
// }

// Write implements io.Writer.
func (s SoftwareTask) Write(p []byte) (n int, err error) {
	// panic("unimplemented")
	if s.onupdate != nil {
		s.onupdate(string(p))
	}
	return len(p), nil
}

type soft_action int

const (
	soft_action_none soft_action = iota
	soft_action_down
	soft_action_install
)

// func (s *software_task) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) (body io.ReadCloser) {
// 	debug.InfoLogf("sw", src, currentSize, totalSize)
// 	return stream

// }
func (s *SoftwareTask) download(dest string, link string) {
	client := grab.NewClient()
	req, err := grab.NewRequest(dest, link)
	var on_error = func(err error) {
		if s.onend != nil {
			s.onend(*s, InstallFail, err)
		}
	}
	if err != nil {
		if s.onend != nil {
			on_error(err)
		}
		return
	}
	resp := client.Do(req)
	var update_progress = func(x string) {
		if s.onupdate != nil {
			s.onupdate(x)
		}
	}

	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			debug.DebugLogf("sw", "%.02f%% complete\n", resp.Progress()*100)
			x := fmt.Sprintf("%.02f%%  %d", resp.Progress()*100, resp.BytesComplete())
			update_progress(x)

		case <-resp.Done:
			if err := resp.Err(); err != nil {
				// ...
				on_error(err)
			} else {
				update_progress(fmt.Sprintf("%s %.02f%%  %d", dest, resp.Progress()*100, resp.BytesComplete()))
				if s.onend != nil {
					os.Chmod(dest, 0755)
					err := Extract(dest, filepath.Dir(dest))
					var rc = InstallFail
					if err == nil {
						rc = InstallSuccess
					}
					s.onend(*s, rc, err)
				}
			}
			return
		}
	}
}

func (s *SoftwareTask) run_install_task(dest string) {
	s.zipdir = dest
	var action = soft_action_none
	cmd := ""
	switch s.Type {
	case pkg_github:
		cmd = s.data
		action = soft_action_down
	case pkg_go:
		cmd = fmt.Sprintf("go  install %s", s.data)
		debug.InfoLog("mason", cmd)
		action = soft_action_install
	case pkg_npm:
		cmd = fmt.Sprintf("npm install --prefx %s %s", dest, s.data)
		debug.InfoLog("mason", cmd)
		action = soft_action_install
	case pkg_pypi:
		cmd = fmt.Sprintf("pip install --target %s %s", dest, s.data)
		debug.InfoLog("mason", cmd)
		action = soft_action_install
	}
	switch action {
	case soft_action_install:
		args := strings.Split(cmd, " ")
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = *s
		cmd.Stderr = *s
		err := cmd.Run()
		if s.onend != nil {
			if err != nil {
				s.onend(*s, InstallFail, err)
			} else {
				s.onend(*s, InstallSuccess, err)
			}
		}
	case soft_action_down:
		s.download(filepath.Join(dest, s.asset_file), cmd)
	}
}

func get_version_account(id string) (version string, account string, t pkgtype) {
	ss := strings.Split(id, "@")
	version = ss[1]
	if strings.HasPrefix(ss[0], "pkg:github") {
		t = pkg_github
		account = strings.TrimPrefix(ss[0], "pkg:github/")
	}
	if strings.HasPrefix(ss[0], "pkg:go") {
		t = pkg_go
		account = strings.TrimPrefix(ss[0], "pkg:golang/")
	}
	if strings.HasPrefix(ss[0], "pkg:npm") {
		t = pkg_npm
		account = strings.TrimPrefix(ss[0], "pkg:npm/")
	}
	if strings.HasPrefix(ss[0], "pkg:pypi/") {
		t = pkg_pypi
		account = strings.TrimPrefix(ss[0], "pkg:pypi/")
		if idx := strings.Index(version, "?"); idx > 0 {
			version = version[:idx]
		}
	}
	return
}

// Asset represents an asset within the source section.
type Asset struct {
	Target []string `yaml:"target"`
	File   string   `yaml:"file"`
	Bin    string   `yaml:"bin"`
}
type plain struct {
	Target any    `yaml:"target"`
	File   string `yaml:"file"`
	Bin    string `yaml:"bin"`
}

func (a *Asset) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var p plain

	if err := unmarshal(&p); err != nil {
		return err
	}

	// Check if target is a single string
	switch t := p.Target.(type) {
	case string:
		a.Target = []string{t}
	case []interface{}:
		for _, v := range t {
			if s, ok := v.(string); ok {
				a.Target = append(a.Target, s)
			}
		}
	default:
		return fmt.Errorf("unexpected type for Target: %T", t)
	}
	a.Bin = p.Bin
	a.File = p.File
	return nil
}

// Schemas represents the schemas section of the YAML configuration.
type Schemas struct {
	LSP string `yaml:"lsp"`
}

// Bin represents the bin section of the YAML configuration.
type Bin struct {
	Clangd string `yaml:"clangd"`
	Key    string
	Value  string
}

func (a *Bin) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var Database map[string]string
	err := unmarshal(&Database)
	if err != nil {
		return err
	}
	for k, v := range Database {
		a.Value = v
		a.Key = k
		break
	}
	return nil
}

// Config represents the entire YAML configuration.
type Config struct {
	Source      Source  `yaml:"source"`
	Schemas     Schemas `yaml:"schemas"`
	Bin         Bin     `yaml:"bin"`
	Name        string  `yaml:"name"`
	Description string  `yaml:"description"`
}

func get_target() string {
	switch runtime.GOOS {
	case "windows":
		return "win_x64"
	case "linux":
		return "linux_x64_gnu"
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "darwin_arm64"
		}
		return "darwin_x64"
	}
	return ""
}
func Load(yamlFile []byte, s string) (task SoftwareTask, err error) {
	// Read the YAML file
	if len(yamlFile) == 0 {
		yamlFile, err = os.ReadFile(s)
		if err != nil {
			debug.DebugLogf("mason", "Error reading YAML file: %v", err)
			return
		}
	}

	// Print the content of the YAML file for debugging
	// fmt.Println(string(yamlFile))

	// Define a variable to hold the parsed data
	var config Config

	// Unmarshal the YAML file into the config struct
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		debug.DebugLogf("mason",
			"Error unmarshaling YAML file: %v", err)
		return
	}
	version, account, pktype := get_version_account(config.Source.ID)
	data := strings.ReplaceAll(string(yamlFile), "{{version}}", version)
	err = yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		debug.DebugLogf("mason",
			"Error unmarshaling YAML file: %v", err)
		return
	}
	var app SoftwareTask
	app.Type = pktype
	switch pktype {
	case pkg_github:
		{
			var download_url_template = "https://github.com/%s/releases/download/%s/%s"
			target := get_target()
			for _, v := range config.Source.Asset {
				for _, t := range v.Target {
					yes := t == target
					if !yes {
						if t == "darwin" {
							yes = strings.Contains(target, "darwin")
						} else if t == "unix" {
							yes = !strings.Contains(target, "win")
						} else if t == "win" {
							yes = strings.Contains(target, "win")
						}
					}
					if yes {
						ss := fmt.Sprintf(download_url_template, account, version, v.File)
						app.data = ss
						app.asset_bin = v.Bin
						app.excute = true
						app.asset_file = v.File
						break
					}
				}
				if len(app.data) > 0 {
					break
				}
			}
		}
	case pkg_pypi:
		app.data = account
	case pkg_go:
		app.data = strings.TrimPrefix(config.Source.ID, "pkg:")
	case pkg_npm:
		pkg := strings.TrimPrefix(config.Source.ID, "pkg:npm/")
		app.data = pkg
	}
	app.Config = config
	task = app
	return
	// println(app.data)
}

type SoftManager struct {
	wk       common.Workdir
	task     []*SoftwareTask
	app      string
	OnResult SoftInstallResult
}
type ToolType int

const (
	ToolLsp_go ToolType = iota
	ToolLsp_clangd
	ToolLsp_py
	ToolLsp_ts
	ToolLsp_rust
	ToolLsp_swift
	ToolLsp_kotlin
)

type soft_config_file struct {
	id  ToolType
	dir string
}

var ToolMap = []soft_config_file{
	{ToolLsp_clangd, "clangd"},
	{ToolLsp_go, "go"},
	{ToolLsp_rust, "rust-analyzer"},
	{ToolLsp_ts, "typescript-language-server"},
	{ToolLsp_kotlin, "kotlin-language-server"},
	{ToolLsp_py, "python-lsp-server"},
	// {ToolLsp_swift, "swift-mesonlsp"},
}

func NewSoftManager(wk common.Workdir) *SoftManager {
	root := filepath.Dir(wk.Configfile)
	app := filepath.Join(root, ".software")
	return &SoftManager{
		wk:  wk,
		app: app,
	}
}
func (v soft_config_file) Load() (ret SoftwareTask, err error) {
	file := fmt.Sprintf("config/%s/package.yaml", v.dir)
	var buf []byte
	buf, err = uiFS.ReadFile(file)
	if err == nil {
		ret, err = Load(buf, "")
		ret.Id = v.id
		return
	}
	return
}

//go:embed  config
var uiFS embed.FS

func (s *SoftManager) GetAll() (ret []SoftwareTask) {
	for _, v := range ToolMap {
		task, err := v.Load()
		task.zipdir = filepath.Join(s.app, v.dir)
		if err == nil {
			ret = append(ret, task)
		}
	}
	return
}

func (mrg *SoftManager) Start(newtask *SoftwareTask, update func(string), onend func(InstallResult, error)) {

	mrg.task = append(mrg.task, newtask)
	newtask.onupdate = update
	newtask.onend = func(s SoftwareTask, result InstallResult, err error) {
		var tasks []*SoftwareTask
		for _, v := range mrg.task {
			if v != newtask {
				tasks = append(tasks, v)
			}
		}
		mrg.task = tasks
		if onend != nil {
			onend(result, err)
		}
	}
	for _, v := range ToolMap {
		if v.id == newtask.Id {
			dest := filepath.Join(mrg.app, v.dir)
			os.MkdirAll(dest, 0755)
			go newtask.run_install_task(dest)
			break
		}
	}
}

// func (s *SoftManager) Run(t ToolType, update func(string)) {
// 	for _, v := range ToolMap {
// 		if v.id == t {
// 			file := fmt.Sprintf("config/%s/package.yaml", v.dir)
// 			dest := filepath.Join(s.app, v.dir)
// 			os.MkdirAll(dest, 0755)
// 			if buf, err := uiFS.ReadFile(file); err == nil {
// 				if task, err := Load(buf, file); err == nil {
// 					new_task := &task
// 					new_task.onupdate = func(s string) {
// 						if update != nil {
// 							update(s)
// 						}
// 						debug.InfoLog("update", s)
// 					}
// 					s.task = append(s.task, new_task)
// 					go task.run_install_task(dest)
// 					task.onend = func(software_task, install_result,err error) {
// 						var tasks []*software_task
// 						for _, v := range s.task {
// 							if v != new_task {
// 								tasks = append(tasks, v)
// 							}
// 						}
// 						s.task = tasks
// 					}
// 				}
// 			} else {
// 				debug.ErrorLog("mason", err)
// 			}
// 		}
// 	}
// }
