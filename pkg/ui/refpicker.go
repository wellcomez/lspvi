package mainui

import (
	// "strings"

	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	fzflib "github.com/reinhrst/fzf-lib"
	"github.com/rivo/tview"
	lsp "github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

func (pk *refpicker) grid(input *tview.InputField) *tview.Grid {
	list := pk.impl.listview
	list.SetBorder(true)
	code := pk.impl.codeprev.view
	pk.impl.codeprev.Load(pk.impl.file.Filename)
	layout := layout_list_edit(list, code, input)
	return layout
}

func layout_list_edit(list tview.Primitive, code tview.Primitive, input *tview.InputField) *tview.Grid {
	layout := tview.NewGrid().
		SetColumns(-1, 24, 16, -1).
		SetRows(-1, 3, 3, 2).
		AddItem(list, 0, 0, 3, 2, 0, 0, false).
		AddItem(code, 0, 2, 3, 2, 0, 0, false).
		AddItem(input, 3, 0, 1, 4, 0, 0, false)
	return layout
}

type refpicker_impl struct {
	file              *lspcore.Symbol_file
	listview          *tview.List
	codeprev          *CodeView
	refs              []ref_with_caller
	listdata          []ref_line
	current_list_data []ref_line
	codeline          []string
	fzf               *fzflib.Fzf
	parent            *fzfmain
	// ref_call_in       []ref_with_callin
}

type refpicker struct {
	impl *refpicker_impl
}

// OnCallInViewChanged implements lspcore.lsp_data_changed.
func (pk refpicker) OnCallInViewChanged(stacks []lspcore.CallStack) {
	panic("unimplemented")
}

// OnCallTaskInViewChanged implements lspcore.lsp_data_changed.
func (pk refpicker) OnCallTaskInViewChanged(stacks *lspcore.CallInTask) {
	panic("unimplemented")
}

// OnCallTaskInViewResovled implements lspcore.lsp_data_changed.
func (pk refpicker) OnCallTaskInViewResovled(stacks *lspcore.CallInTask) {
	panic("unimplemented")
}

// OnCodeViewChanged implements lspcore.lsp_data_changed.
func (pk refpicker) OnCodeViewChanged(file *lspcore.Symbol_file) {
	panic("unimplemented")
}

// OnFileChange implements lspcore.lsp_data_changed.
func (pk refpicker) OnFileChange(file []lsp.Location) {
	panic("unimplemented")
}

type ref_line struct {
	caller string
	loc    lsp.Location
	line   string
	path   string
}

func (ref ref_line) fzf_tring() string {
	return ref.String() + ref.caller
}
func (ref ref_line) String() string {
	return fmt.Sprintf("%s %s:%d", ref.line, ref.path, ref.loc.Range.Start.Line)
}

type ref_with_caller struct {
	loc    lsp.Location
	caller *lspcore.CallStackEntry
}

func (pk refpicker) OnRefenceChanged(ranges lsp.Range, file []lsp.Location) {
	pk.impl.listview.Clear()
	listview := pk.impl.listview
	datafzf := []string{}
	lsp := pk.impl.parent.main.lspmgr.Current
	pk.impl.refs = get_loc_caller(file, lsp)
	for i := range pk.impl.refs {
		caller := pk.impl.refs[i]
		v := caller.loc
		source_file_path := v.URI.AsPath().String()
		data, err := os.ReadFile(source_file_path)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		line := lines[v.Range.Start.Line]
		if len(line) == 0 {
			continue
		}
		gap := 40
		begin := max(0, v.Range.Start.Character-gap)
		end := min(len(line), v.Range.Start.Character+gap)
		path := strings.Replace(v.URI.AsPath().String(), pk.impl.codeprev.main.root, "", -1)
		callinfo := ""
		if caller.caller != nil {
			callinfo = caller_to_listitem(caller.caller, pk.impl.parent.main.root)
		}
		secondline := fmt.Sprintf("%s:%d%s", path, v.Range.Start.Line+1, callinfo)
		r := ref_line{
			caller: caller_to_listitem(caller.caller, pk.impl.parent.main.root),
			loc:    v,
			line:   line,
			path:   path,
		}
		pk.impl.listdata = append(pk.impl.listdata, r)
		datafzf = append(datafzf, r.fzf_tring())
		listview.AddItem(secondline, line[begin:end], 0, func() {
			pk.impl.codeprev.main.OpenFile(v.URI.AsPath().String(), &v)
			pk.impl.parent.hide()
		})
	}
	pk.impl.fzf = fzflib.New(datafzf, fzflib.DefaultOptions())
	pk.impl.current_list_data = pk.impl.listdata
	pk.update_preview()
}

func get_loc_caller(file []lsp.Location, lsp *lspcore.Symbol_file) []ref_with_caller {
	ref_call_in := []ref_with_caller{}
	for _, v := range file {
		stacks, err := lsp.Callin(v, false)
		if err == nil {
			var find *lspcore.CallStackEntry
			for _, s := range stacks {
				find = s.InRange(v)
				if find != nil {
					break
				}
			}
			ref_call_in = append(ref_call_in, ref_with_caller{
				loc:    v,
				caller: find,
			})
		}
	}
	return ref_call_in
}

// OnSymbolistChanged implements lspcore.lsp_data_changed.
func (ref refpicker) OnSymbolistChanged(file *lspcore.Symbol_file, err error) {
	panic("unimplemented")
}

func new_refer_picker(clone lspcore.Symbol_file, v *fzfmain) refpicker {
	main := v.main
	sym := refpicker{
		impl: &refpicker_impl{
			file:     &clone,
			listview: tview.NewList(),
			codeprev: NewCodeView(main),
			codeline: []string{},
			parent:   v,
		},
	}
	sym.impl.codeprev.view.SetBorder(true)
	return sym
}
func (pk *refpicker) load(ranges lsp.Range) {
	pk.impl.file.Handle = *pk
	pk.impl.file.Reference(ranges)
}
func (pk refpicker) handle_key_override(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := pk.impl.listview.InputHandler()
	handle(event, setFocus)
	pk.update_preview()
}

func (pk refpicker) update_preview() {
	cur := pk.impl.listview.GetCurrentItem()
	if cur < len(pk.impl.current_list_data) {
		item := pk.impl.current_list_data[cur]
		pk.impl.codeprev.Load(item.loc.URI.AsPath().String())
		pk.impl.codeprev.gotoline(item.loc.Range.Start.Line)
	}
}

// handle implements picker.
func (pk refpicker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return pk.handle_key_override
}
func (pk refpicker) UpdateQuery(query string) {
	query = strings.ToLower(query)
	listview := pk.impl.listview
	listview.Clear()
	if fzf := pk.impl.fzf; fzf != nil {
		fzf.Search(query)
		var result fzflib.SearchResult
		result = <-fzf.GetResultChannel()
		pk.impl.current_list_data = []ref_line{}
		for _, v := range result.Matches {
			v := pk.impl.listdata[v.HayIndex]
			pk.impl.current_list_data = append(pk.impl.current_list_data, v)
			callinfo := v.caller
			listview.AddItem(
				fmt.Sprintf("%s:%d%s", v.path, v.loc.Range.Start.Line, callinfo),
				v.line, 0, func() {
					pk.impl.codeprev.main.OpenFile(v.loc.URI.AsPath().String(), &v.loc)
					pk.impl.parent.hide()
				})
		}
	} else {
		pk.impl.current_list_data = []ref_line{}
		for i := range pk.impl.listdata {
			v := pk.impl.listdata[i]
			pk.impl.current_list_data = append(pk.impl.current_list_data, v)
			if strings.Contains(strings.ToLower(v.line), query) {
				listview.AddItem(v.path, v.line, 0, func() {
					pk.impl.codeprev.main.OpenFile(v.loc.URI.AsPath().String(), &v.loc)
					pk.impl.parent.hide()
				})
			}
		}
	}
	pk.update_preview()
}
