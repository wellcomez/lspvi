package mainui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type grepresult struct {
	data []grep_output
}
type greppicker struct {
	list     *customlist
	codeprev *CodeView
	main     *mainui
	taskid   int
	result   *grepresult
	parent   *fzfmain
}
func (pk greppicker) update_preview() {
}
func (pk greppicker) handle_key_override(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := pk.list.InputHandler()
	handle(event, setFocus)
	pk.update_preview()
}
// handle implements picker.
func (pk *greppicker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return pk.handle_key_override
}

func (pk *greppicker) new_view(input *tview.InputField) *tview.Grid {
	list := pk.list
	list.SetBorder(true)
	code := pk.codeprev.view
	layout := layout_list_edit(list, code, input)
	return layout
}
func new_greppicker(v *fzfmain) *greppicker {
	main := v.main
	grep := &greppicker{
		list:     new_customlist(),
		codeprev: NewCodeView(main),
		parent:   v,
		main:     main,
	}
	return grep
}
func (grep *greppicker) end(task int, o *grep_output) {
	if task != grep.taskid {
		return
	}
	if grep.result == nil {
		grep.result = &grepresult{
			data: []grep_output{},
		}
	}
	grep.result.data = append(grep.result.data, *o)
	grep.main.app.QueueUpdate(func() {
		grep.list.AddItem(o.destor, []int{}, func() {})
	})

}
func (grep greppicker) UpdateQuery(query string) {
	opt := optionSet{
		// grep_only:true,
		g: true,
	}
	grep.taskid++
	g, err := newGorep(grep.taskid, query, &opt)
	if err != nil {
		return
	}
	g.cb = grep.end
	chans := g.kick(grep.main.root)
	g.report(chans, false)

}
