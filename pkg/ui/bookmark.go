// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	// fzflib "github.com/reinhrst/fzf-lib"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
)

type bookmark_changed interface {
	onsave()
}
type LineMark struct {
	Line    int //linenumber not buffindex begin with 1
	Text    string
	Comment string
}
type bookmarkfile struct {
	Name     string
	LineMark []LineMark
}

func (b *bookmarkfile) after_line_changed(line int, add bool) {
	lines := b.LineMark
	for i := range lines {
		var s = lines[i]
		if s.Line == line {
			b.Add(s.Line, s.Comment, s.Text, false)
			continue
		} else if s.Line > line {
			b.Add(s.Line, s.Comment, s.Text, false)
			if add {
				s.Line++
			} else {
				s.Line--
			}
			b.Add(s.Line, s.Comment, s.Text, true)
		}
	}
}

type proj_bookmark struct {
	Bookmark []bookmarkfile
	path     string
	changed  []bookmark_changed
	root     string
}

func (prj *proj_bookmark) load() error {
	buf, err := os.ReadFile(prj.path)
	if err != nil {
		return err
	}
	var s proj_bookmark
	err = json.Unmarshal(buf, &s)
	if err == nil {
		prj.Bookmark = s.Bookmark
	}
	return err
}
func (prj *proj_bookmark) delete(a ref_line) error {
	for i := range prj.Bookmark {
		v := &prj.Bookmark[i]
		if v.Name == a.path {
			v.Add(a.loc.Range.Start.Line+1, "", a.code, false)
		}
	}
	prj.save()
	for _, v := range prj.changed {
		v.onsave()
	}
	return nil
}
func (prj *proj_bookmark) udpate(bk *bookmarkfile) {
	for i := range prj.Bookmark {
		if prj.Bookmark[i].Name == bk.Name {
			prj.Bookmark[i] = *bk
			break
		}
	}
	prj.save()
}
func (prj *proj_bookmark) save() error {
	buf, err := json.Marshal(prj)
	if err != nil {
		return err
	}
	ret := os.WriteFile(prj.path, buf, 0666)
	if ret == nil && prj.changed != nil {
		for _, v := range prj.changed {
			v.onsave()
		}
	}
	return ret
}
func (prj *proj_bookmark) GetFileBookmark(file string) *bookmarkfile {
	for i := range prj.Bookmark {
		v := &prj.Bookmark[i]
		if v.Name == file {
			return v
		}
	}
	bookmark := bookmarkfile{Name: file}
	prj.Bookmark = append(prj.Bookmark, bookmark)
	return prj.GetFileBookmark(file)
}
func (b *bookmarkfile) Add(line int, comment string, text string, add bool) {
	if add {
		b.LineMark = append(b.LineMark, LineMark{Line: line, Text: text, Comment: comment})

	} else {
		bb := []LineMark{}
		for _, v := range b.LineMark {
			if v.Line != line {
				bb = append(bb, v)
			}
		}
		b.LineMark = bb
	}
}

type bookmark_picker struct {
	impl *bookmark_picker_impl
}

// close implements picker.
func (pk bookmark_picker) close() {
	// pk.impl.cq.CloseQueue()
}

// UpdateQuery implements picker.
func (pk bookmark_picker) UpdateQuery(query string) {
	listview := pk.impl.hlist
	query = strings.ToLower(query)
	listview.Clear()

	listview.Key = query
	pk.impl.fzf.selected = func(data_index int, listindex int) {
		loc := pk.impl.listdata[data_index].loc
		close_bookmark_picker(pk.impl.prev_picker_impl, loc)
	}
	pk.impl.fzf.OnSearch(query, true)
	pk.update_preview()
}
func (pk bookmark_picker) handle_key_override(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := pk.impl.listcustom.InputHandler()
	handle(event, setFocus)
	pk.update_preview()
}

// handle implements picker.
func (pk bookmark_picker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return pk.handle_key_override
}

