package mainui

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strconv"
	"strings"
)

type cmdline struct {
	*view_link
	main  *mainui
	input *tview.InputField
	Vim   *Vim
}

func new_cmdline(main *mainui) *cmdline {
	code := &cmdline{
		view_link: &view_link{
			up: view_fzf,
		},
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
	if command == "cn" {
		get_cmd_actor(cmd.main, vi_quick_next).handle()
		return true
	}
	if command == "cp" {
		get_cmd_actor(cmd.main, vi_quick_prev).handle()
	} else if command == "q" || command == "quit" || command == "q!" || command == "qa" {
		cmd.main.Close()
	} else if command == "h" || command == "help" {
		cmd.main.helpkey(true)
	} else if num, err := strconv.ParseInt(command, 10, 32); err == nil {
		cmd.main.codeview.gotoline(int(num) - 1)
		return true
	} else {
		return false
	}
	return true
}

// func (cmd *cmdline) HandleKeyUnderEscape(event *tcell.EventKey) *tcell.EventKey {
// 	if event.Rune() == 'f' {
// 		cmd.main.OnGrep()
// 	}
// 	return nil
// }

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
	Virtual   bool
	Insert    bool
	Leader    bool
	FindEnter string
}

func (v vimstate) String() string {
	if v.Virtual {
		return "Virtual"
	}
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
func (cmd *cmdline) search_next() {
	vim := cmd.Vim
	cmd.main.OnSearch(vim.vi.FindEnter, false, false)
}
func (cmd *cmdline) search_prev() {
	vim := cmd.Vim
	cmd.main.OnSearch(vim.vi.FindEnter, false, false)
}

// HanldeKey implements vim_mode_handle.
func (v vi_find_handle) HanldeKey(event *tcell.EventKey) bool {
	cmd := v.vi.app.cmdline
	vim := v.vi
	if !cmd.input.HasFocus() {
		if event.Rune() == 'n' {
			cmd.main.OnSearch(vim.vi.FindEnter, false, false)
			return true
		}
		return false
	}
	shouldReturn := handle_backspace(event, cmd)
	if shouldReturn {
		return true
	}
	txt := cmd.input.GetText()
	if event.Key() == tcell.KeyEnter {
		if len(txt) > 1 {
			vim.vi.FindEnter = txt[1:]
		}
		cmd.main.OnSearch(txt[1:], false, false)
		return true
	}
	if len(vim.vi.FindEnter) > 0 {
		if event.Rune() == 'n' {
			cmd.main.OnSearch(vim.vi.FindEnter, false, false)
			return true
		}
	}
	txt = txt + string(event.Rune())
	cmd.input.SetText(txt)
	cmd.main.OnSearch(txt[1:], false, false)
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
	keyseq  []string
	lastcmd string
	init    bool
}
type EscapeHandle struct {
	main  *mainui
	vi    *Vim
	state *escapestate
	input *inputdelay
}

// State implements vim_mode_handle.
func (e EscapeHandle) State() string {
	s := e.state.lastcmd
	if len(e.state.keyseq) > 0 {
		s = strings.Join(e.state.keyseq, "")
	}

	return fmt.Sprintf("escape %s", s)
}

// end
func (e EscapeHandle) end() {
	e.state.end()
}
func (state *escapestate) end() {
	state.init = true
	state.lastcmd = strings.Join(state.keyseq, "")
	state.keyseq = []string{}
}

func (l EscapeHandle) input_cb(word string) {
	if l.state.init {
		return
	}
	l.input.run(word)
}
func (l EscapeHandle) HanldeKey(event *tcell.EventKey) bool {
	if l.state.init {
		l.state.keyseq = []string{}
		l.state.init = false
	}
	ts := l.state.keyseq
	ch := string(event.Rune())

	if c, ok := event_to_keyname[event.Key()]; ok {
		ch = c
	}
	l.state.keyseq = append(ts, string(ch))
	cmdname := strings.Join(l.state.keyseq, "")
	processed, end := l.input.check(cmdname)
	if end {
		l.end()
	} else if l.main.codeview.handle_key_impl(event) == nil {
		l.state.keyseq = []string{string(event.Rune())}
		l.end()
		return true
	}
	return processed
}

type leadstate struct {
	kseq  string
	end   bool
	input *inputdelay
}

type LeaderHandle struct {
	main  *mainui
	vi    *Vim
	state *leadstate
}

func (s *LeaderHandle) inputcb(word string) {
	if !s.state.end {
		s.runcommand(word)
	}
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
	l.main.layout.spacemenu.visible = false
	ch := event.Rune()
	state := l.state
	input := state.input
	if state.end {
		state.kseq = ""
	}
	key := state.kseq + string(ch)
	state.kseq = key
	process, end := input.check(key)
	if end {
		l.end()
		l.vi.EnterEscape()
	}
	return process
}

func (l *LeaderHandle) runcommand(key string) {
	l.main.app.QueueUpdate(func() {
		state := l.state
		state.input.run(key)
		l.end()
	})
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
	//current_view := v.app.get_focus_view_id()
	//if current_view == view_code || current_view == view_cmd
	{
		if v.vi_handle != nil {
			if v.vi_handle.HanldeKey(event) {
				return true, nil
			}
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

// EnterLead jj
func (v *Vim) EnterLead() bool {
	if v.vi.Escape {
		v.vi = vimstate{Leader: true}
		v.vi_handle = nil
		main := v.app
		lead := LeaderHandle{
			main:  main,
			vi:    v,
			state: &leadstate{end: false},
		}
		/*keymap := map[string]func(){
			"f":  main.open_picker_ctrlp,
			"fw": main.codeview.action_grep_word,
			"r":  main.open_picker_refs,
			"h":  main.open_picker_history,
			"o":  main.open_document_symbol_picker,
		}*/
		sss := main.key_map_leader()
		input := &inputdelay{cb: lead.inputcb, cmdlist: sss}
		lead.state.input = input
		v.vi_handle = lead
		v.app.layout.spacemenu.visible = true
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
		v.app.set_viewid_focus(view_code)
	}
	esc := EscapeHandle{
		main:  v.app,
		vi:    v,
		state: &escapestate{init: true},
	}
	main := v.app
	sss := main.key_map_escape()
	inputdelay := inputdelay{
		cmdlist: sss,
		main:    main,
	}
	esc.input = &inputdelay
	esc.input.cb = esc.input_cb
	v.vi_handle = esc
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
