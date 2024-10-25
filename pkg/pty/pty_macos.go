// +build darwin

package pty

import (
	"os/signal"
	"syscall"
)

func (pty *Pty) OsUpdateSize(Rows uint16, Cols uint16) {
	pty.Rows = Rows
	pty.Cols = Cols
	pty.Ch <- syscall.SIGWINCH
}
func (ret Pty)Notify(){
		signal.Notify(ret.Ch, syscall.SIGWINCH)

}