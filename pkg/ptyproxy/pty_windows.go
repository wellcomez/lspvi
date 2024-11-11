package ptyproxy

import (
	"zen108.com/lspvi/pkg/aiopty/pty"
)

func (pty *PtyCmd) OsUpdateSize(Rows uint16, Cols uint16) {
	pty.rows = Rows
	pty.cols = Cols
	go func() {
		pty.set_size_changed <- true
	}()
}
func (ret PtyCmd) Notify() {
	//  signal.Notify(ret.Ch, syscall.SIGWINCH)
}
func get_aiopty_type() aio_option {
	return aio_option{pty.CONPTY, true}
}
