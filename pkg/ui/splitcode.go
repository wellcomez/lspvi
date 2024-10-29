// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

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
	active_codeview *CodeView
	resize          *editor_mouse_resize
}

func (s *CodeSplit) AddCode(d *CodeView) {
	if d == nil {
		return
	}
	s.code_collection[d.id] = d
	s.index = append(s.index, d.id)
	s.last = max(d.id, s.last)
	s.layout.AddItem(d.view, 0, d.Width, false)
}
func (s *CodeSplit) SetActive(v *CodeView) {
	s.active_codeview = v
}
func (s *CodeSplit) New() *CodeView {
	a := NewCodeView(s.main)
	a.id = s.last + 1
	set_view_focus_cb([]view_id{a.id}, s.main)
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
func (c CodeSplit) First() *CodeView {
	if len(c.index) == 0 {
		return nil
	}
	return c.TabView(0)
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
		// code.main.Right_context_menu().remove(code.rightmenu)
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
	SplitCode.resize.remove(code.view_link)
	SplitCode.code_collection = s
	global_file_watch.Remove(code)
}
func SplitRight(code *CodeView) context_menu_item {
	return context_menu_item{item: create_menu_item("SplitRight"), handle: func() {
		if code.id.is_editor_main() {
			code.SplitRight()
		}
	}, hide: !code.id.is_editor_main()}
}

func (code *CodeView) SplitRight() *CodeView {
	codeview2 := create_split_codeview(code)
	codeview2.LoadFileWithLsp(code.Path(), nil, true)
	return codeview2
}

func create_split_codeview(code *CodeView) *CodeView {
	codeview2 := SplitCode.New()
	codeview2.view.SetBorder(true)
	SplitCode.SetActive(codeview2)
	codeview2.view_link.Width = code.Width
	codeview2.view_link.Height = code.Height
	SplitCode.AddCode(codeview2)
	SplitCode.resize.add(codeview2.view_link, SplitCode.resize.LastIndex()+1)
	return codeview2
}
func (code *CodeView) OpenBelow(filename string, line *lsp.Location, focus bool, option *lspcore.OpenOption) {
	code2 := code.main.Codeview2()
	code2.open_file_lspon_line_option(filename, line, focus, option)
	code.main.ActiveTab(view_code_below, true)
}
func (code *CodeView) NewTab(filename string, line *lsp.Location, focus bool, option *lspcore.OpenOption) *CodeView {
	codeview2 := create_split_codeview(code)
	codeview2.open_file_lspon_line_option(filename, line, focus, option)
	return codeview2
}
