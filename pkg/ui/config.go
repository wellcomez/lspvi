// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"os"

	"gopkg.in/yaml.v2"
	"zen108.com/lspvi/pkg/debug"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type highlight struct {
	Search string `yaml:"search"`
}
type color struct {
	Highlight highlight `yaml:"highlight"`
}
type vimmode struct {
	Leadkey string `yaml:"leadkey,omitempty"`
	Enable  *bool  `yaml:"enable,omitempty"`
}
type LspviConfig struct {
	Colorscheme string            `yaml:"colorscheme"`
	Wrap        bool              `yaml:"wrap"`
	Lsp         lspcore.LspConfig `yaml:"lsp"`
	Color       color             `yaml:"color"`
	Vim         *vimmode          `yaml:"vim,omitempty"`
	enablevim   bool
	Keyboard    lspvi_command_map `yaml:"keyboard"`
}

func (ret *LspviConfig) Load() (err error) {
	if buf, e := os.ReadFile(lspviroot.configfile); e != nil {
		debug.ErrorLog("config", err)
		return e
	} else {
		err = yaml.Unmarshal(buf, ret)
		if err == nil {
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
		} else {
			debug.ErrorLog("config", err)
		}
	}
	return
}

func NewLspviconfig() *LspviConfig {
	default_ret := LspviConfig{
		Colorscheme: "darcula",
		Wrap:        false,
		Color:       color{},
		enablevim:   false,
	}
	return &default_ret
}
func (config LspviConfig) Save() error {
	if buf, err := yaml.Marshal(&config); err == nil {
		return os.WriteFile(lspviroot.configfile, buf, 0644)
	} else {
		return err
	}
}
