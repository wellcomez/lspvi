package mainui

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
	"github.com/pgavlin/femto/runtime"

	// fzflib "github.com/reinhrst/fzf-lib"
	"github.com/rivo/tview"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

func (pk *history_picker) grid(input *tview.InputField) *tview.Grid {
	return pk.impl.grid(input)
}

func grid_list_whole_screen(list tview.Primitive, input *tview.InputField) *tview.Grid {
	layout := tview.NewGrid().
		SetColumns(-1, 24, 16, -1).
		SetRows(-1, 3, 3, 2).
		AddItem(list, 0, 0, 3, 4, 0, 0, false).
		AddItem(input, 3, 0, 1, 4, 0, 0, false)

	return layout
}

type history_picker_impl struct {
	*fzflist_impl
	listdata []history_item
}

type history_picker struct {
	impl *history_picker_impl
	fzf  *fzf_on_listview
	// list      *customlist
	// listcheck *GridListClickCheck
}

type color_theme_file struct {
	treesitter bool
	filename   string
	name       string
}
type color_pick_impl struct {
	*fzflist_impl
	data []color_theme_file
}
type color_picker struct {
	impl *color_pick_impl
	fzf  *fzf_on_listview
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
	return "color_picker"
}

func new_color_picker(v *fzfmain) *color_picker {
	impl := &color_pick_impl{
		new_fzflist_impl(nil, v),
		[]color_theme_file{},
	}
	ret := &color_picker{impl: impl}
	dir := "colorscheme/output"
	dirs, err := TreesitterSchemeLoader.ReadDir(dir)
	if err == nil {
		for i := range dirs {
			d := dirs[i]
			a := color_theme_file{
				false,
				filepath.Join(dir, d.Name()),
				d.Name()}
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
	return ret
}

func (pk *color_picker) on_select(c *color_theme_file) {
	code := pk.impl.parent.main.codeview
	code.on_change_color(c)
	pk.impl.parent.hide()
}

// name implements picker.
func (pk history_picker) name() string {
	return "history"
}

// OnSymbolistChanged implements lspcore.lsp_data_changed.
func (ref history_picker) OnSymbolistChanged(file *lspcore.Symbol_file, err error) {
	panic("unimplemented")
}

type history_item struct {
	filepath string
	dispname string
}

func new_history_picker(v *fzfmain) history_picker {
	// list := new_customlist()
	// list.SetBorder(true)
	sym := history_picker{
		impl: &history_picker_impl{
			fzflist_impl: new_fzflist_impl(nil, v),
		},
	}
	sym.impl.set_fuzz(true)
	history := NewHistory(lspviroot.history)
	items := []history_item{}
	close := func(data_index int, listIndex int) {
		v := sym.impl.listdata[data_index]
		path := v.filepath
		parent := sym.impl.parent
		parent.openfile(path)
	}
	for i, h := range history.history_files() {

		dispname := strings.TrimPrefix(h, v.main.root)
		h := history_item{
			filepath: h,
			dispname: dispname,
		}
		index := i
		// fzf_item_strings = append(fzf_item_strings, dispname)
		sym.impl.list.AddItem(h.dispname, "", func() {
			close(index, index)
		})
		items = append(items, h)
	}
	sym.impl.listdata = items
	sym.fzf = new_fzf_on_list(sym.impl.list, true)
	sym.fzf.selected = close
	return sym
}
func (pk history_picker) handle_key_override(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := pk.impl.list.InputHandler()
	handle(event, setFocus)
	pk.update_preview()
}

func (pk history_picker) update_preview() {
}

// handle implements picker.
func (pk history_picker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return pk.handle_key_override
}
func (pk history_picker) UpdateQuery(query string) {
	if len(pk.impl.listdata) == 0 {
		return
	}
	query = strings.ToLower(query)
	listview := pk.impl.list
	listview.Clear()
	listview.Key = query
	pk.fzf.OnSearch(query, true)
}
