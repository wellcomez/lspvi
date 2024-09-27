package mainui

import (
	// "io"
	"bytes"
	"fmt"
	"strings"

	// "encoding/hex"
	"io"
	"log"
	"os/signal"

	// "strings"
	"syscall"
	"time"
	"unicode"

	// "os/exec"

	corepty "github.com/creack/pty"
	// "github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	// "github.com/pgavlin/femto"
	// v100 "golang.org/x/term"
	"zen108.com/lspvi/pkg/pty"
	"zen108.com/lspvi/pkg/term"
)

type terminal_impl struct {
	ptystdio  *pty.Pty
	shellname string
	buf       []byte
	ondata    func(*terminal_impl)
	topline   int
	dest      *terminal.State
	// v100term  *v100.Terminal
	// w, h int
}
type Term struct {
	// *femto.View
	*tview.Box
	imp *terminal_impl
	*view_link
}

// Write implements io.Writer.
func (t Term) Write(p []byte) (n int, err error) {
	// not enough bytes for a full rune
	if n, err := t.imp.v100state(p); err != nil {
		log.Println("vstate 100", err, n)
	} else {
		// log.Println("write", n, hex.EncodeToString(p))
	}
	go func() {
		GlobalApp.QueueUpdateDraw(func() {

		})
	}()
	return len(p), err
}
func (t *terminal_impl) v100state(p []byte) (int, error) {
	var written int
	r := bytes.NewReader(p)
	t.dest.Lock()
	defer t.dest.Unlock()

	state := t.dest
	offsize := t.topline
	log.Println("offscreen size", offsize)
	// _, bottom := state.Cursor()
	// _, height := state.Size()
	// log.Println("bottom", bottom)
	// oldline := state.CurrentCell()
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
	t.topline = len(t.dest.Offscreen)

	log.Println("new offscreen size", t.topline, state.OfflineString(t.topline-1))
	// newline := state.CurrentCell()
	// _, new_bottom := state.Cursor()
	// if new_bottom == height {
	// 	if new_bottom > bottom {

	// 	}
	// }

	col, row := t.dest.Size()

	log.Println(strings.Repeat("o", 80))
	for y := 0; y < len(t.dest.Offscreen); y++ {
		log.Println("<<<<<<", y, state.OfflineString(y))
	}
	log.Println(strings.Repeat("o", 80))

	log.Println(strings.Repeat("n", 80))
	for y := 0; y < row; y++ {
		var line []rune
		for x := 0; x < col; x++ {
			ch, _, _ := t.dest.Cell(x, y)
			line = append(line, ch)
		}
		log.Println(">>>>>", string(line))
	}
	log.Println(strings.Repeat("n", 80))
	return written, nil
}

func NewTerminal(app *tview.Application, shellname string) *Term {
	cmdline := ""
	switch shellname {
	case "bash":
		cmdline = "/usr/bin/sh"
	}

	ret := &Term{tview.NewBox(),
		&terminal_impl{
			nil,
			shellname,
			[]byte{},
			nil,
			0,
			&terminal.State{},
		},
		&view_link{id: view_term},
	}
	t := ret.imp
	t.dest.Init()
	t.dest.DebugLogger = log.Default()
	col := 80
	row := 40
	t.dest.Resize(col, row)
	go func() {
		ptyio := pty.RunNoStdin([]string{cmdline})
		signal.Notify(ptyio.Ch, syscall.SIGWINCH)
		t.ptystdio = ptyio
		ret.UpdateTermSize()
		// v100term := v100.NewTerminal(ptyio.File, "")
		// t.imp.v100term = v100term
		go func() {
			for range ptyio.Ch {
				timer := time.After(100 * time.Millisecond)
				<-timer
				ret.UpdateTermSize()
			}
		}()
		io.Copy(ret, ptyio.File)
	}()
	ret.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		switch action {
		case 14, 13:
			{
				state := t.dest
				if action == 14 {
					t.topline = min(len(state.Offscreen), t.topline+1)
				} else {
					if t.topline < 1 {
						t.topline = 0
					} else {
						t.topline--
					}
				}
				go func() {
					app.QueueUpdateDraw(func() {})
				}()
			}
		}
		return action, event
	})
	return ret
}

