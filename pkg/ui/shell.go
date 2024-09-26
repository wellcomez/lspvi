package mainui

import (
	// "io"
	"io"
	"log"
	"regexp"

	// "os/exec"

	// "github.com/creack/pty"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	v100 "golang.org/x/term"
	"zen108.com/lspvi/pkg/pty"
)

type terminal_impl struct {
	ptystdio  *pty.Pty
	shellname string
	buf       []byte
	ondata    func(*terminal_impl)
	v100term  *v100.Terminal
}
type terminal struct {
	*tview.TextView
	imp *terminal_impl
	*view_link
}

// Read implements io.Reader.
// func (t terminal) Read(p []byte) (n int, err error) {
// 	// panic("unimplemented")
// 	return len(p), nil
// }

// // Write implements io.Writer.
func (t terminal) Write(p []byte) (n int, err error) {
	return t.TextView.Write(p)
}

var ansiEscapeRegex = regexp.MustCompile(`\x1B[@-_][0-?]*[ -/]*[@-~]`)

func filterANSIEscapeCodes(s string) string {
	return ansiEscapeRegex.ReplaceAllString(s, "")
}

func NewTerminal(app *tview.Application, shellname string) *terminal {
	cmdline := ""
	switch shellname {
	case "bash":
		cmdline = "/usr/bin/sh"
	}
	ret := terminal{tview.NewTextView(),
		&terminal_impl{
			nil,
			shellname,
			[]byte{},
			nil,
			nil,
		},
		&view_link{id: view_term},
	}
	// ret.imp.ondata = func(t *terminal_impl) {
	// 	go func() {
	// 		app.QueueUpdateDraw(func() {
	// 			s := filterANSIEscapeCodes(string(t.buf))
	// 			ret.TextView.Write([]byte(s))
	// 			ss := ret.TextView.GetText(true)
	// 			log.Println("shell data", ss, string(t.buf))
	// 		})
	// 	}()
	// }
	go func() {
		ptyio := pty.RunNoStdin([]string{cmdline})
		ret.imp.ptystdio = ptyio
		// v100term := v100.NewTerminal(ptyio.File, "")
		// ret.imp.v100term =v100term
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
