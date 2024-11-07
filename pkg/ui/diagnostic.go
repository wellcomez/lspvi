package mainui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
)

type Diagnostic interface {
}
type editor_diagnostic struct {
	data lsp.PublishDiagnosticsParams
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
func hove_test(root *codetextview, pos mouse_event_pos, event *tcell.EventMouse) {
	if dia := root.code.Dianostic(); !dia.data.IsClear {
		var new_diagnos_hove = func(v lsp.Diagnostic, mouse tcell.EventMouse) {
			root.error = nil
			buff_loc := femto.Loc{
				X: pos.X,
				Y: pos.Y,
			}
			root.hover = new_hover(buff_loc, func() {
				msg := &LspTextView{
					Box:  tview.NewBox(),
					main: root.main,
				}
				ss := []string{" " + v.Message + " ", fmt.Sprintf(" %s %d:%d", root.filename, buff_loc.X, buff_loc.Y+1), v.Source}
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
			if ok && v.Range.Start.Line == pos.Y {
				if hover := root.hover; hover == nil {
					new_diagnos_hove(v, *event)
				} else if hover.Pos.Y != pos.Y||hover.Abort {
					new_diagnos_hove(v, *event)
				}
				break
			}
		}
	}
}
func (c *codetextview) HideHoverIfChanged () {
	if c.hover != nil {
		if c.hover.Pos.Y != c.Cursor.Loc.Y {
			c.hover.Abort = true
			c.error = nil
		}
	}
}