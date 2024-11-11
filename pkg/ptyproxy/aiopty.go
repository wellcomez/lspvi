package ptyproxy

import (
	// "bufio"
	// "fmt"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/iyzyi/aiopty/pty"
	"github.com/iyzyi/aiopty/term"
	"github.com/iyzyi/aiopty/utils/log"
	goterm "golang.org/x/term"
	"zen108.com/lspvi/pkg/debug"
)

type term_stdio struct {
	r, w *os.File
}
type term_stdin struct {
	*term_stdio
}

func new_term_stdio() *term_stdio {
	r, w, _ := os.Pipe()
	return &term_stdio{r: r, w: w}
}
func (t term_stdio) Close() error {
	t.w.Close()
	t.r.Close()
	return nil
}
func (t term_stdin) Write(p []byte) (n int, err error) {
	return t.w.Write(p)
}

type aiopty_pty_interface interface {
	io.ReadWriteCloser
}
type aipty_term_io struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
	pty    *pty.Pty
}

func (t aipty_term_io) Close() error {
	t.stdin.Close()
	t.stdout.Close()
	return nil
}
func (t aipty_term_io) Read(p []byte) (n int, err error) {
	return t.stdout.Read(p)
}

func (t aipty_term_io) Write(p []byte) (n int, err error) {
	if t.pty != nil {
		_, err = t.pty.Write(p)
		return
	} else if t.stdin != nil {
		return t.stdin.Write(p)
	}
	return 0, fmt.Errorf("not target")
}

type term_stdout struct {
	*term_stdio
}

func (t term_stdout) Read(p []byte) (n int, err error) {
	return t.r.Read(p)
}
func (t *term_stdout) start() {
	go func() {
		for {
			buf := make([]byte, 1024)
			n, err := t.r.Read(buf)
			if err != nil {
				log.Error("Failed to read from pty: %v", err)
				return
			}
			print(string(buf[:n]))
		}
	}()
}

type AioPtyCmd struct {
	*PtyCmd
	pty *pty.Pty
}

func (*AioPtyCmd) Kill() (err error) {
	return nil
}
func (*AioPtyCmd) Pid() (ret string) {
	return
}
func (c *AioPtyCmd) UpdateSize(Rows uint16, Cols uint16) {
	if c.rows == Rows && c.cols == Cols {
		return
	}
	c.rows = Rows
	c.cols = Cols
	c.pty.SetSize(&pty.WinSize{
		Cols: Cols,
		Rows: Rows,
	})
}

func NewAioptyPtyCmd(cmdline string) (cmd LspPty) {
	var argv = strings.Split(cmdline, " ")
	Path := argv[0]
	// open a pty with options
	opt := &pty.Options{
		Path: Path,
		Args: argv,
		Dir:  "",
		Env:  nil,
		Size: &pty.WinSize{
			Cols: 120,
			Rows: 30,
		},
		Type: "",
	}
	// term_stdin := &term_stdin{new_term_stdio()}
	term_stdout := &term_stdout{new_term_stdio()}
	p, err := pty.OpenWithOptions(opt)
	if err != nil {
		log.Error("Failed to create pty: %v", err)
		return
	}
	cmd = &AioPtyCmd{
		&PtyCmd{
			file: aipty_term_io{
				// stdin:  term_stdin,
				stdout: term_stdout,
				pty:    p,
			},
			ws_change_signal: make(chan os.Signal, 1),
			set_size_changed: make(chan bool, 1),
			rows:             30,
			cols:             120,
		},
		p,
	}
	go func() {
		defer p.Close()

		// When the terminal window size changes, synchronize the size of the pty
		onSizeChange := func(cols, rows uint16) {
			size := &pty.WinSize{
				Cols: cols,
				Rows: rows,
			}
			p.SetSize(size)
		}
		// term_stdout.start()
		oldState, err := goterm.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			debug.ErrorLog("aiopty", err)
			return
		}
		defer func() { _ = goterm.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.
		defer cmd.File().Close()
		t, err := term.Open(os.Stdin, term_stdout.w, onSizeChange)
		if err != nil {
			log.Error("Failed to enable terminal: %v", err)
			return
		}
		defer t.Close()

		// start data exchange between terminal and pty
		exit := make(chan struct{}, 1)
		// go func() { io.Copy(p, t); exit <- struct{}{} }()
		go func() { io.Copy(t, p); exit <- struct{}{} }()
		<-exit
	}()
	return
}

func NewMain(argv []string) {
	// var sss = aipty_term_io{}
	// var io aiopty_pty_interface = sss

	Path := argv[0]
	// open a pty with options
	opt := &pty.Options{
		Path: Path,
		Args: argv,
		// Args: []string{"ag", "main"},
		Dir: "",
		Env: nil,
		Size: &pty.WinSize{
			Cols: 120,
			Rows: 30,
		},
		Type: "",
	}
	p, err := pty.OpenWithOptions(opt)

	// You can also open a pty simply like this:
	// p, err := pty.Open(path)

	if err != nil {
		log.Error("Failed to create pty: %v", err)
		return
	}
	defer p.Close()

	// When the terminal window size changes, synchronize the size of the pty
	onSizeChange := func(cols, rows uint16) {
		size := &pty.WinSize{
			Cols: cols,
			Rows: rows,
		}
		p.SetSize(size)
	}
	term_stdout := term_stdout{new_term_stdio()}
	defer term_stdout.Close()
	term_stdout.start()
	// scanner:= bufio.NewScanner(stdout)
	// enable terminal
	t, err := term.Open(os.Stdin, term_stdout.w, onSizeChange)
	if err != nil {
		log.Error("Failed to enable terminal: %v", err)
		return
	}
	defer t.Close()

	// start data exchange between terminal and pty
	exit := make(chan struct{}, 2)
	go func() { io.Copy(p, t); exit <- struct{}{} }()
	go func() { io.Copy(t, p); exit <- struct{}{} }()
	<-exit
}

// func main() {
// 	NewMain([]string{"bash"})
// }
