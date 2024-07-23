package lspcore

import (
	"github.com/tectiv3/go-lsp"
)

// CallStackEntry
type CallStackEntry struct {
	Name string
	item lsp.CallHierarchyItem
}

// NewCallStackEntry
func NewCallStackEntry(item lsp.CallHierarchyItem) *CallStackEntry {
	return &CallStackEntry{
		Name: item.Name,
		item: item,
	}
}

type CallInTask struct {
	Name     string
	Allstack []CallStack
	loc      lsp.Location
	lsp      *lsp_base
}
type CallStack struct {
	Items []*CallStackEntry
}

func (c *CallStack) Add(item *CallStackEntry) {
	c.Items = append([]*CallStackEntry{item}, c.Items...)

}
func NewCallStack() *CallStack {
	ret := CallStack{}
	return &ret
}
func NewCallInTask(loc lsp.Location, lsp *lsp_base) *CallInTask {
	name := NewBody(loc).String()
	task := &CallInTask{
		Name: name,
		loc:  loc,
		lsp:  lsp,
	}
	return task
}

func (task *CallInTask) run() error {
	var ret []CallStack
	c1, err := task.lsp.PrepareCallHierarchy(task.loc)
	if err != nil {
		return err
	}
	var stack CallStack
	if len(c1) > 0 {
		stack.Add(NewCallStackEntry(c1[0]))
		c2, err := task.lsp.CallHierarchyIncomingCalls(c1[0])
		if err == nil && len(c2) > 0 {
			stack.Add(NewCallStackEntry(c2[0].From))
		}
	}
	ret = append(ret, stack)
	//sym.Handle.OnCallInViewChanged(ret)
	return nil
}
