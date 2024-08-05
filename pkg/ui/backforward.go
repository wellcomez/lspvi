package mainui

import (
	"encoding/json"
	"os"
)

type filehistory struct {
	Path string
	Line int
}
type History struct {
	datalist []filehistory
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
			var data []filehistory
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
func (h *History) AddToHistory(newdata string) {
	lll := []filehistory{}
	for _, v := range h.datalist {
		if newdata != v.Path {
			lll = append(lll, v)
		}
	}
	h.datalist = h.insertAtFront(lll, filehistory{Path: newdata})

	if h.file != "" {
		buf, err := json.Marshal(h.datalist)
		if err == nil {
			os.WriteFile(h.file, buf, 0644)
		}
	}
}

func (h *History) insertAtFront(slice []filehistory, element filehistory) []filehistory {
	slice = append([]filehistory{element}, slice...)
	return slice
}

type BackForward struct {
	history *History
}

func NewBackForward(h *History) *BackForward {
	return &BackForward{history: h}
}

func (bf *BackForward) GoBack() string {
	bf.history.index++
	bf.history.index = min(len(bf.history.datalist)-1, bf.history.index)
	return bf.history.datalist[bf.history.index].Path
}

func (bf *BackForward) GoForward() string {
	bf.history.index--
	bf.history.index = max(0, bf.history.index)
	return bf.history.datalist[bf.history.index].Path
}
