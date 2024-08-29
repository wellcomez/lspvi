package mainui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
)

type grepresult struct {
	data []grep_output
}
type grep_impl struct {
	result *grepresult
	grep   *gorep
	taskid int
}
type livewgreppicker struct {
	*prev_picker_impl
	list *customlist
	main *mainui
	impl *grep_impl
	qf   func(ref_with_caller) bool
}

// name implements picker.
func (pk *livewgreppicker) name() string {
	return "live grep"
}

// greppicker
type greppicker struct {
	*livewgreppicker
}

// UpdateQuery implements picker.
// Subtle: this method shadows the method (*livewgreppicker).UpdateQuery of greppicker.livewgreppicker.
func (g *greppicker) UpdateQuery(query string) {
	g.livewgreppicker.UpdateQuery(query)
}

// handle implements picker.
// Subtle: this method shadows the method (*livewgreppicker).handle of greppicker.livewgreppicker.
func (g *greppicker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return g.livewgreppicker.handle()
}

// name implements picker.
func (g *greppicker) name() string {
	return "grep word"
}

func (pk livewgreppicker) update_preview() {
	cur := pk.list.GetCurrentItem()
	if cur < len(pk.impl.result.data) {
		item := pk.impl.result.data[cur]
		pk.codeprev.Load(item.fpath)
		pk.codeprev.gotoline(item.lineNumber - 1)
	}
}

func (pk livewgreppicker) handle_key_override(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := pk.list.InputHandler()
	handle(event, setFocus)
	pk.update_preview()
	pk.update_title()
}

// handle implements picker.
func (pk *livewgreppicker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return pk.handle_key_override
}

func (pk *livewgreppicker) grid(input *tview.InputField) *tview.Flex {
	layout := pk.prev_picker_impl.flex(input, 1)
	pk.list_click_check.on_list_selected = func() {
		pk.update_preview()
		pk.update_title()
	}
	return layout
}
func new_grep_picker(v *fzfmain) *greppicker {
	grep := &greppicker{
		livewgreppicker: new_live_grep_picker(v),
	}
	return grep
}
func new_live_grep_picker(v *fzfmain) *livewgreppicker {
	main := v.main
	x := new_preview_picker(v)
	grep := &livewgreppicker{
		prev_picker_impl: x,
		list:             new_customlist(),
		main:             main,
		impl:             &grep_impl{},
	}
	x.use_cusutom_list(grep.list)
	v.Visible = true
	return grep
}
func (grepx *livewgreppicker) update_title() {
	s := fmt.Sprintf("Grep %s %d/%d", grepx.list.Key, grepx.list.GetCurrentItem(), len(grepx.impl.result.data))
	grepx.parent.Frame.SetTitle(s)
}

func (grepx *livewgreppicker) end(task int, o *grep_output) {
	if task != grepx.impl.taskid {
		return
	}
	grep := grepx.impl
	if grep.result == nil {
		grep.result = &grepresult{
			data: []grep_output{},
		}
	}
	// log.Printf("end %d %s", task, o.destor)
	if grepx.qf == nil {
		if !grepx.parent.Visible {
			grep.grep.abort()
			return
		}
		grepx.main.app.QueueUpdate(func() {
			openpreview := len(grep.result.data) == 0
			grep.result.data = append(grep.result.data, *o)
			path := strings.TrimPrefix(o.fpath, grepx.main.root)
			data := fmt.Sprintf("%s:%d %s", path, o.lineNumber, o.line)
			grepx.list.AddItem(data, "", func() {
				grepx.main.OpenFile(o.fpath, nil)
				grepx.parent.hide()
			})
			if openpreview {
				grepx.update_preview()
			}
			grepx.update_title()
			grepx.main.app.ForceDraw()
		})
	} else {
		start := lsp.Position{Line: o.lineNumber - 1, Character: 0}
		end := start
		ref := ref_with_caller{
			Loc: lsp.Location{
				URI: lsp.NewDocumentURI(o.fpath),
				Range: lsp.Range{
					Start: start,
					End:   end,
				},
			},
		}
		if !grepx.qf(ref) {
			grep.grep.abort()
		}
	}

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
	pk.impl.taskid++
	pk.list.Key = query
	pk.list.Clear()
	g, err := newGorep(pk.impl.taskid, query, &opt)
	if err != nil {
		return
	}
	impl := pk.impl
	if impl.grep != nil {
		impl.grep.abort()
	}
	impl.grep = g
	impl.result = &grepresult{}
	g.cb = pk.end
	chans := g.kick(pk.main.root)
	go g.report(chans, false)

}
