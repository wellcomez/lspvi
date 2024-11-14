package mason

import (
	"fmt"
	"os"
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

type software struct {
	Type pkgtype
	data string
}

func (s *software) run(dest string) {
	switch s.Type {
	case pkg_github:
	case pkg_go:
		cmd := fmt.Sprintf("go  install %s", s.data)
		debug.InfoLog("mason", cmd)
	case pkg_npm:
		cmd := fmt.Sprintf("npm install --prefx %s %s", dest, s.data)
		debug.InfoLog("mason", cmd)
	case pkg_pypi:
		cmd := fmt.Sprintf("pip install --target %s %s", dest, s.data)
		debug.InfoLog("mason", cmd)
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
func Load(s string) error {
	// Read the YAML file
	yamlFile, err := os.ReadFile(s)
	if err != nil {
		debug.DebugLogf("mason", "Error reading YAML file: %v", err)
		return err
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
		return err
	}
	version, account, pktype := get_version_account(config.Source.ID)
	data := strings.ReplaceAll(string(yamlFile), "{{version}}", version)
	err = yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		debug.DebugLogf("mason",
			"Error unmarshaling YAML file: %v", err)
		return err
	}
	var app software
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
						return nil
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
	println(app.data)
	return nil
}

type SoftManager struct {
	wk common.Workdir
}

"clangd",          "go"                      ,           "rust-analyzer"   "typescript-language-server"
  "kotlin-language-server"  "python-lsp-server"  "swift-mesonlsp"

func NewSoftManager(wk common.Workdir) *SoftManager {
	return &SoftManager{
		wk: wk,
	}
}
