package mainui

import (
	"fmt"

	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type SearchPos struct {
	X, Y int
}

// GenericSearch struct
type GenericSearch struct {
	indexList    []SearchPos
	view         view_id
	key          string
	currentIndex int
	next_or_prev bool
}

// NewGenericSearch is a constructor function for GenericSearch
func NewGenericSearch(view view_id, key string) *GenericSearch {
	return &GenericSearch{
		indexList:    make([]SearchPos, 0),
		view:         view,
		key:          key,
		currentIndex: 0,
		next_or_prev: true,
	}
}

func (g GenericSearch) Changed(view view_id, key string) bool {
	if g.view != view {
		return true
	}
	if g.key != key {
		return true
	}
	return false
}
func (gs *GenericSearch) GetPrev() SearchPos {
	if len(gs.indexList) == 0 {
		return SearchPos{-1, -1}
	}
	gs.currentIndex = gs.currentIndex - 1 + len(gs.indexList)
	gs.currentIndex %= len(gs.indexList)
	return gs.indexList[gs.currentIndex]
}

// GetNext returns the next index in the indexList
func (gs *GenericSearch) GetNext() SearchPos {
	if len(gs.indexList) == 0 {
		return SearchPos{-1, -1}
	}
	gs.currentIndex++
	gs.currentIndex %= len(gs.indexList)
	return gs.indexList[gs.currentIndex]
}

// GetIndex returns the current index in the indexList
func (gs *GenericSearch) GetIndex() SearchPos {
	if len(gs.indexList) == 0 {
		return SearchPos{-1, -1}
	}
	return gs.indexList[gs.currentIndex]
}

// Add adds an index to the indexList
func (gs *GenericSearch) Add(index SearchPos) {
	gs.indexList = append(gs.indexList, index)
}

// String returns a string representation of the GenericSearch object
func (gs *GenericSearch) String() string {
	return fmt.Sprintf("search %s %d/%d", gs.key, gs.currentIndex, gs.ResultNumber())
}

// ResultNumber returns the number of results in the indexList
func (gs *GenericSearch) ResultNumber() int {
	return len(gs.indexList)
}

func search_on_ui(option search_option, main *mainui) {
	txt := option.txt
	tofzf := option.tofzf
	noloop := option.noloop
	whole := option.whole
	if len(txt) == 0 {
		prev := main.prefocused
		switch prev {
		case view_bookmark:
			main.bookmark_view.OnSearch(txt)
		}
		return
	}
	prefocused := main.prefocused
	switch main.prefocused {
	case view_bookmark, view_callin, view_quickview, view_outline_list:
		{
			prefocused = main.prefocused
		}
	default:
		prefocused = view_code
	}
	changed := true
	if main.searchcontext == nil {
		main.searchcontext = NewGenericSearch(prefocused, txt)
	} else {
		prev := main.searchcontext.next_or_prev
		changed = main.searchcontext.Changed(prefocused, txt) || noloop
		if changed {
			main.searchcontext = NewGenericSearch(prefocused, txt)
			main.searchcontext.next_or_prev = prev
		}
		// if tofzf || !noloop {
		// 	main.cmdline.Vim.EnterGrep(txt)
		// }
	}
	gs := main.searchcontext
	switch prefocused {
	case view_bookmark:
		{
			main.bookmark_view.OnSearch(txt)
		}
	case view_callin:
		{
			main.callinview.OnSearch(txt)
		}
	case view_quickview:
		{
			main.quickview.OnSearch(txt)
		}
	case view_outline_list:
		{
			if changed {
				gs.indexList = main.symboltree.OnSearch(txt)
				main.symboltree.movetonode(gs.GetIndex().Y)
			} else {
				main.symboltree.movetonode(gs.GetNext().Y)
			}
		}
	default:
		{
			if changed {
				gs.indexList = main.codeview.OnSearch(txt, whole)
				pos := gs.GetIndex()
				main.codeview.goto_symbol_location(convert_search_pos_lsprange(pos, gs), true, nil)
				if tofzf {
					locs := convert_to_fzfsearch(gs, main)
					main.ActiveTab(view_quickview, false)
					data := []ref_with_caller{}
					for _, loc := range locs {
						data = append(data, ref_with_caller{
							Loc: loc,
						})
					}
					main.quickview.UpdateListView(data_search, data, lspcore.SymolSearchKey{Key: txt})
				}
			} else {
				var pos SearchPos
				if gs.next_or_prev {
					pos = gs.GetNext()
				} else {
					pos = gs.GetPrev()
				}
				main.codeview.goto_symbol_location(convert_search_pos_lsprange(pos, gs), true, nil)
			}
			main.page.update_title(gs.String())
		}
	}
}

func convert_search_pos_lsprange(loc SearchPos, gs *GenericSearch) lsp.Range {
	x := lsp.Range{
		Start: lsp.Position{
			Line:      loc.Y,
			Character: loc.X,
		},
		End: lsp.Position{
			Line:      loc.Y,
			Character: loc.X + len(gs.key),
		},
	}
	return x
}

func convert_to_fzfsearch(gs *GenericSearch, main *mainui) []lsp.Location {
	var locs []lsp.Location
	for _, loc := range gs.indexList {
		x := convert_search_pos_lsprange(loc, gs)
		loc := lsp.Location{
			URI:   lsp.NewDocumentURI(main.codeview.Path()),
			Range: x,
		}
		locs = append(locs, loc)
	}
	return locs
}
