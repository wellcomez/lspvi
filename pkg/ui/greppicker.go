package mainui

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type grepresult struct {
	data []grep_output
}
type grep_impl struct {
	result        *grepresult
	temp          *grepresult
	grep          *gorep
	taskid        int
	key           string
	fzf_on_result *fzf_on_listview
}
type livewgreppicker struct {
	*prev_picker_impl
	grep_list_view *customlist
	main           *mainui
	impl           *grep_impl
	qf             func(bool, ref_with_caller) bool
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

// close implements picker.
func (g *greppicker) close() {
	g.livewgreppicker.close()
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
		var key = event.Key()
		if key == tcell.KeyEnter && !focused {
			g.impl.fzf_on_result = nil
			g.parent.input.SetLabel(">")
			g.livewgreppicker.UpdateQuery(g.query)
		} else if key == tcell.KeyCtrlS {
			g.Save()
		} else {
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
		if cur < 0 {
			return
		}
	}
	if pk.impl.result == nil {
		return
	}
	if cur < len(pk.impl.result.data) {
		item := pk.impl.result.data[cur]
		pk.codeprev.Load2Line(item.fpath, item.lineNumber-1)
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
func new_grep_picker(v *fzfmain,code *CodeView) *greppicker {
	grep := &greppicker{
		livewgreppicker: new_live_grep_picker(v,code),
	}
	grep.not_live = true
	return grep
}
func new_live_grep_picker(v *fzfmain,code *CodeView) *livewgreppicker {
	main := v.main
	x := new_preview_picker(v, code)
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
	if grepx.impl.result == nil {
		return
	}
	index := grepx.grep_list_view.GetCurrentItem()
	x := len(grepx.impl.result.data)
	if x > 0 {
		index = index + 1
	}
	s := fmt.Sprintf("Grep %s %d/%d", grepx.grep_list_view.Key, index, x)
	grepx.parent.update_dialog_title(s)
}

func (parent *fzfmain) update_dialog_title(s string) {
	UpdateTitleAndColor(parent.Frame.Box, s)
}
func (grepx *livewgreppicker) grep_to_list() {
	grep := grepx.impl
	openpreview := len(grep.result.data) == 0
	tmp := grep.temp
	if tmp == nil {
		return
	}
	grep.temp = nil
	grep.result.data = append(grep.result.data, tmp.data...)
	for _, o := range tmp.data {
		path := strings.TrimPrefix(o.fpath, grepx.main.root)
		data := fmt.Sprintf("%s:%d %s", path, o.lineNumber, o.line)
		grepx.grep_list_view.AddItem(data, "", func() {
			loc := convert_grep_info_location(&o)
			grepx.main.OpenFile(o.fpath, &loc)
			grepx.parent.hide()
		})
	}
	if openpreview {
		grepx.update_preview()
	}
	grepx.update_title()
	grepx.main.app.ForceDraw()
}
func (grepx *livewgreppicker) end(task int, o *grep_output) {
	if task != grepx.impl.taskid {
		return
	}
	if o == nil {
		if grepx.qf != nil {
			grepx.qf(true, ref_with_caller{})
		} else if grepx.not_live {
			grepx.main.app.QueueUpdate(func() {
				grepx.grep_to_list()
				grepx.impl.fzf_on_result = new_fzf_on_list(grepx.grep_list_view, true)
				grepx.impl.fzf_on_result.selected = func(dataindex, listindex int) {
					o := grepx.impl.result.data[dataindex]
					loc := convert_grep_info_location(&o)
					grepx.main.OpenFile(o.fpath, &loc)
					grepx.parent.hide()
				}
			})
		}
		return
	}
	grep := grepx.impl
	if grep.result == nil {
		grep.result = &grepresult{
			data: []grep_output{},
		}
	}
	if grep.temp == nil {
		grep.temp = &grepresult{
			data: []grep_output{},
		}
	}
	// log.Printf("end %d %s", task, o.destor)
	if grepx.qf == nil {
		if !grepx.parent.Visible {
			grep.grep.abort()
			return
		}
		grep.temp.data = append(grep.temp.data, *o)
		if len(grep.result.data) > 10 {
			if len(grep.temp.data) < 50 {
				return
			}
		}
		grepx.main.app.QueueUpdate(func() {
			grepx.grep_to_list()
		})
	} else {
		ref := o.to_ref_caller()
		if !grepx.qf(false, ref) {
			grep.grep.abort()
		}
	}

}

func convert_grep_info_location(o *grep_output) lsp.Location {
	loc := lsp.Location{
		URI: lsp.NewDocumentURI(o.fpath),
		Range: lsp.Range{
			Start: lsp.Position{Line: o.lineNumber - 1, Character: 0},
			End:   lsp.Position{Line: o.lineNumber - 1, Character: 0},
		},
	}
	return loc
}

func (o *grep_output) to_ref_caller() ref_with_caller {
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
	return ref
}

func (pk livewgreppicker) Save() {
	Result := search_reference_result{}
	data := qf_history_data{Type: data_grep_word,
		Key:  lspcore.SymolSearchKey{Key: pk.impl.key},
		Date: time.Now().Unix(),
	}
	for _, v := range pk.impl.result.data {
		Result.Refs = append(Result.Refs, v.to_ref_caller())
	}
	data.Result = Result
	main := pk.main
	save_qf_uirefresh(main, data)
}

func (pk livewgreppicker) UpdateQuery(query string) {
	// pk.defer_input.start(query)
	pk.__updatequery(query)
}

// close implements picker.
func (pk *livewgreppicker) close() {
	pk.stop_grep()
}

func (pk *livewgreppicker) __updatequery(query string) {
	if len(query) == 0 {
		pk.stop_grep()
		return
	}
	opt := optionSet{
		grep_only: true,
		g:         true,
		wholeword: true,
	}
	pk.impl.taskid++
	pk.impl.key = query
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

func (pk livewgreppicker) stop_grep() {
	if pk.impl != nil && pk.impl.grep != nil {
		pk.impl.grep.abort()
	}
}
