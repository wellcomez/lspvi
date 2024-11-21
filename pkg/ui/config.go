// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"embed"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
	"zen108.com/lspvi/pkg/debug"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type highlight struct {
	Search string `yaml:"search"`
}
type ColorConfig struct {
	Highlight  *highlight `yaml:"highlight,omitempty"`
	Cursorline *string    `yaml:"cursorline,omitempty"`
}
type vimmode struct {
	Leadkey string `yaml:"leadkey,omitempty"`
	Enable  *bool  `yaml:"enable,omitempty"`
}
type LspviConfig struct {
	Colorscheme string            `yaml:"colorscheme"`
	Wrap        bool              `yaml:"wrap"`
	Lsp         lspcore.LspConfig `yaml:"lsp"`
	Color       ColorConfig       `yaml:"color"`
	Vim         *vimmode          `yaml:"vim,omitempty"`
	enablevim   bool
	Keyboard    lspvi_command_map `yaml:"keyboard"`
}

//go:embed  config
var uiFS embed.FS

func is_file_ok(path string) bool {
	fi, e := os.Stat(path)
	return e == nil && !fi.IsDir()
}
func (ret *LspviConfig) Load(prjroot string) (err error) {
	filename := lspviroot.Configfile
	err = ret.newMethod(filename, true)
	var prj LspviConfig
	if e := prj.newMethod(filepath.Join(prjroot, ".lspvi.yaml"), false); e == nil {
		ret.merge(&prj)
	}
	return
}

func (ret *LspviConfig) merge(prj *LspviConfig) {
	ret.Lsp.Merge(&prj.Lsp)
}
func (ret *LspviConfig) newMethod(filename string, created bool) (err error) {
	var buf []byte
	if _, err = os.Stat(filename); err != nil {
		if created {
			buf = load_internal_data(buf, err)
			os.WriteFile(filename, buf, 0644)
		} else {
			return
		}
	} else {
		if buf, err = os.ReadFile(filename); err != nil {
			if created {
				buf = load_internal_data(buf, err)
				os.WriteFile(filename, buf, 0644)
			} else {
				return
			}
		}
	}

	err = yaml.Unmarshal(buf, ret)
	if err == nil {
		if created {
			if ret.Vim == nil {
				ret.enablevim = true
				ret.Vim = &vimmode{
					Leadkey: "space",
					Enable:  &ret.enablevim,
				}
			} else {
				ret.enablevim = true
				if ret.Vim.Enable != nil {
					ret.enablevim = *ret.Vim.Enable
				} else {
					ret.enablevim = true
				}
			}
		}
	} else {
		debug.ErrorLog("config", err)
	}
	return
}

func load_internal_data(buf []byte, err error) []byte {
	buf, err = uiFS.ReadFile("config/config.yaml")
	if err != nil {
		debug.ErrorLog("config", "embed", err)
	} else {
		os.WriteFile(lspviroot.Configfile, buf, 0644)
	}
	return buf
}

func NewLspviconfig() *LspviConfig {
	disable := false
	lang_config := &lspcore.LangConfig{}
	default_ret := LspviConfig{
		Colorscheme: "dracula",
		Wrap:        false,
		Lsp: lspcore.LspConfig{
			C:          lang_config,
			Golang:     lang_config,
			Py:         lang_config,
			Rust:       lang_config,
			Typescript: lang_config,
			Javascript: lang_config,
			Swift:      lang_config,
		},
		Color: ColorConfig{},
		Vim: &vimmode{
			Enable:  &disable,
			Leadkey: "space",
		},
		enablevim: false,
	}
	return &default_ret
}
func (config LspviConfig) Save() error {
	if buf, err := yaml.Marshal(&config); err == nil {
		return os.WriteFile(lspviroot.Configfile, buf, 0644)
	} else {
		return err
	}
}
