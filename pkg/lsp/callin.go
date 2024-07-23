package lspcore

import (
	"fmt"
	// "log"
	// "strings"

	"github.com/tectiv3/go-lsp"
)

func key(call lsp.CallHierarchyItem) string {
	return fmt.Sprintf("%s%s%d%d%d%d", call.Name, call.URI.String(), call.Range.End.Line, call.Range.End.Character, call.Range.Start.Line, call.Range.Start.Character)
}

// CallStackEntry
type CallStackEntry struct {
	Name      string
	Item      lsp.CallHierarchyItem
	PtrSymobl *Symbol
}

func (c CallStackEntry) DirName() string {
	if c.PtrSymobl != nil {
		if len(c.PtrSymobl.classname) > 0 {
			s := fmt.Sprintf("%s::%s", c.PtrSymobl.classname, c.PtrSymobl.SymInfo.Name)
			return s
		}
	}
	return c.Name
}
func (c CallStackEntry) DisplayName() string {
	if c.PtrSymobl != nil {
		if len(c.PtrSymobl.classname) > 0 {
			s := fmt.Sprintf("%s::%s", c.PtrSymobl.classname, c.PtrSymobl.SymInfo.Name)
			return fmt.Sprintf("%s %s:%d", s, c.Item.URI.AsPath().String(), c.Item.Range.Start.Line)
		}
	}
	return fmt.Sprintf("%s %s:%d", c.Name, c.Item.URI.AsPath().String(), c.Item.Range.Start.Line)
}

// NewCallStackEntry
func NewCallStackEntry(item lsp.CallHierarchyItem) *CallStackEntry {
	return &CallStackEntry{
		Name: item.Name,
		Item: item,
	}
}

type CallInTask struct {
	Name     string
	Allstack []*CallStack
	loc      lsp.Location
	lsp      lspclient
	set      map[string]bool
	cb       *func(task CallInTask)
}

type CallStack struct {
	Items    []*CallStackEntry
	resovled bool
}

func (c *CallStack) Add(item *CallStackEntry) {
	// c.Items = append([]*CallStackEntry{item}, c.Items...)
	c.Items = append(c.Items, item)

}
func NewCallStack() *CallStack {
	ret := CallStack{resovled: false}
	return &ret
}
func NewCallInTask(loc lsp.Location, lsp lspclient) *CallInTask {
	name := NewBody(loc).String()
	task := &CallInTask{
		Name: name,
		loc:  loc,
		lsp:  lsp,
	}
	task.set = make(map[string]bool)
	return task
}
func (c CallInTask) Dir() string {
	for _, v := range c.Allstack {
		if v.resovled && len(v.Items) > 0 {
			a := v.Items[len(v.Items)-1]
			return a.DirName()
		}
	}
	return c.Name
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
		if task.set[key(cc.From)] {
			continue
		}
		top := &callchain{
			data:   cc.From,
			parent: parent,
			level:  parent.level + 1,
		}
		task.set[key(cc.From)] = true
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
		task.set[key(item)] = true
		task.addchild(top, &leaf)
		for _, v := range leaf.set {
			stacks := &CallStack{resovled: false}
			for v != nil {
				stacks.Add(NewCallStackEntry(v.data))
				v = v.parent
			}
			task.Allstack = append(task.Allstack, stacks)
		}
	}
	return nil
}

type class_resolve_task struct {
	callstack *CallStack
	wklsp     *LspWorkspace
}

// Run jj
func (c *class_resolve_task) Run() error {
	for _, v := range c.callstack.Items {
		c.resolve(v)
	}
	c.callstack.resovled = true
	return nil
}
func (c *class_resolve_task) resolve(entry *CallStackEntry) {
	// sss :=entry.DisplayName()
	// if strings.Index(sss,"ExecuteNavigationEvent")>-1{
	// 	log.Println("xxxxxxxxx")
	// }
	sym, _ := c.wklsp.find_from_stackentry(entry)
	if sym != nil {
		entry.PtrSymobl = sym
	}
}
