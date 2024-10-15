package mainui

import (
	"fmt"
	"sync"

	"github.com/pgavlin/femto"
	"github.com/tectiv3/go-lsp"
	"zen108.com/lspvi/pkg/debug"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type code_change_cheker struct {
	lineno   int
	next     string
	cur      string
	undo_top *femto.Element
	redo_top *femto.Element
}

func new_code_change_checker(code *CodeView) code_change_cheker {
	lineno := code.view.Cursor.Loc.Y
	next := get_line_content(lineno+1, code.view.Buf)
	cur := get_line_content(lineno, code.view.Buf)
	if code.diff != nil {
		if len(code.diff.bufer) == 0 {
			Buf := code.view.Buf
			end := Buf.LinesNum()
			code.diff = &Differ{Buf.Lines(0, end), -1}
		}
	}
	return code_change_cheker{lineno: lineno, next: next, cur: cur,
		undo_top: code.view.Buf.UndoStack.Top,
		redo_top: code.view.Buf.RedoStack.Top,
	}
}

const tag = "Triggers Event textDocument/didChange"

func (check *code_change_cheker) CheckRedo(code *CodeView) []lspcore.CodeChangeEvent {
	var ret []lspcore.CodeChangeEvent
	var events []femto.TextEvent
	var redo = code.view.Buf.RedoStack.Top
	debug.DebugLog(tag, " stack REDO stack  top", code.view.Buf.RedoStack.Size)
	for {
		if redo != nil && redo != check.redo_top {
			events = append([]femto.TextEvent{*redo.Value}, events...)
			redo = redo.Next
		} else {
			break
		}

	}
	for _, v := range events {
		a := code.LspContentFullChangeEvent()
		a.Full = false
		a = ParserEvent(a, &v)
		code.on_content_changed(a)
		ret = append(ret, a)
	}
	check.UpdateLineChange(code)
	return ret
}

func (check *code_change_cheker) after(code *CodeView) []lspcore.CodeChangeEvent {
	undo := code.view.Buf.UndoStack
	debug.DebugLog(tag, " stack change top", undo.Size)
	var stack = code.view.Buf.UndoStack.Top
	var ret []lspcore.CodeChangeEvent
	var events []femto.TextEvent
	ele := stack
	for {
		if ele != nil && ele != check.undo_top {
			events = append([]femto.TextEvent{*ele.Value}, events...)
			ele = ele.Next
		} else {
			break
		}

	}
	for _, v := range events {
		a := code.LspContentFullChangeEvent()
		a.Full = false
		a = ParserEvent(a, &v)
		code.on_content_changed(a)
		ret = append(ret, a)
	}
	check.UpdateLineChange(code)
	return ret
}

func ParserEvent(change lspcore.CodeChangeEvent, event *femto.TextEvent) lspcore.CodeChangeEvent {
	var name string
	var Type lspcore.TextChangeType
	switch event.EventType {
	case femto.TextEventInsert:
		Type = lspcore.TextChangeTypeInsert
		name = "insert"
	case femto.TextEventRemove:
		Type = lspcore.TextChangeTypeDeleted
		name = "remove"
	case femto.TextEventReplace:
		Type = lspcore.TextChangeTypeReplace
		name = "replace"
	default:
	}
	for _, v := range event.Deltas {
		Start := lsp.Position{
			Line: v.Start.Y, Character: v.Start.X}
		End := lsp.Position{
			Line: v.End.Y, Character: v.End.X}
		Text := v.Text

		switch event.EventType {
		case femto.TextEventInsert:
			End = Start
		case femto.TextEventRemove:
			Text = ""
		default:
		}

		a := lspcore.TextChangeEvent{
			Type: Type,
			Text: Text,
			Time: event.Time,
			Range: lsp.Range{
				Start: Start,
				End:   End,
			},
		}
		change.Events = append(change.Events, a)
		debug.DebugLog(tag, name,
			"\nStart", a.Range.Start, " End:", a.Range.End,
			fmt.Sprintf("\nlen=%d [%s] ", len(a.Text), a.Text),
			a.Time.UnixMilli(),
		)
	}
	return change
}
func (check *code_change_cheker) UpdateLineChange(code *CodeView) {
	top := code.view.Buf.UndoStack.Top
	var changeline = []int{}

	seen := make(map[int]bool)
	for {
		if top != nil {
			for _, v := range top.Value.Deltas {
				for _, y := range []int{v.Start.Y} {
					if _, ok := seen[y]; !ok {
						changeline = append(changeline, y)
						seen[y] = true
					}
				}
			}
			top = top.Next
		} else {
			break
		}
	}
	var marks = []LineMark{}
	for _, v := range changeline {
		marks = append(marks, LineMark{Line: v + 1})
	}
	code.view.linechange.LineMark = marks
}

// func (check *code_change_cheker) newMethod(code *CodeView) (bool, int) {
// 	after_lineno := code.view.Cursor.Loc.Y
// 	next := check.next
// 	lineno := check.lineno
// 	after_cur := get_line_content(after_lineno, code.view.Buf)
// 	if after_lineno+1 == lineno {
// 		code.view.bookmark.after_line_changed(lineno, false)
// 		code.udpate_modified_lines(lineno)
// 		return true, lineno
// 	} else if after_lineno == lineno {
// 		if after_cur == next {
// 			code.view.bookmark.after_line_changed(lineno+1, false)
// 			code.udpate_modified_lines(lineno)
// 		} else if after_cur != check.cur {
// 			code.udpate_modified_lines(lineno)
// 			return true, lineno
// 		}
// 	} else if after_lineno == lineno+1 {
// 		code.view.bookmark.after_line_changed(lineno+1, true)
// 		code.udpate_modified_lines(lineno + 1)
// 		return true, after_lineno
// 	}
// 	return false, 0
// }

type lspchange struct {
	code  *CodeView
	event lspcore.CodeChangeEvent
}
type lspchange_queue struct {
	wait_queue []lspchange
	lspchange  chan int
	lock       sync.Mutex
	start      bool
}

var lsp_queue = lspchange_queue{
	lspchange: make(chan int, 100),
}

func (q *lspchange_queue) AddQuery(c *CodeView, event lspcore.CodeChangeEvent) {
	if len(event.File) == 0 {
		debug.DebugLog("lsp_queue", ".AddQuery: empty filename")
	}
	if !q.start {
		q.start = true
		go q.Worker()
	}
	q.lock.Lock()
	defer q.lock.Unlock()
	// for i := range q.wait_queue {
	// 	v := q.wait_queue[i]
	// 	if v.code == c {
	// 		debug.DebugLog("lspchange_queue", "skip")
	// 		v.event.Full = true
	// 		return
	// 	}
	// }
	q.wait_queue = append(q.wait_queue, lspchange{c, event})
	if len(q.wait_queue) > 0 {
		q.lspchange <- len(q.wait_queue)
	}

}
func (q *lspchange_queue) Worker() {
	for {
		select {
		case <-q.lspchange:
			data := empty_queue(q)
			for _, v := range data {
				v.code.update_ts(v.event)
			}
		}
	}
}

func empty_queue(q *lspchange_queue) []lspchange {
	var ret []lspchange
	q.lock.Lock()
	ret = q.wait_queue
	q.wait_queue = []lspchange{}
	q.lock.Unlock()
	return ret
}

func (code *CodeView) on_content_changed(event lspcore.CodeChangeEvent) {
	lsp_queue.AddQuery(code, event)
}
func (code *CodeView) update_ts(event lspcore.CodeChangeEvent) {
	// event.Full = true
	data := code.GetBuffData()
	if event.Full {
		event.Data = data
	}
	if code.lspsymbol != nil {
		code.lspsymbol.NotifyCodeChange(event)
	}
	var ts = event
	ts.Data = data
	var new_ts = lspcore.GetNewTreeSitter(code.Path(), ts)
	new_ts.Init(func(ts *lspcore.TreeSitter) {
		if !ts.IsMe(code.Path()) {
			return
		}
		if code.lspsymbol != nil {
			code.lspsymbol.LspLoadSymbol()
		}
		on_treesitter_update(code, ts)
	})
}

func (code *CodeView) GetBuffData() []byte {
	data := []byte{}
	for i := 0; i < code.view.Buf.LinesNum(); i++ {
		data = append(data, code.view.Buf.LineBytes(i)...)
		data = append(data, '\n')
	}
	return data
}
