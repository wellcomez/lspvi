package mainui

import (
	// "io"
	"io"
	"log"
	"regexp"

	// "os/exec"

	// "github.com/creack/pty"
	// "github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/pgavlin/femto"
	v100 "golang.org/x/term"
	"zen108.com/lspvi/pkg/pty"
)

const (
	keyCtrlC     = 3
	keyCtrlD     = 4
	keyCtrlU     = 21
	keyEnter     = '\r'
	keyEscape    = 27
	keyBackspace = 127
	keyUnknown   = 0xd800 /* UTF-16 surrogate area */ + iota
	keyUp
	keyDown
	keyLeft
	keyRight
	keyAltLeft
	keyAltRight
	keyHome
	keyEnd
	keyDeleteWord
	keyDeleteLine
	keyClearScreen
	keyPasteStart
	keyPasteEnd
)

type terminal_impl struct {
	ptystdio  *pty.Pty
	shellname string
	buf       []byte
	ondata    func(*terminal_impl)
	v100term  *v100.Terminal
}
type terminal struct {
	*femto.View
	imp *terminal_impl
	*view_link
}

var re = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// Write implements io.Writer.
func (t terminal) Write(p []byte) (n int, err error) {
	check := false
	if len(p) == 3 {
		if p[0] == 0x8 && p[2] == 0x8 && p[1] == 0x20 {
			// ret := t.View
			b := t.imp.buf
			t.imp.buf = b[0 : len(b)-1]
			t.View.Backspace()
			check = true
		}
	}
	if !check {
		p1 := re.ReplaceAll(p, []byte{})
		t.imp.buf = append(t.imp.buf, p1...)
		go func() {
			GlobalApp.QueueUpdateDraw(func() {

				t.View.OpenBuffer(femto.NewBufferFromString(string(t.imp.buf), ""))
				t.Cursor.Loc = femto.Loc{
					X: 0,
					Y: t.View.Buf.LinesNum() - 1,
				}
				t.View.EndOfLine()
			})
		}()
	}
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
	ret := terminal{femto.NewView(femto.NewBufferFromString("", "")),
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

	ret.View.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if ret.imp.ptystdio != nil {
			ch := event.Rune()
			switch event.Key() {
			case tcell.KeyLeft:
				ch = keyLeft
			case tcell.KeyRight:
				ch = keyRight
			case tcell.KeyUp:
				ch = keyUp
			case tcell.KeyDown:
				ch = keyDown
			}
			n, e := ret.imp.v100term.Write([]byte{byte(ch)})
			if e == nil {
				log.Println(n, e)
			}
		}
		return nil
	})
	return &ret
}
