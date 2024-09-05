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
	result        *grepresult
	grep          *gorep
	taskid        int
	fzf_on_result *fzf_on_listview
}
type livewgreppicker struct {
	*prev_picker_impl
	grep_list_view *customlist
	main           *mainui
	impl           *grep_impl
	qf             func(ref_with_caller) bool
	not_live       bool
}

// name implements picker.
func (pk *livewgreppicker) name() string {
	return "live grep"
}

// greppicker
type greppicker struct {
	*livewgreppicker
	query string
}

// UpdateQuery implements picker.
// Subtle: this method shadows the method (*livewgreppicker).UpdateQuery of greppicker.livewgreppicker.
func (g *greppicker) UpdateQuery(query string) {
	g.query = query
	if g.impl.fzf_on_result != nil {
		g.impl.fzf_on_result.OnSearch(query, true)
		g.grep_list_view.Key = query
	}
}

// handle implements picker.
// Subtle: this method shadows the method (*livewgreppicker).handle of greppicker.livewgreppicker.
func (g *greppicker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		focused := g.grep_list_view.HasFocus()
		if event.Key() == tcell.KeyEnter && !focused {
			g.impl.fzf_on_result = nil
			g.parent.input.SetText(">")
			g.livewgreppicker.UpdateQuery(g.query)
		} else {
			var key = event.Key()
			if key == tcell.KeyDown || key == tcell.KeyUp {
				g.grep_list_view.List.Focus(nil)
			}
			g.livewgreppicker.handle_key_override(event, nil)
		}
	}
}

// name implements picker.
func (g *greppicker) name() string {
	return "grep word"
}

func (pk livewgreppicker) update_preview() {
	cur := pk.grep_list_view.GetCurrentItem()
	if pk.impl.fzf_on_result != nil {
		cur = pk.impl.fzf_on_result.get_data_index(cur)
	}
	if cur < len(pk.impl.result.data) {
		item := pk.impl.result.data[cur]
		pk.codeprev.Load(item.fpath)
		pk.codeprev.gotoline(item.lineNumber - 1)
	}
}

func (pk livewgreppicker) handle_key_override(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := pk.grep_list_view.InputHandler()
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
	grep.not_live = true
	return grep
}
func new_live_grep_picker(v *fzfmain) *livewgreppicker {
	main := v.main
	x := new_preview_picker(v)
	grep := &livewgreppicker{
		prev_picker_impl: x,
		grep_list_view:   new_customlist(false),
		main:             main,
		impl:             &grep_impl{},
		not_live:         false,
	}
	x.use_cusutom_list(grep.grep_list_view)
	v.Visible = true
	return grep
}
func (grepx *livewgreppicker) update_title() {
	index := grepx.grep_list_view.GetCurrentItem()

	x := len(grepx.impl.result.data)
	if x > 0 {
		index = index + 1
	}
	s := fmt.Sprintf("Grep %s %d/%d", grepx.grep_list_view.Key, index, x)
	grepx.parent.Frame.SetTitle(s)
}

func (grepx *livewgreppicker) end(task int, o *grep_output) {
	if task != grepx.impl.taskid {
		return
	}
	if o == nil {
		if grepx.not_live {
			grepx.impl.fzf_on_result = new_fzf_on_list(grepx.grep_list_view, true)
		}
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
			grepx.grep_list_view.AddItem(data, "", func() {
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
	if len(query) == 0 {
		return
	}
	opt := optionSet{
		grep_only: true,
		g:         true,
	}
	pk.impl.taskid++
	pk.grep_list_view.Key = query
	pk.grep_list_view.Clear()
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
