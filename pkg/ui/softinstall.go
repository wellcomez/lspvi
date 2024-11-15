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
	UpdateColorFzfList(u.fzf).SetCurrentItem(0)
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
	ret.list.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
	})
	return
}

func (ret *softwarepicker) updatelist(selectindex []int) {
	for i := range selectindex {
		v := ret.app[i]
		s := v.TaskState()
		ret.list.AddItem(s, "", nil)
	}
	ret.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'i' {
			x := ret.list.GetCurrentItem()
			a := ret.app[x]
			software.Start(&a, func(s string) {
				ret.list.SetItemText(x, a.TaskState()+s, "")
				go ret.parent.main.App().QueueUpdate(func() {
					ret.parent.main.App().ForceDraw()
				})
			}, func(i mason.InstallResult, err error) {
				ret.list.SetItemText(x, a.TaskState(), "")
			})
			return nil
		}
		return event
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
