package mainui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"zen108.com/lspvi/pkg/debug"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type symbol_picker_impl struct {
	*prev_picker_impl
	fzf             *fzf_on_listview
	symbol          *FzfSymbolFilter
	list_index_data map[int]int
}
type current_document_pick struct {
	impl *symbol_picker_impl
}

// UpdateQuery implements picker.
func (c current_document_pick) UpdateQuery(query string) {
	fzf := c.impl.fzf
	fzf.OnSearch(query, false)
	filter := make(map[int]bool)
	selected_postion := make(map[int][]int)
	for i, v := range fzf.selected_index {
		filter[v] = true
		selected_postion[v] = fzf.selected_postion[i]
		debug.DebugLog("selected ", fzf.selected_text[i], v)
	}
	list := c.impl.listview
	list.Clear()
	var list_index_data = make(map[int]int)
	var list_index = 0
	for _, n := range c.impl.symbol.ClassObject {
		mm := []SymbolFzf{}
		for _, v := range n.Member {
			if _, ok := filter[v.index]; ok {
				mm = append(mm, v)
			}
		}
		if len(mm) == 0 {
			if _, ok := filter[n.index]; ok {
				s := set_list_item(n, 0, selected_postion[n.index])
				list_index_data[list_index] = n.index
				list_index++
				list.AddItem(s, "", nil)
			}
		} else {
			s := set_list_item(n, 0, nil)
			list.AddItem(s, "", nil)
			list_index_data[list_index] = n.index
			list_index++
			for _, v := range mm {
				s := set_list_item(v, 1, selected_postion[v.index])
				list.AddItem(s, "", nil)
				list_index_data[list_index] = v.index
				list_index++
			}
		}
	}
	c.impl.list_index_data = list_index_data
}

// close implements picker.
func (c current_document_pick) close() {
}

// handle implements picker.
func (c current_document_pick) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		c.impl.listview.InputHandler()(event, setFocus)
	}
}

type SymbolFzf struct {
	sym    lspcore.Symbol
	index  int
	Member []SymbolFzf
}
type FzfSymbolFilter struct {
	ClassObject      []SymbolFzf
	names            []string
	filter           map[int]bool
	selected_postion map[int][]int
	object_index     map[int]SymbolFzf
}

func NewSymbolListFzf(sym *lspcore.Symbol_file) *FzfSymbolFilter {
	ret := &FzfSymbolFilter{[]SymbolFzf{}, []string{}, nil, nil, nil}
	ret.Convert(sym)
	return ret
}

func (f *FzfSymbolFilter) GetSymbolFile(key string) *lspcore.Symbol_file {
	ret := []*lspcore.Symbol{}
	for _, v := range f.ClassObject {
		member := []lspcore.Symbol{}
		for _, vv := range v.Member {
			if _, ok := f.filter[vv.index]; ok {
				member = append(member, v.sym)
			}
		}
		var sss = v.sym
		root := &sss
		if _, yes := f.filter[v.index]; yes {
			root.Members = member
			ret = append(ret, root)
		} else if len(member) > 0 {
			root.Members = member
			ret = append(ret, root)
		}
	}
	return &lspcore.Symbol_file{
		Class_object: ret,
	}
}
func (f *FzfSymbolFilter) Convert(symbol *lspcore.Symbol_file) {
	if symbol == nil {
		return
	}
	id := 0
	class_object := []SymbolFzf{}
	names := []string{}
	var ojbect_index = make(map[int]SymbolFzf)
	classobject := symbol.Class_object
	for _, v := range classobject {
		member := []SymbolFzf{}
		for i := range v.Members {
			sym := SymbolFzf{v.Members[i], id, nil}
			ojbect_index[sym.index] = sym
			id++
			names = append(names, sym.sym.SymInfo.Name)
			member = append(member, sym)
		}
		var sss = SymbolFzf{*v, id, nil}
		names = append(names, sss.sym.SymInfo.Name)
		ojbect_index[sss.index] = sss
		id++
		sss.Member = member
		class_object = append(class_object, sss)
	}
	f.ClassObject = class_object
	f.names = names
	f.object_index = ojbect_index
}

// name implements picker.
func (c current_document_pick) name() string {
	return "symbol"
}
func new_current_document_picker(v *fzfmain, symbol *lspcore.Symbol_file) current_document_pick {
	impl := &symbol_picker_impl{
		prev_picker_impl: new_preview_picker(v),
	}
	list := new_customlist(false)
	sym := current_document_pick{
		impl: impl,
	}
	impl.symbol = NewSymbolListFzf(symbol)
	var list_index_data = make(map[int]int)
	var list_index = 0
	for _, v := range impl.symbol.ClassObject {
		list.AddItem(set_list_item(v, 0, nil), "", nil)
		list_index_data[list_index] = v.index
		list_index++
		if len(v.Member) > 0 {
			for _, m := range v.Member {
				list.AddItem(set_list_item(m, 1, nil), "", nil)
				list_index_data[list_index] = v.index
				list_index++
			}
		}
	}
	impl.list_index_data = list_index_data
	list.SetSelectedFunc(func(index int, s1 string, s2 string, r rune) {
		if sym := impl.get_current_item_symbol(index); sym != nil {
			file := sym.sym.SymInfo.Location.URI.AsPath().String()
			v.main.OpenFileHistory(file, &sym.sym.SymInfo.Location)
			v.hide()
		}

	})
	list.SetChangedFunc(func(index int, s string, s2 string, r rune) {
		impl.UpdatePrev(index)
	})

	impl.listview = list
	impl.fzf = new_fzf_on_list_data(list, impl.symbol.names, true)
	list.SetCurrentItem(0)
	impl.UpdatePrev(-1)
	return sym
}

func (impl *symbol_picker_impl) UpdatePrev(index int) {
	if sym := impl.get_current_item_symbol(index); sym != nil {
		file := sym.sym.SymInfo.Location.URI.AsPath().String()
		impl.cq.LoadFileNoLsp(file, sym.sym.SymInfo.Location.Range.Start.Line)
	}
}

func (impl *symbol_picker_impl) get_current_item_symbol(i int) *SymbolFzf {
	if i == -1 {
		i = impl.listview.GetCurrentItem()
	}
	if data_index, ok := impl.list_index_data[i]; ok {
		if sym, ok := impl.symbol.object_index[data_index]; ok {
			return &sym
		}
	}
	return nil
}

func set_list_item(v SymbolFzf, prefix int, posistion []int) string {
	space := strings.Repeat("\t\t", prefix)
	icon := v.sym.Icon()
	query := global_theme
	cc := tview.Styles.PrimaryTextColor
	if query != nil {
		if style, err := query.get_lsp_color(v.sym.SymInfo.Kind); err == nil {
			fg, _, _ := style.Decompose()
			cc = fg
		}
	}
	// if false {
	if posistion != nil {
		s := fzf_color(posistion, v.sym.SymInfo.Name)
		return fmt.Sprintf("%s%s %s", space, icon, s)
	} else {
		t := fmt.Sprintf("%s%s %s", space, icon, v.sym.SymInfo.Name)
		t = fmt_color_string(t, cc)
		return t
	}
}