// name implements picker.
func (pk bookmark_picker) name() string {
	// panic("unimplemented")
	return "bookmark"
}

type bookmark_picker_impl struct {
	*prev_picker_impl
	fzf   *fzf_on_listview
	hlist *customlist
}

func get_list_item(v ref_line, root string) (string, string) {
	path := v.line + ":" + strings.ReplaceAll(v.path, root, "")
	return fmt.Sprintf("%s **%s**", path, v.caller), strings.TrimLeft(v.code, " \t")
}

type bookmark_edit struct {
	*fzflist_impl
	cb func(s string)
}

// close implements picker.
func (b bookmark_edit) close() {
}

// UpdateQuery implements picker.
func (b bookmark_edit) UpdateQuery(query string) {

}

func (pk bookmark_edit) handle_key_override(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	if event.Key() == tcell.KeyEnter {
		pk.cb(pk.parent.input.GetText())
		pk.parent.hide()
	}
}

// handle implements picker.
func (b bookmark_edit) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	// panic("unimplemented")
	return b.handle_key_override
}

// name implements picker.
func (b bookmark_edit) name() string {
	return "add bookmark comment"
}
func (pk *bookmark_edit) grid(input *tview.InputField) *tview.Grid {
	return pk.fzflist_impl.grid(input)
}

// func new_bookmark_editor(v *fzfmain, cb func(string), code *CodeView) bookmark_edit {
// 	var line = code.view.Cursor.Loc.Y + 1
// 	line1 := code.view.Buf.Line(line - 1)
// 	ret := bookmark_edit{
// 		fzflist_impl: new_fzflist_impl(nil, v),
// 		cb:           cb,
// 	}
// 	ret.fzflist_impl.list.AddItem(line1, code.Path(), nil)
// 	v.create_dialog_content(ret.grid(v.input), ret)
// 	return ret
// }

// new_bookmark_picker
func new_bookmark_picker(v *fzfmain, bookmark *proj_bookmark) bookmark_picker {
	impl := &bookmark_picker_impl{
		prev_picker_impl: new_preview_picker(v),
	}
	sym := bookmark_picker{
		impl: impl,
	}
	// sym.impl.codeprev.view.SetBorder(true)

	impl.hlist, impl.listdata = init_bookmark_list(bookmark, func(i int) {
		loc := sym.impl.listdata[i].loc
		close_bookmark_picker(impl.prev_picker_impl, loc)
	})
	impl.hlist.fuzz = true
	impl.use_cusutom_list(impl.hlist)
	fzf := new_fzf_on_list(sym.impl.hlist, sym.impl.hlist.fuzz)
	sym.impl.fzf = fzf
	sym.UpdateQuery("")
	return sym
}

func init_bookmark_list(bookmark *proj_bookmark, selected func(int)) (*customlist, []ref_line) {
	hlist := new_customlist(false)
	listdata := reload_bookmark_list(bookmark, hlist, selected)
	return hlist, listdata
}

func reload_bookmark_list(bookmark *proj_bookmark, hlist *customlist, selected func(int)) []ref_line {
	hlist.Clear()
	marks := bookmark.Bookmark
	listdata := []ref_line{}
	for _, file := range marks {
		for _, v := range file.LineMark {
			ref := ref_line{
				line:   fmt.Sprintf("%d", v.Line),
				path:   file.Name,
				caller: v.Comment,
				code:   v.Text,
				loc: lsp.Location{
					URI: lsp.NewDocumentURI(file.Name),
					Range: lsp.Range{
						Start: lsp.Position{Line: v.Line - 1},
						End:   lsp.Position{Line: v.Line - 1},
					},
				},
			}
			listdata = append(listdata, ref)
		}
	}
	root := bookmark.root
	for i, v := range listdata {
		a, b := get_list_item(v, root)
		index := i
		hlist.AddItem(fmt.Sprintf("**%-03d** %s", i+1, a), fmt.Sprintf("   %s", b), func() { selected(index) })
	}
	return listdata
}

