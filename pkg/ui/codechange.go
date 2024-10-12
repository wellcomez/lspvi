package mainui

import (
	"log"
	"sync"

	"github.com/pgavlin/femto"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type code_change_cheker struct {
	lineno int
	next   string
	cur    string
	old    *femto.TextEvent
	new    *femto.TextEvent
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
	event := code.view.Buf.UndoStack.Peek()
	return code_change_cheker{lineno: lineno, next: next, cur: cur, old: event}
}

func (check *code_change_cheker) after(code *CodeView) lspcore.CodeChangeEvent {
	event := code.view.Buf.UndoStack.Peek()
	check.new = event
	ret := lspcore.CodeChangeEvent{}
	if check.old != check.new {
		if event == nil {
			return ret
		}
		name := ""
		switch event.EventType {
		case femto.TextEventInsert:
			name = "insert"
			newFunction1(&ret, lspcore.TextChangeTypeInsert, event)
		case femto.TextEventRemove:
			name = "remove"
			newFunction1(&ret, lspcore.TextChangeTypeDeleted, event)
		case femto.TextEventReplace:
			name = "replace"
			newFunction1(&ret, lspcore.TextChangeTypeReplace, event)
		default:
		}
		if len(ret.Events) > 0 {
			code.on_content_changed(ret)
		}
		log.Println("editor event", name, event.Deltas)
	}
	return ret
}

func newFunction1(ret *lspcore.CodeChangeEvent, changetype lspcore.TextChangeType, event *femto.TextEvent) {
	for _, v := range event.Deltas {
		a := lspcore.TextChangeEvent{
			Type: changetype,
			Text: v.Text,
			Range: lsp.Range{
				Start: lsp.Position{
					Line: v.Start.Y, Character: v.Start.X},
				End: lsp.Position{
					Line: v.End.Y, Character: v.End.X},
			},
		}
		ret.Events = append(ret.Events, a)
	}
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
	if !q.start {
		q.start = true
		go q.Worker()
	}
	q.lock.Lock()
	defer q.lock.Unlock()
	for i := range q.wait_queue {
		v := q.wait_queue[i]
		if v.code == c {
			log.Println("skip")
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
		on_treesitter_update(code, ts)
	})
}
