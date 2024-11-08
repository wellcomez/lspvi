package ptyproxy
func (pty *PtyCmd) OsUpdateSize(Rows uint16, Cols uint16) {
	pty.Rows = Rows
	pty.Cols = Cols
	go func ()  {
		pty.wch <- true 
	}()
}
func (ret PtyCmd)Notify(){
//  signal.Notify(ret.Ch, syscall.SIGWINCH)
}