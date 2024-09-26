package mainui

import (
	// "io"
	"bytes"
	"io"
	"log"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	// "os/exec"

	corepty "github.com/creack/pty"
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
	w, h      int
}
type terminal struct {
	*femto.View
	imp *terminal_impl
	*view_link
}

var re = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// Write implements io.Writer.
func (t terminal) Write(p []byte) (n int, err error) {
	retlen := len(p)
	check := false

	bash_1 := []uint8{
		27, 91, 63, 50, 48, 48, 52, 108, 13,
	}
	bash2 := []byte{
		27, 91, 63, 50, 48, 48, 52, 104, 27, 93, 48, 59,
	}
	data3 := []byte{
		27, 91, 63, 50, 48, 48, 52, 104, 32,
	}
	p = replace_sub_array(p, bash2)
	p = replace_sub_array(p, bash_1)
	p = replace_sub_array(p, data3)

	backetptn := []byte{0x8, 0x20, 0x8}

	// "\b\x1b[K"
	bash_backet := []byte{
		8, 27, 91, 75,
	}
	check = pth_match(p, backetptn) || pth_match(p, bash_backet)
	if check {
		b := t.imp.buf
		t.imp.buf = b[0 : len(b)-1]
		t.View.Backspace()
	}

	if !check {
		p1 := re.ReplaceAll(p, []byte{})
		t.imp.buf = append(t.imp.buf, p1...)
	}
	go func() {
		GlobalApp.QueueUpdateDraw(func() {
			linecout := t.View.Buf.LinesNum()
			if linecout > 1000 {
				buf := []byte{}
				for i := linecout - 500; i < linecout; i++ {
					b := t.View.Buf.LineBytes(i)
					b = append(b, []byte("\r\n")...)
					buf = append(buf, b...)
				}
				t.imp.buf = buf
			}
			t.View.OpenBuffer(femto.NewBufferFromString(string(t.imp.buf), ""))
			t.Cursor.Loc = femto.Loc{
				X: 0,
				Y: t.View.Buf.LinesNum() - 1,
			}
			t.Buf.Settings["cursorline"] = false
			t.Buf.Settings["ruler"] = false
			ss := t.View.Buf.Line(t.Cursor.Loc.Y)
			log.Println(ss)
			t.View.EndOfLine()
			line := t.View.Buf.LinesNum()
			_, _, w, h := t.GetInnerRect()
			if t.imp.w != w || t.imp.h != h {
				t.imp.v100term.SetSize(w, h)
				t.imp.w = w
				t.imp.h = h
			}
			if line > h && h > 0 {
				t.Topline = t.View.Buf.LinesNum() - h
			}
		})
	}()
	return retlen, nil
}

func replace_sub_array(p []byte, bash2 []byte) []byte {
	index := bytes.Index(p, bash2)
	if index != -1 {
		p = append(p[:index], p[index+len(bash2):]...)
	}
	return p
}

func pth_match(p []byte, backetptn []byte) bool {
	if len(p) >= len(backetptn) && bytes.Equal(p[0:len(backetptn)], backetptn) {
		return true
	}
	return false
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
			0, 0,
		},
		&view_link{id: view_term},
	}
	ret.SetColorscheme(global_theme.colorscheme)
	// view:=ret.View
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
		if err := corepty.Setsize(ptyio.File, &corepty.Winsize{Rows: 100, Cols: 200}); err != nil {
			log.Printf("error resizing pty: %s", err)
		}
		signal.Notify(ptyio.Ch, syscall.SIGWINCH)
		ret.imp.ptystdio = ptyio
		v100term := v100.NewTerminal(ptyio.File, "")
		ret.imp.v100term = v100term
		go func() {
			for range ptyio.Ch {
				timer := time.After(500 * time.Millisecond)
				<-timer
				_, _, w, h := ret.GetRect()
				if err := corepty.Setsize(ptyio.File, &corepty.Winsize{Rows: uint16(h), Cols: uint16(w)}); err != nil {
					log.Printf("error resizing pty: %s", err)
				}
			}
		}()
		io.Copy(ret, ptyio.File)
	}()
	ret.View.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		switch action {
		case 14, 13:
			{
				gap := 1
				if action == 14 {
					ret.ScrollDown(gap)
				} else {
					ret.ScrollUp(gap)
				}
				go func() {
					app.QueueUpdateDraw(func() {})
				}()
			}
		}
		return action, event
	})
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
