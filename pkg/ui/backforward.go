package mainui

import (
	"encoding/json"
	"os"

	"github.com/tectiv3/go-lsp"
)

type backforwarditem struct {
	Path string
	Line int
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

func (h *History) UpdateLine(path string, linenum int) {
	for i := range h.datalist {
		line := &h.datalist[i]
		if line.Path == path {
			line.Line = linenum
			return
		}
	}
}
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
func (h *History) AddToHistory(newdata string, loc *lsp.Location) {
	lll := h.datalist
	line := 0
	if loc != nil {
		line = loc.Range.Start.Line
	}
	h.datalist = h.insertAtFront(lll, backforwarditem{Path: newdata, Line: line})

	if h.file != "" {
		buf, err := json.Marshal(h.datalist)
		if err == nil {
			os.WriteFile(h.file, buf, 0644)
		}
	}
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
	bf.history.index++
	bf.history.index = min(len(bf.history.datalist)-1, bf.history.index)
	return bf.history.datalist[bf.history.index]
}

func (bf *BackForward) GoForward() backforwarditem {
	bf.history.index--
	bf.history.index = max(0, bf.history.index)
	return bf.history.datalist[bf.history.index]
}