func (t Term) UpdateTermSize() {
	ptyio := t.imp.ptystdio
	_, _, w, h := t.GetRect()
	t.imp.dest.Resize(w, h)
	if err := corepty.Setsize(ptyio.File, &corepty.Winsize{Rows: uint16(h), Cols: uint16(w)}); err != nil {
		log.Printf("error resizing pty: %s", err)
	}
}
func (t *Term) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return t.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		_, _, width, height := t.GetRect()
		log.Println("term", "width", width, "height", height)
		if t.imp.ptystdio != nil {
			var n int
			var err error
			ptyio := t.imp.ptystdio.File
			if buf := t.TypedKey(event); buf != nil {
				n, err = ptyio.Write(buf.buf)
			} else {
				n, err = ptyio.Write([]byte{byte(event.Rune())})
			}
			if err != nil {
				log.Println(n, err)
			}
		}
	})
}
func (termui *Term) Draw(screen tcell.Screen) {
	termui.Box.DrawForSubclass(screen, termui)
	t := termui.imp
	t.dest.Lock()
	defer t.dest.Unlock()
	posx, posy, width, height := termui.GetInnerRect()
	cols, rows := t.dest.Size()
	bottom := posy + height
	top := posy
	default_fg, default_bg, _ := global_theme.get_default_style().Decompose()
	state := t.dest
	offline := state.Offscreen
	total_offscreen_len := len(offline)
	offlines_to_draw := 0
	lineno_offset := 5

	if t.topline > 0 {
		offlines_to_draw = (total_offscreen_len - t.topline)
		if offlines_to_draw > 0 {
			for y := 0; y < offlines_to_draw; y++ {
				line_index := total_offscreen_len - y - 1
				screenY := posy + y
				if screenY < top {
					break
				}
				if screenY < bottom {
					if lineno_offset > 0 {
						sss := fmt.Sprintf("%4d", line_index)
						for i, v := range sss {
							screen.SetContent(posx+i, screenY, rune(v), nil, tcell.StyleDefault)
						}
					}
					for x := 0; x < cols-lineno_offset; x++ {
						ch, fg, bg := t.dest.OfflineCell(x, line_index)
						style := get_style_from_fg_bg(bg, default_bg, fg, default_fg)
						screenX := posx + x + lineno_offset
						screen.SetContent(screenX, screenY, ch, nil, style)
					}
				}
			}
		}
	}
	for y := 0; y < rows; y++ {
		screenY := posy + y + offlines_to_draw
		if screenY >= bottom {
			break
		}
		if lineno_offset > 0 {
			sss := fmt.Sprintf("%4d", y+total_offscreen_len)
			for i, v := range sss {
				screen.SetContent(posx+i, screenY, rune(v), nil,
					tcell.StyleDefault.Foreground(default_fg).Background(default_bg))
			}
		}
		for x := 0; x < cols-lineno_offset; x++ {
			ch, fg, bg := t.dest.Cell(x, y)
			style := get_style_from_fg_bg(bg, default_bg, fg, default_fg)
			screenX := posx + x + lineno_offset
			screen.SetContent(screenX, screenY, ch, nil, style)
		}
	}
	if offlines_to_draw > 0 {
		screen.HideCursor()
	} else {
		x, y := t.dest.Cursor()
		screen.ShowCursor(posx+x+lineno_offset, posy+y)
	}
	if width != cols || height != rows {
		go func() {
			t.ptystdio.UpdateSize(uint16(width), uint16(height))
		}()
	}
}

func get_style_from_fg_bg(bg terminal.Color, default_bg tcell.Color, fg terminal.Color, default_fg tcell.Color) tcell.Style {
	style := tcell.StyleDefault
	if bg == terminal.DefaultBG {
		style = style.Background(default_bg)
	} else if bg < 256 {
		style = style.Foreground(tcell.ColorValid + tcell.Color(bg))
	} else {

	}
	if fg == terminal.DefaultFG {
		style = style.Foreground(default_fg)
	} else if fg < 256 {
		style = style.Foreground(tcell.ColorValid + tcell.Color(fg))
	} else {

	}
	return style
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
