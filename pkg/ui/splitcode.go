package mainui

import (
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type CodeSplit struct {
	code_collection map[view_id]*CodeView
	last            view_id
	layout          *flex_area
	main            *mainui
	index           []view_id
}

func (s *CodeSplit) AddCode(d *CodeView) {
	if d == nil {
		return
	}
	s.code_collection[d.id] = d
	s.index = append(s.index, d.id)
	s.last = max(d.id, s.last)
	s.layout.AddItem(d.view, 0, 1, false)
}
func (s *CodeSplit) New() *CodeView {
	a := NewCodeView(s.main)
	global_file_watch.AddReciever(a)
	a.id = s.last + 1
	s.AddCode(a)
	return a
}
func NewCodeSplit(d *CodeView) *CodeSplit {
	code := make(map[view_id]*CodeView)
	ret := &CodeSplit{
		code_collection: code,
	}
	ret.AddCode(d)
	return ret
}

var SplitCode = NewCodeSplit(nil)

func (c CodeSplit) TabIndex(vid view_id) int {
	for i, v := range c.index {
		if v == vid {
			return i
		}
	}
	return -1
}
func (c CodeSplit) Last() *CodeView {
	ind := len(c.index) - 1
	return c.TabView(ind)
}
func (c CodeSplit) TabView(index int) *CodeView {
	if len(c.index) == 0 {
		return nil
	}
	if index < 0 || index >= len(c.index) {
		return nil
	}
	viewid := c.index[index]
	if code, ok := c.code_collection[viewid]; ok {
		return code
	}
	return nil
}

func SplitClose(code *CodeView) context_menu_item {
	return context_menu_item{item: create_menu_item("Close"), handle: func() {
		SplitCode.Remove(code)
		code.main.right_context_menu.remove(code.rightmenu)
	}, hide: !(code.id > view_code)}
}

func (SplitCode *CodeSplit) Remove(code *CodeView) {
	SplitCode.layout.RemoveItem(code.view)
	id := code.id
	var s = make(map[view_id]*CodeView)
	for k, v := range SplitCode.code_collection {
		if v != code {
			s[k] = v
		}
	}
	for i, v := range SplitCode.index {
		if v == id {
			SplitCode.index = append(SplitCode.index[:i], SplitCode.index[i+1:]...)
		}
	}
	SplitCode.code_collection = s
	global_file_watch.Remove(code)
}
func SplitRight(code *CodeView) context_menu_item {
	main := code.main
	return context_menu_item{item: create_menu_item("SplitRight"), handle: func() {
		if code.id == view_code_below {
			return
		}
		if code.id < view_code {
			return
		}

		codeview2 := SplitCode.New()
		codeview2.view.SetBorder(true)
		main.right_context_menu.add(codeview2.rightmenu)
		codeview2.open_file_line(code.Path(), nil, true)
	}}
}
func (codeview2 *CodeView) open_file_line(filename string, line *lsp.Location, focus bool) {
	main := codeview2.main
	codeview2.LoadAndCb(filename, func() {
		codeview2.view.SetTitle(codeview2.Path())
		if line != nil {
			codeview2.goto_loation(line.Range, codeview2.id != view_code_below)
		}
		go main.async_lsp_open(filename, func(sym *lspcore.Symbol_file) {
			codeview2.lspsymbol = sym
			if focus && codeview2.id != view_code_below {
				if sym == nil {
					main.symboltree.Clear()
				}
			}
		})
		if codeview2.id == view_code_below {
			go func() {
				main.app.QueueUpdateDraw(func() {
					main.tab.ActiveTab(view_code_below, true)
				})
			}()
		}
	})
}
