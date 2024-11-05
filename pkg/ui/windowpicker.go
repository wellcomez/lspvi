package mainui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type uipicker struct {
	*fzflist_impl
	fzf *fzf_on_listview
}

// UpdateQuery implements picker.
func (u *uipicker) UpdateQuery(query string) {
	// panic("unimplemented")
	u.fzf.OnSearch(query, false)
	UpdateColorFzfList(u.fzf).SetCurrentItem(0)
}
func (pk *uipicker) grid(input *tview.InputField) *tview.Grid {
	return pk.fzflist_impl.grid(input)
}

// close implements picker.
func (u *uipicker) close() {
	// panic("unimplemented")
}

// handle implements picker.
func (u *uipicker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	// panic("unimplemented")
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		u.list.InputHandler()(event, setFocus)
	}
}

// name implements picker.
func (u *uipicker) name() string {
	return "ui"
}

func new_uipciker(dialog *fzfmain) (ret *uipicker) {
	ret = &uipicker{
		fzflist_impl: new_fzflist_impl(dialog),
	}
	windows_id := []view_id{view_outline_list, view_file, view_cmd}
	windows_id = append(windows_id, tab_view_id...)
	a := []view_id{}
	for k, _ := range SplitCode.code_collection {
		a = append(a, k)
	}
	windows_id = append(a, windows_id...)
	data := []string{}
	for _, v := range windows_id {
		link := dialog.main.to_view_link(v)
		hide := ""
		if link != nil && link.Hide {
			hide = "(hidden)"
		}
		if v >= view_code {
			var code = SplitCode.code_collection[v]
			active := ""
			if s := SplitCode.active_codeview; s != nil && s.vid() == v {
				active = "*"
			}
			name := fmt.Sprintf("%d %s%s", v-view_code+1, code.Path(), active) + " " + hide
			ret.list.AddItem(name, "", nil)
			data = append(data, name)
		} else {
			x := v.getname() + " " + hide
			ret.list.AddItem(x, "", nil)
			data = append(data, x)
		}
	}
	ret.fzf = new_fzf_on_list_data(ret.list, data, true)
	ret.list.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		var vid = windows_id[i]
		var is_tab = false
		for _, a := range tab_view_id {
			if a == vid {
				is_tab = true
				break

			}
		}
		if link := dialog.main.to_view_link(vid); link != nil && link.Hide {
			dialog.main.toggle_view(vid)
		}
		if is_tab {
			dialog.main.ActiveTab(vid, true)
		} else {
			dialog.main.set_viewid_focus(vid)
			if vid >= view_code {
				code := SplitCode.code_collection[vid]
				code.Acitve()
				SplitCode.active_codeview = code
			}
		}
		dialog.main.App().ForceDraw()
		dialog.hide()
	})
	return
}
