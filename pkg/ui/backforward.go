package mainui

import (
	"os"
	"strings"
)

type History struct {
	datalist []string
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
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				history.AddToHistory(strings.TrimSpace(line))
			}
		}
	}
	return history
}

func (h *History) AddToHistory(newdata string) {
	var lll []string
	for _, line := range h.datalist {
		if newdata != line {
			lll = append(lll, line)
		}
	}
	h.datalist = h.insertAtFront(lll, newdata)

	if h.file != "" {
		os.WriteFile(h.file, []byte(strings.Join(h.datalist, "\n")), 0644)
	}
}

func (h *History) insertAtFront(slice []string, element string) []string {
	slice = append([]string{element}, slice...)
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
	return bf.history.datalist[bf.history.index]
}

func (bf *BackForward) GoForward() string {
	bf.history.index--
	bf.history.index = max(0, bf.history.index)
	return bf.history.datalist[bf.history.index]
}
