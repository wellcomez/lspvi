package mainui

import (
	// "io"
	"bytes"
	"encoding/hex"
	"io"
	"log"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"
	"unicode"

	// "os/exec"

	corepty "github.com/creack/pty"
	// "github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	// "github.com/pgavlin/femto"
	v100 "golang.org/x/term"
	"zen108.com/lspvi/pkg/pty"
	"zen108.com/lspvi/pkg/term"
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
type Term struct {
	// *femto.View
	*tview.Box
	imp *terminal_impl
	*view_link
	dest *terminal.State
	// dest *terminal.State
}

var re = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func extractStr(t *terminal.State, x0, x1, row int) string {
	var s []rune
	for i := x0; i <= x1; i++ {
		c, _, _ := t.Cell(i, row)
		s = append(s, c)
	}
	return string(s)
}

// Write implements io.Writer.
func (t Term) Write(p []byte) (n int, err error) {
	// not enough bytes for a full rune
	if n, err := t.v100state(p); err != nil {
		log.Println("vstate 100", err)
	} else {
		log.Println("write", n, hex.EncodeToString(p))
	}
	return len(p), err
}
func (t *Term) v100state(p []byte) (int, error) {
	var written int
	r := bytes.NewReader(p)
	t.dest.Lock()
	defer t.dest.Unlock()
	for {
		c, sz, err := r.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return written, err
		}
		written += sz
		if c == unicode.ReplacementChar && sz == 1 {
			if r.Len() == 0 {
				// not enough bytes for a full rune
				return written - 1, nil
			}
			log.Println("invalid utf8 sequence")
			continue
		}
		t.dest.Put(c)
	}
	col, row := t.dest.Size()
	log.Println(strings.Repeat("-", 80))
	for y := 0; y < row; y++ {
		var line []rune
		for x := 0; x < col; x++ {
			ch, _, _ := t.dest.Cell(x, y)
			line = append(line, ch)
		}
		log.Println(">", string(line))

	}
	log.Println(strings.Repeat("+", 80))
	return written, nil
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

func NewTerminal(app *tview.Application, shellname string) *Term {
	cmdline := ""
	switch shellname {
	case "bash":
		cmdline = "/usr/bin/sh"
	}

	t := Term{tview.NewBox(),
		&terminal_impl{
			nil,
			shellname,
			[]byte{},
			nil,
			nil,
			0, 0,
		},
		&view_link{id: view_term},
		&terminal.State{},
	}
	t.dest.Init()
	t.dest.DebugLogger = log.Default()
	col := 80
	row := 40
	t.dest.Resize(col, row)
	go func() {
		ptyio := pty.RunNoStdin([]string{cmdline})
		if err := corepty.Setsize(ptyio.File, &corepty.Winsize{Rows: uint16(row), Cols: uint16(col)}); err != nil {
			log.Printf("error resizing pty: %s", err)
		}
		signal.Notify(ptyio.Ch, syscall.SIGWINCH)
		t.imp.ptystdio = ptyio
		v100term := v100.NewTerminal(ptyio.File, "")
		t.imp.v100term = v100term
		go func() {
			for range ptyio.Ch {
				timer := time.After(500 * time.Millisecond)
				<-timer
				_, _, w, h := t.GetRect()
				t.dest.Resize(w, h)
				if err := corepty.Setsize(ptyio.File, &corepty.Winsize{Rows: uint16(h), Cols: uint16(w)}); err != nil {
					log.Printf("error resizing pty: %s", err)
				}
			}
		}()
		io.Copy(t, ptyio.File)
	}()
	t.SetDrawFunc(func(screen tcell.Screen, posx, posy, width, height int) (int, int, int, int) {
		log.Println("term", "width", width, "height", height)
		cols, rows := t.dest.Size()
		for y := 0; y < rows; y++ {
			for x := 0; x < cols; x++ {
				ch, bg, fg := t.dest.Cell(x, y)
				screen.SetContent(posx+x, posy+y, ch, nil, tcell.StyleDefault.Foreground(tcell.Color(fg)).Background(tcell.Color(bg)))
			}
		}
		return posx, posy, width, height
	})
	t.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		// switch action {
		// case 14, 13:
		// 	{
		// 		gap := 1
		// 		if action == 14 {
		// 			t.ScrollDown(gap)
		// 		} else {
		// 			t.ScrollUp(gap)
		// 		}
		// 		go func() {
		// 			app.QueueUpdateDraw(func() {})
		// 		}()
		// 	}
		// }
		return action, event
	})
	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		_, _, width, height := t.GetRect()
		log.Println("term", "width", width, "height", height)
		if t.imp.ptystdio != nil {
			var n int
			var err error
			if buf := t.TypedKey(event); buf != nil {
				n, err = t.imp.v100term.Write(buf.buf)
			} else {
				n, err = t.imp.v100term.Write([]byte{byte(event.Rune())})
			}
			if err != nil {
				log.Println(n, err)
			}
		}
		return nil
	})
	return &t
}

