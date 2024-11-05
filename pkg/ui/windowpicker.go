package mainui

import (
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

func new_uipciker(v *fzfmain) (ret *uipicker) {
	ret = &uipicker{
		fzflist_impl: new_fzflist_impl(v),
	}
	windows_id := []view_id{view_outline_list}
	windows_id = append(windows_id, tab_view_id...)
	for k, _ := range SplitCode.code_collection {
		windows_id = append(windows_id, k)
	}
	data := []string{}
	for _, v := range windows_id {
		if v >= view_code {
			var code = SplitCode.code_collection[v]
			name := code.Path()
			ret.list.AddItem(name, "", nil)
			data = append(data, name)
		} else {
			ret.list.AddItem(v.getname(), "", nil)
			data = append(data, v.getname())
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
		if is_tab {
			v.main.ActiveTab(vid, true)
		} else {
			v.main.set_viewid_focus(vid)
			if vid >= view_code {
				SplitCode.code_collection[vid].Acitve()
			}
		}
		v.hide()
	})
	return
}
