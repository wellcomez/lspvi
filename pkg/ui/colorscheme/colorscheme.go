package main

import (
	"fmt"
	"strings"
	// "log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
	// "gopkg.in/yaml.v2"
)

// Define a struct to match the YAML structure
type Item struct {
	Group      string `yaml:"Group"`
	Foreground string `yaml:"foreground"`
	Background string `yaml:"background"`
}
type File struct {
	Data []Item `yaml:"Data"`
}

func main() {
	// Sample YAML data as a string
	x := "/home/z/dev/lsp/goui/pkg/ui/colorscheme/"
	dirs, err := os.ReadDir(x)
	if err != nil {
		return
	}
	dir := filepath.Join(x, "output")
	_, err = os.Stat(dir)
	if err != nil {

		err = os.Mkdir(dir, 0755)

		if err != nil {
			fmt.Println(err)
			return
		}
	}
	for _, v := range dirs {
		var data File
		path := filepath.Join(x, v.Name())
		name := v.Name()
		index := strings.Index(name, ".")
		if index < 0 {
			continue
		}
		name = name[:index]
		if filepath.Ext(path) != ".yml" {
			continue
		}
		buf, err := os.ReadFile(path)
		if err != nil {
			fmt.Println(err)
			continue
		}
		err = yaml.Unmarshal(buf, &data)

		if err != nil {
			fmt.Println(path, err)
			continue
		}
		if len(data.Data) == 0 {
			fmt.Println(path, "empty")
			continue
		}
		ret := []string{}
		bg := "#263238"
		for _, v := range data.Data {
			if v.Group == "Normal" {
				if len(v.Background) > 0 {
					bg = v.Background
				}
				break
			}
		}
		for _, v := range data.Data {
			s := fmt.Sprintf("color-link %s \"%s,%s\"", strings.ToLower(v.Group), v.Foreground, bg)
			ret = append(ret, s)
		}
		sss := strings.Join(ret, "\n")
		filename := filepath.Join(dir, fmt.Sprintf("%s.micro", name))
		err = os.WriteFile(filename, []byte(sss), 0666)
		if err != nil {
			fmt.Println(path, err)
		}
	}
}
