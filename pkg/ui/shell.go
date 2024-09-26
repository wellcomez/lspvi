package mainui

import (
	"io"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"zen108.com/lspvi/pkg/pty"
)

var buf = make(chan []byte)

type terminal_impl struct {
	ptystdio  *pty.Pty
	shellname string
	buf       []byte
	ondata    func(*terminal_impl)
}
type terminal struct {
	*tview.TextView
	imp *terminal_impl
	*view_link
}

// Write implements io.Writer.
func (t terminal) Write(p []byte) (n int, err error) {
	t.imp.buf = append(t.imp.buf, p...)
	return len(buf), nil
}

func NewTerminal(app *tview.Application, shellname string) *terminal {
	cmdline := ""
	switch shellname {
	case "bash":
		cmdline = "/usr/bin/bash"
	}
	ptystdio := pty.Ptymain([]string{cmdline})
	ptystdio.Cols = 80
	ptystdio.Rows = 24
	ret := terminal{tview.NewTextView(), &terminal_impl{ptystdio, shellname, []byte{}, nil}, &view_link{id: view_term}}
	ret.imp.ondata = func(t *terminal_impl) {
		go func() {
			app.QueueUpdateDraw(func() {
				ret.Write(t.buf)
			})
		}()
	}
	io.Copy(ret, ptystdio.File)
	ret.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		ret.imp.ptystdio.File.Write([]byte{byte(event.Rune())})
		return nil
	})
	return &ret
}
