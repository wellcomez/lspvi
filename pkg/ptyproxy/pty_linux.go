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
	pty.Ch <- syscall.SIGWINCH
}
func (ret PtyCmd) Notify() {
	signal.Notify(ret.Ch, syscall.SIGWINCH)

}
