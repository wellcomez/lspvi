package mainui

import (
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
	old      *femto.TextEvent
	new      *femto.TextEvent
	stacktop int
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

	event, top := statckstatus(code)
	return code_change_cheker{lineno: lineno, next: next, cur: cur, old: event, stacktop: top}
}

const tag = "editor event"

func (check *code_change_cheker) after(code *CodeView) lspcore.CodeChangeEvent {
	event, top := statckstatus(code)
	debug.DebugLog(tag, " stack change top", top, check.stacktop)
	check.new = event
	ret := code.LspContentFullChangeEvent()
	ret.Full = false
	if check.old != check.new {
		if event == nil {
			return ret
		}
		ret = ParserEvent(ret, event)
		if len(ret.Events) > 0 {
			code.on_content_changed(ret)
		}
	}
	return ret
}

func statckstatus(code *CodeView) (event *femto.TextEvent, top int) {
	event = code.view.Buf.UndoStack.Peek()
	top = code.view.Buf.UndoStack.Size
	return event, top
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
		a := lspcore.TextChangeEvent{
			Type: Type,
			Text: v.Text,
			Range: lsp.Range{
				Start: lsp.Position{
					Line: v.Start.Y, Character: v.Start.X},
				End: lsp.Position{
					Line: v.End.Y, Character: v.End.X},
			},
		}
		change.Events = append(change.Events, a)
		debug.DebugLog(tag, name, event.Deltas)
	}
	return change
}
func (check *code_change_cheker) newMethod(code *CodeView) (bool, int) {
	after_lineno := code.view.Cursor.Loc.Y
	next := check.next
	lineno := check.lineno
	after_cur := get_line_content(after_lineno, code.view.Buf)
	if after_lineno+1 == lineno {
		code.view.bookmark.after_line_changed(lineno, false)
		code.udpate_modified_lines(lineno)
		return true, lineno
	} else if after_lineno == lineno {
		if after_cur == next {
			code.view.bookmark.after_line_changed(lineno+1, false)
			code.udpate_modified_lines(lineno)
		} else if after_cur != check.cur {
			code.udpate_modified_lines(lineno)
			return true, lineno
		}
	} else if after_lineno == lineno+1 {
		code.view.bookmark.after_line_changed(lineno+1, true)
		code.udpate_modified_lines(lineno + 1)
		return true, after_lineno
	}
	return false, 0
}

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
	lspchange: make(chan int),
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
	for i := range q.wait_queue {
		v := q.wait_queue[i]
		if v.code == c {
			debug.DebugLog("lspchange_queue", "skip")
			v.event.Full = true
			return
		}
	}
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
	data := []byte{}
	for i := 0; i < code.view.Buf.LinesNum(); i++ {
		data = append(data, code.view.Buf.LineBytes(i)...)
		data = append(data, '\n')
	}
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
		if code.lspsymbol != nil {
			code.lspsymbol.LspLoadSymbol()
		}
		on_treesitter_update(code, ts)
	})
}
