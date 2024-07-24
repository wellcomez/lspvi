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
	if command == "q" || command == "quit" || command == "q!" || command == "qa" {
		cmd.main.Close()
	}
}

func (cmd *cmdline) HandleKeyUnderEscape(event *tcell.EventKey) *tcell.EventKey {
	if event.Rune() == 'f' {
		cmd.main.OnGrep()
	}
	return nil
}
func (cmd *cmdline) Keyhandle(event *tcell.EventKey) *tcell.EventKey {
	vim := cmd.vim
	if cmd.vim.vi.Find || cmd.vim.vi.Command {
		txt := cmd.view.GetText()
		if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
			if len(txt) > 1 {
				txt = txt[0 : len(txt)-1]
				cmd.view.SetText(txt)
				vim.vi.FindEnter = ""
			}
			return nil
		}
		if vim.vi.Command || vim.vi.Find {
      if !cmd.view.HasFocus(){
        return nil
      }
			if event.Key() == tcell.KeyEnter {
				if len(txt) > 1 {
					vim.vi.FindEnter = txt[1:]
				}
				if vim.vi.Command {
					cmd.OnComand(vim.vi.FindEnter)
				} else if vim.vi.Find {
					cmd.main.OnSearch(txt[1:], false)
				}
				return nil
			}
		}
		if vim.vi.Find && len(vim.vi.FindEnter) > 0 {
			if event.Rune() == 'n' {
				cmd.main.OnSearch(vim.vi.FindEnter, false)
				return nil
			}
		}
		txt = txt + string(event.Rune())
		cmd.view.SetText(txt)
		if cmd.vim.vi.Find {
			cmd.main.OnSearch(txt[1:], false)
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
	FindEnter string
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
		FindEnter: "",
	}
}

// NewVim creates a new vim instance.
func NewVim(app *mainui) *vim {
	return &vim{
		app: app,
		vi:  NewVimState(),
	}
}

// MoveFocus moves the focus in the application.
func (v *vim) MoveFocus() {
	v.app.MoveFocus()
}

// EnterGrep
func (v *vim) EnterGrep(txt string) {
	v.vi = vimstate{Find: true}
	v.vi.FindEnter = txt
	v.app.cmdline.SetValue("/" + txt)
}

// EnterFind enters find mode.
func (v *vim) EnterFind() bool {
	if v.vi.Escape {
		v.MoveFocus()
		v.vi = vimstate{Find: true}
		v.app.cmdline.SetValue("/")
		return true
	} else {
		return false
	}
}

// EnterInsert enters insert mode.
func (v *vim) EnterInsert() bool {
	if v.vi.Escape {
		v.vi = vimstate{Insert: true}
		return true
	} else {
		return false
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
	v.app.SavePrevFocus()
}

// EnterCommand enters command mode.
func (v *vim) EnterCommand() bool {
	if v.vi.Escape {
		v.vi = vimstate{Command: true}
		v.MoveFocus()
		v.app.cmdline.view.Focus(nil)
		v.app.cmdline.SetValue(":")
		return true
	}
	return false
}
