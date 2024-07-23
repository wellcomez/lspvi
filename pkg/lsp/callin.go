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

type CallStack struct {
	Items []*CallStackEntry
}

func (c *CallStack) Add(item *CallStackEntry) {
	c.Items = append(c.Items, item)
}
func NewCallStack() *CallStack {
	ret := CallStack{}
	return &ret
}

type LspCallInRecord struct {
	name string
	data []CallStack
}
