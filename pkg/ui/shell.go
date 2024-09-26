package mainui

import (
	// "io"
	"io"
	"log"

	// "os/exec"

	// "github.com/creack/pty"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	// "golang.org/x/term"
	"zen108.com/lspvi/pkg/pty"
)

var inputdata = make(chan []byte)

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

// Read implements io.Reader.
func (t terminal) Read(p []byte) (n int, err error) {
	// panic("unimplemented")
	return len(p), nil
}

// Write implements io.Writer.
func (t terminal) Write(p []byte) (n int, err error) {
	t.imp.buf = p
	t.imp.ondata(t.imp)
	return len(p), nil
	// return len(inputdata), nil
}

func NewTerminal(app *tview.Application, shellname string) *terminal {
	// cmdline := ""
	// switch shellname {
	// case "bash":
	// 	cmdline = "/usr/bin/bash"
	// 	cmdline = "/usr/bin/ls"
	// }
	ret := terminal{tview.NewTextView(), &terminal_impl{nil, shellname, []byte{}, nil}, &view_link{id: view_term}}
	ret.imp.ondata = func(t *terminal_impl) {
		go func() {
			app.QueueUpdateDraw(func() {
				ret.TextView.Write(t.buf)
				ss := ret.TextView.GetText(true)
				log.Println("shell data", ss, string(t.buf))
			})
		}()
	}
	go func() {
		ptyio := pty.RunNoStdin([]string{"/usr/bin/bash"})
		ret.imp.ptystdio = ptyio
		io.Copy(ret, ptyio.File)
	}()
	ret.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if ret.imp.ptystdio != nil {
			n, e := ret.imp.ptystdio.File.Write([]byte{byte(event.Rune())})
			if e == nil {
				log.Println(n, e)
			}
		}
		return event
	})
	return &ret
}
