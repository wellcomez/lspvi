package mainui

import (
	"encoding/json"
	"os"

	"github.com/rivo/tview"
)

type LineMark struct {
	Line int
	Text string
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
func (b *bookmarkfile) Add(line int, text string, add bool) {
	if add {
		b.LineMark = append(b.LineMark, LineMark{Line: line, Text: text})

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
type bookmark_picker_impl struct {
	*prev_picker_impl
}

// new_bookmark_picker
func new_bookmark_picker(v *fzfmain) bookmark_picker {
	main := v.main
	sym := bookmark_picker{
		impl: &bookmark_picker_impl{
			new_prev_picker(main,v),
		},
	}
	sym.impl.codeprev.view.SetBorder(true)
	return sym
}
func (pk bookmark_picker) update_preview() {
	cur := pk.impl.listview.GetCurrentItem()
	if cur < len(pk.impl.current_list_data) {
		item := pk.impl.current_list_data[cur]
		pk.impl.codeprev.Load(item.loc.URI.AsPath().String())
		pk.impl.codeprev.gotoline(item.loc.Range.Start.Line)
	}
}
func (pk *bookmark_picker) grid(input *tview.InputField) *tview.Grid {
	list := pk.impl.listview
	list.SetBorder(true)
	code := pk.impl.codeprev.view
	layout := layout_list_edit(list, code, input)
	pk.impl.list_click_check = NewGridListClickCheck(layout, list, 2)
	pk.impl.list_click_check.on_list_selected = func() {
		pk.update_preview()
	}
	// , func() {
	// 	sym.impl.parent.hide()
	// })
	return layout
}
