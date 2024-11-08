package ptyproxy
func (pty *Pty) OsUpdateSize(Rows uint16, Cols uint16) {
	pty.Rows = Rows
	pty.Cols = Cols
	go func ()  {
		pty.wch <- true 
	}()
}
func (ret Pty)Notify(){
//  signal.Notify(ret.Ch, syscall.SIGWINCH)
}