package mainui

import (
	"encoding/json"
	"fmt"
	// "log"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	// fzflib "github.com/reinhrst/fzf-lib"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
)

type LineMark struct {
	Line    int
	Text    string
	Comment string
}
type bookmarkfile struct {
	Name     string
	LineMark []LineMark
}
type proj_bookmark struct {
	Bookmark []bookmarkfile
	path     string
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
func (prj *proj_bookmark) save() error {
	buf, err := json.Marshal(prj)
	if err != nil {
		return err
	}
	return os.WriteFile(prj.path, buf, 0666)
}
func (prj *proj_bookmark) GetFileBookmark(file string) *bookmarkfile {
	for i, _ := range prj.Bookmark {
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

// UpdateQuery implements picker.
func (pk bookmark_picker) UpdateQuery(query string) {
	listview := pk.impl.hlist
	query = strings.ToLower(query)
	listview.Clear()

	listview.Key = query
	pk.impl.fzf.selected = func(i int) {
		loc := pk.impl.listdata[i].loc
		close_bookmark_picker(pk.impl.prev_picker_impl, loc)
	}
	pk.impl.fzf.OnSearch(query, true)
	pk.update_preview()
}
func (pk bookmark_picker) handle_key_override(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := pk.impl.listview.InputHandler()
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

func new_bookmark_editor(v *fzfmain, cb func(string)) bookmark_edit {
	main := v.main
	code := main.codeview
	var line = code.view.Cursor.Loc.Y + 1
	line1 := code.view.Buf.Line(line - 1)
	ret := bookmark_edit{
		fzflist_impl: new_fzflist_impl(nil, v),
		cb:           cb,
	}
	ret.fzflist_impl.list.AddItem(line1, code.filename, nil)
	v.create_dialog_content(ret.grid(v.input), ret)
	return ret
}

// new_bookmark_picker
func new_bookmark_picker(v *fzfmain) bookmark_picker {
	impl := &bookmark_picker_impl{
		prev_picker_impl: new_preview_picker(v),
	}
	sym := bookmark_picker{
		impl: impl,
	}
	sym.impl.codeprev.view.SetBorder(true)

	impl.hlist, impl.listdata = init_bookmark_list(v.main, func(i int) {
		loc := sym.impl.listdata[i].loc
		close_bookmark_picker(impl.prev_picker_impl, loc)
	})
	impl.hlist.fuzz = true
	impl.use_cusutom_list(impl.hlist)
	fzf := new_fzf_on_list(sym.impl.hlist, sym.impl.hlist.fuzz)
	sym.impl.fzf = fzf
	return sym
}

func init_bookmark_list(main *mainui, selected func(int)) (*customlist, []ref_line) {
	hlist := new_customlist()
	bookmark := main.bookmark
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
	root := main.root
	for i, v := range listdata {
		a, b := get_list_item(v, root)
		index := i
		hlist.AddItem(a, b, func() { selected(index) })
	}
	return hlist, listdata
}

func close_bookmark_picker(impl *prev_picker_impl, loc lsp.Location) {
	impl.open_location(loc)
	impl.parent.hide()
}
func (pk bookmark_picker) update_preview() {
	pk.impl.update_preview()
}
func (pk *bookmark_picker) grid(input *tview.InputField) *tview.Flex {
	return pk.impl.flex(input, 2)
}
