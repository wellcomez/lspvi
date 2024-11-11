package ptyproxy
func (pty *PtyCmd) OsUpdateSize(Rows uint16, Cols uint16) {
	pty.rows = Rows
	pty.cols = Cols
	go func ()  {
		pty.set_size_changed <- true 
	}()
}
func (ret PtyCmd)Notify(){
//  signal.Notify(ret.Ch, syscall.SIGWINCH)
}