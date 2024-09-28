package mainui

import (
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

func SplitClose(code *CodeView) context_menu_item {
	return context_menu_item{item: create_menu_item("Close"), handle: func() {
		SplitCode.layout.RemoveItem(code.view)
		var s = make(map[view_id]*CodeView)
		for k, v := range SplitCode.code_collection {
			if v != code {
				s[k] = v
			}
		}
		for i, v := range SplitCode.index {
			if v == code.id {
				SplitCode.index = append(SplitCode.index[:i], SplitCode.index[i+1:]...)
			}
		}
		SplitCode.code_collection = s
		code.main.right_context_menu.remove(code.rightmenu)
	}, hide: !(code.id > view_code)}
}
func SplitRight(code *CodeView) context_menu_item {
	main := code.main
	return context_menu_item{item: create_menu_item("SplitRight"), handle: func() {
		if code == main.codeview {
			codeview2 := SplitCode.New()
			codeview2.view.SetBorder(true)
			main.right_context_menu.add(codeview2.rightmenu)
			codeview2.LoadAndCb(code.filename, func() {
				codeview2.view.SetTitle(codeview2.filename)
				go main.async_lsp_open(code.filename, func(sym *lspcore.Symbol_file) {
					codeview2.lspsymbol = sym
				})
				go func() {
					main.app.QueueUpdateDraw(func() {
						main.tab.ActiveTab(view_code_below, true)
					})
				}()
			})
		}
	}}
}
