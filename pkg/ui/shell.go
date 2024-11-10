// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	// "io"
	"bytes"
	"fmt"

	// "encoding/hex"
	"io"
	"log"

	// "strings"
	"time"
	"unicode"

	// "os/exec"

	// corepty "github.com/creack/pty"
	// "github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	// "github.com/pgavlin/femto"
	// v100 "golang.org/x/term"
	"zen108.com/lspvi/pkg/debug"
	"zen108.com/lspvi/pkg/ptyproxy"
	terminal "zen108.com/lspvi/pkg/term"
)

type terminal_pty struct {
	ptystdio  *ptyproxy.PtyCmd
	shellname string
	ondata    func(*terminal_pty)
	topline   int
	dest      *terminal.State
	ui        tview.Primitive
}

func (t terminal_pty) displayname() string {
	pid := ""
	if t.ptystdio != nil && t.ptystdio.Cmd != nil {
		pid = fmt.Sprintf("%d",
			t.ptystdio.Cmd.Process.Pid)
	}
	return fmt.Sprintf("%s-%s", t.shellname, pid)
}

type line []rune

type Term struct {
	// *femto.View
	*tview.Box
	current *terminal_pty
	*view_link
	termlist []*terminal_pty
	sel      selectarea
	// right_context *term_right_menu
	main *mainui
}
type ptyread struct {
	term *terminal_pty
}

func (ty ptyread) Write(p []byte) (n int, err error) {
	return ty.term.Write(p)
}

func (t *terminal_pty) Kill() error {
	if t.ptystdio.Cmd != nil {
		return t.ptystdio.Cmd.Process.Kill()
	}
	return fmt.Errorf("not cmd")
}

// Write implements io.Writer.
func (t *terminal_pty) Write(p []byte) (n int, err error) {
	// not enough bytes for a full rune
	if n, err := t.v100state(p); err != nil {
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
func (t *terminal_pty) v100state(p []byte) (int, error) {
	var written int
	r := bytes.NewReader(p)
	t.dest.Lock()
	defer t.dest.Unlock()

	// state := t.dest
	// offsize := t.topline
	// log.Println("offscreen size", offsize)
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
			// log.Println("invalid utf8 sequence")
			continue
		}
		t.dest.Put(c)
	}
	t.topline = len(t.dest.Offscreen)

	// log.Println("new offscreen size", t.topline, state.OfflineString(t.topline-1))
	// newline := state.CurrentCell()
	// _, new_bottom := state.Cursor()
	// if new_bottom == height {
	// 	if new_bottom > bottom {

	// 	}
	// }

	// col, row := t.dest.Size()

	// log.Println(strings.Repeat("o", 80))
	// for y := 0; y < len(t.dest.Offscreen); y++ {
	// log.Println("<<<<<<", y, state.OfflineString(y))
	// }
	// log.Println(strings.Repeat("o", 80))

	// log.Println(strings.Repeat("n", 80))
	// for y := 0; y < row; y++ {
	// 	var line []rune
	// 	for x := 0; x < col; x++ {
	// 		ch, _, _ := t.dest.Cell(x, y)
	// 		line = append(line, ch)
	// 	}
	// 	// log.Println(">>>>>", string(line))
	// }
	// log.Println(strings.Repeat("n", 80))
	return written, nil
}

type term_right_menu struct {
	view      *Term
	menu_item *menudata
}

func (menu term_right_menu) getbox() *tview.Box {
	// menu.view.
	if menu.view.main.tab.activate_tab_id == view_term {
		return menu.view.Box
	}
	return nil
}

func (menu term_right_menu) menuitem() []context_menu_item {
	return menu.menu_item.menu_item
}

func (menu term_right_menu) on_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	if action == tview.MouseRightClick {
		return tview.MouseConsumed, nil
	}
	if action == tview.MouseLeftClick {
		view_term.setfocused(menu.view.main)
	}
	return tview.MouseConsumed, nil
}
func NewTerminal(main *mainui, app *tview.Application, shellname string) *Term {

	ret := &Term{tview.NewBox(),
		nil,
		&view_link{id: view_term},
		[]*terminal_pty{},
		selectarea{},
		main,
	}
	right_context := term_right_menu{
		view: ret,
		menu_item: &menudata{[]context_menu_item{
			{item: cmditem{Cmd: cmdactor{desc: "Copy "}}, handle: func() {
				// ret.qfh.Delete(ret.GetCurrentItem())
				s := ret.sel.GetSelection()
				main.CopyToClipboard(s)
			}},
		}},
		// main:      main,
	}
	term := ret.new_pty(shellname, nil)
	ret.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if a, e := main.Right_context_menu().handle_menu_mouse_action(action, event, right_context, ret.Box); a == tview.MouseConsumed {
			return a, e
		}
		return ret.handle_mouse(action, app, event)
	})
	ret.current = term
	ret.termlist = append(ret.termlist, term)
	return ret
}
func (t *Term) ListTerm() []string {
	var ret = []string{}
	for _, v := range t.termlist {
		ret = append(ret, v.displayname())
	}
	return ret
}

func (ret *Term) new_pty(shellname string, cb func(bool)) *terminal_pty {
	cmdline := ""
	switch shellname {
	case "bash":
		{
			cmdline = "bash"
		}
	case "zsh":
		cmdline = "/usr/bin/zsh"
	case "sh":
		cmdline = "/usr/bin/sh"
	}
	term := &terminal_pty{
		nil,
		shellname,
		nil,
		0,
		&terminal.State{},
		ret,
	}

	term.start_pty(cmdline, func(yes bool, t *terminal_pty) {
		if !yes {
			for i, v := range ret.termlist {
				if v == t {
					ret.termlist = append(ret.termlist[:i], ret.termlist[i+1:]...)
					break
				}
			}
		}
		if cb != nil {
			cb(yes)
		}
	})
	return term
}

