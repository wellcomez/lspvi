// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"fmt"
	// "log"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
	"github.com/pgavlin/femto/runtime"
	"github.com/rivo/tview"
	"zen108.com/lspvi/pkg/treesittertheme"
)

type color_theme_file struct {
	treesitter bool
	// filename   string
	name string
}
type color_pick_impl struct {
	*fzflist_impl
	data []color_theme_file
}
type theme_changer interface {
	on_change_color(name string)
}
type color_picker struct {
	impl *color_pick_impl
	fzf  *fzf_on_listview
	main theme_changer
	// code *CodeView
}

// close implements picker.
func (pk *color_picker) close() {
	// panic("unimplemented")
}

func (pk *color_picker) grid(input *tview.InputField) *tview.Grid {
	return pk.impl.grid(input)
}

// UpdateQuery implements picker.
func (c *color_picker) UpdateQuery(query string) {
	c.fzf.OnSearch(query, false)
	UpdateColorFzfList(c.fzf)
}
func (pk color_picker) handle_key_override(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := pk.impl.list.InputHandler()
	handle(event, setFocus)
}

func (pk color_picker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return pk.handle_key_override
}

func (c *color_picker) name() string {
	return "color picker"
}

func new_color_picker(v *fzfmain) *color_picker {
	impl := &color_pick_impl{
		new_fzflist_impl(v),
		[]color_theme_file{},
	}
	ret := &color_picker{impl: impl, main: v.main}
	dirs, err := treesittertheme.GetTheme()
	fzfdata := []string{}
	if err == nil {
		index := 1
		for i := range dirs {
			d := dirs[i]
			a := color_theme_file{
				treesitter: true,
				name:       d}
			a.name = a.name[:strings.Index(a.name, ".")]
			ret.impl.data = append(ret.impl.data, a)
			x := fmt.Sprintf("%-4d. %-30s *ts", index, a.name)
			fzfdata = append(fzfdata, x)
			impl.list.AddItem(x, "", func() {
				ret.on_select(&a)
			})
			index++
		}
		files := runtime.Files.ListRuntimeFiles(femto.RTColorscheme)
		for _, v := range files {
			a := color_theme_file{
				name:       v.Name(),
				treesitter: false,
			}
			ret.impl.data = append(ret.impl.data, a)
			x := fmt.Sprintf("%-4d. %-30s ", index, a.name)
			fzfdata = append(fzfdata, x)
			impl.list.AddItem(x, "", func() {
				ret.on_select(&a)
			})
			index++
		}
	}
	ret.fzf = new_fzf_on_list_data(ret.impl.list, fzfdata, true)
	impl.list.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		a := ret.impl.data[ret.fzf.get_data_index(i)]
		ret.on_select(&a)
	})
	for i, v := range ret.impl.data {
		if v.name == global_theme.name {
			ret.impl.list.SetCurrentItem(i)
			break
		}
	}
	return ret
}

func (pk *color_picker) on_select(c *color_theme_file) {
	pk.main.on_change_color(c.name)
	pk.impl.parent.hide()
}