func (t *Term) Draw(screen tcell.Screen) {
}
type inputbuf struct {
	buf        []byte
	bufferMode bool
}

func (s *inputbuf) Write(buf []byte) (int, error) {
	s.buf = buf
	return len(buf), nil
}
func (t *inputbuf) typeCursorKey(key *tcell.EventKey) {
	cursorPrefix := byte('[')
	if t.bufferMode {
		cursorPrefix = 'O'
	}

	switch key.Key() {
	case tcell.KeyUp:
		_, _ = t.Write([]byte{asciiEscape, cursorPrefix, 'A'})
	case tcell.KeyDown:
		_, _ = t.Write([]byte{asciiEscape, cursorPrefix, 'B'})
	case tcell.KeyLeft:
		_, _ = t.Write([]byte{asciiEscape, cursorPrefix, 'D'})
	case tcell.KeyRight:
		_, _ = t.Write([]byte{asciiEscape, cursorPrefix, 'C'})
	}
}

const (
	asciiBell      = 7
	asciiBackspace = 8
	asciiEscape    = 27

	noEscape = 5000
	tabWidth = 8
)

func (t *Term) TypedKey(e *tcell.EventKey) *inputbuf {
	// if t.keyboardState.shiftPressed {
	// 	t.keyTypedWithShift(e)
	// 	return
	// }
	var in inputbuf
	switch e.Key() {
	// case tcell.KeyReturn:
	// 	_, _ = in.Write([]byte{'\r'})
	case tcell.KeyEnter:
		_, _ = in.Write([]byte{'\r'})
		// if t.newLineMode {
		// 	_, _ = in.Write([]byte{'\r'})
		// 	return
		// }
		_, _ = in.Write([]byte{'\n'})
	case tcell.KeyTab:
		_, _ = in.Write([]byte{'\t'})
	case tcell.KeyF1:
		_, _ = in.Write([]byte{asciiEscape, 'O', 'P'})
	case tcell.KeyF2:
		_, _ = in.Write([]byte{asciiEscape, 'O', 'Q'})
	case tcell.KeyF3:
		_, _ = in.Write([]byte{asciiEscape, 'O', 'R'})
	case tcell.KeyF4:
		_, _ = in.Write([]byte{asciiEscape, 'O', 'S'})
	case tcell.KeyF5:
		_, _ = in.Write([]byte{asciiEscape, '[', '1', '5', '~'})
	case tcell.KeyF6:
		_, _ = in.Write([]byte{asciiEscape, '[', '1', '7', '~'})
	case tcell.KeyF7:
		_, _ = in.Write([]byte{asciiEscape, '[', '1', '8', '~'})
	case tcell.KeyF8:
		_, _ = in.Write([]byte{asciiEscape, '[', '1', '9', '~'})
	case tcell.KeyF9:
		_, _ = in.Write([]byte{asciiEscape, '[', '2', '0', '~'})
	case tcell.KeyF10:
		_, _ = in.Write([]byte{asciiEscape, '[', '2', '1', '~'})
	case tcell.KeyF11:
		_, _ = in.Write([]byte{asciiEscape, '[', '2', '3', '~'})
	case tcell.KeyF12:
		_, _ = in.Write([]byte{asciiEscape, '[', '2', '4', '~'})
	case tcell.KeyEscape:
		_, _ = in.Write([]byte{asciiEscape})
	case tcell.KeyBackspace:
		_, _ = in.Write([]byte{asciiBackspace})
	case tcell.KeyDelete:
		_, _ = in.Write([]byte{asciiEscape, '[', '3', '~'})
	case tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight:
		in.typeCursorKey(e)
	case tcell.KeyPgUp:
		_, _ = in.Write([]byte{asciiEscape, '[', '5', '~'})
	case tcell.KeyPgDn:
		_, _ = in.Write([]byte{asciiEscape, '[', '6', '~'})
	case tcell.KeyHome:
		_, _ = in.Write([]byte{asciiEscape, 'O', 'H'})
	case tcell.KeyInsert:
		_, _ = in.Write([]byte{asciiEscape, '[', '2', '~'})
	case tcell.KeyEnd:
		_, _ = in.Write([]byte{asciiEscape, 'O', 'F'})
	default:
		return nil
	}
	return &in
}