func close_bookmark_picker(impl *prev_picker_impl, loc lsp.Location) {
	impl.open_location(loc)
	impl.parent.hide()
	// impl.cq.CloseQueue()
}
func (pk bookmark_picker) update_preview() {
	pk.impl.update_preview()
}
func (pk *bookmark_picker) grid(input *tview.InputField) *tview.Flex {
	return pk.impl.flex(input, 2)
}

type bookmark_view struct {
	*view_link
	list          *customlist
	data          []ref_line
	Name          string
	fzf           *fzf_on_listview
	menuitem      []context_menu_item
	right_context bk_menu_context
	bookmark      *proj_bookmark
	code          MainService
	yes           func() bool
}

func (bk *bookmark_view) onsave() {
	b := bk
	b.list.Clear()
	b.list.SetChangedFunc(nil)
	b.list.SetSelectedFunc(nil)
	b.data = reload_bookmark_list(b.bookmark, b.list, func(i int) {
		b.onclick(i)
	})
	b.fzf = new_fzf_on_list(b.list, true)
}
func (bk *bookmark_view) OnSearch(txt string) {
	bk.list.Key = txt
	old := bk.fzf.OnSearch(txt, true)
	if len(txt) > 0 {
		highlight_listitem_search_key(old, bk.list, txt)
	}
	bk.fzf.selected = func(dataindex int, listindex int) {
		loc := bk.data[dataindex].loc
		bk.code.OpenFileHistory(loc.URI.AsPath().String(), &loc)
	}
}
func new_bookmark_view(bookmark *proj_bookmark, code MainService, yes func() bool) *bookmark_view {
	ret := &bookmark_view{
		view_link: &view_link{id: view_bookmark, up: view_code, left: view_uml, right: view_outline_list},
		Name:      view_bookmark.getname(),
		list:      new_customlist(false),
		code:      code,
		bookmark:  bookmark,
		yes:       yes,
	}
	ret.right_context = bk_menu_context{
		qk: ret,
	}
	ret.menuitem = []context_menu_item{
		{item: create_menu_item("Delete"), handle: func() {
			idnex := ret.fzf.get_data_index(ret.list.GetCurrentItem())
			if idnex < 0 {
				return
			}
			r := ret.data[idnex]
			ret.bookmark.delete(r)
		}},
	}
	ret.update_redraw()
	return ret
}

func (ret *bookmark_view) update_redraw() {
	go func() {
		GlobalApp.QueueUpdateDraw(func() {
			ret.loaddata()
		})
	}()
}

func (ret *bookmark_view) loaddata() {
	// main := ret.main
	ret.data = reload_bookmark_list(ret.bookmark, ret.list, func(i int) {

		ret.onclick(i)
	})
	ret.list.SetChangedFunc(func(i int, mainText, secondaryText string, shortcut rune) {
		loc := ret.data[i].loc
		ret.code.OpenFileHistory(loc.URI.AsPath().String(), &loc)
	})

	if ret.bookmark.changed == nil {
		ret.bookmark.changed = []bookmark_changed{ret}
	} else {
		ret.bookmark.changed = append(ret.bookmark.changed, ret)
	}
	ret.fzf = new_fzf_on_list(ret.list, true)

}

type bk_menu_context struct {
	qk *bookmark_view
}

func (menu bk_menu_context) on_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	return action, event
}

// getbox implements context_menu_handle.
func (menu bk_menu_context) getbox() *tview.Box {
	if menu.qk != nil {
		yes := menu.qk.yes
		if yes() {
			return menu.qk.list.Box
		}
	}
	return nil
}

// menuitem implements context_menu_handle.
func (menu bk_menu_context) menuitem() []context_menu_item {
	return menu.qk.menuitem
}
func (ret *bookmark_view) onclick(i int) {
	loc := ret.data[i].loc
	ret.list.SetCurrentItem(i)
	ret.code.OpenFileHistory(loc.URI.AsPath().String(), &loc)
}
