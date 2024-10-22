// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/reinhrst/fzf-lib"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type workspace_query_picker_impl struct {
	*prev_picker_impl
	symbol *lspcore.Symbol_file
	list   *customlist
	query  string
	// sym    []lsp.SymbolInformation
}
type workspace_query_picker struct {
	impl *workspace_query_picker_impl
}

// close implements picker.
func (pk *workspace_query_picker) close() {
	// pk.impl.cq.CloseQueue()
}

// name implements picker.
func (pk *workspace_query_picker) name() string {
	return "workspace symbol"
}

func (pk *workspace_query_picker) on_query_ok(query string, arg []lsp.SymbolInformation) {
	if len(arg) == 0 {
		return
	}
	pk.impl.parent.app.QueueUpdateDraw(func() {
		opt := fzf.DefaultOptions()
		opt.Fuzzy = true
		ss := []string{}
		for _, v := range arg {
			ss = append(ss, v.Name)
		}
		fzf := fzf.New(ss, opt)
		fzf.Search(query)
		result := <-fzf.GetResultChannel()
		pk.impl.list.Clear()
		pk.impl.list.SetCurrentItem(0)
		pk.impl.list.Key = pk.impl.query
		var sym []lsp.SymbolInformation
		for _, m := range result.Matches {
			if m.Score < 50{
				continue
			}
			i := m.HayIndex
			v := arg[i]
			sym =append(sym, v)
			index := i
			filename := v.Location.URI.AsPath().String()
			var fg tcell.Color
			query := global_theme
			if query != nil {
				if style, err := query.get_lsp_color(v.Kind); err == nil {
					fg, _, _ = style.Decompose()
				}
			}
			name := convert_string_colortext(m.Positions, v.Name, fg, tcell.ColorYellow)
			colors := []colortext{
				{fmt.Sprintf("%-10s", strings.ReplaceAll(v.Kind.String(), "SymbolKind:", "")), fg},
				// {fmt.Sprintf("%-30s ", v.Name), fg},
			}
			colors = append(colors, name...)
			if n := 30 - len(v.Name); n > 0 {
				colors = append(colors, colortext{
					strings.Repeat(" ", n), 0,
				})
			}
			colors = append(colors, colortext{" ", 0})
			file := colortext{filepath.Base(filename), 0}
			colors = append(colors, file)

			pk.impl.list.AddColorItem(colors, nil, func() {
				sym := sym[index]
				main := pk.impl.parent.main
				main.OpenFileHistory(sym.Location.URI.AsPath().String(), &sym.Location)
				pk.impl.parent.hide()
			})
		}
		pk.impl.list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
			pk.update_preview(sym)
			if index < len(sym) {
				v := sym[index]
				filename := v.Location.URI.AsPath().String()
				pk.impl.list.SetTitle(trim_project_filename(filename, global_prj_root))
			}
		})
	})
}

// UpdateQuery implements picker.
func (pk *workspace_query_picker) UpdateQuery(query string) {
	if pk.impl.symbol == nil {
		return
	}
	pk.impl.query = query
	go func() {
		symbol, _ := pk.impl.symbol.WorkspaceQuery(query)
		if pk.impl.query == query {
			pk.on_query_ok(query, symbol)
		}
	}()
}
func (pk workspace_query_picker) update_preview(sym []lsp.SymbolInformation) {
	cur := pk.impl.list.GetCurrentItem()
	if cur < len(sym) {
		item := sym[cur]
		pk.impl.PrevOpen(item.Location.URI.AsPath().String(),
			item.Location.Range.Start.Line)
	}
}
func (pk workspace_query_picker) handle_key_override(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := pk.impl.list.InputHandler()
	handle(event, setFocus)
}

// handle implements picker.
func (pk *workspace_query_picker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return pk.handle_key_override
}

func new_workspace_symbol_picker(v *fzfmain, code CodeEditor) *workspace_query_picker {
	ret := &workspace_query_picker{
		impl: &workspace_query_picker_impl{
			prev_picker_impl: new_preview_picker(v),
			symbol:           code.LspSymbol(),
			list:             new_customlist(false),
		},
	}
	ret.impl.prev_picker_impl.use_cusutom_list(ret.impl.list)

	return ret
}
func (pk *workspace_query_picker) grid() *tview.Flex {
	return pk.impl.flex(pk.impl.parent.input, 1)
}
