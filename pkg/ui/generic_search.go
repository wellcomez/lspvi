// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"fmt"

	"github.com/tectiv3/go-lsp"
	hlresult "zen108.com/lspvi/pkg/highlight/result"
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
	option       search_option
}

// NewGenericSearch is a constructor function for GenericSearch
func NewGenericSearch(view view_id, option search_option) *GenericSearch {
	return &GenericSearch{
		indexList:    make([]SearchPos, 0),
		view:         view,
		key:          option.txt,
		currentIndex: 0,
		next_or_prev: true,
		option:       option,
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
	// tofzf := option.tofzf
	noloop := option.noloop
	if len(option.txt) == 0 {
		prev := main.prefocused
		switch prev {
		case view_bookmark:
			main.bookmark_view.OnSearch(option.txt)
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
		prefocused = main.current_editor().vid()
	}
	changed := true
	if main.searchcontext == nil {
		main.searchcontext = NewGenericSearch(prefocused, option)
	} else {
		prev := main.searchcontext.next_or_prev
		changed = main.searchcontext.Changed(prefocused, txt) || noloop
		if changed {
			main.searchcontext = NewGenericSearch(prefocused, option)
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
			code := main.current_editor()
			code.code_search(main, main.quickview, changed)
			main.page.update_title(gs.String())
		}
	}
}

func (code *CodeView) code_search(main MainService, qf *quick_view, changed bool) {
	var gs *GenericSearch = main.Searchcontext()
	if changed {
		gs.indexList = code.OnSearch(gs.key, gs.option.whole)
		pos := gs.GetIndex()
		code.goto_search_result(pos)
		if gs.option.tofzf {
			locs := convert_to_fzfsearch(gs, main)
			main.ActiveTab(view_quickview, false)
			data := []ref_with_caller{}
			for _, loc := range locs {
				data = append(data, ref_with_caller{
					Loc: loc,
				})
			}
			grep := QueryOption{}
			grep.Wholeword = gs.option.whole
			grep.Query = gs.option.txt
			qf.UpdateListView(data_search, data, SearchKey{&lspcore.SymolSearchKey{Key: gs.option.txt}, &grep})
		}
	} else {
		var pos SearchPos
		if gs.next_or_prev {
			pos = gs.GetNext()
		} else {
			pos = gs.GetPrev()
		}
		code.goto_search_result(pos)
	}
}
func (code *CodeView) goto_search_result(pos SearchPos) {
	gs := code.main.Searchcontext()
	code.view.Buf.UpdateCurrent(hlresult.MatchPosition{Begin: pos.X, End: pos.X + len(gs.key), Y: pos.Y})
	code.goto_location_no_history(convert_search_pos_lsprange(pos, gs), true, nil)
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

func convert_to_fzfsearch(gs *GenericSearch, main MainService) []lsp.Location {
	var locs []lsp.Location
	for _, loc := range gs.indexList {
		x := convert_search_pos_lsprange(loc, gs)
		loc := lsp.Location{
			URI:   lsp.NewDocumentURI(main.current_editor().Path()),
			Range: x,
		}
		locs = append(locs, loc)
	}
	return locs
}
