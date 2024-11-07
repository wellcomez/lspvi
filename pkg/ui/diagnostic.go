package mainui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type Diagnostic interface {
}
type editor_diagnostic struct {
	data  lsp.PublishDiagnosticsParams
	index int
}

func (e *editor_diagnostic) Next(yes bool) (ret lsp.Diagnostic, err error) {
	count := len(e.data.Diagnostics)
	if count == 0 {
		err = fmt.Errorf("empty")
		return
	}
	ret = e.data.Diagnostics[e.index]
	next := 1
	if !yes {
		next = -1
	}
	e.index = (count + next) % count
	return
}
func NewLspDiagnostic(diags lsp.PublishDiagnosticsParams) *editor_diagnostic {
	return &editor_diagnostic{
		data: diags,
	}
}

type project_diagnostic struct {
	data []editor_diagnostic
}

func (prj *project_diagnostic) Find(path string) (ret *editor_diagnostic) {
	for _, v := range prj.data {
		if v.data.URI.AsPath().String() == path {
			ret = &v
			return
		}
	}
	return
}

func (prj *project_diagnostic) Update(diags lsp.PublishDiagnosticsParams) {
	for i, v := range prj.data {
		if v.data.URI.AsPath().String() == diags.URI.AsPath().String() {
			if diags.IsClear {
				data := prj.data[:i]
				data = append(data, prj.data[i+1:]...)
				prj.data = data
				return
			} else {
				prj.data[i].data = diags
				return
			}
		}
	}
	if !diags.IsClear {
		prj.data = append(prj.data, *NewLspDiagnostic(diags))
	}
}
func hove_test(root *codetextview, move bool, pos mouse_event_pos, event *tcell.EventMouse) {
	if dia := root.code.Dianostic(); !dia.data.IsClear {
		var new_diagnos_hove = func(v lsp.Diagnostic, mouse tcell.EventMouse) {
			root.error = nil
			buff_loc := femto.Loc{
				X: pos.X,
				Y: pos.Y,
			}
			root.hover = new_hover(buff_loc, move, func() {
				msg := &LspTextView{
					Box:  tview.NewBox(),
					main: root.main,
				}
				tag := ErrorTag(v)
				ss := []string{
					fmt.Sprintf("%s %s %s %s ", tag, v.Message, v.Source, v.Code),
					fmt.Sprintf(" %s %d:%d ", filepath.Base(root.code.FileName()),
						v.Range.Start.Character,
						buff_loc.Y+1)}
				msg.Load(strings.Join(ss, "\n"), root.code.Path())
				go func() {
					root.main.App().QueueUpdate(func() {
						w := 0
						h := len(msg.lines)
						for _, v := range msg.lines {
							w = max(w, len(v))
						}
						ss := &mouse
						_, y := ss.Position()
						edit_x, _, _, _ := root.GetInnerRect()
						x := v.Range.Start.Character + int(root.lineNumOffset())
						msg.SetRect(x+edit_x, y+1, w, h)
						root.error = msg
					})
				}()
			})
		}
		for i := range dia.data.Diagnostics {
			v := dia.data.Diagnostics[i]
			var ok = v.Severity == lsp.DiagnosticSeverityError || v.Severity == lsp.DiagnosticSeverityWarning
			if ok && v.Range.Start.Line == pos.Y && !move {
				if hover := root.hover; hover == nil {
					new_diagnos_hove(v, *event)
				} else if hover.Pos.Y != pos.Y || hover.Abort {
					new_diagnos_hove(v, *event)
				}
				break
			}
		}
	}
}

func ErrorTag(v lsp.Diagnostic) string {
	tag := "I"
	switch v.Severity {
	case lsp.DiagnosticSeverityError:
		tag = "E"
	case lsp.DiagnosticSeverityHint:
		tag = "H"
	case lsp.DiagnosticSeverityWarning:
		tag = "W"
	case lsp.DiagnosticSeverityInformation:
		tag = "I"
	}
	return tag
}
func (c *codetextview) HideHoverIfChanged() {
	if c.hover != nil {
		if c.hover.Pos.Y != c.Cursor.Loc.Y {
			c.hover.Abort = true
			c.error = nil
		}
	}
}

type diagnospicker struct {
	*prev_picker_impl
	fzf *fzf_on_listview
	qk  *quick_view_data
}

// UpdateQuery implements picker.
func (d *diagnospicker) UpdateQuery(query string) {
	// panic("unimplemented")
	d.fzf.OnSearch(query, false)
	UpdateColorFzfList(d.fzf).SetCurrentItem(0)
}

// close implements picker.
func (d *diagnospicker) close() {
	// panic("unimplemented")
}

// handle implements picker.
func (d *diagnospicker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyUp, tcell.KeyDown, tcell.KeyEnter:
			d.listcustom.InputHandler()(event, setFocus)
		}
	}
}

// name implements picker.
func (d *diagnospicker) name() string {
	return ("Diagnostics")
}

func new_diagnospicker_picker(dialog *fzfmain) (pk *diagnospicker) {
	x := new_preview_picker(dialog)
	pk = &diagnospicker{prev_picker_impl: x}
	x.listcustom = new_customlist(false)
	main := dialog.main
	dia := main.Dialogsize()
	var refs []ref_with_caller

	for _, v := range dia.data {
		for _, d := range v.data.Diagnostics {
			tag := ErrorTag(d)
			var ref = ref_with_caller{
				Loc: lsp.Location{
					URI:   v.data.URI,
					Range: d.Range,
				},
				Caller: &lspcore.CallStackEntry{
					Name: fmt.Sprintf("%s %s %s %s", tag, d.Message, d.Code, d.Source),
					Item: lsp.CallHierarchyItem{Kind: lsp.SymbolKindEnum},
				},
			}
			refs = append(refs, ref)
		}
	}
	var qkdata = new_quikview_data(main, data_grep_word, "", &SearchKey{&lspcore.SymolSearchKey{Key: "errrors"}, nil}, refs, false)
	qkdata.tree_to_listemitem()
	tree := qkdata.build_flextree_data(5)
	data := tree.ListColorString()
	for _, s := range data {
		pk.listcustom.AddColorItem(s.line, nil, nil)
	}
	pk.qk = qkdata
	next_index := -1
	pk.listcustom.SetChangedFunc(func(index int, s1, s2 string, r rune) {
		next_index = index
		if data, err := qkdata.get_data(index); err == nil {
			pk.PrevOpen(data.Loc.URI.AsPath().String(), data.Loc.Range.Start.Line)
		}
	})
	pk.listcustom.SetSelectedFunc(func(index int, s1, s2 string, r rune) {
		if index != next_index {
			return
		}
		if data, err := qkdata.get_data(index); err == nil {
			pk.parent.open_in_edior(data.Loc)
		}
	})
	sss := []string{}
	for _, v := range data {
		sss = append(sss, v.plaintext())
	}
	pk.fzf = new_fzf_on_list_data(pk.listcustom, sss, true)
	return
}
