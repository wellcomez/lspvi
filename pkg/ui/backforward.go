package mainui

import (
	"encoding/json"
	"log"

	// "log"
	"os"

	"github.com/tectiv3/go-lsp"
)

func (path backforwarditem) GetLocation() lsp.Location {
	return lsp.Location{URI: lsp.NewDocumentURI(path.Path), Range: lsp.Range{
		Start: lsp.Position{Line: path.Pos.Line, Character: 0},
		End:   lsp.Position{Line: path.Pos.Line, Character: 0},
	},
	}
}

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
func (h *History) history_files() []backforwarditem {
	ret := []backforwarditem{}
	for _, r := range h.datalist {
		added := false
		for _, s := range ret {
			if r.Path == s.Path {
				added = true
				break
			}
		}
		if !added {
			ret = append(ret, r)
		}
	}
	return ret
}

type EditorPosition struct {
	Line   int
	Offset int
}

func NewEditorPosition(Line int) *EditorPosition {
	return &EditorPosition{
		Line:   Line,
		Offset: 0,
	}
}
func (h *History) SaveToHistory(code CodeEditor) {
	pos := code.EditorPosition()
	if pos == nil {
		return
	}
	if len(h.datalist) > 0 {
		h.datalist[h.index].Pos = *pos
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
	if newdata == "" {
		return
	}
	debug_print_bf("add before", h)
	lll := h.datalist[h.index:]
	pos := EditorPosition{
		Line: 0,
	}
	if loc != nil {
		pos = *loc
	}
	newhistory := backforwarditem{Path: newdata, Pos: pos}
	h.datalist = append([]backforwarditem{newhistory}, lll...)
	h.index = 0
	debug_print_bf("add after", h)
	h.save_to_file()
}

func debug_print_bf(prefix string, h *History) {
	return
	if len(h.datalist) > 0 {
		tag := "HISTORY"
		item := h.datalist[h.index]
		log.Println("index:", h.index, tag, prefix, "len", len(h.datalist), item.Path, ":", item.Pos.Line)
		for i := range h.datalist {
			flag := ""
			if i == h.index {
				flag = "*"
			}
			log.Println("	", tag, flag, "<", i, ">", h.datalist[i].Path, h.datalist[i].Pos.Line)
		}
	}
}

// func (h *History) insertAtFront(slice []backforwarditem, element backforwarditem) []backforwarditem {
// 	slice = append([]backforwarditem{element}, slice...)
// 	return slice
// }

type BackForward struct {
	history *History
}

func NewBackForward(h *History) *BackForward {
	return &BackForward{history: h}
}
func (bf *BackForward) Last() backforwarditem {
	if len(bf.history.datalist) > 0 {
		return bf.history.datalist[0]
	}
	return backforwarditem{}
}

func (bf *BackForward) GoBack() backforwarditem {
	if len(bf.history.datalist) > 0 {
		bf.history.index++
		bf.history.index = min(len(bf.history.datalist)-1, bf.history.index)
		ret := bf.history.datalist[bf.history.index]
		debug_print_bf("Goback", bf.history)
		return ret
	}
	return backforwarditem{}
}

func (bf *BackForward) GoForward() backforwarditem {
	if len(bf.history.datalist) == 0 {
		return backforwarditem{}
	}
	bf.history.index--
	bf.history.index = max(0, bf.history.index)
	ret := bf.history.datalist[bf.history.index]
	debug_print_bf("GoForward", bf.history)
	return ret
}
