// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"gopkg.in/yaml.v2"
	"os"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type highlight struct {
	Search string `yaml:"search"`
}
type color struct {
	Highlight highlight `yaml:"highlight"`
}
type LspviConfig struct {
	Colorscheme string            `yaml:"colorscheme"`
	Wrap        bool              `yaml:"wrap"`
	Lsp         lspcore.LspConfig `yaml:"lsp"`
	Color       color             `yaml:"color"`
}

func (config LspviConfig) Load() (*LspviConfig, error) {
	buf, err := os.ReadFile(lspviroot.configfile)
	default_ret := LspviConfig{
		Colorscheme: "darcula",
		Wrap:        false,
		Color:       color{},
	}
	if err != nil {
		return &default_ret, err

	}
	var ret LspviConfig
	if yaml.Unmarshal(buf, &ret) == nil {
		return &ret, nil
	}
	return &default_ret, err
}
func (config LspviConfig) Save() error {
	if buf, err := yaml.Marshal(&config); err == nil {
		return os.WriteFile(lspviroot.configfile, buf, 0644)
	} else {
		return err
	}
}
