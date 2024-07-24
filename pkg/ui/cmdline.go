package mainui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type cmdline struct {
	main *mainui
	view *tview.InputField
	vim  *vim
}

func new_cmdline(main *mainui) *cmdline {
	ret := &cmdline{
		main: main,
		view: tview.NewInputField(),
	}
	ret.view.SetBorder(true)
	root := ret.view
	root.SetMouseCapture(ret.handle_mouse)
	root.SetFieldBackgroundColor(tcell.ColorBlack)
	root.SetInputCapture(ret.Keyhandle)
	ret.vim = NewVim(main)
	return ret
}
func (cmd *cmdline) OnComand(command string) {
	command = strings.TrimRight(command, "\r")
	command = strings.TrimRight(command, "\n")
	if command == "q" {
		cmd.main.app.Stop()
	}
}

func (cmd *cmdline) Keyhandle(event *tcell.EventKey) *tcell.EventKey {
	if cmd.vim.vi.Find || cmd.vim.vi.Command {
		txt := cmd.view.GetText()
		txt = txt + string(event.Rune())
		if cmd.vim.vi.Command {
			if event.Key() == tcell.KeyEnter {
				cmd.OnComand(txt[1:])
				return nil
			}
		}
		cmd.view.SetText(txt)
		if cmd.vim.vi.Find {
			cmd.main.OnSearch(txt[1:])
			return nil
		}
	}
	switch event.Key() {
	case tcell.KeyEnter:
	}
	return event
}

func (ret *cmdline) handle_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	return action, event
}

func (cmd *cmdline) SetValue(value string) {
	cmd.view.SetText(value)
}
func (cmd cmdline) Value() string {
	return cmd.view.GetText()
}
func (cmd cmdline) Clear() {
	cmd.view.SetText("")
}

// vimstate structure
type vimstate struct {
	Escape    bool
	Find      bool
	Command   bool
	Insert    bool
	FindEnter *string
}

// vim structure
type vim struct {
	app *mainui
	vi  vimstate
}

// NewVimState creates a new vimstate instance.
func NewVimState() vimstate {
	return vimstate{
		Escape:    false,
		Find:      false,
		Command:   false,
		Insert:    false,
		FindEnter: nil,
	}
}

// NewVim creates a new vim instance.
func NewVim(app *mainui) *vim {
	return &vim{
		app: app,
		vi:  NewVimState(),
	}
}

// CheckInput checks and updates the input based on the current state.
func (v *vim) CheckInput() string {
	if v.vi.FindEnter != nil {
		if len(*v.vi.FindEnter) > len(v.app.cmdline.Value()) {
			v.vi.FindEnter = nil
		} else {
			v.app.cmdline.SetValue(*v.vi.FindEnter)
		}
	}
	return v.app.cmdline.Value()
}

// MoveFocus moves the focus in the application.
func (v *vim) MoveFocus() {
	v.app.MoveFocus()
}

// EnterFind enters find mode.
func (v *vim) EnterFind() {
	if v.vi.Escape {
		v.MoveFocus()
		v.vi = vimstate{Find: true}
		v.app.cmdline.SetValue("/")
	}
}

// EnterInsert enters insert mode.
func (v *vim) EnterInsert() {
	if v.vi.Escape {
		v.vi = vimstate{Insert: true}
	}
}

// EnterEscape enters escape mode.
func (v *vim) EnterEscape() {
	v.app.cmdline.Clear()
	v.vi = vimstate{Escape: true}

	f := v.app.app.GetFocus()
	if f == v.app.cmdline.view {
		v.app.codeview.view.Focus(nil)
	}
}

// EnterCommand enters command mode.
func (v *vim) EnterCommand() {
	if v.vi.Escape {
		v.vi = vimstate{Command: true}
		v.MoveFocus()
		v.app.cmdline.view.Focus(nil)
		v.app.cmdline.SetValue(":")
	}
}
