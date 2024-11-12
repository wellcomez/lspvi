package main

import (
	"fmt"
	"strconv"
	"strings"

	// "log"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"gopkg.in/yaml.v2"
	// "gopkg.in/yaml.v2"
)

func hexToRGB(hex string) (int32, int32, int32, error) {
	if len(hex) != 7 || hex[0] != '#' {
		return 0, 0, 0, fmt.Errorf("invalid hex color code")
	}

	r, err := strconv.ParseInt(hex[1:3], 16, 0)
	if err != nil {
		return 0, 0, 0, err
	}

	g, err := strconv.ParseInt(hex[3:5], 16, 0)
	if err != nil {
		return 0, 0, 0, err
	}

	b, err := strconv.ParseInt(hex[5:7], 16, 0)
	if err != nil {
		return 0, 0, 0, err
	}

	return int32(r), int32(g), int32(b), nil
}

// Define a struct to match the YAML structure
type Item struct {
	Group      string `yaml:"Group"`
	Foreground string `yaml:"foreground"`
	Background string `yaml:"background"`
	Reverse    *bool  `yaml:"reverse"`
}
type File struct {
	Data []Item `yaml:"Data"`
}

func main() {
	// Sample YAML data as a string
	x := "/home/z/dev/lsp/goui/pkg/treesittertheme/colorscheme/"
	// os.getcwd()
	x = "."
	x = "/Users/jialaizhu/dev/lspvi/pkg/treesittertheme/colorscheme"
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
		bg := ""
		for _, v := range data.Data {
			if v.Group == "Normal" {
				if len(v.Background) > 0 {
					bg = v.Background
				}
				break
			}
		}
		bg = strings.ToUpper(bg)
		if len(bg) < 7 {
			continue
		}
		for _, v := range data.Data {
			b := bg
			if len(v.Background) > 0 {
				b = v.Background
			} else if v.Reverse != nil && *v.Reverse {
				b = newFunction(b).CSS()
			}
			Foreground := v.Foreground
			// if strings.ToLower(v.Group) == "cursorline" {
			// 	Foreground = b
			// 	b = ""
			// }
			var make6 = func(c string) string {
				if len(c) < 7 && len(c) > 1 {
					c = "#" + strings.Repeat("0", 7-len(c)) + c[1:]
				}
				return c
			}
			// if b == "#0" {
			// 	b = "#000000"
			// }

			b = make6(b)
			Foreground = make6(Foreground)
			s := fmt.Sprintf("color-link %s \"%s,%s\"", strings.ToLower(v.Group), strings.ToUpper(Foreground), strings.ToUpper(b))
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

func newFunction(hex string) (cell tcell.Color) {
	r, g, b, _ := hexToRGB(hex)
	bb := tcell.NewRGBColor(r, g, b)
	_, ccc, _ := tcell.StyleDefault.Background(bb).Decompose()
	cell = ccc
	return
}
