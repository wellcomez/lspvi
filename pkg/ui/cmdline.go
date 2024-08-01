package mainui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type cmdline struct {
	main  *mainui
	input *tview.InputField
	Vim   *Vim
}

func new_cmdline(main *mainui) *cmdline {
	code := &cmdline{
		main:  main,
		input: tview.NewInputField(),
	}
	code.input.SetBorder(true)
	input := code.input
	input.SetMouseCapture(code.handle_mouse)
	input.SetFieldBackgroundColor(tcell.ColorBlack)
	input.SetInputCapture(code.Keyhandle)
	code.Vim = NewVim(main)
	return code
}
func (cmd *cmdline) OnComand(command string) bool {
	command = strings.TrimRight(command, "\r")
	command = strings.TrimRight(command, "\n")
	if command == "q" || command == "quit" || command == "q!" || command == "qa" {
		cmd.main.Close()
		return true
	}
	num, err := strconv.ParseInt(command, 10, 32)
	if err == nil {
		cmd.main.codeview.gotoline(int(num) - 1)
		return true
	}
	return false
}

func (cmd *cmdline) HandleKeyUnderEscape(event *tcell.EventKey) *tcell.EventKey {
	if event.Rune() == 'f' {
		cmd.main.OnGrep()
	}
	return nil
}

func (cmd *cmdline) Keyhandle(event *tcell.EventKey) *tcell.EventKey {
	yes, event := cmd.Vim.VimKeyModelMethod(event)
	if yes {
		return nil
	}
	return event
}

func (ret *cmdline) handle_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	return action, event
}

func (cmd *cmdline) SetValue(value string) {
	cmd.input.SetText(value)
}
func (cmd cmdline) Value() string {
	return cmd.input.GetText()
}
func (cmd cmdline) Clear() {
	cmd.input.SetText("")
}

// vimstate structure
type vimstate struct {
	Escape    bool
	Find      bool
	Command   bool
	Insert    bool
	Leader    bool
	FindEnter string
}

func (v vimstate) String() string {
	if v.Escape {
		return "escape"
	}
	if v.Command {
		return "command"
	}
	if v.Insert {
		return "insert"
	}
	if v.Find {
		return "find"
	}
	if v.Leader {
		return "leader"
	}
	return "none"

}

type vim_mode_handle interface {
	HanldeKey(event *tcell.EventKey) bool
	State() string
	end()
}
type vi_command_mode_handle struct {
	main *mainui
	vi   *Vim
}

// end implements vim_mode_handle.
func (v vi_command_mode_handle) end() {
}

// State implements vim_mode_handle.
func (v vi_command_mode_handle) State() string {
	return "command"
}

func (v vi_command_mode_handle) HanldeKey(event *tcell.EventKey) bool {
	cmd := v.vi.app.cmdline
	vim := cmd.Vim
	if handle_backspace(event, cmd) {
		return true
	}
	txt := cmd.input.GetText()
	if event.Key() == tcell.KeyEnter {
		if len(txt) > 1 {
			vim.vi.FindEnter = txt[1:]
		}
		if cmd.OnComand(vim.vi.FindEnter) {
			cmd.Vim.EnterEscape()
		}
		return true
	}
	txt = txt + string(event.Rune())
	cmd.input.SetText(txt)
	return true
}

type vi_find_handle struct {
	main *mainui
	vi   *Vim
}

// end implements vim_mode_handle.
func (v vi_find_handle) end() {

}

// State implements vim_mode_handle.
func (v vi_find_handle) State() string {
	return "find " + v.vi.vi.FindEnter
}

