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
func RangeAfter(r1 lsp.Range, r2 lsp.Range) bool {
	if r1.Start.Line > r2.Start.Line {
		return true
	}
	if r1.Start.Line == r2.Start.Line {
		return r1.Start.Character > r2.Start.Character
	}
	return false

}
func RangeBefore(r1 lsp.Range, r2 lsp.Range) bool {
	if r1.Start.Line < r2.Start.Line {
		return true
	}
	if r1.Start.Line == r2.Start.Line {
		return r1.Start.Character < r2.Start.Character
	}
	return false

}
func (c CallStackEntry) InRange(loc lsp.Location) bool {
	small := loc.Range
	big := c.Item.Range
	if loc.URI.String() == c.Item.URI.String() {
		// return RangeBefore(big, small) && RangeAfter(big, small)
		return RangeBefore(big, small)
	}
	return false
}

// NewCallStackEntry
func NewCallStackEntry(item lsp.CallHierarchyItem) *CallStackEntry {
	return &CallStackEntry{
		Name: item.Name,
		Item: item,
	}
}

var callstack_task_id = 0

type CallInTask struct {
	Name     string
	Allstack []*CallStack
	loc      lsp.Location
	lsp      lspclient
	set      map[string]bool
	UID      int
	// cb       *func(task CallInTask)
}

func (task CallInTask) TreeNodeid() string {
	return string(task.UID)
}

var callstack_id = 0

type CallStack struct {
	Items    []*CallStackEntry
	resovled bool
	UID      int
}

func (c *CallStack) InRange(loc lsp.Location) []*CallStackEntry {
	ret := []*CallStackEntry{}
	for _, v := range c.Items {
		if v.InRange(loc) {
			ret = append(ret, v)
		}
	}
	return ret
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
	callstack_task_id++
	task := &CallInTask{
		Name: name,
		loc:  loc,
		lsp:  lsp,
		UID:  callstack_task_id,
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
			callstack_id++
			stacks := &CallStack{resovled: false, UID: callstack_id}
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
