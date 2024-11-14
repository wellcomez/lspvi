package mason

import (
	"embed"
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

type software_task struct {
	Type   pkgtype
	data   string
	onend  func()
	Config Config
	Id     ToolType
}

// Write implements io.Writer.
func (s software_task) Write(p []byte) (n int, err error) {
	// panic("unimplemented")
	return len(p), nil
}

type soft_action int

const (
	soft_action_none soft_action = iota
	soft_action_down
	soft_action_install
)

func (s *software_task) download(dest string) {
}

func (s *software_task) run(dest string) {
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
		cmd.Run()
	case soft_action_down:
		s.download(cmd)
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
}

// Config represents the entire YAML configuration.
type Config struct {
	Source  Source  `yaml:"source"`
	Schemas Schemas `yaml:"schemas"`
	Bin     Bin     `yaml:"bin"`
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
func Load(yamlFile []byte, s string) (task software_task, err error) {
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
	var app software_task
	app.Type = pktype
	switch pktype {
	case pkg_github:
		{
			var download_url_template = "https://github.com/%s/releases/download/%s/%s"
			target := get_target()
			for _, v := range config.Source.Asset {
				for _, t := range v.Target {
					if t == target {
						ss := fmt.Sprintf(download_url_template, account, version, v.File)
						app.data = ss
						println(ss)
						return
					}
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
	task.Config = config
	task = app
	return
	// println(app.data)
}

type SoftManager struct {
	wk   common.Workdir
	task []*software_task
	app  string
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
	{ToolLsp_swift, "swift-mesonlsp"},
}

func NewSoftManager(wk common.Workdir) *SoftManager {
	app := filepath.Join(wk.Root, "app")
	return &SoftManager{
		wk:  wk,
		app: app,
	}
}
func (v soft_config_file) Load() (ret software_task, err error) {
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

func (s *SoftManager) GetAll() (ret []software_task) {
	for _, v := range ToolMap {
		task, err := v.Load()
		if err == nil {
			ret = append(ret, task)
		}
	}
	return
}

func (s *SoftManager) Run(t ToolType) {
	for _, v := range ToolMap {
		if v.id == t {
			file := fmt.Sprintf("config/%s/package.yaml", v.dir)
			dest := filepath.Join(s.app, v.dir)
			os.MkdirAll(dest, 0755)
			if buf, err := uiFS.ReadFile(file); err != nil {
				if task, err := Load(buf, file); err == nil {
					new_task := &task
					s.task = append(s.task, new_task)
					go task.run(dest)
					task.onend = func() {
						var tasks []*software_task
						for _, v := range s.task {
							if v != new_task {
								tasks = append(tasks, v)
							}
						}
						s.task = tasks
					}
				}
			}
		}
	}
}
