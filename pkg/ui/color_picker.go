package mainui

import (
	"fmt"
	"log"
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
	c.fzf.OnSearch(query, true)
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
		new_fzflist_impl(nil, v),
		[]color_theme_file{},
	}
	ret := &color_picker{impl: impl, main: v.main}
	dirs, err := treesittertheme.GetTheme()
	if err == nil {
		for i := range dirs {
			d := dirs[i]
			a := color_theme_file{
				treesitter: true,
				name:       d}
			a.name = a.name[:strings.Index(a.name, ".")]
			ret.impl.data = append(ret.impl.data, a)
			impl.list.AddItem(fmt.Sprintf("%-30s *ts", a.name), "", func() {
				ret.on_select(&a)
			})
		}
		files := runtime.Files.ListRuntimeFiles(femto.RTColorscheme)
		for _, v := range files {
			a := color_theme_file{
				name:       v.Name(),
				treesitter: false,
			}
			ret.impl.data = append(ret.impl.data, a)
			impl.list.AddItem(fmt.Sprintf("%-30s ", a.name), "", func() {
				ret.on_select(&a)
			})
		}
	}
	ret.fzf = new_fzf_on_list(ret.impl.list, true)
	ret.fzf.selected = func(dataindex, listindex int) {
		a := ret.impl.data[dataindex]
		log.Println(a)
		ret.on_select(&a)
	}
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
