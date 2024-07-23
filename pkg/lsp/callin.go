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
	Allstack []*CallStack
	loc      lsp.Location
	lsp      *lsp_base
	set      map[lsp.Range]bool
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

type callchain struct {
	parent *callchain
	data   lsp.CallHierarchyItem
	level  int
}
type added struct {
	set []*callchain
}

func (task *CallInTask) addchild(parent *callchain, leaf *added) error {
	child, err := task.lsp.CallHierarchyIncomingCalls(parent.data)
	if err != nil || parent.level > 5 {
		leaf.set = append(leaf.set, parent)
		return err
	}
	add := false
	for _, cc := range child {
		if task.set[cc.From.Range] {
			continue
		}
		top := &callchain{
			data:   cc.From,
			parent: parent,
			level:  parent.level + 1,
		}
		task.addchild(top, leaf)
		add = true
	}
	if !add {
		leaf.set = append(leaf.set, parent)
	}
	return nil
}
func (task *CallInTask) run() error {
	c1, err := task.lsp.PrepareCallHierarchy(task.loc)
	if err != nil {
		return err
	}
	for _, item := range c1 {
		var leaf added
		top := &callchain{
			data:   item,
			parent: nil,
			level:  0,
		}
		task.set[item.Range] = true
		task.addchild(top, &leaf)
		for _, v := range leaf.set {
			stacks := &CallStack{}
			for v != nil {
				stacks.Add(NewCallStackEntry(v.data))
				v = v.parent
			}
			task.Allstack = append(task.Allstack, stacks)
		}
	}
	return nil
}
