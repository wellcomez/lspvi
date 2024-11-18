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
	"zen108.com/lspvi/pkg/devicon"
	"zen108.com/lspvi/pkg/ui/common"
	nerd "zen108.com/lspvi/pkg/ui/icon"
)

// Source represents the source section of the YAML configuration.
type Source struct {
	ID             string   `yaml:"id"`
	Asset          []Asset  `yaml:"asset"`
	Build          []Build  `yaml:"build"`
	Extra_packages []string `yaml:"extra_packages"`
}
type pkgtype int

const (
	pkg_github pkgtype = iota
	pkg_go
	pkg_npm
	pkg_pypi
	pkg_nuget
)

type InstallResult int

const (
	InstallSuccess InstallResult = iota
	InstallFail
	InstallBreak
)

type SoftInstallResult func(SoftwareTask, InstallResult, error)

func rune_string(r rune) string {
	return fmt.Sprintf("%c", r)
}
func (v SoftwareTask) Executable() string {
	bin, _ := v.get_bin()
	if bin.Path != "" {
		return bin.Path
	}
	if is_file_ok(bin.Download) {
		return bin.Download
	}
	return ""
}
func (v SoftwareTask) TaskState(state string) string {
	status := " Not installed"
	check := rune_string(nerd.Nf_seti_checkbox_unchecked)
	yes, _ := v.get_bin()
	installed := ">[?]"
	if len(yes.Path) > 0 {
		installed = ">" + yes.Path
		check = rune_string(nerd.Nf_seti_checkbox)
	}
	download := ""
	cmd, action := v.get_cmd()
	if state != "" {
		cmd = state
	}
	if action == soft_action_down {
		if !yes.DownloadOk {
			download = " " + rune_string(nerd.Nf_fa_download) + " " + cmd
		} else {
			download = yes.Download
		}
	} else {
		download = " " + rune_string(nerd.Nf_cod_debug_rerun) + " " + cmd
	}
	status = fmt.Sprintf("%s %s", installed, download)
	return fmt.Sprintf("%s %s %s %s", check, v.Icon.Icon, v.Config.Name, status)
}

type SoftwareTask struct {
	Type pkgtype
	// data string

	Icon devicon.Icon
	data []SubTask
	//for asset field
	// asset_bin  string
	// asset_file string
	build    Build
	assert   Asset
	excute   bool
	onend    SoftInstallResult
	onupdate func(string)
	Config   Config
	Id       ToolType
	zipdir   string
	Desc     string
}

func IsExecutableInPath(executable string) bool {
	_, err := exec.LookPath(executable)
	return err == nil
}

type Executable struct {
	Path     string
	Download string
	// Url        string
	DownloadOk bool
}

func (s SoftwareTask) get_bin() (bin Executable, err error) {
	// bin.Url = s.data
	if key := s.Config.Bin.GetValue("{{source.build.bin.lsp}}"); len(key) > 0 {
		bin.Download = s.assert.File
		path := filepath.Join(s.zipdir, s.assert.File)
		if is_file_ok(path) {
			bin.DownloadOk = true
		}
	} else if key := s.Config.Bin.GetValue("{{source.asset.bin}}"); len(key) > 0 {
		if len(s.assert.Bin) > 0 {
			path := filepath.Join(s.zipdir, s.assert.Bin)
			bin.Path, _ = exec.LookPath(filepath.Base(s.assert.Bin))
			bin.Download = path
			bin.DownloadOk = false
			if is_file_ok(path) {
				bin.DownloadOk = true
				return
			}
			if len(bin.Path) > 0 {
				return
			}
		}
	}
	for _, v := range s.Config.Bin.data {
		bin.Path, err = exec.LookPath(v.Key)
		if err == nil {
			return
		}
	}
	return
}

