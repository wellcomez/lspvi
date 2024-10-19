// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import "github.com/gdamore/tcell/v2"

type InsertHandle struct {
	main     MainService
	codeview CodeEditor
}

// HanldeKey implements vim_mode_handle.
func (i InsertHandle) HanldeKey(event *tcell.EventKey) bool {
	i.codeview.handle_key(event)
	return true
}

// State implements vim_mode_handle.
func (i InsertHandle) State() string {
	// panic("unimplemented")
	return "Insert"
}

// end implements vim_mode_handle.
func (i InsertHandle) end() {
	// panic("unimplemented")
}
