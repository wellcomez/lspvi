package lspcore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUMLTask(t *testing.T) {
	root := "/home/z/.lspvi/goui/export/uml/get_call_json_filename"
	file := filepath.Join(root, "callstack.json")
	if task, err := NewCallInTaskFromFile(file); err == nil {
		data := task.Allstack[0].Uml(false)
		os.WriteFile("test.uml", []byte(data), 0666)
	}

}
