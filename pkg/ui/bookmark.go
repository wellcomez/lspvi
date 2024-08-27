package mainui

import (
	"encoding/json"
	"fmt"

	// "log"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	fzflib "github.com/reinhrst/fzf-lib"
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
	query = strings.ToLower(query)
	listview := pk.impl.listview
	listview.Clear()
	var result fzflib.SearchResult
	if fzf := pk.impl.fzf; fzf != nil {
		fzf.Search(query)
		result = <-fzf.GetResultChannel()
		pk.impl.current_list_data = []ref_line{}
		for _, v := range result.Matches {
			v := pk.impl.listdata[v.HayIndex]
			pk.impl.current_list_data = append(pk.impl.current_list_data, v)
			a, b := get_list_item(v)
			listview.AddItem(
				a, b,
				0, func() {
					pk.impl.codeprev.main.OpenFile(v.loc.URI.AsPath().String(), &v.loc)
					pk.impl.parent.hide()
				})
		}
	} else {
		pk.impl.current_list_data = []ref_line{}
		for i := range pk.impl.listdata {
			v := pk.impl.listdata[i]
			pk.impl.current_list_data = append(pk.impl.current_list_data, v)
			if strings.Contains(strings.ToLower(v.line), query) {
				a, b := get_list_item(v)
				listview.AddItem(a, b, 0, func() {
					pk.impl.codeprev.main.OpenFile(v.loc.URI.AsPath().String(), &v.loc)
					pk.impl.parent.hide()
				})
			}
		}
	}
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
	fzf *fzflib.Fzf
}

func get_list_item(v ref_line) (string, string) {
	return v.line + ":" + v.path, v.caller
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
	sym := bookmark_picker{
		impl: &bookmark_picker_impl{
			prev_picker_impl: new_preview_picker(v),
			fzf:              nil,
		},
	}
	sym.impl.codeprev.view.SetBorder(true)
	marks := v.main.bookmark.Bookmark
	impl := sym.impl
	for _, file := range marks {
		for _, v := range file.LineMark {
			ref := ref_line{line: fmt.Sprintf("%d", v.Line), path: file.Name, caller: v.Comment + v.Text, loc: lsp.Location{
				URI:   lsp.NewDocumentURI(file.Name),
				Range: lsp.Range{Start: lsp.Position{Line: v.Line - 1}},
			},
			}
			impl.listdata = append(impl.listdata, ref)
		}
	}
	impl.current_list_data = impl.listdata
	datafzf := []string{}
	for _, v := range impl.listdata {
		datafzf = append(datafzf, v.path+":"+v.line+v.caller)
		a, b := get_list_item(v)
		impl.listview.AddItem(a, b, 0, nil)
	}
	option := fzflib.DefaultOptions()
	option.CaseMode = fzflib.CaseIgnore
	sym.impl.fzf = fzflib.New(datafzf, option)
	return sym
}
func (pk bookmark_picker) update_preview() {
	pk.impl.update_preview()
}
func (pk *bookmark_picker) grid(input *tview.InputField) *tview.Flex {
	return pk.impl.flex(input, 2)
}
