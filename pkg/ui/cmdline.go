package mainui

import (
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type cmdline struct {
	main *mainui
	input *tview.InputField
	Vim  *Vim
}

func new_cmdline(main *mainui) *cmdline {
	code := &cmdline{
		main: main,
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
	vim := cmd.Vim
	if cmd.Vim.vi.Find || cmd.Vim.vi.Command {
		txt := cmd.input.GetText()
		if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
			if len(txt) > 1 {
				txt = txt[0 : len(txt)-1]
				cmd.input.SetText(txt)
				vim.vi.FindEnter = ""
			}
			return nil
		}
		if vim.vi.Command || vim.vi.Find {
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
		if !cmd.input.HasFocus() {
			x, y, x1, y1 := cmd.main.app.GetFocus().GetRect()

			log.Println("wrong focus", x, y, x1, y1, "id", cmd.main.GetFocusViewId(), cmd.Vim.String())
			return event
		}
		if vim.vi.Find && len(vim.vi.FindEnter) > 0 {
			if event.Rune() == 'n' {
				cmd.main.OnSearch(vim.vi.FindEnter, false)
				return nil
			}
		}
		txt = txt + string(event.Rune())
		cmd.input.SetText(txt)
		if cmd.Vim.vi.Find {
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
		return "Escape"
	}
	if v.Command {
		return "Command"
	}
	if v.Insert {
		return "Insert"
	}
	if v.Find {
		return "Find"
	}
	if v.Leader{
		return "Leader"
	}
	return "none"

}

// Vim structure
type Vim struct {
	app *mainui
	vi  vimstate
}

func (v Vim) String() string {
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
	v.vi = vimstate{Find: true}
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
	if v.vi.Escape {
		v.ExitEnterEscape()
	}
	return false, event
}

// EnterFind enters find mode.
func (v *Vim) EnterFind() bool {
	if v.vi.Escape {
		v.MoveFocus()
		v.vi = vimstate{Find: true}
		v.app.cmdline.SetValue("/")
		return true
	} else {
		return false
	}
}
func (v *Vim) EnterLead() bool {
	if v.vi.Escape {
		v.vi = vimstate{Leader: true}
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
}
// EnterEscape enters escape mode.
func (v *Vim) EnterEscape() {
	v.app.cmdline.Clear()
	v.vi = vimstate{Escape: true}

	f := v.app.app.GetFocus()
	if f == v.app.cmdline.input {
		v.app.codeview.view.Focus(nil)
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
		return true
	}
	return false
}
