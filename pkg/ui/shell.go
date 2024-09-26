package mainui

import (
	"io"
	"log"
	"os/exec"

	"github.com/creack/pty"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	// "zen108.com/lspvi/pkg/pty"
)

var buf = make(chan []byte)

type terminal_impl struct {
	// ptystdio  *pty.Pty
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
	t.imp.buf = p
	t.imp.ondata(t.imp)
	return len(buf), nil
}

func NewTerminal(app *tview.Application, shellname string) *terminal {
	cmdline := ""
	switch shellname {
	case "bash":
		cmdline = "/usr/bin/bash"
		cmdline = "/usr/bin/ls"
	}
	ret := terminal{tview.NewTextView(), &terminal_impl{shellname, []byte{}, nil}, &view_link{id: view_term}}
	ret.imp.ondata = func(t *terminal_impl) {
		go func() {
			app.QueueUpdateDraw(func() {
				ret.TextView.Write(t.buf)
				ss := ret.TextView.GetText(true)
				log.Println("shell data", ss, string(t.buf))
			})
		}()
	}
	c := exec.Command(cmdline)
	f, err := pty.Start(c)
	pty.Setsize(f, &pty.Winsize{Cols: 100, Rows: 80})
	io.Copy(ret, f)
	if err != nil {
		log.Println(err)
	}
	ret.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// f.Write([]byte{byte(event.Rune())})
		n, err := f.WriteString("ls\n")
		log.Println(n, err)
		return nil
	})
	return &ret
}
