package main

import (
	// "bufio"
	// "fmt"
	"io"
	"os"

	"github.com/iyzyi/aiopty/pty"
	"github.com/iyzyi/aiopty/term"
	"github.com/iyzyi/aiopty/utils/log"
)
type term_stdio struct {
	r, w *os.File
}
type term_stdin struct {
	*term_stdio
}
func new_term_stdio() *term_stdio{
	r, w, _ := os.Pipe()
	return &term_stdio{ r: r, w: w}
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
}
func (t term_stdout) Read(p []byte) (n int, err error) {
	return t.r.Read(p)
}
func new_aipty_term_io() *aipty_term_io {
	return &aipty_term_io{
		stdin:  term_stdin{
			new_term_stdio(),
		},
		stdout: term_stdout{
			new_term_stdio(),
		},
	}
}
func (t *aipty_term_io) Read(p []byte) (n int, err error) {
	return t.stdout.Read(p)
}

func (t *aipty_term_io) Write(p []byte) (n int, err error) {
	return t.stdin.Write(p)
}

func (t *aipty_term_io) Close() error {
	t.stdin.Close()
	t.stdout.Close()
	return nil
}

type term_stdout struct {
	*term_stdio
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

func main() {
	NewMain([]string{"bash"})
}
