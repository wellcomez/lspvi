package mainui

import (
	"encoding/json"
	// "log"
	"os"
)

type backforwarditem struct {
	Path string
	Pos  EditorPosition
}
type History struct {
	datalist []backforwarditem
	file     string
	index    int
}

func NewHistory(file string) *History {
	history := &History{
		file:  file,
		index: 0,
	}
	if file != "" {
		content, err := os.ReadFile(file)
		if err == nil {
			var data []backforwarditem
			if json.Unmarshal(content, &data) == nil {
				history.datalist = data
			}
		}
	}
	return history
}

//	func (h *History) UpdateLine(path string, linenum int) {
//		for i := range h.datalist {
//			line := &h.datalist[i]
//			if line.Path == path {
//				line.Line = linenum
//				return
//			}
//		}
//	}
func (h *History) history_files() []string {
	ret := []string{}
	for _, r := range h.datalist {
		added := false
		for _, s := range ret {
			if r.Path == s {
				added = true
				break
			}
		}
		if !added {
			ret = append(ret, r.Path)
		}
	}
	return ret
}

type EditorPosition struct {
	Line   int
	Offset int
}

func NewEditorPosition(Line int, text *CodeView) *EditorPosition {
	return &EditorPosition{
		Line:   Line,
		Offset: 0,
	}
}
func (h *History) SaveToHistory(code *CodeView) {
	line := code.view.Cursor.Loc.Y + code.view.Topline
	pos := EditorPosition{
		Line:   line,
		Offset: code.view.Topline,
	}
	if len(h.datalist) > 0 {
		h.datalist[h.index].Pos = pos
		h.save_to_file()
	}
}

func (h *History) save_to_file() {
	if h.file != "" {
		buf, err := json.Marshal(h.datalist)
		if err == nil {
			os.WriteFile(h.file, buf, 0644)
		}
	}
}

// AddToHistory
func (h *History) AddToHistory(newdata string, loc *EditorPosition) {
	lll := h.datalist
	pos := EditorPosition{
		Line: 0,
	}
	if loc != nil {
		pos = *loc
	}
	h.datalist = h.insertAtFront(lll, backforwarditem{Path: newdata, Pos: pos})
	// log.Printf("add history %v", h.datalist[0])
	h.save_to_file()
}

func (h *History) insertAtFront(slice []backforwarditem, element backforwarditem) []backforwarditem {
	slice = append([]backforwarditem{element}, slice...)
	return slice
}

type BackForward struct {
	history *History
}

func NewBackForward(h *History) *BackForward {
	return &BackForward{history: h}
}

func (bf *BackForward) GoBack() backforwarditem {
	if len(bf.history.datalist) > 0 {
		bf.history.index++
		bf.history.index = min(len(bf.history.datalist)-1, bf.history.index)
		return bf.history.datalist[bf.history.index]
	}
	return backforwarditem{}
}

func (bf *BackForward) GoForward() backforwarditem {
	if len(bf.history.datalist) == 0 {
		return backforwarditem{}
	}
	bf.history.index--
	bf.history.index = max(0, bf.history.index)
	return bf.history.datalist[bf.history.index]
}
