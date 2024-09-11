package mainui

import (
	"os"

	"gopkg.in/yaml.v2"
)

type LspviConfig struct {
	Colorscheme string `yaml:"colorscheme"`
	Wrap        bool `yaml:"wrap"`
}

func (config LspviConfig) Load() (*LspviConfig, error) {
	buf, err := os.ReadFile(lspviroot.configfile)
	default_ret:=LspviConfig{
		Colorscheme:        "darcula",
		Wrap: false,
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
	var ret LspviConfig
	if buf, err := yaml.Marshal(&ret); err == nil {
		return os.WriteFile(lspviroot.configfile, buf, 0644)
	} else {
		return err
	}
}
