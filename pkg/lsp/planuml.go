package lspcore

import (
	"fmt"
	// "log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"zen108.com/lspvi/pkg/debug"
)

type PlanUmlBin struct {
	jarPath string
	javaCmd string
}

func NewPlanUmlBin() (*PlanUmlBin, error) {
	jarPath := "plantuml-1.2024.6.jar"
	exec, err := os.Executable()
	if err != nil {
		return nil, err
	}
	jarPath = filepath.Join(
		filepath.Dir(exec), jarPath)
	javaCmd := findJavaBinary()
	if _, err := os.Stat(jarPath); os.IsNotExist(err) {
		debug.DebugLog(DebugTag, err)
		return nil, err
	}
	return &PlanUmlBin{jarPath: jarPath, javaCmd: javaCmd}, nil
}

func (p *PlanUmlBin) Convert(uml string) (output_utxt, output_uml string, ret error) {

	if p.javaCmd == "" {
		return "", "", fmt.Errorf("exception java not found")
	}
	root := filepath.Dir(uml)
	// if _, err := os.Stat(p.javaCmd); os.IsNotExist(err) {
	// 	return fmt.Errorf("exception java not found")
	// }
	cmd := exec.Command(p.javaCmd, "-jar", p.jarPath, uml)
	if err := cmd.Run(); err == nil {
		output_uml = filepath.Join(root, strings.Split(filepath.Base(uml), ".")[0]+".png")
	} else {
		ret = err
	}
	cmd = exec.Command(p.javaCmd, "-jar", p.jarPath, uml, "-utxt")
	if err := cmd.Run(); err == nil {
		output_utxt = filepath.Join(root, strings.Split(filepath.Base(uml), ".")[0]+".utxt")
	} else {
		ret = err
	}
	return output_utxt, output_uml, ret
}

// findJavaBinary is a placeholder function to find the Java binary.
// You should implement this function according to your needs.
func findJavaBinary() string {
	// Implement logic to find the Java binary path here.
	// This could be a simple check for the "java" command in PATH,
	// or a more complex search through known installation directories.
	return "java" // Placeholder value
}

type export_root struct {
	wk  *WorkSpace
	Dir string
}

func checkDirExists(dirPath string) bool {
	_, err := os.Stat(dirPath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	// 其他类型的错误
	return false
}
func NewExportRoot(wk *WorkSpace) (*export_root, error) {
	ret := &export_root{wk: wk}
	err := ret.Init()
	return ret, err
}
func (d *export_root) Init() error {
	d.Dir = filepath.Join(d.wk.Export, "uml")
	if checkDirExists(d.Dir) {
		return nil
	}
	err := os.Mkdir(d.Dir, 0755)
	return err
}
func (d export_root) SaveMD(dir, name, content string) (string, error) {
	newdir := filepath.Join(d.Dir, dir)
	os.Mkdir(newdir, 0755)
	filename := filepath.Join(newdir, name+".md")
	err := os.WriteFile(filename, []byte(content), 0644)
	return filename, err
}
func (d export_root) SavePlanUml(dir, name, content string) (string, error) {
	newdir := filepath.Join(d.Dir, dir)
	os.Mkdir(newdir, 0755)
	filename := filepath.Join(newdir, name+".puml")
	err := os.WriteFile(filename, []byte(content), 0644)
	return filename, err
}
