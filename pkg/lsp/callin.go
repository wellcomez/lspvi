package lspcore

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	// "log"
	// "strings"

	"github.com/tectiv3/go-lsp"
)

func key(call lsp.CallHierarchyItem) string {
	return fmt.Sprintf("%s%s%d%d%d%d", call.Name, call.URI.String(), call.Range.End.Line, call.Range.End.Character, call.Range.Start.Line, call.Range.Start.Character)
}

// CallStackEntry
type CallStackEntry struct {
	Name           string
	Item           lsp.CallHierarchyItem
	PtrSymobl      *Symbol
	FromRanges     []lsp.Range
	ReferencePlace []lsp.Location
}

func (c CallStackEntry) DirName() string {
	if c.PtrSymobl != nil {
		if len(c.PtrSymobl.Classname) > 0 {
			s := fmt.Sprintf("%s::%s", c.PtrSymobl.Classname, c.PtrSymobl.SymInfo.Name)
			return s
		}
	}
	return c.Name
}
func (c CallStackEntry) DisplayName() string {
	if c.PtrSymobl != nil {
		if len(c.PtrSymobl.Classname) > 0 {
			s := fmt.Sprintf("%s::%s", c.PtrSymobl.Classname, c.PtrSymobl.SymInfo.Name)
			return fmt.Sprintf("%s %s:%d", s, c.Item.URI.AsPath().String(), c.Item.Range.Start.Line+1)
		}
	}
	return fmt.Sprintf("%s %s:%d", c.Name, c.Item.URI.AsPath().String(), c.Item.Range.Start.Line+1)
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
func (c CallStackEntry) IsCaller(loc lsp.Location) bool {
	// small := loc.Range
	// big := c.Item.Range
	if loc.URI.String() == c.Item.URI.String() {
		// return RangeBefore(big, small) && RangeAfter(big, small)
		// return RangeBefore(big, small)
		return true
	}
	return false
}

// NewCallStackEntry
func NewCallStackEntry(item lsp.CallHierarchyItem, fromRanges []lsp.Range, referenceplace []lsp.Location) *CallStackEntry {
	return &CallStackEntry{
		Name:           item.Name,
		Item:           item,
		FromRanges:     fromRanges,
		ReferencePlace: referenceplace,
	}
}

var callstack_task_id = 0

type CallInTask struct {
	Name       string
	Allstack   []*CallStack
	Loc        lsp.Location
	lsp        lspclient
	set        map[string]bool
	UID        int
	TraceLevel int
	// cb       *func(task CallInTask)
}

func (task CallInTask) TreeNodeid() string {
	return fmt.Sprintf("%d", task.UID)
	// return string(task.UID)
}
func (task *CallInTask) Delete(root string) error {
	fielname := filepath.Join(root, task.Dir())
	return os.RemoveAll(fielname)
}
func (task *CallInTask) Save(root string) error {
	fielname := task.get_call_json_filename(root)
	buf, err := json.Marshal(task)
	if err != nil {
		log.Println(err)
		return err
	}
	err = os.WriteFile(fielname, buf, os.ModePerm)
	if err != nil {
		log.Println(err)
		return err
	}
	return err
}

func (task *CallInTask) get_call_json_filename(root string) string {
	fielname := filepath.Join(root, task.Dir(), "callstack.json")
	return fielname
}

var callstack_id = 0

type CallStack struct {
	Items    []*CallStackEntry
	Resovled bool
	UID      int
}

func shortfuncitonname(c *CallStackEntry) string {
	if c.PtrSymobl != nil {
		if len(c.PtrSymobl.Classname) > 0 {
			s := fmt.Sprintf("%s::%s", c.PtrSymobl.Classname, c.PtrSymobl.SymInfo.Name)
			return fmt.Sprintf("%s :%d", s, c.Item.Range.Start.Line+1)
		}
	}
	return fmt.Sprintf("%s :%d", c.Name, c.Item.Range.Start.Line+1)
}
func (stack *CallStack) Get_callchain_name(index int, global_prj_root string) string {
	first := stack.Items[0]
	last := stack.Items[len(stack.Items)-1]
	nodename := fmt.Sprintf("%d.  [%d] %s <- %s", index,
		len(stack.Items),
		strings.ReplaceAll(shortfuncitonname(last), global_prj_root, ""),
		strings.ReplaceAll(shortfuncitonname(first), global_prj_root, ""))
	return nodename
}
func (c *CallStack) InRange(loc lsp.Location) *CallStackEntry {
	var ret *CallStackEntry = nil
	line := loc.Range.Start.Line
	for _, v := range c.Items {
		if v.IsCaller(loc) {
			if ret == nil {
				ret = v
			} else {
				if line-ret.Item.Range.Start.Line > line-v.Item.Range.Start.Line {
					ret = v
				}
			}
		}
	}
	return ret
}
func (c *CallStack) Insert(item lsp.CallHierarchyItem, cc lsp.CallHierarchyIncomingCall) {
	new_one := NewCallStackEntry(cc.From, cc.FromRanges, []lsp.Location{})
	if len(c.Items) > 0 {
		a := c.Items[0]
		locations := []lsp.Location{}
		for i := range cc.FromRanges {
			locations = append(locations, lsp.Location{
				Range: cc.FromRanges[i],
				URI:   cc.From.URI,
			})
		}
		a.ReferencePlace = append(a.ReferencePlace, locations...)
		c.Items = append([]*CallStackEntry{new_one, a}, c.Items[1:]...)
	} else {
		c.Items = append([]*CallStackEntry{new_one}, c.Items...)
	}
}
func (c *CallStack) Add(item *CallStackEntry) {
	// c.Items = append([]*CallStackEntry{item}, c.Items...)
	c.Items = append(c.Items, item)

}
func NewCallStack() *CallStack {
	ret := CallStack{Resovled: false}
	return &ret
}
func NewCallInTask(loc lsp.Location, lsp lspclient, level int) *CallInTask {
	name := ""
	if body, err := NewBody(loc); err == nil {
		name = body.String()
	} else {
		log.Println(err)
	}
	callstack_task_id++
	task := &CallInTask{
		Name:       name,
		Loc:        loc,
		lsp:        lsp,
		UID:        callstack_task_id,
		TraceLevel: level,
	}
	task.set = make(map[string]bool)
	return task
}
func (c CallInTask) Dir() string {
	for _, v := range c.Allstack {
		if v.Resovled && len(v.Items) > 0 {
			a := v.Items[len(v.Items)-1]
			return a.DirName()
		}
	}
	return c.Name
}

type callchain struct {
	parent         *callchain
	data           lsp.CallHierarchyItem
	level          int
	fromRanges     []lsp.Range
	ReferencePlace []lsp.Location
}
type added struct {
	set []*callchain
}

var CallMaxLevel = 5

func (task *CallInTask) addchild(parent *callchain, leaf *added) error {
	child, err := task.lsp.CallHierarchyIncomingCalls(parent.data)
	if err != nil || parent.level >= task.TraceLevel {
		leaf.set = append(leaf.set, parent)
		return err
	}
	add := false
	onechild := len(child) == 1
	for _, cc := range child {
		if !onechild {
			if task.set[key(cc.From)] {
				continue
			}
		}
		locations := []lsp.Location{}
		for i := range cc.FromRanges {
			locations = append(locations, lsp.Location{
				Range: cc.FromRanges[i],
				URI:   cc.From.URI,
			})
		}
		parent.ReferencePlace = append(parent.ReferencePlace, locations...)
		top := &callchain{
			data:           cc.From,
			parent:         parent,
			level:          parent.level + 1,
			fromRanges:     cc.FromRanges,
			ReferencePlace: []lsp.Location{},
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
	if task.lsp == nil {
		return errors.New("lsp is nil")
	}
	c1, err := task.lsp.PrepareCallHierarchy(task.Loc)
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
			stacks := &CallStack{Resovled: false, UID: callstack_id}
			for v != nil {
				stacks.Add(NewCallStackEntry(v.data, v.fromRanges, v.ReferencePlace))
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
	c.callstack.Resovled = true
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