func is_file_ok(path string) bool {
	fi, e := os.Stat(path)
	return e == nil && !fi.IsDir()
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
type SubTask interface {
	Run(*SoftwareTask)
}
type SubTaskCmd struct {
	main    *SoftwareTask
	cmdline string
}
type SubTaskDownload struct {
	main *SoftwareTask
	dest string
	link string
}

func (s *SubTaskDownload) Run(m *SoftwareTask) {
	s.main = m
	s.download(s.dest, s.link)
}
func (sub *SubTaskDownload) download(dest string, link string) {
	s := sub.main
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

func (s *SoftwareTask) run_ide_stnstall_task() {
	changedir := false
	switch s.Type {
	case pkg_npm:
		changedir = true
	}
	current, _ := os.Getwd()
	defer func() {
		if changedir {
			os.Chdir(current)
		}
	}()
	if changedir {
		os.Chdir(s.zipdir)
	}
	task := s.get_sub_task()
	for _, v := range task {
		v.Run(s)
	}
}

func (s *SoftwareTask) NewSubCmd(cmdline string) *SubTaskCmd {
	s.Desc = cmdline
	d := SubTaskCmd{s, cmdline}
	return &d
}
func (sub *SubTaskCmd) Run(s *SoftwareTask) {
	sub.main = s
	sub.run_cmdline(sub.cmdline)
}

func (sub *SubTaskCmd) run_cmdline(cmdline string) bool {
	maintask := sub.main
	ssss := strings.Split(cmdline, " ")
	maintask.Write([]byte(cmdline))
	var args []string
	for _, v := range ssss {
		if v == "" {
			continue
		} else {
			args = append(args, v)
		}
	}
	ss, err :=
		exec.LookPath(args[0])
	if err != nil {
		debug.ErrorLog("mason", err)
		maintask.onend(*maintask, InstallFail, err)
		return true
	}
	cmd := exec.Command(ss, args[1:]...)
	cmd.Stdout = *maintask
	cmd.Stderr = *maintask
	err = cmd.Run()
	if maintask.onend != nil {
		if err != nil {
			maintask.onend(*maintask, InstallFail, err)
			return false
		} else {
			maintask.onend(*maintask, InstallSuccess, err)
		}
	}
	return true
}

func (s *SoftwareTask) NewDownloadTask(link string) *SubTaskDownload {

	dest := s.zipdir
	s.Desc = link
	t := SubTaskDownload{s, filepath.Join(dest, s.assert.File), link}
	return &t
}

//	func  get_sub_task(Type pkgtype, data string,s SoftwareTask ) (ret SubTask) {
//		// var cmd string
//		// var  action soft_action
//		switch Type {
//		case pkg_go:
//			cmd := fmt.Sprintf("npm install  %s", data)
//				return s.NewSubCmd(cmd)
//			// action = soft_action_install
//		case pkg_pypi:
//			cmd := fmt.Sprintf("pip  %s", data)
//				return s.NewSubCmd(cmd)
//		case pkg_nuget:
//			// cmd := fmt.Sprintf("nuget %s %s", dest, data)
//			return  s.NewSubCmd(cmd)
//		}
//		return ret
//	}
func (s *SoftwareTask) get_sub_task() (ret []SubTask) {
	return s.data
}

// func run_cmdline(data string)
// var cmd string
// var  action soft_action
// dest := s.zipdir
// switch s.Type {
// case pkg_github:
// 	cmd := s.data
// 	// action = soft_action_down
// 	ret = append(ret, s.NewDownloadTask(cmd))
// case pkg_go:
// 	cmd := fmt.Sprintf("go  install %s", s.data)
// 	ret = append(ret, s.NewSubCmd(cmd))
// 	// action = soft_action_install
// case pkg_npm:
// 	// cmd = fmt.Sprintf("npm install --prefx %s %s", dest, s.data)
// 	cmd := fmt.Sprintf("npm install  %s", s.data)
// 	ret = append(ret, s.NewSubCmd(cmd))
// 	// action = soft_action_install
// case pkg_pypi:
// 	cmd := fmt.Sprintf("pip  %s", s.data)
// 	ret = append(ret, s.NewSubCmd(cmd))
// case pkg_nuget:
// 	cmd := fmt.Sprintf("nuget %s %s", dest, s.data)
// 	ret = append(ret, s.NewSubCmd(cmd))
// }
// return ret
// }

func (s *SoftwareTask) get_cmd() (cmd string, action soft_action) {
	cmd = s.Desc
	switch s.Type {
	case pkg_github:
		action = soft_action_down
	case pkg_go:
		action = soft_action_install
	case pkg_npm:
		action = soft_action_install
	case pkg_pypi:
		action = soft_action_install
	case pkg_nuget:
		action = soft_action_install
	}
	return
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
	if strings.HasPrefix(ss[0], "pkg:nuget/") {
		t = pkg_nuget
		account = strings.TrimPrefix(ss[0], "pkg:nuget/")
		if idx := strings.Index(version, "?"); idx > 0 {
			version = version[:idx]
		}
	}
	return
}

type BuildBin struct {
	Lsp string `yaml:"lsp"`
	Dap string `yaml:"dap"`
}
type buildanystruct struct {
	Target any      `yaml:"target"`
	Run    string   `yaml:"run"`
	Bin    BuildBin `yaml:"bin"`
}
type Build struct {
	Target []string `yaml:"target"`
	Run    string   `yaml:"file"`
	Bin    BuildBin `yaml:"bin"`
}

func (a *Build) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var p buildanystruct

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
	a.Bin.Lsp = strings.TrimPrefix(a.Bin.Lsp, "exec:")
	a.Bin.Dap = strings.TrimPrefix(a.Bin.Dap, "exec:")
	a.Run = p.Run
	return nil
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
type BinDefine struct {
	Key   string
	Value string
}

// Bin represents the bin section of the YAML configuration.
type Bin struct {
	Clangd string `yaml:"clangd"`
	data   []BinDefine
}

func (b *Bin) GetValue(value string) string {
	for _, v := range b.data {
		if v.Value == value {
			return v.Key
		}
	}
	return ""
}

func (a *Bin) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var Database map[string]string
	err := unmarshal(&Database)
	if err != nil {
		return err
	}
	for k, v := range Database {
		a.data = append(a.data, BinDefine{Value: v, Key: k})
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

func match_arch(target string) bool {
	switch runtime.GOARCH {
	case "amd64":
		return strings.Contains(target, "x64") || !strings.Contains(target, "arm64")
	case "arm64":
		return strings.Contains(target, "arm64") || !strings.Contains(target, "x64")
	case "arm":
		debug.DebugLog("mason", "arm", "not support")
	case "386":
		debug.DebugLog("mason", "386", "not support")
	}
	return false
}
func match_os(target string) bool {
	switch runtime.GOOS {
	case "windows":
		return strings.Contains(target, "win")
	case "linux":
		return strings.Contains(target, "linux") || strings.Contains(target, "unix")
	case "darwin":
		return strings.Contains(target, "darwin") || strings.Contains(target, "unix")
	}
	return false
}
func match_target(targets []string) (ret string, err error) {
	for _, v := range targets {
		if match_arch(v) && match_os(v) {
			ret = v
			return
		}
	}
	err = fmt.Errorf("not found target")
	return
}

//	func get_target() string {
//		switch runtime.GOOS {
//		case "windows":
//			return "win_x64"
//		case "linux":
//			return "linux_x64_gnu"
//		case "darwin":
//			if runtime.GOARCH == "arm64" {
//				return "darwin_arm64"
//			}
//			return "darwin_x64"
//		}
//		return ""
//	}
func Load(yamlFile []byte, s string) (app SoftwareTask, err error) {
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
	app.Type = pktype
	switch pktype {
	case pkg_github:
		{
			var download_url_template = "https://github.com/%s/releases/download/%s/%s"
			for _, v := range config.Source.Build {
				var download_url_template = "https://github.com/%s/archive/refs/tags/%s.zip"
				if _, err := match_target(v.Target); err == nil {
					ss := fmt.Sprintf(download_url_template, account, version)
					app.data = append(app.data, app.NewDownloadTask(ss))
					app.build = v
					filename := strings.Split(account, "/")[1]
					app.assert.File = fmt.Sprintf("%s-%s.zip", filename, version[1:])
					app.excute = true
					break
				}
			}
			if len(app.data) > 0 {
				break
			}
			for _, v := range config.Source.Asset {
				if _, err := match_target(v.Target); err == nil {
					ss := fmt.Sprintf(download_url_template, account, version, v.File)
					app.assert = v
					app.excute = true
					app.data = append(app.data, app.NewDownloadTask(ss))
					break
				}
			}
		}
	case pkg_pypi:
		data := account
		cmd := fmt.Sprintf("pip  %s", data)
		app.data = append(app.data, app.NewSubCmd(cmd))
	case pkg_go:
		data := strings.TrimPrefix(config.Source.ID, "pkg:golang/")
		cmd := fmt.Sprintf("go  install %s", data)
		app.data = append(app.data, app.NewSubCmd(cmd))
	case pkg_npm:
		pkg := strings.TrimPrefix(config.Source.ID, "pkg:npm/")
		if len(config.Source.Extra_packages) > 0 {
			packages := []string{pkg}
			packages = append(packages, config.Source.Extra_packages...)
			pkg = strings.Join(packages, " ")
		}
		data := pkg
		// dest:=app.zipdir
		cmd := fmt.Sprintf("npm install   %s", data)
		app.data = append(app.data, app.NewSubCmd(cmd))
	}
	app.Config = config
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
	ToolLsp_bash
	ToolLsp_cmake
	ToolLsp_java_jedi
	ToolLsp_lua
	ToolLsp_vue
	ToolLsp_csharp
	ToolLsp_java
)

type soft_package_file struct {
	id   ToolType
	dir  string
	icon devicon.Icon
}

func get_icon(file string) devicon.Icon {
	if ret, err := devicon.FindIconPath(file); err == nil {
		return ret
	}
	return devicon.Icon{Icon: fmt.Sprintf("%c", nerd.Nf_cod_file_binary)}
}

var ToolMap = []soft_package_file{
	{ToolLsp_clangd, "clangd", get_icon(".cpp")},
	{ToolLsp_go, "go", get_icon(".go")},
	{ToolLsp_rust, "rust-analyzer", get_icon(".rs")},
	{ToolLsp_ts, "typescript-language-server", get_icon(".ts")},
	{ToolLsp_kotlin, "kotlin-language-server", get_icon(".kt")},
	{ToolLsp_py, "python-lsp-server", get_icon(".py")},
	{ToolLsp_bash, "bash-language-server", get_icon(".sh")},
	{ToolLsp_cmake, "cmake-language-server", get_icon("cmakelists.txt")},
	{ToolLsp_java_jedi, "jedi-language-server", get_icon(".java")},
	{ToolLsp_lua, "lua-language-server", get_icon(".lua")},
	{ToolLsp_vue, "vue-language-server", get_icon(".vue")},
	{ToolLsp_csharp, "csharp-language-server", get_icon(".cs")},
	{ToolLsp_java, "java-language-server", get_icon(".java")},
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
func (v soft_package_file) Load() (ret SoftwareTask, err error) {
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

func (s *SoftManager) FindLsp(id ToolType) (task SoftwareTask, err error) {
	for _, v := range ToolMap {
		if v.id != id {
			continue
		}
		task, err = v.Load()
		if err == nil {
			task.zipdir = filepath.Join(s.app, v.dir)
			task.Icon = v.icon
			return
		}
	}
	err = fmt.Errorf("not found software")
	return
}
func (s *SoftManager) GetAll() (ret []SoftwareTask) {
	for _, v := range ToolMap {
		task, err := v.Load()
		task.zipdir = filepath.Join(s.app, v.dir)
		task.Icon = v.icon
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
			dest := newtask.zipdir
			os.MkdirAll(dest, 0755)
			go newtask.run_ide_stnstall_task()
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
