package ptyproxy

import (
	"github.com/iyzyi/aiopty/pty"
	"github.com/iyzyi/aiopty/term"
	"zen108.com/lspvi/pkg/debug"

	// "github.com/iyzyi/aiopty/utils/log"
	"io"
	"os"
)

func NewAioPty(Args []string) {
	// open a pty with options
	opt := &pty.Options{
		Path: Args[0],
		Args: Args,
		Dir:  "",
		Env:  nil,
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
		debug.ErrorLogf("aiopty", "Failed to create pty: %v", err)
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

	// enable terminal
	t, err := term.Open(os.Stdin, os.Stdout, onSizeChange)
	if err != nil {
		debug.ErrorLogf("aiopty", "Failed to enable terminal: %v", err)
		return
	}
	defer t.Close()

	// start data exchange between terminal and pty
	exit := make(chan struct{}, 2)
	go func() { io.Copy(p, t); exit <- struct{}{} }()
	go func() { io.Copy(t, p); exit <- struct{}{} }()
	<-exit
}