func (ret *Term) handle_mouse(action tview.MouseAction, app *tview.Application, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	t := ret.current
	if t == nil {
		return action, event
	}
	drawit := ret.sel.handle_mouse_selection(action, event)
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
			drawit = true

		}
	}
	if drawit {
		go func() {
			app.QueueUpdateDraw(func() {})
		}()
	}
	return action, event
}

func (term *terminal_pty) start_pty(cmdline string, end func(bool, *terminal_pty)) {
	term.dest.Init()
	term.dest.DebugLogger = log.Default()
	col := 80
	row := 40
	term.dest.Resize(col, row)
	go func() {
		ptyio := ptyproxy.RunNoStdin([]string{cmdline})
		if ptyio == nil {
			debug.ErrorLog("terminal ", "ptyio=nil", cmdline)
			return
		}
		term.ptystdio = ptyio
		term.ptystdio.Notify()
		term.UpdateTermSize()
		go func() {
			for range ptyio.Ch {
				timer := time.After(100 * time.Millisecond)
				<-timer
				term.UpdateTermSize()
			}
		}()
		go func() {
			for range ptyio.Wch {
				timer := time.After(100 * time.Millisecond)
				<-timer
				term.UpdateTermSize()
			}
		}()
		if end != nil {
			end(true, term)
		}
		io.Copy(ptyread{term}, ptyio.File())
		if end != nil {
			end(false, term)
		}
	}()
}

func (term *terminal_pty) UpdateTermSize() {
	ptyio := term.ptystdio
	_, _, w, h := term.ui.GetRect()
	term.dest.Resize(w, h)
	ptyio.UpdateSize(uint16(h), uint16(w))
}
func (t *Term) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return t.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		_, _, width, height := t.GetRect()
		log.Println("term", "width", width, "height", height)
		if t.current.ptystdio != nil {
			var n int
			var err error
			ptyio := t.current.ptystdio.File()
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

type term_line_drawer struct {
	default_fg, default_bg tcell.Color
	posx, posy             int
	lineno_offset, cols    int
	sel                    *selectarea
	selection_style        tcell.Style
}

func (d *term_line_drawer) Draw(screen tcell.Screen, index, screenY int, style tcell.Style, OfflineCell func(x, y int) (ch rune, fg, bg terminal.Color)) {
	if d.lineno_offset > 0 {
		sss := fmt.Sprintf("%4d", index)
		for i, v := range sss {
			if d.sel.In(d.posx+i, screenY) {
				screen.SetContent(d.posx+i, screenY, rune(v), nil, style)
			} else {
				screen.SetContent(d.posx+i, screenY, rune(v), nil, style)
			}
		}
	}
	_, selbg, _ := d.selection_style.Decompose()
	for x := 0; x < d.cols-d.lineno_offset; x++ {
		ch, fg, bg := OfflineCell(x, index)
		style := get_style_from_fg_bg(bg, d.default_bg, fg, d.default_fg)
		screenX := d.posx + x + d.lineno_offset
		if d.sel.HasSelection() && d.sel.In(screenX, screenY) {
			s := d.sel.text[screenY-d.sel.start.Y]
			s = append(s, ch)
			d.sel.text[screenY-d.sel.start.Y] = s
			screen.SetContent(screenX, screenY, ch, nil, style.Background(selbg))
		} else {
			screen.SetContent(screenX, screenY, ch, nil, style)
		}
	}
}

func (termui *Term) Draw(screen tcell.Screen) {
	termui.Box.DrawForSubclass(screen, termui)
	t := termui.current
	t.dest.Lock()
	defer t.dest.Unlock()
	posx, posy, width, height := termui.GetInnerRect()
	cols, rows := t.dest.Size()
	termui.sel.cols = cols
	bottom := posy + height
	default_fg, default_bg, _ := global_theme.get_default_style().Decompose()
	state := t.dest
	offline := state.Offscreen
	total_offscreen_len := len(offline)
	offlines_to_draw := 0
	lineno_offset := 0
	default_theme_style := tcell.StyleDefault.Foreground(default_fg).Background(default_bg)
	sel_style := default_theme_style
	if s := global_theme.select_style(); s != nil {
		sel_style = *s
	}
	var draw = term_line_drawer{
		default_fg, default_bg,
		posx, posy,
		lineno_offset, cols,
		&termui.sel,
		sel_style,
	}
	if termui.sel.HasSelection() {
		// _, selbg, _ := d.selection_style.Decompose()
		// log.Println("selection", termui.sel.start, termui.sel.end,
		// 	"(", posx, posy, posx+width, posy+height, ")")
	}
	if t.topline >= 0 {
		offlines_to_draw = (total_offscreen_len - t.topline)
		if offlines_to_draw > 0 {
			screenY := posy
			for index := total_offscreen_len - offlines_to_draw; index < total_offscreen_len; index++ {
				if screenY < bottom {

					draw.Draw(screen, index, screenY, tcell.StyleDefault, t.dest.OfflineCell)
				} else {
					break
				}
				screenY++
			}
		}
	}
	for y := 0; y < rows; y++ {
		screenY := posy + y + offlines_to_draw
		if screenY >= bottom {
			break
		}
		draw.Draw(screen, y, screenY, default_theme_style, t.dest.Cell)
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
