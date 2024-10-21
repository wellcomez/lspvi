// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3
package mainui

import (
	"sync"

	"github.com/tectiv3/go-lsp"
	"zen108.com/lspvi/pkg/debug"
	fileloader "zen108.com/lspvi/pkg/ui/fileload"
)

type CodeOpenQueueStatus int

const (
	CodeOpenQueueStatusRunning = iota
	CodeOpenQueueStatusStop
)

type arg_main_openhistory struct {
	filename string
	line     *lsp.Location
}
type arg_openbuf struct {
	data     []byte
	filename string
}
type arg_open_nolsp struct {
	filename string
	line     int
}
type EditorOpenArgument struct {
	open_no_lsp       *arg_open_nolsp
	Range             *lsp.Range
	openbuf           *arg_openbuf
	main_open_history *arg_main_openhistory
}

type CodeOpenQueue struct {
	mutex      sync.Mutex
	close      chan bool
	open       chan bool
	data       *EditorOpenArgument
	editor     CodeEditor
	open_count int
	req_count  int
	main       MainService
	status     CodeOpenQueueStatus
}

func NewCodeOpenQueue(editor CodeEditor, main MainService) *CodeOpenQueue {
	ret := &CodeOpenQueue{
		close:  make(chan bool),
		open:   make(chan bool, 100),
		editor: editor,
		main:   main,
	}
	go ret.Work()
	return ret
}
func (queue *CodeOpenQueue) CloseQueue() {
	go close_queue(queue)
}

func close_queue(queue *CodeOpenQueue) {
	switch queue.status {
	case CodeOpenQueueStatusRunning:
		{
			queue.status = CodeOpenQueueStatusStop
			queue.close <- true
		}
	}
}

func (q *CodeOpenQueue) LoadFileNoLsp(filename string, line int) {
	q.enqueue(EditorOpenArgument{open_no_lsp: &arg_open_nolsp{filename, line}})
}
func (q *CodeOpenQueue) OpenFileHistory(filename string, line *lsp.Location) {
	q.enqueue(EditorOpenArgument{main_open_history: &arg_main_openhistory{filename: filename, line: line}})
}
func (queue *CodeOpenQueue) enqueue(req EditorOpenArgument) {
	if queue.status == CodeOpenQueueStatusStop {
		return
	}
	queue.replace_data(req)
	queue.open <- true
	debug.DebugLog("cqdebug", "enqueue", ":", queue.skip(), "open", queue.open_count, "req", queue.req_count)
}
func (queue *CodeOpenQueue) dequeue() *EditorOpenArgument {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	d := queue.data
	queue.data = nil
	return d
}
func (queue *CodeOpenQueue) replace_data(req EditorOpenArgument) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	queue.data = &req
	queue.req_count++
}
func (queue *CodeOpenQueue) Work() {
	for {
		select {
		case <-queue.close:
			debug.DebugLog("cqdebug", "cq Close-------")
			return
		case <-queue.open:
			{
				skip := queue.skip()
				debug.DebugLog("cqdebug", "denqueue", ":", skip, "open", queue.open_count, "req", queue.req_count)
				if !skip {
					e := queue.dequeue()
					if e != nil {
						if e.Range != nil {
							queue.editor.goto_location_no_history(*e.Range, false, nil)
						} else if e.openbuf != nil {
							queue.editor.LoadBuffer(fileloader.NewDataFileLoad( e.openbuf.data,  e.openbuf.filename))
						} else if e.open_no_lsp != nil {
							queue.editor.LoadFileNoLsp(e.open_no_lsp.filename, e.open_no_lsp.line)
						} else if e.main_open_history != nil {
							queue.main.OpenFileHistory(e.main_open_history.filename, e.main_open_history.line)
						}
						queue.open_count++
					}
				}
				if queue.status == CodeOpenQueueStatusStop {
					return
				}
			}
		}
	}
}

func (queue *CodeOpenQueue) skip() bool {
	skip := false
	if queue.editor != nil && queue.editor.IsLoading() {
		skip = true
	}
	if queue.main != nil && queue.main.current_editor().IsLoading() {
		skip = true
	}
	return skip
}
