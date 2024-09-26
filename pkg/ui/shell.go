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

var re = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// Write implements io.Writer.
func (t terminal) Write(p []byte) (n int, err error) {
	p1 := re.ReplaceAll(p, []byte{})
	go func() {
		GlobalApp.QueueUpdateDraw(func() {
			t.TextView.ScrollToEnd()
		})
	}()
	t.TextView.Write(p1)
	return len(p), nil
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
		v100term := v100.NewTerminal(ptyio.File, "")
		ret.imp.v100term = v100term
		io.Copy(ret, ptyio.File)
	}()
	ret.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if ret.imp.ptystdio != nil {
			switch event.Key() {
			case tcell.KeyBackspace2, tcell.KeyBackspace:
				t := ret.GetText(false)
				t = t[0 : len(t)-1]
				ret.SetText(t)
			}
			n, e := ret.imp.v100term.Write([]byte{byte(event.Rune())})
			if e == nil {
				log.Println(n, e)
			}
		}
		return event
	})
	return &ret
}
