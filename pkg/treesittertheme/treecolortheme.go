package treesittertheme

import (
	"embed"
	"path/filepath"
)

//go:embed  colorscheme/output
var TreesitterSchemeLoader embed.FS

func LoadTreesitterTheme(theme string) ([]byte, error) {
	path := filepath.Join("colorscheme", "output", theme+".micro")
	buf, err := TreesitterSchemeLoader.ReadFile(path)
	return buf, err
}
func GetTheme() ([]string, error) {
	ret := []string{}
	dir := "colorscheme/output"
	dirs, err := TreesitterSchemeLoader.ReadDir(dir)
	if err != nil {
		return ret, err
	}
	for i := range dirs {
		d := dirs[i]
		ret = append(ret, d.Name())
	}
	return ret, nil
}