// HanldeKey implements vim_mode_handle.
func (v vi_find_handle) HanldeKey(event *tcell.EventKey) bool {
	cmd := v.vi.app.cmdline
	vim := v.vi
	shouldReturn := handle_backspace(event, cmd)
	if shouldReturn {
		return true
	}
	txt := cmd.input.GetText()
	if event.Key() == tcell.KeyEnter {
		if len(txt) > 1 {
			vim.vi.FindEnter = txt[1:]
		}
		cmd.main.OnSearch(txt[1:], false)
		return true
	}
	if len(vim.vi.FindEnter) > 0 {
		if event.Rune() == 'n' {
			cmd.main.OnSearch(vim.vi.FindEnter, false)
			return true
		}
	}
	txt = txt + string(event.Rune())
	cmd.input.SetText(txt)
	cmd.main.OnSearch(txt[1:], false)
	return true
}

func handle_backspace(event *tcell.EventKey, cmd *cmdline) bool {
	txt := cmd.input.GetText()
	vim := cmd.Vim
	if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
		if len(txt) > 1 {
			txt = txt[0 : len(txt)-1]
			cmd.input.SetText(txt)
			vim.vi.FindEnter = ""
		}
		return true
	}
	return false
}

type escapestate struct {
	keyseq []string
	init   bool
}
type EscapeHandle struct {
	main  *mainui
	vi    *Vim
	state *escapestate
}

// State implements vim_mode_handle.
func (e EscapeHandle) State() string {
	return fmt.Sprintf("escape %s", strings.Join(e.state.keyseq, ""))
}

// end
func (e EscapeHandle) end() {
	e.state.init = true
}

func (l EscapeHandle) HanldeKey(event *tcell.EventKey) bool {
	if len(l.state.keyseq) == 0 {
		if l.main.codeview.handle_key_impl(event) == nil {
			l.end()
			return true
		}
	}
	if l.state.init {
		l.state.keyseq = []string{}
		l.state.init = false
	}
	ts := l.state.keyseq
	ch := string(event.Rune())
	const ctrlw = "c-w"
	const left = "left"
	const right = "right"
	const up = "up"
	const down = "down"
	var mm = map[tcell.Key]string{
		tcell.KeyCtrlW: ctrlw,
		tcell.KeyLeft:  left,
		tcell.KeyRight: right,
		tcell.KeyUp:    up,
		tcell.KeyDown:  down,
	}
	if c, ok := mm[event.Key()]; ok {
		ch = c
	}
	l.state.keyseq = append(ts, string(ch))
	commandmap := map[string]func(){}

	move_up := func() {
		l.main.move_up_window()
	}
	commandmap[ctrlw+"k"] = move_up
	commandmap[ctrlw+up] = move_up

	move_down := func() {
		l.main.move_down_window()
	}
	commandmap[ctrlw+down] = move_down
	commandmap[ctrlw+"j"] = move_down

	commandmap[ctrlw+left] = func() {
		l.main.move_left_window()
	}
	commandmap[ctrlw+right] = func() {
		l.main.move_right_window()
	}
	commandmap["gg"] = func() {
		l.main.codeview.gotoline(0)
		l.end()
	}
	commandmap["gd"] = func() {
		l.main.codeview.action_goto_define()
	}
	commandmap["gr"] = func() {
		l.main.codeview.action_get_refer()
		l.end()
	}
	commandmap["G"] = func() {
		l.main.codeview.gotoline(-1)
		l.end()
	}
	if fun, ok := commandmap[strings.Join(l.state.keyseq, "")]; ok {
		fun()
		l.end()
	}

	return true
}

type leadstate struct {
	kseq string
	end  bool
}
type LeaderHandle struct {
	main  *mainui
	vi    *Vim
	state *leadstate
}

// end implements vim_mode_handle.
func (s LeaderHandle) end() {
	s.state.end = true
}

// State implements vim_mode_handle.
func (l LeaderHandle) State() string {
	return fmt.Sprintf("leader %s", l.state.kseq)
}

// HanldeKey implements vim_mode_handle.
func (l LeaderHandle) HanldeKey(event *tcell.EventKey) bool {
	ch := event.Rune()
	if l.state.end {
		l.state.kseq = ""
	}
	key := l.state.kseq + string(ch)
	l.state.kseq = key

	if key == "r" {
		l.main.OpenDocumntRef()
		l.end()
	}
	if key == "o" {
		l.main.OpenDocumntFzf()
		l.end()
		return true
	}
	return false
}

