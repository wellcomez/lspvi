package lspcore

import "testing"

func TestUMLTask(t *testing.T) {

	file := "/home/z/.lspvi/goui/export/uml/CallInTask::get_call_json_filename/callstack.json"
	if task, err := NewCallInTaskFromFile(file); err == nil {
		task.Allstack[1].Uml(true)
	}

}
