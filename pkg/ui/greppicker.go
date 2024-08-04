package mainui

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type grepresult struct {
	data []grep_output
}
type livewgreppicker struct {
	list        *customlist
	codeprev    *CodeView
	main        *mainui
	taskid      int
	result      *grepresult
	parent      *fzfmain
	grep        *gorep
	defer_input *key_defer_input
}

func (pk livewgreppicker) update_preview() {
}
func (pk livewgreppicker) handle_key_override(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := pk.list.InputHandler()
	handle(event, setFocus)
	pk.update_preview()
}

// handle implements picker.
func (pk *livewgreppicker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return pk.handle_key_override
}

func (pk *livewgreppicker) new_view(input *tview.InputField) *tview.Grid {
	list := pk.list
	list.SetBorder(true)
	code := pk.codeprev.view
	layout := layout_list_edit(list, code, input)
	return layout
}
func new_live_grep_picker(v *fzfmain) *livewgreppicker {
	main := v.main
	grep := &livewgreppicker{
		list:        new_customlist(),
		codeprev:    NewCodeView(main),
		parent:      v,
		main:        main,
		defer_input: &key_defer_input{},
	}
	grep.defer_input.cb = grep.__updatequery
	return grep
}
func (grep *livewgreppicker) end(task int, o *grep_output) {
	if task != grep.taskid {
		return
	}
	if grep.result == nil {
		grep.result = &grepresult{
			data: []grep_output{},
		}
	}
	// log.Printf("end %d %s", task, o.destor)
	grep.result.data = append(grep.result.data, *o)
	if !grep.parent.Visible {
		grep.grep.abort()
		return
	}
	grep.main.app.QueueUpdate(func() {
		grep.list.AddItem(o.destor, []int{}, func() {})
		grep.main.app.ForceDraw()
	})

}

type key_defer_input struct {
	buf_three_char string
	cb             func(string)
}

func (k *key_defer_input) __start() {
	after := time.After(200 * time.Millisecond)
	<-after
	if k.cb != nil {
		k.cb(k.empty_buffer())
	}
}
func (k *key_defer_input) start(key string) (string, bool) {
	k.buf_three_char = k.buf_three_char + key
	if len(k.buf_three_char) >= 3 {
		a := k.empty_buffer()
		return a, true
	}
	go k.__start()
	return "", false
}

func (k *key_defer_input) empty_buffer() string {
	a := k.buf_three_char
	k.buf_three_char = ""
	return a
}

func (pk livewgreppicker) UpdateQuery(query string) {
	// pk.defer_input.start(query)
	pk.__updatequery(query)
}
func (pk livewgreppicker) __updatequery(query string) {
	opt := optionSet{
		grep_only: true,
		g:         true,
	}
	pk.taskid++
	pk.list.Key = query
	pk.list.Clear()
	g, err := newGorep(pk.taskid, query, &opt)
	if err != nil {
		return
	}
	if pk.grep != nil {
		pk.grep.abort()
	}
	pk.grep = g
	g.cb = pk.end
	chans := g.kick(pk.main.root)
	go g.report(chans, false)

}