// Vim structure
type Vim struct {
	app       *mainui
	vi        vimstate
	vi_handle vim_mode_handle
}

func (v Vim) String() string {
	if v.vi_handle != nil {
		return v.vi_handle.State() + " "
	}
	return v.vi.String()
}

// NewVimState creates a new vimstate instance.
func NewVimState() vimstate {
	return vimstate{
		Escape:    false,
		Find:      false,
		Command:   false,
		Insert:    false,
		FindEnter: "",
	}
}

// NewVim creates a new vim instance.
func NewVim(app *mainui) *Vim {
	return &Vim{
		app: app,
		vi:  NewVimState(),
	}
}

// MoveFocus moves the focus in the application.
func (v *Vim) MoveFocus() {
	v.app.MoveFocus()
}

// EnterGrep
func (v *Vim) EnterGrep(txt string) {
	v._enter_find_mode()
	v.vi.FindEnter = txt
	v.app.cmdline.SetValue("/" + txt)
}
func (v *Vim) VimKeyModelMethod(event *tcell.EventKey) (bool, *tcell.EventKey) {
	if event.Rune() == ':' {
		if v.EnterCommand() {
			return true, nil
		}
	}
	if event.Rune() == leadkey {
		if v.EnterLead() {
			return true, nil
		}
	}
	if event.Rune() == 'i' {
		if v.EnterInsert() {
			return true, nil
		}
	}
	if event.Rune() == '/' {
		if v.EnterFind() {
			return true, nil
		}
	}
	if event.Key() == tcell.KeyEscape {
		v.EnterEscape()
		return true, nil
	}
	// if v.vi.Escape {
	// 	v.ExitEnterEscape()
	// }
	if v.vi_handle != nil {
		if v.vi_handle.HanldeKey(event) {
			return true, nil
		}
	}
	return false, event
}

// EnterFind enters find mode.
func (v *Vim) EnterFind() bool {
	if v.vi.Escape {
		v.MoveFocus()
		v._enter_find_mode()
		return true
	} else {
		return false
	}
}

func (v *Vim) _enter_find_mode() {
	v.vi = vimstate{Find: true}
	a := vi_find_handle{
		main: v.app,
		vi:   v,
	}
	v.vi_handle = a
	v.app.cmdline.SetValue("/")
}
func (v *Vim) EnterLead() bool {
	if v.vi.Escape {
		v.vi = vimstate{Leader: true}
		v.vi_handle = nil
		v.vi_handle = LeaderHandle{
			main:  v.app,
			vi:    v,
			state: &leadstate{end: false},
		}
		return true
	} else {
		return false
	}
}

// EnterInsert enters insert mode.
func (v *Vim) EnterInsert() bool {
	if v.vi.Escape {
		v.vi = vimstate{Insert: true}
		return true
	} else {
		return false
	}
}

func (v *Vim) ExitEnterEscape() {
	v.vi = vimstate{}
	v.vi_handle = nil
}

// EnterEscape enters escape mode.
func (v *Vim) EnterEscape() {
	v.app.cmdline.Clear()
	v.vi = vimstate{Escape: true}

	f := v.app.app.GetFocus()
	if f == v.app.cmdline.input {
		v.app.codeview.view.Focus(nil)
	}
	v.vi_handle = EscapeHandle{
		main:  v.app,
		vi:    v,
		state: &escapestate{init: true},
	}
	v.app.SavePrevFocus()
}

// EnterCommand enters command mode.
func (v *Vim) EnterCommand() bool {
	if v.vi.Escape {
		v.vi = vimstate{Command: true}
		v.MoveFocus()
		v.app.cmdline.input.Focus(nil)
		v.app.cmdline.SetValue(":")
		v.vi_handle = vi_command_mode_handle{
			main: v.app,
			vi:   v,
		}
		return true
	}
	return false
}
