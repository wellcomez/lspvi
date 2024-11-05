// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/reinhrst/fzf-lib"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	"zen108.com/lspvi/pkg/debug"
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

type ListSymbolItem struct {
	sym        *lsp.SymbolInformation
	Positions  []int
	sort       int
	colorsname []colortext
}
type ListSymbolItemGroup struct {
	item  ListSymbolItem
	child []ListSymbolItem
}

func (pk *workspace_query_picker) on_query_ok(query string, arg []lsp.SymbolInformation) {
	pk.on_query_1(query, arg)
}

func (pk *workspace_query_picker) on_query_1(query string, arg []lsp.SymbolInformation) {
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
			if m.Score < 50 {
				continue
			}
			v := arg[m.HayIndex]
			sym = append(sym, v)
			index := len(sym) - 1
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
				{fmt.Sprintf("%-10s", strings.ReplaceAll(v.Kind.String(), "SymbolKind:", "")), fg, 0},
				// {fmt.Sprintf("%-30s ", v.Name), fg},
			}
			colors = append(colors, name...)
			if n := 30 - len(v.Name); n > 0 {
				colors = append(colors, colortext{
					strings.Repeat(" ", n), 0,
					0})
			}
			colors = append(colors, colortext{" ", 0, 0})
			file := colortext{filepath.Base(filename), 0, 0}
			colors = append(colors, file)

			pk.impl.list.AddColorItem(colors, nil, func() {
				sym := sym[index]
				pk.impl.parent.open_in_edior(sym.Location)
			})
		}
		pk.impl.list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
			pk.update_preview(sym, index)
			if index < len(sym) {
				v := sym[index]
				filename := v.Location.URI.AsPath().String()
				pk.impl.list.SetTitle(trim_project_filename(filename, global_prj_root))
			}
		})
	})
}
func (pk *workspace_query_picker) newMethod(arg []lsp.SymbolInformation, query string) {
	if len(arg) == 0 {
		return
	}
	debug.DebugLog("workspace_query_picker", query, len(arg))
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

		var Symbols []ListSymbolItem
		var class_group = make(map[string]ListSymbolItemGroup)
		for _, m := range result.Matches {
			if m.Score < 50 {
				continue
			}
			i := m.HayIndex
			sym := arg[i]
			var fg tcell.Color
			query := global_theme
			if query != nil {
				if style, err := query.get_lsp_color(sym.Kind); err == nil {
					fg, _, _ = style.Decompose()
				}
			}
			SymbolNames := convert_string_colortext(m.Positions, sym.Name, fg, tcell.ColorYellow)
			new_one := ListSymbolItem{&sym, m.Positions, m.Score, SymbolNames}
			switch sym.Kind {
			case lsp.SymbolKindClass, lsp.SymbolKindStruct:
				class_group[sym.Name] = ListSymbolItemGroup{new_one, []ListSymbolItem{}}
			default:
				added := false
				for k, v := range class_group {
					var prefix = k + "."
					if strings.HasPrefix(sym.Name, prefix) {

						added = true
						new_one.newMethod(prefix)
						a := append(class_group[k].child, new_one)
						v.child = a
						class_group[k] = v
					}
				}
				if !added {
					class_group[sym.Name] = ListSymbolItemGroup{new_one, []ListSymbolItem{}}
				}

			}
		}
		var sss = []ListSymbolItemGroup{}
		for k := range class_group {
			sss = append(sss, class_group[k])
		}
		sort.Slice(sss, func(i, j int) bool {
			return sss[i].item.sort > sss[j].item.sort
		})
		for i := range sss {
			Symbols = append(Symbols, sss[i].item)
			Symbols = append(Symbols, sss[i].child...)
		}

		for i := range Symbols {
			index := i
			m := Symbols[i]
			sym := Symbols[i].sym
			filename := sym.Location.URI.AsPath().String()

			var fg tcell.Color
			query := global_theme
			if query != nil {
				if style, err := query.get_lsp_color(sym.Kind); err == nil {
					fg, _, _ = style.Decompose()
				}
			}

			colors := []colortext{
				{fmt.Sprintf("%-10s", strings.ReplaceAll(sym.Kind.String(), "SymbolKind:", "")), fg, 0},
			}
			name_length := 0
			for _, v := range m.colorsname {
				name_length = name_length + len(v.text)
			}
			colors = append(colors, m.colorsname...)
			if n := 30 - name_length; n > 0 {
				colors = append(colors, colortext{
					strings.Repeat(" ", n), 0,
					0})
			}
			colors = append(colors, colortext{" ", 0, 0})
			file := colortext{filepath.Base(filename), 0, 0}
			colors = append(colors, file)

			pk.impl.list.AddColorItem(colors, nil, func() {
				sym := Symbols[index]
				pk.impl.parent.open_in_edior(sym.sym.Location)
			})
		}
		pk.impl.list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
			pk.update_preview_2(Symbols)
			if index < len(Symbols) {
				v := Symbols[index]
				filename := v.sym.Location.URI.AsPath().String()
				pk.impl.list.SetTitle(trim_project_filename(filename, global_prj_root))
			}
		})
	})
}

func (new_one *ListSymbolItem) newMethod(prefix string) {
	var indx int
	for i := range new_one.colorsname {
		s := new_one.colorsname[i]
		if len(s.text) == 0 {
			continue
		}
		break_it := !(indx < len(prefix))
		if !break_it {
			begin := 0
			for i := range s.text {
				break_it := !(indx < len(prefix))
				if !break_it {
					if prefix[indx] == s.text[i] {
						indx++
						begin++
					} else {
						break_it = true
						break
					}
				}
			}
			new_one.colorsname[i].text = s.text[begin:]
			if break_it {
				break
			}
		} else {
			break
		}
	}
}

// UpdateQuery implements picker.
func (pk *workspace_query_picker) UpdateQuery(query string) {
	if pk.impl.symbol == nil {
		return
	}
	pk.impl.query = query
	pk.impl.list.Clear()
	go func() {
		symbol, _ := pk.impl.symbol.WorkspaceQuery(query)
		if pk.impl.query == query {
			pk.on_query_ok(query, symbol)
		}
	}()
}
func (pk workspace_query_picker) update_preview_2(sym []ListSymbolItem) {
	cur := pk.impl.list.GetCurrentItem()
	if cur < len(sym) {
		item := sym[cur]
		pk.impl.PrevOpen(item.sym.Location.URI.AsPath().String(),
			item.sym.Location.Range.Start.Line)
	}
}
func (pk workspace_query_picker) update_preview(sym []lsp.SymbolInformation, cur int) {
	// cur := pk.impl.list.GetCurrentItem()
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
