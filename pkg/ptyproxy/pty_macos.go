// +build darwin

package pty

import (
	"os/signal"
	"syscall"
)

func (pty *Pty) OsUpdateSize(Rows uint16, Cols uint16) {
	pty.Rows = Rows
	pty.Cols = Cols
	pty.set_size_changed <- true
}
func (ret Pty)Notify(){
		signal.Notify(ret.ws_change_signal, syscall.SIGWINCH)

}