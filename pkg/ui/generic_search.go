package mainui

import (
	"fmt"
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
