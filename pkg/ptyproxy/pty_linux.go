//go:build linux
// +build linux

package ptyproxy

import (
	"os/signal"
	"syscall"
)

func (pty *PtyCmd) OsUpdateSize(Rows uint16, Cols uint16) {
	pty.Rows = Rows
	pty.Cols = Cols
	// pty.Ch <- syscall.SIGWINCH
	pty.set_size_changed <- true
}
func (ret PtyCmd) Notify() {
	signal.Notify(ret.ws_change_signal, syscall.SIGWINCH)

}
