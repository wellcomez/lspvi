package mainui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"zen108.com/lspvi/pkg/mason"
	"zen108.com/lspvi/pkg/ui/common"
	nerd "zen108.com/lspvi/pkg/ui/icon"
)

var software *mason.SoftManager

type softwarepicker struct {
	*fzflist_impl
	fzf *fzf_on_listview
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
	apps := software.GetAll()
	ret = &softwarepicker{
		fzflist_impl: new_fzflist_impl(dialog),
	}
	data := []string{}
	for i := range apps {
		v := apps[i]
		status := TaskState(v)
		ret.list.AddItem(fmt.Sprintf("%s %s", v.Config.Name, status), "", nil)
		data = append(data, v.Config.Name)
	}
	ret.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'i' {
			x := ret.list.GetCurrentItem()
			a := apps[x]
			software.Start(&a, func(s string) {
				ret.list.SetItemText(x, fmt.Sprintf("%s %s", a.Config.Name, s), "")
				go dialog.main.App().QueueUpdate(func() {
					dialog.main.App().ForceDraw()
				})
			}, func(i mason.InstallResult, err error) {
				ret.list.SetItemText(x, fmt.Sprintf("%s %s", a.Config.Name, TaskState(a)), "")
			})
			return nil
		}
		return event
	})
	ret.fzf = new_fzf_on_list_data(ret.list, data, true)
	ret.list.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
	})
	return
}

func TaskState(v mason.SoftwareTask) string {
	status := " Not installed"
	if yes, err := v.GetBin(); err == nil {
		installed := ">[X]"
		if len(yes.Path) > 0 {
			installed = " > " + yes.Path
		}
		download := ""
		if !yes.DownloadOk {
			download = ">" + rune_string(nerd.Nf_fa_download)
		} else {
			download = ">" + yes.Download
		}
		status = fmt.Sprintf("%s %s", installed, download)
	}
	return status
}
