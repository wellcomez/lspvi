package mainui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"zen108.com/lspvi/pkg/mason"
	"zen108.com/lspvi/pkg/ui/common"
)

var software *mason.SoftManager

type softwarepicker struct {
	*fzflist_impl
	fzf *fzf_on_listview
	app []mason.SoftwareTask
}

// UpdateQuery implements picker.
func (u *softwarepicker) UpdateQuery(query string) {
	// panic("unimplemented")
	u.fzf.OnSearch(query, false)
	UpdateColorFzfList(u.fzf)
	u.list.Clear()
	u.updatelist(u.fzf.selected_index)
}
func (pk *softwarepicker) grid(input *tview.InputField) *tview.Grid {
	return pk.fzflist_impl.grid(input)
}

// close implements picker.
func (u *softwarepicker) close() {
	// panic("unimplemented")
}

// handle implements picker.
func (u *softwarepicker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	// panic("unimplemented")

	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyDown, tcell.KeyUp:
			u.list.InputHandler()(event, setFocus)
		}
	}
}

// name implements picker.
func (u *softwarepicker) name() string {
	return "ui"
}

func NewSoftwarepciker(dialog *fzfmain) (ret *softwarepicker, err error) {
	if software == nil {
		if k, err := common.NewMkWorkdir(global_prj_root); err == nil {
			// software = mason.NewSoftManager(wk)
			software = mason.NewSoftManager(k)
		}
	}
	if software == nil {
		err = fmt.Errorf("can not create workdir")
		return
	}
	ret = &softwarepicker{
		fzflist_impl: new_fzflist_impl(dialog),
	}
	ret.app = software.GetAll()
	data := []string{}
	selectindex := []int{}
	for i := range ret.app {
		v := ret.app[i]
		selectindex = append(selectindex, i)
		data = append(data, v.Config.Name)
	}
	ret.updatelist(selectindex)
	ret.fzf = new_fzf_on_list_data(ret.list, data, true)
	last := -1
	selected := -11
	ret.list.SetChangedFunc(func(i int, s1, s2 string, r rune) {
		last = i
	})
	ret.list.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		if selected == i {
			return
		}
		if last == i {
			ret.run_start_i(i)
			selected = i
		}
	})
	return
}

func (ret *softwarepicker) updatelist(selectindex []int) {
	for i := range selectindex {
		v := ret.app[i]
		s := v.TaskState("")
		var c colorstring
		c.add_color_text(colortext{text: s})
		ret.list.AddColorItem(c.line, nil, nil)
	}
	ret.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'i' {
			x := ret.list.GetCurrentItem()
			ret.run_start_i(x)
			return nil
		}
		return event
	})
}

func (ret *softwarepicker) run_start_i(x int) {
	dataindex := x
	if i := ret.fzf.get_data_index(x); i != -1 {
		dataindex = i
	}
	a := ret.app[dataindex]
	set_text := func(s string, active int) {
		go ret.parent.main.App().QueueUpdateDraw(func() {
			ret.parent.main.update_log_view(s)
		})
		var c colorstring
		text := a.TaskState(s)
		if active == 1 {
			c.add_color_text(colortext{text: text, color: tcell.ColorGreen, bg: tcell.ColorWhite})
			ret.list.active_index = append(ret.list.active_index, x)
			ret.list.SetColorItem(x, c.line, nil)
		} else if active == -1 {
			c.add_color_text(colortext{text: text})
			var ss = []int{}
			for _, v := range ret.list.active_index {
				if v != x {
					ss = append(ss, v)
				}
			}
			ret.list.active_index = ss
			ret.list.SetColorItem(x, c.line, nil)
		}
		refreshlist(ret)
	}
	set_text("Starting", 1)
	software.Start(&a, func(s string) {
		set_text(s, 0)
		refreshlist(ret)
	}, func(i mason.InstallResult, err error) {
		if err != nil {
			set_text(err.Error(), -1)
		} else {
			set_text("", -1)
		}
		refreshlist(ret)
	})
}

func refreshlist(ret *softwarepicker) {
	go ret.parent.main.App().QueueUpdate(func() {
		ret.parent.main.App().ForceDraw()
	})
}

// func TaskState(v mason.SoftwareTask) string {
// 	status := " Not installed"
// 	check := rune_string(nerd.Nf_seti_checkbox_unchecked)
// 	yes, _ := v.GetBin()
// 	installed := ">[?]"
// 	if len(yes.Path) > 0 {
// 		installed = ">" + yes.Path
// 		check = rune_string(nerd.Nf_seti_checkbox)
// 	}
// 	download := ""
// 	if !yes.DownloadOk {
// 		download = ">" + rune_string(nerd.Nf_fa_download) + " " + yes.Url
// 	} else {
// 		download = yes.Download
// 	}
// 	status = fmt.Sprintf("%s %s", installed, download)
// 	return fmt.Sprintf("%s %s %s %s", check, v.Icon.Icon, v.Config.Name, status)
// }
